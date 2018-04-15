package daemon

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/samoslab/nebula/client/config"
	client "github.com/samoslab/nebula/client/provider_client"
	"github.com/samoslab/nebula/provider/node"
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
	mclient    mpb.MatadataServiceClient
	NodeId     []byte
	TempDir    string
	log        *logrus.Logger
	cfg        *config.ClientConfig
	serverConn *grpc.ClientConn
}

// NewClientManager create manager
func NewClientManager(log *logrus.Logger, trackerServer string, cfg *config.ClientConfig) (*ClientManager, error) {
	if trackerServer == "" {
		return nil, errors.New("tracker server nil")
	}
	if cfg == nil {
		return nil, errors.New("client config nil")
	}
	c := &ClientManager{}
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		return nil, err
	}
	fmt.Printf("tracker server %s\n", trackerServer)
	//defer conn.Close()
	c.serverConn = conn

	c.mclient = mpb.NewMatadataServiceClient(conn)
	c.log = log
	c.TempDir = cfg.TempDir
	c.NodeId = cfg.Node.NodeId
	c.cfg = cfg
	return c, nil
}

// Shutdown shutdown tracker connection
func (c *ClientManager) Shutdown() {
	c.serverConn.Close()
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

	log.Infof("check file exists rsp:%+v\n", rsp)
	if rsp.GetCode() == 0 {
		log.Infof("upload success %s %s", filename, rsp.GetErrMsg())
		return nil
	}
	// 1 can upload
	if rsp.GetCode() != 1 {
		return fmt.Errorf("%d:%s", rsp.GetCode(), rsp.GetErrMsg())
	}

	log.Infof("start upload file %s", filename)
	switch rsp.GetStoreType() {
	case mpb.FileStoreType_MultiReplica:
		log.Infof("upload manner is multi-replication\n")
		partitions, err := c.uploadFileByMultiReplica(req, rsp)
		if err != nil {
			return err
		}
		return c.UploadFileDone(req, partitions)
	case mpb.FileStoreType_ErasureCode:
		log.Infof("upload manner erasure\n")
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

	}
	return nil
}

func (c *ClientManager) CheckFileExists(filename string) (*mpb.CheckFileExistReq, *mpb.CheckFileExistResp, error) {
	log := c.log
	hash, err := util_hash.Sha1File(filename)
	if err != nil {
		return nil, nil, err
	}
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, nil, err
	}
	dir, _ := filepath.Split(filename)
	ctx := context.Background()
	req := &mpb.CheckFileExistReq{}
	req.FileSize = uint64(fileInfo.Size())
	req.Interactive = true
	req.NewVersion = false
	req.Parent = &mpb.FilePath{&mpb.FilePath_Path{dir}}
	req.FileHash = hash
	req.NodeId = c.NodeId
	req.FileName = filename
	req.Timestamp = uint64(time.Now().UTC().Unix())
	mtime, err := GetFileModTime(filename)
	if err != nil {
		return nil, nil, err
	}
	req.FileModTime = uint64(mtime)
	if fileInfo.Size() < ReplicaFileSize {
		fileData, err := util_hash.GetFileData(filename)
		if err != nil {
			log.Errorf("get file %s data error %v", filename, err)
			return nil, nil, err
		}
		req.FileData = fileData
	}
	err = req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return nil, nil, err
	}
	log.Infof("checkfileexist req:%s\n", req.GetFileHash())
	rsp, err := c.mclient.CheckFileExist(ctx, req)
	return req, rsp, err
}

// MkFolder create folder
func (c *ClientManager) MkFolder(filepath string, folders []string, node *node.Node) (bool, error) {
	log := c.log
	ctx := context.Background()
	req := &mpb.MkFolderReq{}
	req.Parent = &mpb.FilePath{&mpb.FilePath_Path{filepath}}
	req.Folder = folders
	req.NodeId = c.NodeId
	req.Timestamp = uint64(time.Now().UTC().Unix())
	err := req.SignReq(node.PriKey)
	if err != nil {
		return false, err
	}
	log.Infof("make folder req:%+v", req.GetFolder())
	rsp, err := c.mclient.MkFolder(ctx, req)
	if rsp.GetCode() != 0 {
		return false, fmt.Errorf("%s", rsp.GetErrMsg())
	}
	log.Infof("make folder response:%+v", rsp)
	return true, nil
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
	err = client.Store(pclient, fileInfo.FileName, ha.GetAuth(), ha.GetTicket(), pro.GetTimestamp(), fileInfo.FileHash, uint64(fileInfo.FileSize), first)
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

func (c *ClientManager) uploadFileToReplicaProvider(pro *mpb.ReplicaProvider, fileInfo HashFile) ([]byte, error) {
	log := c.log
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	log.Infof("upload to provider %s\n", server)
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Errorf("RPC Dail failed: %v", err)
		return nil, err
	}
	defer conn.Close()
	pclient := pb.NewProviderServiceClient(conn)
	log.Debugf("upload fileinfo %+v", fileInfo)
	log.Debugf("provider auth %x", pro.GetAuth())
	log.Debugf("provider ticket %+v", pro.GetTicket())
	log.Debugf("provider nodeid %x", pro.GetNodeId())
	log.Debugf("provider time %d", pro.GetTimestamp())

	err = client.Store(pclient, fileInfo.FileName, pro.GetAuth(), pro.GetTicket(), pro.GetTimestamp(), fileInfo.FileHash, uint64(fileInfo.FileSize), true)
	if err != nil {
		log.Errorf("upload error %v\n", err)
		return nil, err
	}

	log.Infof("upload file %s success", fileInfo.FileName)

	return pro.GetNodeId(), nil
}

func (c *ClientManager) uploadFileByMultiReplica(req *mpb.CheckFileExistReq, rsp *mpb.CheckFileExistResp) ([]*mpb.StorePartition, error) {

	fileInfo := HashFile{}
	fileInfo.FileName = req.FileName
	fileInfo.FileSize = int64(req.FileSize)
	fileInfo.FileHash = req.FileHash
	fileInfo.SliceIndex = 0

	block := &mpb.StoreBlock{}
	block.Hash = fileInfo.FileHash
	block.Size = uint64(fileInfo.FileSize)
	block.BlockSeq = uint32(fileInfo.SliceIndex)
	block.Checksum = false
	block.StoreNodeId = [][]byte{}
	for _, pro := range rsp.GetProvider() {
		proID, err := c.uploadFileToReplicaProvider(pro, fileInfo)
		if err != nil {
			return nil, err
		}
		block.StoreNodeId = append(block.StoreNodeId, proID)
	}

	partition := &mpb.StorePartition{}
	partition.Block = append(partition.Block, block)
	partitions := []*mpb.StorePartition{partition}
	fmt.Printf("partitions %+v\n", partitions)
	return partitions, nil
}

func (c *ClientManager) UploadFileDone(reqCheck *mpb.CheckFileExistReq, partitions []*mpb.StorePartition) error {
	req := &mpb.UploadFileDoneReq{}
	req.Version = 1
	req.NodeId = c.NodeId
	req.FileHash = reqCheck.GetFileHash()
	req.FileSize = reqCheck.GetFileSize()
	req.FileName = reqCheck.GetFileName()
	req.FileModTime = reqCheck.GetFileModTime()

	req.Parent = reqCheck.GetParent()
	req.Interactive = reqCheck.GetInteractive()
	req.NewVersion = reqCheck.GetNewVersion()

	req.Timestamp = uint64(time.Now().UTC().Unix())
	req.Partition = partitions
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}
	ctx := context.Background()
	c.log.Infof("upload file done req:%s", req.GetFileHash())
	ufdrsp, err := c.mclient.UploadFileDone(ctx, req)
	if err != nil {
		return err
	}
	c.log.Infof("upload done code: %d", ufdrsp.GetCode())
	if ufdrsp.GetCode() != 0 {
		return fmt.Errorf("%s", ufdrsp.GetErrMsg())
	}
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
	rsaPrikey, err := x509.ParsePKCS1PrivateKey([]byte(c.cfg.PrivateKey))
	if err != nil {
		return nil, err
	}
	err = req.SignReq(rsaPrikey)
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
	rsaPrikey, err := x509.ParsePKCS1PrivateKey([]byte(c.cfg.PrivateKey))
	if err != nil {
		return err
	}
	err = req.SignReq(rsaPrikey)
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
