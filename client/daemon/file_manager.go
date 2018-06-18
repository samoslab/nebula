package daemon

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	collectClient "github.com/samoslab/nebula/client/collector_client"

	"github.com/samoslab/nebula/client/common"
	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/order"
	client "github.com/samoslab/nebula/client/provider_client"
	pb "github.com/samoslab/nebula/provider/pb"
	mpb "github.com/samoslab/nebula/tracker/metadata/pb"
	util_file "github.com/samoslab/nebula/util/file"
	util_hash "github.com/samoslab/nebula/util/hash"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
)

var (
	// ReplicaFileSize using replication if file size less than
	ReplicaFileSize = int64(8 * 1024)

	// PartitionMaxSize  max size of one partition
	PartitionMaxSize = int64(256 * 1024 * 1024)

	// DefaultTempDir default temporary file
	DefaultTempDir = "/tmp/nebula_client"
)

// ClientManager client manager
type ClientManager struct {
	mclient    mpb.MatadataServiceClient
	NodeId     []byte
	TempDir    string
	Log        logrus.FieldLogger
	cfg        *config.ClientConfig
	serverConn *grpc.ClientConn
	PM         *common.ProgressManager
	OM         *order.OrderManager
}

// NewClientManager create manager
func NewClientManager(log logrus.FieldLogger, webcfg config.Config, cfg *config.ClientConfig) (*ClientManager, error) {
	if webcfg.TrackerServer == "" {
		return nil, errors.New("tracker server nil")
	}
	if webcfg.CollectServer == "" {
		return nil, errors.New("collect server nil")
	}
	if cfg == nil {
		return nil, errors.New("client config nil")
	}
	conn, err := grpc.Dial(webcfg.TrackerServer, grpc.WithInsecure())
	if err != nil {
		log.Errorf("RPC Dial failed: %s", err.Error())
		return nil, err
	}
	log.Infof("tracker server %s", webcfg.TrackerServer)

	om := order.NewOrderManager(webcfg.TrackerServer, log, cfg.Node.PriKey, cfg.Node.NodeId)

	c := &ClientManager{
		serverConn: conn,
		//collectConn:   connCollect,
		Log:     log,
		cfg:     cfg,
		TempDir: DefaultTempDir,
		NodeId:  cfg.Node.NodeId,
		PM:      common.NewProgressManager(),
		mclient: mpb.NewMatadataServiceClient(conn),
		OM:      om,
	}

	// set collect node
	collectClient.NodePtr = cfg.Node

	log.Infof("temp dir %s", c.TempDir)
	if _, err := os.Stat(c.TempDir); os.IsNotExist(err) {
		//create the dir.
		if err := os.MkdirAll(c.TempDir, 0744); err != nil {
			panic(err)
		}
	}

	collectClient.Start(webcfg.CollectServer)

	return c, nil
}

// Shutdown shutdown tracker connection
func (c *ClientManager) Shutdown() {
	c.serverConn.Close()
	collectClient.Stop()
}

func (c *ClientManager) getPingTime(pro *mpb.BlockProviderAuth) int {
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	timeStart := time.Now().Unix()
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		c.Log.Errorf("RPC Dial failed: %s", err.Error())
		return 99999
	}
	defer conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pclient := pb.NewProviderServiceClient(conn)
	req := &pb.PingReq{
		Version: common.Version,
	}
	_, err = pclient.Ping(ctx, req)
	if err != nil {
		return 99999
	}
	timeEnd := time.Now().Unix()
	return int(timeEnd - timeStart)
}

// PingProvider ping provider
func (c *ClientManager) PingProvider(pros []*mpb.BlockProviderAuth, needNum int) ([]*mpb.BlockProviderAuth, error) {
	//todo if provider ip is same
	type SortablePro struct {
		Pro   *mpb.BlockProviderAuth
		Delay int
	}

	sortPros := []SortablePro{}
	// TODO can ping concurrent
	for _, bpa := range pros {
		pingTime := c.getPingTime(bpa)
		sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: pingTime})
	}

	sort.Slice(sortPros, func(i, j int) bool { return sortPros[i].Delay < sortPros[j].Delay })

	availablePros := []*mpb.BlockProviderAuth{}
	for _, proInfo := range sortPros {
		availablePros = append(availablePros, proInfo.Pro)
	}

	//return availablePros[0:needNum], nil
	return pros, nil
}

// UploadDir upload all files in dir to provider
func (c *ClientManager) UploadDir(parent string, interactive, newVersion bool) error {
	log := c.Log
	if !filepath.IsAbs(parent) {
		return fmt.Errorf("path %s must absolute", parent)
	}
	dirs, files, err := GetDirsAndFiles(parent)
	if err != nil {
		return err
	}
	log.Debugf("dirs %+v", dirs)
	log.Debugf("files %+v", files)
	for _, dpair := range dirs {
		_, err := c.MkFolder(dpair.Parent, []string{dpair.Name}, interactive)
		if err != nil {
			return err
		}
	}
	for _, fname := range files {
		err := c.UploadFile(fname, interactive, newVersion)
		if err != nil {
			return nil
		}
	}
	return nil
}

// UploadFile upload file to provider
func (c *ClientManager) UploadFile(filename string, interactive, newVersion bool) error {
	log := c.Log
	req, rsp, err := c.CheckFileExists(filename, interactive, newVersion)
	if err != nil {
		return common.StatusErrFromError(err)
	}

	log.Infof("check file exists rsp code:%d", rsp.GetCode())
	if rsp.GetCode() == 0 {
		log.Infof("upload success %s %s", filename, rsp.GetErrMsg())
		return nil
	}
	// 1 can upload
	if rsp.GetCode() != 1 {
		return fmt.Errorf("%d:%s", rsp.GetCode(), rsp.GetErrMsg())
	}

	log.Infof("[upload file] %s", filename)
	switch rsp.GetStoreType() {
	case mpb.FileStoreType_MultiReplica:
		log.Infof("upload manner is multi-replication")
		c.PM.SetProgress(filename, 0, req.FileSize)
		partitions, err := c.uploadFileByMultiReplica(filename, req, rsp)
		if err != nil {
			return err
		}
		return c.UploadFileDone(req, partitions)
	case mpb.FileStoreType_ErasureCode:
		log.Infof("upload manner is erasure")
		partFiles := []string{}
		var err error
		fileSize := int64(req.GetFileSize())
		if fileSize > PartitionMaxSize {
			chunkSize, chunkNum := GetChunkSizeAndNum(fileSize, PartitionMaxSize)
			partFiles, err = FileSplit(c.TempDir, filename, fileSize, chunkSize, int64(chunkNum))
			if err != nil {
				return err
			}
		} else {
			partFiles = append(partFiles, filename)
		}

		log.Infof("file %s need split to %d partitions", req.GetFileName(), len(partFiles))

		dataShards := int(rsp.GetDataPieceCount())
		verifyShards := int(rsp.GetVerifyPieceCount())

		log.Infof("prepare response gave %d dataShards, %d verifyShards", dataShards, verifyShards)

		fileInfos := []common.PartitionFile{}

		realSizeAfterRS := int64(0)
		for _, fname := range partFiles {
			fileSlices, err := c.onlyFileSplit(fname, dataShards, verifyShards)
			if err != nil {
				return err
			}
			fileInfos = append(fileInfos, common.PartitionFile{
				FileName:       fname,
				Pieces:         fileSlices,
				OriginFileName: req.FileName,
				OriginFileHash: req.FileHash,
				OriginFileSize: req.FileSize,
			})
			log.Infof("file %s need split to %d blocks", fname, len(fileSlices))
			for _, fs := range fileSlices {
				log.Debugf("erasure block files %s index %d", fs.FileName, fs.SliceIndex)
				c.PM.SetPartitionMap(fs.FileName, filename)
				realSizeAfterRS += fs.FileSize
			}
			c.PM.SetPartitionMap(fname, filename)
		}

		c.PM.SetProgress(filename, 0, uint64(realSizeAfterRS))

		// delete temporary file
		defer func() {
			for _, partInfo := range fileInfos {
				for _, slice := range partInfo.Pieces {
					deleteTemporaryFile(log, slice.FileName)
				}
				if len(fileInfos) != 1 {
					deleteTemporaryFile(log, partInfo.FileName)
				}
			}
		}()

		ufpr := &mpb.UploadFilePrepareReq{
			Version:   common.Version,
			NodeId:    req.NodeId,
			FileHash:  req.FileHash,
			FileSize:  req.FileSize,
			Timestamp: common.Now(),
			Partition: make([]*mpb.SplitPartition, len(partFiles)),
		}
		block := 0
		for i, partInfo := range fileInfos {
			phslist := []*mpb.PieceHashAndSize{}
			for j, slice := range partInfo.Pieces {
				phs := &mpb.PieceHashAndSize{
					Hash: slice.FileHash,
					Size: uint32(slice.FileSize),
				}
				phslist = append(phslist, phs)
				log.Debugf("%s %dth piece", slice.FileName, j)
				block++
			}
			ufpr.Partition[i] = &mpb.SplitPartition{phslist}
			log.Debugf("%s is %dth partitions", partInfo.FileName, i)
		}
		log.Infof("upload request has %d partitions, %d pieces", len(ufpr.Partition), block)
		err = ufpr.SignReq(c.cfg.Node.PriKey)
		if err != nil {
			return err
		}

		ctx := context.Background()
		log.Infof("send prepare request for %s", req.GetFileName())
		ufprsp, err := c.mclient.UploadFilePrepare(ctx, ufpr)
		if err != nil {
			log.Errorf("UploadFilePrepare error %v", err)
			return common.StatusErrFromError(err)
		}

		rspPartitions := ufprsp.GetPartition()
		log.Infof("upload prepare response partitions count:%d", len(rspPartitions))

		if len(rspPartitions) == 0 {
			return fmt.Errorf("only 0 partitions, not correct")
		}

		for i, part := range rspPartitions {
			auth := part.GetProviderAuth()
			for _, pa := range auth {
				log.Debugf("partition %d, server %s, port %d hashauth %d", i, pa.GetServer(), pa.GetPort(), len(pa.GetHashAuth()))
			}
		}

		partitions := []*mpb.StorePartition{}
		for i, partInfo := range fileInfos {

			partition, err := c.uploadFileBatchByErasure(ufpr, rspPartitions[i], partInfo, dataShards)
			if err != nil {
				return err
			}
			log.Debugf("partition %d has %d store blocks", i, len(partition.GetBlock()))
			partitions = append(partitions, partition)
		}
		log.Infof("there are %d store partitions", len(partitions))

		return c.UploadFileDone(req, partitions)

	}
	return nil
}

func deleteTemporaryFile(log logrus.FieldLogger, filename string) {
	log.Debugf("delete file %s", filename)
	if err := os.Remove(filename); err != nil {
		log.Errorf("delete %s failed, error %v", filename, err)
	}
}

func (c *ClientManager) CheckFileExists(filename string, interactive, newVersion bool) (*mpb.CheckFileExistReq, *mpb.CheckFileExistResp, error) {
	log := c.Log
	hash, err := util_hash.Sha1File(filename)
	if err != nil {
		return nil, nil, err
	}
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, nil, err
	}
	dir, fname := filepath.Split(filename)
	ctx := context.Background()
	req := &mpb.CheckFileExistReq{
		Version:     common.Version,
		FileSize:    uint64(fileInfo.Size()),
		Interactive: interactive,
		NewVersion:  newVersion,
		Parent:      &mpb.FilePath{&mpb.FilePath_Path{dir}},
		FileHash:    hash,
		NodeId:      c.NodeId,
		FileName:    fname,
		Timestamp:   common.Now(),
	}
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
	log.Infof("check file exist req:%s", req.GetFileName())
	rsp, err := c.mclient.CheckFileExist(ctx, req)
	return req, rsp, err
}

// MkFolder create folder
func (c *ClientManager) MkFolder(filepath string, folders []string, interactive bool) (bool, error) {
	log := c.Log
	ctx := context.Background()
	req := &mpb.MkFolderReq{
		Version:     common.Version,
		Parent:      &mpb.FilePath{&mpb.FilePath_Path{filepath}},
		Folder:      folders,
		NodeId:      c.NodeId,
		Interactive: interactive,
		Timestamp:   common.Now(),
	}
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return false, err
	}
	log.Infof("make folder :%+v, parent:%s", req.GetFolder(), filepath)
	rsp, err := c.mclient.MkFolder(ctx, req)
	if err != nil {
		return false, common.StatusErrFromError(err)
	}
	if rsp.GetCode() != 0 {
		if strings.Contains(rsp.GetErrMsg(), "System error: pq: duplicate key value") {
			log.Warning("folder exists %s", rsp.GetErrMsg())
			return true, nil
		}
		return false, fmt.Errorf("%s", rsp.GetErrMsg())
	}
	log.Infof("make folder response code:%d", rsp.GetCode())
	return true, nil
}

func (c *ClientManager) onlyFileSplit(filename string, dataNum, verifyNum int) ([]common.HashFile, error) {
	fileSlices, err := RsEncoder(c.Log, c.TempDir, filename, dataNum, verifyNum)
	if err != nil {
		c.Log.Errorf("reed se error %v", err)
		return nil, err
	}
	return fileSlices, nil

}

func (c *ClientManager) uploadFileBatchByErasure(req *mpb.UploadFilePrepareReq, rspPartition *mpb.ErasureCodePartition, partFile common.PartitionFile, dataShards int) (*mpb.StorePartition, error) {
	log := c.Log
	partition := &mpb.StorePartition{}
	partition.Block = []*mpb.StoreBlock{}
	phas := rspPartition.GetProviderAuth()
	providers, err := c.PingProvider(phas, len(phas))
	if err != nil {
		return nil, err
	}

	var errResult []error
	wg := sync.WaitGroup{}
	var mutex sync.Mutex
	for i, pro := range providers {
		wg.Add(1)
		checksum := i >= dataShards
		uploadParas := &common.UploadParameter{
			OriginFileHash: partFile.OriginFileHash,
			OriginFileSize: partFile.OriginFileSize,
			HF:             partFile.Pieces[i],
			Checksum:       checksum,
		}
		go func(pro *mpb.BlockProviderAuth, tm uint64, uploadPara *common.UploadParameter) {
			defer wg.Done()
			server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
			block, err := c.uploadFileToErasureProvider(pro, tm, uploadParas)
			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				log.Errorf("upload file %s error %v", uploadPara.HF.FileName, err)
				errResult = append(errResult, err)
				return
			}
			partition.Block = append(partition.Block, block)
			log.Debugf("%s upload to privider %s success", uploadParas.HF.FileName, server)
		}(pro, rspPartition.GetTimestamp(), uploadParas)
	}
	wg.Wait()
	if len(errResult) != 0 {
		return partition, errResult[0]
	}
	return partition, nil
}

func getOneOfPartition(pro *mpb.ErasureCodePartition) *mpb.BlockProviderAuth {
	pa := pro.GetProviderAuth()[0]
	return pa
}

func (c *ClientManager) uploadFileToErasureProvider(pro *mpb.BlockProviderAuth, tm uint64, uploadPara *common.UploadParameter) (*mpb.StoreBlock, error) {
	log := c.Log
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Errorf("RPC Dial failed: %s", err.Error())
		return nil, err
	}
	defer conn.Close()
	pclient := pb.NewProviderServiceClient(conn)

	ha := pro.GetHashAuth()[0]
	err = client.StorePiece(log, pclient, uploadPara, ha.GetAuth(), ha.GetTicket(), tm, c.PM)
	if err != nil {
		return nil, err
	}
	block := &mpb.StoreBlock{
		Hash:        uploadPara.HF.FileHash,
		Size:        uint64(uploadPara.HF.FileSize),
		BlockSeq:    uint32(uploadPara.HF.SliceIndex),
		Checksum:    uploadPara.Checksum,
		StoreNodeId: [][]byte{},
	}
	block.StoreNodeId = append(block.StoreNodeId, []byte(pro.GetNodeId()))

	return block, nil
}

func (c *ClientManager) uploadFileToReplicaProvider(pro *mpb.ReplicaProvider, uploadPara *common.UploadParameter) ([]byte, error) {
	log := c.Log
	fileInfo := uploadPara.HF
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Errorf("RPC Dail failed: %v", err)
		return nil, err
	}
	defer conn.Close()
	pclient := pb.NewProviderServiceClient(conn)
	log.Debugf("upload file %s hash %x size %d to %s", fileInfo.FileName, fileInfo.FileHash, fileInfo.FileSize, server)

	err = client.StorePiece(log, pclient, uploadPara, pro.GetAuth(), pro.GetTicket(), pro.GetTimestamp(), c.PM)
	if err != nil {
		log.Errorf("upload error %v", err)
		return nil, err
	}

	log.Infof("upload file %s success", fileInfo.FileName)

	return pro.GetNodeId(), nil
}

func (c *ClientManager) uploadFileByMultiReplica(filename string, req *mpb.CheckFileExistReq, rsp *mpb.CheckFileExistResp) ([]*mpb.StorePartition, error) {

	fileInfo := common.HashFile{
		FileName:   filename,
		FileSize:   int64(req.FileSize),
		FileHash:   req.FileHash,
		SliceIndex: 0,
	}
	uploadPara := &common.UploadParameter{
		OriginFileHash: req.FileHash,
		OriginFileSize: req.FileSize,
		HF:             fileInfo,
	}

	block := &mpb.StoreBlock{
		Hash:        fileInfo.FileHash,
		Size:        uint64(fileInfo.FileSize),
		BlockSeq:    uint32(fileInfo.SliceIndex),
		Checksum:    false,
		StoreNodeId: [][]byte{},
	}
	c.PM.SetPartitionMap(filename, filename)
	for _, pro := range rsp.GetProvider() {
		proID, err := c.uploadFileToReplicaProvider(pro, uploadPara)
		if err != nil {
			return nil, err
		}
		block.StoreNodeId = append(block.StoreNodeId, proID)
	}

	partition := &mpb.StorePartition{}
	partition.Block = append(partition.Block, block)
	partitions := []*mpb.StorePartition{partition}
	return partitions, nil
}

func (c *ClientManager) UploadFileDone(reqCheck *mpb.CheckFileExistReq, partitions []*mpb.StorePartition) error {
	req := &mpb.UploadFileDoneReq{
		Version:     common.Version,
		NodeId:      c.NodeId,
		FileHash:    reqCheck.GetFileHash(),
		FileSize:    reqCheck.GetFileSize(),
		FileName:    reqCheck.GetFileName(),
		FileModTime: reqCheck.GetFileModTime(),
		Parent:      reqCheck.GetParent(),
		Interactive: reqCheck.GetInteractive(),
		NewVersion:  reqCheck.GetNewVersion(),
		Timestamp:   common.Now(),
		Partition:   partitions,
	}
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}
	ctx := context.Background()
	c.Log.Infof("upload file %s done request", req.GetFileName())
	ufdrsp, err := c.mclient.UploadFileDone(ctx, req)
	if err != nil {
		return common.StatusErrFromError(err)
	}
	c.Log.Infof("upload done code: %d", ufdrsp.GetCode())
	if ufdrsp.GetCode() != 0 {
		return fmt.Errorf("%s", ufdrsp.GetErrMsg())
	}
	return nil
}

// ListFiles list files on dir
func (c *ClientManager) ListFiles(path string, pageSize, pageNum uint32, sortType string, ascOrder bool) ([]*DownFile, error) {
	c.Log.Infof("path %s, size %d, num %d, sortype %s, asc %v", path, pageSize, pageNum, sortType, ascOrder)
	req := &mpb.ListFilesReq{
		Version:   common.Version,
		Timestamp: common.Now(),
		NodeId:    c.NodeId,
		PageSize:  pageSize,
		PageNum:   pageNum,
	}
	switch sortType {
	case "name":
		req.SortType = mpb.SortType_Name
	case "size":
		req.SortType = mpb.SortType_Size
	case "modtime":
		req.SortType = mpb.SortType_ModTime
	default:
		req.SortType = mpb.SortType_Name
	}
	req.Parent = &mpb.FilePath{&mpb.FilePath_Path{path}}
	req.AscOrder = ascOrder
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	c.Log.Infof("list request path %s", req.Parent)
	rsp, err := c.mclient.ListFiles(ctx, req)

	if err != nil {
		return nil, common.StatusErrFromError(err)
	}

	if rsp.GetCode() != 0 {
		return nil, fmt.Errorf("errmsg %s", rsp.GetErrMsg())
	}
	fileLists := []*DownFile{}
	for _, info := range rsp.GetFof() {
		hash := hex.EncodeToString(info.GetFileHash())
		id := hex.EncodeToString(info.GetId())
		df := &DownFile{
			ID:       id,
			FileHash: hash,
			FileName: info.GetName(),
			Folder:   info.GetFolder(),
			FileSize: info.GetFileSize()}
		fileLists = append(fileLists, df)
	}
	return fileLists, nil
}

// DownloadDir download dir
func (c *ClientManager) DownloadDir(path string) error {
	log := c.Log
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path %s must absolute", path)
	}
	errResult := []error{}
	page := uint32(1)
	for {
		// list 1 page 100 items order by name
		downFiles, err := c.ListFiles(path, 100, page, "name", true)
		if err != nil {
			return err
		}
		if len(downFiles) == 0 {
			break
		}
		// next page
		page++
		for _, fileInfo := range downFiles {
			currentFile := filepath.Join(path, fileInfo.FileName)
			if fileInfo.Folder {
				log.Infof("create folder %s", currentFile)
				if _, err := os.Stat(currentFile); os.IsNotExist(err) {
					os.Mkdir(currentFile, 0744)
				}
				err = c.DownloadDir(currentFile)
				if err != nil {
					log.Errorf("recursive download %s failed %v", currentFile, err)
					return err
				}
			} else {
				log.Infof("start download %s", currentFile)
				if fileInfo.FileSize == 0 {
					log.Infof("only create %s because file size is 0", fileInfo.FileName)
					saveFile(currentFile, []byte{})
				} else {
					err = c.DownloadFile(currentFile, fileInfo.FileHash, fileInfo.FileSize)
					if err != nil {
						log.Errorf("download file %s error %v", currentFile, err)
						errResult = append(errResult, fmt.Errorf("%s %v", currentFile, common.StatusErrFromError(err)))
						//return err
					}
				}
			}
		}
	}
	if len(errResult) > 0 {
		for _, err := range errResult {
			log.Errorf("download error: %v", err)
		}
		return errResult[0]
	}
	return nil
}

// DownloadFile download file
func (c *ClientManager) DownloadFile(downFileName string, filehash string, fileSize uint64) error {
	log := c.Log
	fileHash, err := hex.DecodeString(filehash)
	if err != nil {
		return err
	}
	req := &mpb.RetrieveFileReq{
		Version:   common.Version,
		NodeId:    c.NodeId,
		Timestamp: common.Now(),
		FileHash:  fileHash,
		FileSize:  fileSize,
	}
	err = req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}
	c.PM.SetProgress(downFileName, 0, req.FileSize)

	ctx := context.Background()
	log.Infof("download file request hash:%x, size %d", fileHash, fileSize)
	rsp, err := c.mclient.RetrieveFile(ctx, req)
	if err != nil {
		return common.StatusErrFromError(err)
	}
	if rsp.GetCode() != 0 {
		return fmt.Errorf("%s", rsp.GetErrMsg())
	}

	// tiny file
	if filedata := rsp.GetFileData(); filedata != nil {
		saveFile(downFileName, filedata)
		c.PM.SetProgress(downFileName, req.FileSize, req.FileSize)
		log.Infof("download tiny file %s", downFileName)
		return nil
	}

	partitions := rsp.GetPartition()
	log.Infof("there is %d partitions", len(partitions))
	partitionCount := len(partitions)
	if partitionCount == 1 {
		blockCount := len(partitions[0].GetBlock())
		if blockCount == 1 { // 1 partition 1 block is multiReplica
			log.Infof("file %s is multi replication files", downFileName)
			// for progress stats
			for _, block := range partitions[0].GetBlock() {
				c.PM.SetPartitionMap(hex.EncodeToString(block.GetHash()), downFileName)
			}
			_, _, _, _, err := c.saveFileByPartition(downFileName, partitions[0], rsp.GetTimestamp(), req.FileHash, req.FileSize, true)
			return err
		}
	}

	// erasure files handle by below codes

	// for progress stats
	realSizeAfterRS := uint64(0)
	for i, partition := range partitions {
		for j, block := range partition.GetBlock() {
			c.PM.SetPartitionMap(hex.EncodeToString(block.GetHash()), downFileName)
			realSizeAfterRS += block.GetSize()
			fmt.Printf("partition %d block %d hash %x size %d checksum %v seq %d\n", i, j, block.Hash, block.Size, block.Checksum, block.BlockSeq)
		}
	}
	c.PM.SetProgress(downFileName, 0, realSizeAfterRS)

	if len(partitions) == 1 {
		partition := partitions[0]
		datas, paritys, failedCount, middleFiles, err := c.saveFileByPartition(downFileName, partition, rsp.GetTimestamp(), req.FileHash, req.FileSize, false)
		if failedCount > paritys {
			log.Errorf("file %s cannot be recoved!!!", downFileName)
			return err
		}
		if err != nil {
			log.Errorf("save file by partition error: %v, but file still can be recoverd", err)
		}

		// delete middle files
		defer func() {
			for _, file := range middleFiles {
				deleteTemporaryFile(log, file)
			}
		}()

		log.Infof("dataShards %d, parityShards %d, failedCount %d", datas, paritys, failedCount)

		_, onlyFileName := filepath.Split(downFileName)
		tempDownFileName := filepath.Join(c.TempDir, onlyFileName)
		err = RsDecoder(log, tempDownFileName, "", int64(req.FileSize), datas, paritys)
		if err != nil {
			return err
		}

		defer func() {
			// delete file in case rename failed
			if util_file.Exists(tempDownFileName) {
				deleteTemporaryFile(log, tempDownFileName)
			}
		}()

		if err := os.Rename(tempDownFileName, downFileName); err != nil {
			return err
		}

		return nil
	}
	partFiles := []string{}
	for i, partition := range partitions {
		partFileName := fmt.Sprintf("%s.part.%d", downFileName, i)
		datas, paritys, failedCount, middleFiles, err := c.saveFileByPartition(partFileName, partition, rsp.GetTimestamp(), req.FileHash, req.FileSize, false)
		if failedCount > paritys {
			log.Errorf("middle file %s cannot be recoved!!!", partFileName)
			return err
		}
		if err != nil {
			log.Errorf("save file by partition error %v, but file still can be recoverd", err)
		}
		log.Infof("dataShards %d, parityShards %d, failedCount %d", datas, paritys, failedCount)
		_, onlyFileName := filepath.Split(partFileName)
		tempDownFileName := filepath.Join(c.TempDir, onlyFileName)
		// file real size can be calcauted by filesize and partition number
		partitionFileSize := ReverseCalcuatePartFileSize(int64(req.FileSize), len(partitions), i)
		log.Infof("partition %d, size %d", i, partitionFileSize)
		err = RsDecoder(log, tempDownFileName, "", int64(partitionFileSize), datas, paritys)
		if err != nil {
			return err
		}

		partFiles = append(partFiles, tempDownFileName)
		// delete middle files
		for _, file := range middleFiles {
			deleteTemporaryFile(log, file)
		}

	}

	defer func() {
		for _, file := range partFiles {
			deleteTemporaryFile(log, file)
		}
	}()

	log.Infof("file %s is erasure files", downFileName)
	if err := FileJoin(downFileName, partFiles); err != nil {
		log.Errorf("file %s join failed, part files %+v", downFileName, partFiles)
		return err
	}
	return nil
}

func (c *ClientManager) saveFileByPartition(filename string, partition *mpb.RetrievePartition, tm uint64, fileHash []byte, fileSize uint64, multiReplica bool) (int, int, int, []string, error) {
	log := c.Log
	log.Infof("there is %d blocks", len(partition.GetBlock()))
	dataShards := 0
	parityShards := 0
	failedCount := 0
	middleFiles := []string{}
	errArray := []string{}
	for _, block := range partition.GetBlock() {
		if block.GetChecksum() {
			parityShards++
		} else {
			dataShards++
		}
		nodes := block.GetStoreNode()
		node := nodes[0]
		server := fmt.Sprintf("%s:%d", node.GetServer(), node.GetPort())
		tempFileName := filename
		if !multiReplica {
			_, onlyFileName := filepath.Split(filename)
			tempFileName = filepath.Join(c.TempDir, fmt.Sprintf("%s.%d", onlyFileName, block.GetBlockSeq()))
		}
		log.Infof("[part file] %s, hash %x retrieve from %s", tempFileName, block.GetHash(), server)
		conn, err := grpc.Dial(server, grpc.WithInsecure(), grpc.WithTimeout(3*time.Second), grpc.WithBlock())
		if err != nil {
			log.Errorf("RPC Dial %s failed, : %s", server, err.Error())
			log.Errorf("[part file] %s  retrieve failed", tempFileName)
			errArray = append(errArray, err.Error())
			middleFiles = append(middleFiles, tempFileName)
			failedCount++
			continue
		}
		pclient := pb.NewProviderServiceClient(conn)

		err = client.Retrieve(log, pclient, tempFileName, node.GetAuth(), node.GetTicket(), tm, fileHash, block.GetHash(), fileSize, block.GetSize(), c.PM)
		if err != nil {
			failedCount++
			errArray = append(errArray, err.Error())
			conn.Close()
			log.Errorf("[part file] %s  retrieve failed", tempFileName)
			middleFiles = append(middleFiles, tempFileName)
			continue
		}
		log.Infof("[part file] %s  retrieve success", tempFileName)
		middleFiles = append(middleFiles, tempFileName)
		conn.Close()
	}

	if len(errArray) > 0 {
		errRtn := fmt.Errorf("%s", strings.Join(errArray, "\n"))
		return dataShards, parityShards, failedCount, middleFiles, errRtn
	}
	return dataShards, parityShards, failedCount, middleFiles, nil
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

// RemoveFile download file
func (c *ClientManager) RemoveFile(target string, recursive bool, isPath bool) error {
	log := c.Log
	req := &mpb.RemoveReq{
		Version:   common.Version,
		NodeId:    c.NodeId,
		Timestamp: common.Now(),
		Recursive: recursive,
	}
	if isPath {
		req.Target = &mpb.FilePath{&mpb.FilePath_Path{target}}
	} else {
		id, err := hex.DecodeString(target)
		if err != nil {
			return err
		}
		log.Infof("delete file by id %s, binary id %s", target, id)
		req.Target = &mpb.FilePath{&mpb.FilePath_Id{id}}
	}

	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}

	log.Infof("remove file :%+v, recursion %v", req.Target, req.GetRecursive())
	rsp, err := c.mclient.Remove(context.Background(), req)
	if err != nil {
		return common.StatusErrFromError(err)
	}
	log.Infof("remove file rsp code :%d msg: %s", rsp.GetCode(), rsp.GetErrMsg())
	if rsp.GetCode() != 0 {
		return fmt.Errorf("%s", rsp.GetErrMsg())
	}
	return nil

}

// GetProgress returns progress rate
func (c *ClientManager) GetProgress(files []string) (map[string]float64, error) {
	return c.PM.GetProgress(files)
}
