package daemon

import (
	"context"
	"crypto/rsa"
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
	"github.com/samoslab/nebula/client/util/filetype"

	"github.com/samoslab/nebula/client/common"
	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/order"
	client "github.com/samoslab/nebula/client/provider_client"
	pb "github.com/samoslab/nebula/provider/pb"
	mpb "github.com/samoslab/nebula/tracker/metadata/pb"
	"github.com/samoslab/nebula/util/aes"
	util_file "github.com/samoslab/nebula/util/file"
	util_hash "github.com/samoslab/nebula/util/hash"
	rsalong "github.com/samoslab/nebula/util/rsa"
	"github.com/sirupsen/logrus"

	"github.com/samoslab/nebula/client/register"
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
	mclient       mpb.MatadataServiceClient
	NodeId        []byte
	TempDir       string
	Log           logrus.FieldLogger
	cfg           *config.ClientConfig
	serverConn    *grpc.ClientConn
	PM            *common.ProgressManager
	OM            *order.OrderManager
	Root          string
	SpaceM        *SpaceManager
	TrackerPubkey *rsa.PublicKey
	PubkeyHash    []byte
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

	rsaPubkey, pubkeyHash, err := register.GetPublicKey(webcfg.TrackerServer)
	if err != nil {
		return nil, err
	}

	om := order.NewOrderManager(webcfg.TrackerServer, log, cfg.Node.PriKey, cfg.Node.NodeId)

	spaceM := NewSpaceManager()
	for _, sp := range cfg.Space {
		spaceM.AddSpace(sp.SpaceNo, sp.Password, sp.Home)
	}

	c := &ClientManager{
		serverConn:    conn,
		Log:           log,
		cfg:           cfg,
		TempDir:       os.TempDir(),
		NodeId:        cfg.Node.NodeId,
		PM:            common.NewProgressManager(),
		mclient:       mpb.NewMatadataServiceClient(conn),
		OM:            om,
		SpaceM:        spaceM,
		TrackerPubkey: rsaPubkey,
		PubkeyHash:    pubkeyHash,
	}

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

// SetRoot set user root directory
func (c *ClientManager) SetRoot(path string) error {
	if !util_file.Exists(path) {
		return fmt.Errorf("%s not exists", path)
	}
	c.Root = path
	c.cfg.Root = path
	// todo save root into config
	return config.SaveClientConfig(c.cfg.SelfFileName, c.cfg)
}

// SetPassword set user privacy space password
func (c *ClientManager) SetPassword(sno uint32, password string) error {
	err := c.SpaceM.SetSpacePasswd(sno, password)
	if err != nil {
		return err
	}
	c.cfg.Space[sno].Password = password
	return config.SaveClientConfig(c.cfg.SelfFileName, c.cfg)
}

func (c *ClientManager) getPingTime(ip string, port uint32) int {
	server := fmt.Sprintf("%s:%d", ip, port)
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

// UsingBestProvider ping provider
func (c *ClientManager) UsingBestProvider(pros []*mpb.BlockProviderAuth, needNum int) ([]*mpb.BlockProviderAuth, error) {
	//todo if provider ip is same
	type SortablePro struct {
		Pro         *mpb.BlockProviderAuth
		Delay       int
		OriginIndex int
	}

	log := c.Log

	sortPros := []SortablePro{}
	// TODO can ping concurrent
	pingResultMap := map[int]int{}
	var pingResultMutex sync.Mutex

	var wg sync.WaitGroup

	for i, bpa := range pros {
		wg.Add(1)
		go func(i int, bpa *mpb.BlockProviderAuth) {
			defer wg.Done()
			pingTime := c.getPingTime(bpa.GetServer(), bpa.GetPort())
			pingResultMutex.Lock()
			defer pingResultMutex.Unlock()
			pingResultMap[i] = pingTime
		}(i, bpa)
	}
	wg.Wait()
	for i, bpa := range pros {
		pingTime, _ := pingResultMap[i]
		sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: pingTime, OriginIndex: i})
	}

	// TODO need consider Spare , Spare = false is backup provider
	//sort.Slice(sortPros, func(i, j int) bool { return sortPros[i].Delay < sortPros[j].Delay })
	workPros := []SortablePro{}
	backupPros := []SortablePro{}
	for _, proInfo := range sortPros {
		if proInfo.Pro.GetSpare() {
			workPros = append(workPros, proInfo)
		} else {
			backupPros = append(backupPros, proInfo)
		}
	}

	workedNum := len(workPros)
	backupNum := len(backupPros)

	backupMap := createBackupProvicer(workedNum, backupNum)

	availablePros := []*mpb.BlockProviderAuth{}
	for _, proInfo := range workPros {
		if proInfo.Delay == 99999 {
			// provider cannot connect , choose one from backup
			log.Errorf("provider %v cannot connected")
			if backupNum == 0 {
				log.Errorf("no backup provider")
				return nil, fmt.Errorf("one of provider cannot connected and no backup provider")
			}
			choosed := chooseBackupProvicer(proInfo.OriginIndex, backupMap)
			if choosed == -1 {
				log.Errorf("no availbe provider")
				return nil, fmt.Errorf("no more backup provider can be choosed")
			}
			availablePros = append(availablePros, backupPros[choosed].Pro)
		} else {
			availablePros = append(availablePros, proInfo.Pro)
		}
	}

	return availablePros[0:needNum], nil
	//return pros, nil
}

type IndexStatus struct {
	Index int
	Used  bool
}

func chooseBackupProvicer(current int, backupMap map[int][]IndexStatus) int {
	choosed := -1
	if arr, ok := backupMap[current]; ok {
		for i, _ := range arr {
			if !arr[i].Used {
				choosed = arr[i].Index
				arr[i].Used = true
				backupMap[current] = arr
				return choosed
			}
		}
	}
	return choosed
}

func createBackupProvicer(workedNum, backupNum int) map[int][]IndexStatus {
	// workedNum = 40 , backupNum = 10
	// span = 40 /10 * 2 = 8 nextGroup = 10 /2 = 5
	// 0-7 --> [0, 5] ; 8-15 --> [1, 6] ; 16-23 --> [2, 7] ; 24-31 -->[3, 8]; 32-39 --> [4, 9]
	backupMap := map[int][]IndexStatus{}
	span := (workedNum / backupNum) * 2
	nextGroup := backupNum / 2
	for i := 0; i < workedNum; i++ {
		backupMap[i] = append(backupMap[i], IndexStatus{Index: i / span, Used: false})
		backupMap[i] = append(backupMap[i], IndexStatus{Index: i/span + nextGroup, Used: false})
	}

	return backupMap
}

// BestRetrieveNode ping retrieve node
func (c *ClientManager) BestRetrieveNode(pros []*mpb.RetrieveNode) *mpb.RetrieveNode {
	//todo if provider ip is same
	type SortablePro struct {
		Pro   *mpb.RetrieveNode
		Delay int
	}

	sortPros := []SortablePro{}
	for _, bpa := range pros {
		pingTime := c.getPingTime(bpa.GetServer(), bpa.GetPort())
		sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: pingTime})
	}

	sort.Slice(sortPros, func(i, j int) bool { return sortPros[i].Delay < sortPros[j].Delay })

	availablePros := []*mpb.RetrieveNode{}
	for _, proInfo := range sortPros {
		availablePros = append(availablePros, proInfo.Pro)
	}

	return availablePros[0]
}

// UploadDir upload all files in dir to provider
func (c *ClientManager) UploadDir(parent, dest string, interactive, newVersion bool, sno uint32) error {
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
	// replace parent by dest, parent is D://work, dest = /cloud, D://work/abc.txt -> /cloud/abc.txt
	newDirs := []DirPair{}
	newFiles := []string{}
	for _, dir := range dirs {
		actualDir := strings.Replace(dir.Parent, parent, dest, 1)
		newDP := DirPair{
			Parent: actualDir,
			Name:   dir.Name,
		}
		newDirs = append(newDirs, newDP)
	}
	for _, file := range files {
		actualFile := strings.Replace(file, parent, dest, 1)
		newFiles = append(newFiles, actualFile)
	}
	for _, dpair := range newDirs {
		_, err := c.MkFolder(dpair.Parent, []string{dpair.Name}, interactive, sno)
		if err != nil {
			return err
		}
	}
	for _, fname := range newFiles {
		err := c.UploadFile(fname, dest, interactive, newVersion, sno)
		if err != nil {
			return nil
		}
	}
	return nil
}

// UploadFile upload file to provider
func (c *ClientManager) UploadFile(filename, dest string, interactive, newVersion bool, sno uint32) error {
	log := c.Log
	password, err := c.SpaceM.GetSpacePasswd(sno)
	if err != nil {
		log.Errorf("get encrypt key of space no %d error %v", sno, err)
		return err
	}
	req, rsp, err := c.CheckFileExists(filename, dest, interactive, newVersion, password, sno)
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
		// encrypt file
		encry, err := c.SpaceM.GetSpacePasswd(sno)
		if err != nil {
			log.Errorf("get encrypt key of space no %d error %v", sno, err)
			return err
		}
		if len(encry) != 0 {
			_, onlyFileName := filepath.Split(filename)
			encryptedFileName := filepath.Join(c.TempDir, onlyFileName)
			err := aes.EncryptFile(filename, encry, encryptedFileName)
			if err != nil {
				log.Errorf("encrypt %s error %v", filename, err)
				return err
			}
			filename = encryptedFileName
			defer func() {
				deleteTemporaryFile(log, encryptedFileName)
			}()
		}
		partitions, err := c.uploadFileByMultiReplica(filename, req, rsp)
		if err != nil {
			return err
		}
		return c.UploadFileDone(req, partitions, password)
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
			fileSlices, err := c.onlyFileSplit(fname, dataShards, verifyShards, sno)
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

		return c.UploadFileDone(req, partitions, password)

	}
	return nil
}

func deleteTemporaryFile(log logrus.FieldLogger, filename string) {
	log.Debugf("delete file %s", filename)
	if err := os.Remove(filename); err != nil {
		log.Errorf("delete %s failed, error %v", filename, err)
	}
}

func (c *ClientManager) CheckFileExists(filename, dest string, interactive, newVersion bool, password []byte, sno uint32) (*mpb.CheckFileExistReq, *mpb.CheckFileExistResp, error) {
	log := c.Log
	hash, err := util_hash.Sha1File(filename)
	if err != nil {
		return nil, nil, err
	}
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return nil, nil, err
	}
	fileType := filetype.FileType(filename)
	_, fname := filepath.Split(filename)
	ctx := context.Background()
	encryptKey, err := rsalong.EncryptLong(c.TrackerPubkey, password, 256)
	if err != nil {
		return nil, nil, err
	}
	req := &mpb.CheckFileExistReq{
		Version:       common.Version,
		FileSize:      uint64(fileInfo.Size()),
		Interactive:   interactive,
		NewVersion:    newVersion,
		Parent:        &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{dest}, SpaceNo: sno},
		FileHash:      hash,
		NodeId:        c.NodeId,
		FileName:      fname,
		Timestamp:     common.Now(),
		FileType:      fileType.Value,
		EncryptKey:    encryptKey,
		PublicKeyHash: c.PubkeyHash,
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
		if len(password) != 0 {
			fileData, err = aes.Encrypt(fileData, password)
			if err != nil {
				log.Errorf("encrypt %s error %v", filename, err)
				return nil, nil, err
			}
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
func (c *ClientManager) MkFolder(filepath string, folders []string, interactive bool, sno uint32) (bool, error) {
	log := c.Log
	ctx := context.Background()
	req := &mpb.MkFolderReq{
		Version:     common.Version,
		Parent:      &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{filepath}, SpaceNo: sno},
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

func (c *ClientManager) onlyFileSplit(filename string, dataNum, verifyNum int, sno uint32) ([]common.HashFile, error) {
	log := c.Log
	fileSlices, err := RsEncoder(c.Log, c.TempDir, filename, dataNum, verifyNum)
	if err != nil {
		log.Errorf("reedsolomon encoder error %v", err)
		return nil, err
	}
	encry, err := c.SpaceM.GetSpacePasswd(sno)
	if err != nil {
		log.Errorf("get encrypt key of space no %d error %v", sno, err)
		return nil, err
	}
	if len(encry) != 0 {
		for _, ff := range fileSlices {
			err := aes.EncryptFile(ff.FileName, encry, ff.FileName)
			if err != nil {
				log.Errorf("encrypt %s error %v", filename, err)
				return nil, err
			}

		}
	}
	return fileSlices, nil

}

func (c *ClientManager) uploadFileBatchByErasure(req *mpb.UploadFilePrepareReq, rspPartition *mpb.ErasureCodePartition, partFile common.PartitionFile, dataShards int) (*mpb.StorePartition, error) {
	log := c.Log
	partition := &mpb.StorePartition{}
	partition.Block = []*mpb.StoreBlock{}
	phas := rspPartition.GetProviderAuth()
	providers, err := c.UsingBestProvider(phas, len(phas))
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

func (c *ClientManager) UploadFileDone(reqCheck *mpb.CheckFileExistReq, partitions []*mpb.StorePartition, password []byte) error {
	encryptKey, err := rsalong.EncryptLong(c.TrackerPubkey, password, 256)
	if err != nil {
		return err
	}
	req := &mpb.UploadFileDoneReq{
		Version:       common.Version,
		NodeId:        c.NodeId,
		FileHash:      reqCheck.GetFileHash(),
		FileSize:      reqCheck.GetFileSize(),
		FileName:      reqCheck.GetFileName(),
		FileModTime:   reqCheck.GetFileModTime(),
		Parent:        reqCheck.GetParent(),
		Interactive:   reqCheck.GetInteractive(),
		NewVersion:    reqCheck.GetNewVersion(),
		Timestamp:     common.Now(),
		Partition:     partitions,
		EncryptKey:    encryptKey,
		PublicKeyHash: c.PubkeyHash,
	}
	err = req.SignReq(c.cfg.Node.PriKey)
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
func (c *ClientManager) ListFiles(path string, pageSize, pageNum uint32, sortType string, ascOrder bool, sno uint32) (*FilePages, error) {
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
	req.Parent = &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{path}, SpaceNo: sno}
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

	return &FilePages{
		Total: rsp.GetTotalRecord(),
		Files: fileLists,
	}, nil
}

// DownloadDir download dir
func (c *ClientManager) DownloadDir(path, destDir string, sno uint32) error {
	log := c.Log
	if !filepath.IsAbs(path) {
		return fmt.Errorf("path %s must absolute", path)
	}
	errResult := []error{}
	page := uint32(1)
	for {
		// list 1 page 100 items order by name
		downFiles, err := c.ListFiles(path, 100, page, "name", true, sno)
		if err != nil {
			return err
		}
		if downFiles.Total == 0 {
			break
		}
		// next page
		page++
		for _, fileInfo := range downFiles.Files {
			currentFile := filepath.Join(destDir, fileInfo.FileName)
			if fileInfo.Folder {
				log.Infof("create folder %s", currentFile)
				if _, err := os.Stat(currentFile); os.IsNotExist(err) {
					os.Mkdir(currentFile, 0744)
				}
				err = c.DownloadDir(currentFile, destDir, sno)
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
					err = c.DownloadFile(currentFile, destDir, fileInfo.FileHash, fileInfo.FileSize, sno)
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
func (c *ClientManager) DownloadFile(downFileName, destDir, filehash string, fileSize uint64, sno uint32) error {
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
		SpaceNo:   sno,
	}
	err = req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}
	_, fileName := filepath.Split(downFileName)
	downFileName = filepath.Join(destDir, fileName)
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

	encryptKey := rsp.GetEncryptKey()
	password, err := rsalong.DecryptLong(c.cfg.Node.PriKey, encryptKey, 256)
	if err != nil {
		return err
	}
	// tiny file
	if filedata := rsp.GetFileData(); filedata != nil {
		if len(password) != 0 {
			filedata, err = aes.Decrypt(filedata, password)
			if err != nil {
				log.Errorf("decrypted %s error %v", downFileName, err)
				return err
			}
		}
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
			if err != nil {
				return err
			}
			if len(password) != 0 {
				return aes.DecryptFile(downFileName, password, downFileName)
			}
			return nil
		}
	}

	// erasure files handle by below codes

	// for progress stats
	realSizeAfterRS := uint64(0)
	for i, partition := range partitions {
		for j, block := range partition.GetBlock() {
			c.PM.SetPartitionMap(hex.EncodeToString(block.GetHash()), downFileName)
			realSizeAfterRS += block.GetSize()
			log.Infof("partition %d block %d hash %x size %d checksum %v seq %d\n", i, j, block.Hash, block.Size, block.Checksum, block.BlockSeq)
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

		if len(password) != 0 {
			for _, file := range middleFiles {
				if err := aes.DecryptFile(file, password, file); err != nil {
					return err
				}
			}
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
		partFileName := fmt.Sprintf("%s.%s.%d", downFileName, TEMP_NAMESPACE, i)
		datas, paritys, failedCount, middleFiles, err := c.saveFileByPartition(partFileName, partition, rsp.GetTimestamp(), req.FileHash, req.FileSize, false)
		if failedCount > paritys {
			log.Errorf("middle file %s cannot be recoved!!!", partFileName)
			return err
		}
		if err != nil {
			log.Errorf("save file by partition error %v, but file still can be recoverd", err)
		}
		if len(password) != 0 {
			for _, file := range middleFiles {
				if err := aes.DecryptFile(file, password, file); err != nil {
					return err
				}
			}
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
		node := c.BestRetrieveNode(block.GetStoreNode())
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

// RemoveFile remove file
func (c *ClientManager) RemoveFile(target string, recursive bool, isPath bool, sno uint32) error {
	log := c.Log
	req := &mpb.RemoveReq{
		Version:   common.Version,
		NodeId:    c.NodeId,
		Timestamp: common.Now(),
		Recursive: recursive,
	}
	if isPath {
		req.Target = &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{target}, SpaceNo: sno}
	} else {
		id, err := hex.DecodeString(target)
		if err != nil {
			return err
		}
		log.Infof("delete file by id %s, binary id %s", target, id)
		req.Target = &mpb.FilePath{OneOfPath: &mpb.FilePath_Id{id}, SpaceNo: sno}
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

// MoveFile move file
func (c *ClientManager) MoveFile(source, dest string, sno uint32) error {
	log := c.Log
	req := &mpb.MoveReq{
		Version:   common.Version,
		NodeId:    c.NodeId,
		Timestamp: common.Now(),
	}
	id, err := hex.DecodeString(source)
	if err != nil {
		return err
	}
	log.Infof("rename file by id %s, binary id %s", source, id)
	req.Source = &mpb.FilePath{OneOfPath: &mpb.FilePath_Id{id}, SpaceNo: sno}
	req.Dest = dest

	err = req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}

	log.Infof("move file from %+v, to %v", source, dest)
	rsp, err := c.mclient.Move(context.Background(), req)
	if err != nil {
		return common.StatusErrFromError(err)
	}
	log.Infof("move file rsp code :%d msg: %s", rsp.GetCode(), rsp.GetErrMsg())
	if rsp.GetCode() != 0 {
		return fmt.Errorf("%s", rsp.GetErrMsg())
	}
	return nil

}

// GetProgress returns progress rate
func (c *ClientManager) GetProgress(files []string) (map[string]float64, error) {
	return c.PM.GetProgress(files)
}
