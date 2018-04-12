package daemon

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/samoslab/nebula/client/config"
	client "github.com/samoslab/nebula/client/provider_client"
	pb "github.com/samoslab/nebula/provider/pb"
	mpb "github.com/samoslab/nebula/tracker/metadata/pb"
	util_hash "github.com/samoslab/nebula/util/hash"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

var (
	// ReplicaFileSize using replication if file size less than
	ReplicaFileSize  = int64(8 * 1024)
	PartitionMaxSize = int64(256 * 1024 * 1024)
)

// ClientManager client manager
type ClientManager struct {
	mclient mpb.MatadataServiceClient
	NodeId  []byte
	TempDir string
	log     *logrus.Logger
	cfg     *config.ClientConfig
}

// NewClientManager create manager
func NewClientManager(log *logrus.Logger, trackerServer string, cfg *config.ClientConfig) (*ClientManager, error) {
	c := &ClientManager{}
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		return nil, err
	}
	defer conn.Close()

	c.mclient = mpb.NewMatadataServiceClient(conn)
	c.log = log
	c.TempDir = cfg.TempDir
	c.NodeId = []byte(cfg.NodeId)
	c.cfg = cfg
	return c, nil
}

// PingProvider ping provider
func (c *ClientManager) PingProvider(pro []*mpb.ErasureCodePartition) ([]*mpb.ErasureCodePartition, error) {
	return pro, nil
}

func (c *ClientManager) ConnectProvider() error {
	return nil
}

// UploadFile upload file to provider
func (c *ClientManager) UploadFile(filename string) error {
	log := c.log
	req, rsp, err := c.CheckFileExists(filename)
	if err != nil {
		return err
	}
	if rsp.GetCode() != 0 {
		return nil
	}
	log.Info("upload file %s", filename)
	// partition files if size > 256M
	// 如果分区了，文件大小，不分区前，也该检测到是否存在？
	switch rsp.GetStoreType() {
	case mpb.FileStoreType_ErasureCode:
		partFiles := []string{}
		var err error
		fileSize := int64(req.GetFileSize())
		if fileSize > PartitionMaxSize {
			chunkNum := int(math.Ceil(float64(fileSize) / float64(PartitionMaxSize)))
			chunkSize := fileSize / int64(chunkNum)
			partFiles, err = FileSplit(c.TempDir, filename, chunkSize)
			if err != nil {
				return err
			}
		} else {
			partFiles = append(partFiles, filename)
		}

		fileInfos := []MyPart{}

		for _, fname := range partFiles {
			fileSlices, err := c.OnlyFileSplit(fname, int(rsp.GetDataPieceCount()), int(rsp.GetVerifyPieceCount()))
			if err != nil {
				return err
			}
			fileInfos = append(fileInfos, MyPart{Filename: fname, Pieces: fileSlices})
		}

		ufpr := &mpb.UploadFilePrepareReq{}
		ufpr.Version = 1
		ufpr.FileHash = req.FileHash
		ufpr.Timestamp = uint64(time.Now().UTC().Unix())
		ufpr.NodeId = req.NodeId
		ufpr.FileSize = req.FileSize
		ufpr.Partition = make([]*mpb.SplitPartition, len(partFiles))
		// todo delete temp file
		for i, partInfo := range fileInfos {
			phslist := []*mpb.PieceHashAndSize{}
			for _, slice := range partInfo.Pieces {
				phs := &mpb.PieceHashAndSize{}
				phs.Hash = slice.FileHash
				phs.Size = uint32(slice.FileSize)
				phslist = append(phslist, phs)
			}
			ufpr.Partition[i].Piece = phslist
		}

		ctx := context.Background()
		ufprsp, err := c.mclient.UploadFilePrepare(ctx, ufpr)
		if err != nil {
			log.Errorf("UploadFilePrepare error %v", err)
			return err
		}

		partitions := []*mpb.StorePartition{}
		for _, partInfo := range fileInfos {

			partition, err := c.uploadFileBatchByErasure(ufpr, ufprsp, partInfo.Pieces)
			if err != nil {
				return err
			}
			partitions = append(partitions, partition)
		}

		return c.UploadFileDone(req, partitions)

	case mpb.FileStoreType_MultiReplica:
		return c.uploadFileByMultiReplica(req, rsp)
	}
	return nil
}

func (c *ClientManager) CheckFileExists(filename string) (*mpb.CheckFileExistReq, *mpb.CheckFileExistResp, error) {
	log := c.log
	hash, err := util_hash.Sha1File(filename)
	if err != nil {
		log.Errorf("sha1 file %s error %v", filename, err)
		return nil, nil, err
	}
	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Errorf("stat file %s error %v", filename, err)
		return nil, nil, err
	}
	ctx := context.Background()
	req := &mpb.CheckFileExistReq{}
	req.FileSize = uint64(fileInfo.Size())
	req.FileHash = hash
	req.NodeId = c.NodeId
	if fileInfo.Size() < ReplicaFileSize {
		fileData, err := util_hash.GetFileData(filename)
		if err != nil {
			log.Errorf("get file data error %v", err)
			return nil, nil, err
		}
		req.FileData = fileData
	}
	rsp, err := c.mclient.CheckFileExist(ctx, req)
	return req, rsp, err
}

func (c *ClientManager) OnlyFileSplit(filename string, dataNum, verifyNum int) ([]HashFile, error) {
	// Split file and hash
	// todo delete temp file
	fileSlices, err := RsEncoder(c.TempDir, filename, dataNum, verifyNum)
	if err != nil {
		c.log.Errorf("reed se error %v", err)
		return nil, err
	}
	return fileSlices, nil

}

func (c *ClientManager) uploadFileBatchByErasure(req *mpb.UploadFilePrepareReq, rsp *mpb.UploadFilePrepareResp, hashFiles []HashFile) (*mpb.StorePartition, error) {
	partition := &mpb.StorePartition{}
	providers, err := c.PingProvider(rsp.GetPartition())
	if err != nil {
		return nil, err
	}

	for i, pro := range providers {
		if i == 0 {
			block, err := c.uploadFileToErasureProvider(pro, hashFiles[i], true)
			if err != nil {
				return nil, err
			}
			partition.Block = append(partition.Block, block)
		} else {
			block, err := c.uploadFileToErasureProvider(pro, hashFiles[i], false)
			if err != nil {
				return nil, err
			}
			partition.Block = append(partition.Block, block)
		}
	}
	return partition, nil
}

func getOneOfPartition(pro *mpb.ErasureCodePartition) *mpb.BlockProviderAuth {
	pa := pro.GetProviderAuth()[0]
	return pa
}

func (c *ClientManager) uploadFileToErasureProvider(pro *mpb.ErasureCodePartition, fileInfo HashFile, first bool) (*mpb.StoreBlock, error) {
	block := &mpb.StoreBlock{}
	onePartition := getOneOfPartition(pro)
	server := fmt.Sprintf("%s:%d", onePartition.GetServer(), onePartition.GetPort())
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		return nil, err
	}
	defer conn.Close()
	pclient := pb.NewProviderServiceClient(conn)

	ha := onePartition.GetHashAuth()[0]
	err = client.Store(pclient, fileInfo.FileName, ha.GetAuth(), ha.GetTicket(), fileInfo.FileHash, uint64(fileInfo.FileSize), first)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	block.Hash = fileInfo.FileHash
	block.Size = uint64(fileInfo.FileSize)
	block.BlockSeq = uint32(fileInfo.SliceIndex)
	block.Checksum = true
	block.StoreNodeId = [][]byte{}
	block.StoreNodeId = append(block.StoreNodeId, []byte(onePartition.GetNodeId()))

	return block, nil
}

func (c *ClientManager) uploadFileToReplicaProvider(pro *mpb.ReplicaProvider, fileInfo HashFile) error {
	conn, err := grpc.Dial(pro.GetServer(), grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		c.log.Errorf("RPC Dail failed: %v", err)
		return err
	}
	defer conn.Close()
	pclient := pb.NewProviderServiceClient(conn)

	err = client.Store(pclient, fileInfo.FileName, pro.GetAuth(), pro.GetTicket(), fileInfo.FileHash, uint64(fileInfo.FileSize), true)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func (c *ClientManager) uploadFileByMultiReplica(req *mpb.CheckFileExistReq, rsp *mpb.CheckFileExistResp) error {

	fileInfo := HashFile{}
	fileInfo.FileName = req.FileName
	fileInfo.FileSize = int64(req.FileSize)
	fileInfo.FileHash = req.FileHash
	fileInfo.SliceIndex = 0
	for _, pro := range rsp.GetProvider() {
		c.uploadFileToReplicaProvider(pro, fileInfo)
	}

	return nil
}

func (c *ClientManager) UploadFileDone(reqCheck *mpb.CheckFileExistReq, partitions []*mpb.StorePartition) error {
	req := &mpb.UploadFileDoneReq{}
	req.Version = 1
	req.NodeId = reqCheck.GetNodeId()
	req.FileHash = reqCheck.GetFileHash()
	req.FileSize = reqCheck.GetFileSize()

	//Todo
	req.Parent = &mpb.FilePath{&mpb.FilePath_Path{"/folder1/folder2"}}
	req.FileModTime = 1

	var err error
	req.Sign, err = SignatureMessage(c.cfg.PrivateKey, []byte(""))
	if err != nil {
		return err
	}
	req.Timestamp = uint64(time.Now().UTC().Unix())
	req.Partition = partitions
	//for _, partition := range partitions {
	//req.Partition = append(req.Partition, partition)
	//}
	ctx := context.Background()
	ufdrsp, err := c.mclient.UploadFileDone(ctx, req)
	if err != nil {
		return err
	}
	fmt.Printf("done: %d\n", ufdrsp.GetCode())
	return nil
}

func (c *ClientManager) ListFiles() (*mpb.ListFilesResp, error) {
	req := &mpb.ListFilesReq{}
	req.Version = 1
	req.Timestamp = uint64(time.Now().UTC().Unix())
	req.NodeId = c.NodeId
	req.PageSize = 10
	req.PageNum = 1
	req.SortType = mpb.SortType_Name
	req.AscOrder = true
	var err error
	req.Sign, err = SignatureMessage(c.cfg.PrivateKey, []byte(""))
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	rsp, err := c.mclient.ListFiles(ctx, req)
	if err != nil {
		return nil, err
	}
	if rsp.GetCode() != 0 {
		return nil, fmt.Errorf("errmsg %s", rsp.GetErrMsg())
	}
	return rsp, nil
}

// DownloadFile download file
func (c *ClientManager) DownloadFile(fileInfo HashFile) error {
	req := &mpb.RetrieveFileReq{}
	req.Version = 1
	req.NodeId = c.NodeId
	req.Timestamp = uint64(time.Now().UTC().Unix())
	req.FileHash = []byte(fileInfo.FileHash)
	req.FileSize = uint64(fileInfo.FileSize)
	var err error
	req.Sign, err = SignatureMessage(c.cfg.PrivateKey, []byte(""))
	if err != nil {
		return err
	}
	ctx := context.Background()
	rsp, err := c.mclient.RetrieveFile(ctx, req)
	if err != nil {
		return err
	}
	if rsp.GetCode() == 1 {
		return errors.New("get file info failed")
	}

	// tiny file
	if filedata := rsp.GetFileData(); filedata != nil {
		fmt.Printf("filedata %s", filedata)
		saveFile(fileInfo.FileName, filedata)
		return nil
	}

	for i, partition := range rsp.GetPartition() {
		c.log.Infof("partition has %d ", i)
		c.saveFileByPartition(fileInfo.FileName, partition)
	}

	err = RsDecoder(fileInfo.FileName, "", 4, 2)
	if err != nil {
		return err
	}

	return nil
}

func (c *ClientManager) saveFileByPartition(filename string, partition *mpb.RetrievePartition) error {
	fmt.Printf("there is %d blocks", len(partition.GetBlock()))
	dataShards := 0
	parityShards := 0
	for _, block := range partition.GetBlock() {
		if block.GetChecksum() {
			dataShards += 1
		} else {
			parityShards += 1
		}
		node := block.GetStoreNode()
		if len(node) > 1 {
			fmt.Printf("node > 1")
		}
		node1 := node[0]
		server := fmt.Sprintf("%s:%d", node1.GetServer(), node1.GetPort())
		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if err != nil {
			fmt.Printf("RPC Dial failed: %s", err.Error())
			return err
		}
		defer conn.Close()
		pclient := pb.NewProviderServiceClient(conn)

		tempFilePrefix := filepath.Join(c.TempDir, filename)
		tempFileName := fmt.Sprintf("%s.%d", tempFilePrefix, block.GetBlockSeq())
		err = client.Retrieve(pclient, tempFileName, node1.GetAuth(), node1.GetTicket(), block.GetHash())
		if err != nil {
			return err
		}
	}

	// rs code
	return nil
}

func saveFile(fileName string, content []byte) error {
	// open output file
	fo, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fo.Close()
	if _, err := fo.Write(content); err != nil {
		return err
	}
	return nil
}

// SignatureMessage sign message
func SignatureMessage(privateKey string, message []byte) ([]byte, error) {
	rsaPrikey, err := x509.ParsePKCS1PrivateKey([]byte(privateKey))
	if err != nil {
		return nil, err
	}
	rng := rand.Reader
	hashed := sha256.Sum256(message)
	sign, err := rsa.SignPKCS1v15(rng, rsaPrikey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, err
	}

	return sign, nil
}
