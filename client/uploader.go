package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/klauspost/reedsolomon"
	"github.com/sirupsen/logrus"
	client "github.com/spolabs/nebula/client/provider_client"
	pb "github.com/spolabs/nebula/provider/pb"
	mpb "github.com/spolabs/nebula/tracker/metadata/pb"
	util_hash "github.com/spolabs/nebula/util/hash"

	"google.golang.org/grpc"
)

type ClientManager struct {
	mclient mpb.MatadataServiceClient
	NodeId  []byte
	TempDir string
	log     *logrus.Logger
}

func NewClientManager(log *logrus.Logger) (*ClientManager, error) {
	c := &ClientManager{}
	conn1, err := grpc.Dial("127.0.0.1:8080", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		return nil, err
	}
	defer conn1.Close()

	c.mclient = mpb.NewMatadataServiceClient(conn1)
	c.log = log
	return c, nil
}

func (c *ClientManager) PingProvider(pro []*mpb.ErasureCodeProvider) ([]*mpb.ErasureCodeProvider, error) {
	return pro, nil
}

func (c *ClientManager) ConnectProvider() error {
	return nil
}

func (c *ClientManager) UploadFile(filename string) error {
	hash, err := util_hash.Sha1File(filename)
	log := c.log
	if err != nil {
		log.Errorf("sha1 file %s error %v", filename, err)
		return err
	}
	fileInfo, err := os.Stat(filename)
	if err != nil {
		log.Errorf("stat file %s error %v", filename, err)
		return err
	}
	ctx := context.Background()
	req := &mpb.CheckFileExistReq{}
	req.FileSize = uint64(fileInfo.Size())
	req.FileHash = hash
	req.NodeId = c.NodeId
	// TODO
	if fileInfo.Size() < 8*1024 {
		fileData, err := util_hash.GetFileData(filename)
		if err != nil {
			log.Errorf("get file data error %v", err)
			return err
		}
		req.FileData = fileData
	}
	rsp, err := c.mclient.CheckFileExist(ctx, req)
	if err != nil {
		log.Errorf("check file exists error %v", err)
		return err
	}

	log.Infof("check file exists response %+v", rsp)
	if rsp.GetCode() == 0 {
		log.Errorf("check file exists error %v", err)
		return nil
	}

	if rsp.StoreType == mpb.FileStoreType_ErasureCode {
		ufpr := &mpb.UploadFilePrepareReq{}
		ufpr.Version = 1
		ufpr.FileHash = req.FileHash
		ufpr.Timestamp = uint64(time.Now().UTC().Unix())
		ufpr.NodeId = req.NodeId
		ufpr.FileSize = req.FileSize
		ufpr.Piece = make([]*mpb.PieceHashAndSize, rsp.GetDataPieceCount()+rsp.GetVerifyPieceCount())
		// Split file and hash
		// todo delete temp file
		fileSlices, err := rsEncoder(c.TempDir, req.FileName, int(rsp.GetDataPieceCount()), int(rsp.GetVerifyPieceCount()))
		if err != nil {
			log.Errorf("reed se error %v", err)
			return err
		}

		for i, slice := range fileSlices {
			phs := &mpb.PieceHashAndSize{}
			phs.Hash = slice.FileHash
			phs.Size = uint32(slice.FileSize)
			ufpr.Piece[i] = phs
		}

		ufprsp, err := c.mclient.UploadFilePrepare(ctx, ufpr)
		if err != nil {
			log.Errorf("UploadFilePrepare error %v", err)
			return err
		}
		partition, err := c.UploadFileBatchByErasure(ufpr, ufprsp, fileSlices)
		if err != nil {
			return err
		}
		return c.UploadFileDone(ufpr, partition)
	} else if rsp.StoreType == mpb.FileStoreType_MultiReplica {
		return c.UploadFileByMultiReplica(req, rsp)
	} else {
		log.Error("unsupport type")
		return errors.New("no support")
	}

	return nil
}

func (c *ClientManager) UploadFileBatchByErasure(req *mpb.UploadFilePrepareReq, rsp *mpb.UploadFilePrepareResp, hashFiles []HashFile) (*mpb.Partition, error) {
	partition := &mpb.Partition{}
	providers, err := c.PingProvider(rsp.GetProvider())
	if err != nil {
		return nil, err
	}

	for i, pro := range providers {
		if i == 0 {
			block, err := c.UploadFileToErasureProvider(pro, hashFiles[i], true)
			if err != nil {
				return nil, err
			}
			partition.Block = append(partition.Block, block)
		} else {
			block, err := c.UploadFileToErasureProvider(pro, hashFiles[i], false)
			if err != nil {
				return nil, err
			}
			partition.Block = append(partition.Block, block)
		}
	}
	return partition, nil
}

func (c *ClientManager) UploadFileToErasureProvider(pro *mpb.ErasureCodeProvider, fileInfo HashFile, first bool) (*mpb.Block, error) {
	block := &mpb.Block{}
	conn, err := grpc.Dial(pro.GetServer(), grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		return nil, err
	}
	defer conn.Close()
	pclient := pb.NewProviderServiceClient(conn)

	ha := pro.GetHashAuth()[0]
	err = client.Store(pclient, fileInfo.FileName, ha.GetAuth(), ha.GetTicket(), fileInfo.FileHash, uint64(fileInfo.FileSize), first)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	block.Hash = fileInfo.FileHash
	block.Size = uint32(fileInfo.FileSize)
	block.BlockSeq = uint32(fileInfo.SliceIndex)
	block.Checksum = true
	block.StoreNodeId = [][]byte{}
	block.StoreNodeId = append(block.StoreNodeId, []byte(pro.GetNodeId()))

	return block, nil
}

func (c *ClientManager) UploadFileToReplicaProvider(pro *mpb.ReplicaProvider, fileInfo HashFile) error {
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

func (c *ClientManager) UploadFileByMultiReplica(req *mpb.CheckFileExistReq, rsp *mpb.CheckFileExistResp) error {

	fileInfo := HashFile{}
	fileInfo.FileName = req.FileName
	fileInfo.FileSize = int64(req.FileSize)
	fileInfo.FileHash = req.FileHash
	fileInfo.SliceIndex = 0
	for _, pro := range rsp.GetProvider() {
		c.UploadFileToReplicaProvider(pro, fileInfo)
	}

	return nil
}

func (c *ClientManager) UploadFileDone(req *mpb.UploadFilePrepareReq, partition *mpb.Partition) error {
	ufdr := &mpb.UploadFileDoneReq{}
	ufdr.Version = 1
	ufdr.NodeId = req.GetNodeId()
	ufdr.FileHash = req.FileHash
	ufdr.FileSize = req.FileSize
	//Todo
	//ufdr.Sign =
	ufdr.Timestamp = uint64(time.Now().UTC().Unix())
	ufdr.Partition = append(ufdr.Partition, partition)
	ctx := context.Background()
	ufdrsp, err := c.mclient.UploadFileDone(ctx, ufdr)
	if err != nil {
		return err
	}
	fmt.Printf("done: %d\n", ufdrsp.GetDone())
	return nil
}

type HashFile struct {
	FileSize   int64
	FileName   string
	FileHash   []byte
	SliceIndex int
}

func rsEncoder(outDir, fname string, dataShards, parShards int) ([]HashFile, error) {
	enc, err := reedsolomon.NewStream(dataShards, parShards)
	checkErr(err)

	fmt.Println("Opening", fname)
	f, err := os.Open(fname)
	checkErr(err)

	instat, err := f.Stat()
	checkErr(err)

	shards := dataShards + parShards
	out := make([]*os.File, shards)

	// Create the resulting files.
	dir, file := filepath.Split(fname)
	if outDir != "" {
		dir = outDir
	}
	for i := range out {
		outfn := fmt.Sprintf("%s.%d", file, i)
		fmt.Println("Creating", outfn)
		out[i], err = os.Create(filepath.Join(dir, outfn))
		checkErr(err)
	}

	// Split into files.
	data := make([]io.Writer, dataShards)
	for i := range data {
		data[i] = out[i]
	}
	// Do the split
	err = enc.Split(f, data, instat.Size())
	checkErr(err)

	// Close and re-open the files.
	input := make([]io.Reader, dataShards)

	for i := range data {
		out[i].Close()
		f, err := os.Open(out[i].Name())
		checkErr(err)
		input[i] = f
		defer f.Close()
	}

	// Create parity output writers
	parity := make([]io.Writer, parShards)
	for i := range parity {
		parity[i] = out[dataShards+i]
		defer out[dataShards+i].Close()
	}

	// Encode parity
	err = enc.Encode(input, parity)
	checkErr(err)
	fmt.Printf("File split into %d data + %d parity shards.\n", dataShards, parShards)
	result := []HashFile{}
	for i := range out {
		outfn := fmt.Sprintf("%s.%d", file, i)
		hash, err := util_hash.Sha1File(outfn)
		if err != nil {
			panic(err)
		}
		fileInfo, err := os.Stat(outfn)
		if err != nil {
			panic(err)
		}
		fmt.Printf("filename %s, hash %+v ,size %d\n", outfn, hash, fileInfo.Size())
		hf := HashFile{}
		hf.FileHash = hash
		hf.FileName = outfn
		hf.FileSize = fileInfo.Size()
		hf.SliceIndex = i
		result = append(result, hf)
	}
	return result, nil
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		os.Exit(2)
	}
}
