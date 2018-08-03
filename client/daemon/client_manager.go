package daemon

import (
	"context"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	collectClient "github.com/samoslab/nebula/client/collector_client"
	"github.com/samoslab/nebula/client/progress"
	"github.com/samoslab/nebula/util/filetype"

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
	"google.golang.org/grpc/status"
)

const (
	// InfoForEncrypt info for encrypt
	InfoForEncrypt = "this is a interesting info for encrypt msg"
	// SysFile store password hash file
	SysFile = ".nebula"

	// ClientDBName client db name
	ClientDBName = "data.db"
)

var (
	// ReplicaFileSize using replication if file size less than
	ReplicaFileSize = int64(8 * 1024)

	// PartitionMaxSize  max size of one partition
	PartitionMaxSize = int64(256 * 1024 * 1024)

	// DefaultTempDir default temporary file
	DefaultTempDir = "/tmp/nebula_client"

	// ReplicaNum number of muliti-replication
	ReplicaNum    = 5
	MinReplicaNum = 3
)

// DownFile list files format, used when download file
type DownFile struct {
	ID        string `json:"id"`
	FileSize  uint64 `json:"filesize"`
	FileName  string `json:"filename"`
	FileHash  string `json:"filehash"`
	Folder    bool   `json:"folder"`
	FileType  string `json:"filetype"`
	ModTime   uint64 `json:"modtime"`
	Extension string `json:"extension"`
}

// FilePages list file
type FilePages struct {
	Total uint32      `json:"total"`
	Files []*DownFile `json:"files"`
}

// ClientManager client manager
type ClientManager struct {
	NodeId        []byte
	store         *store
	TempDir       string
	PubkeyHash    []byte
	Root          string
	MsgCount      uint32
	MsgChan       chan string
	done          chan struct{}
	quit          chan struct{}
	TaskChan      chan TaskInfo
	SpaceM        *SpaceManager
	webcfg        config.Config
	TrackerPubkey *rsa.PublicKey
	serverConn    *grpc.ClientConn
	Log           logrus.FieldLogger
	OM            *order.OrderManager
	FileTypeMap   filetype.SupportType
	cfg           *config.ClientConfig
	PM            *progress.ProgressManager
	mclient       mpb.MatadataServiceClient
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
	//conn, err := grpc.Dial(webcfg.TrackerServer, grpc.WithBlock(), grpc.WithInsecure(), grpc.WithKeepaliveParams(keepalive.ClientParameters{
	//	Time:                50 * time.Millisecond,
	//	Timeout:             100 * time.Millisecond,
	//	PermitWithoutStream: true,
	//}))
	conn, err := grpc.Dial(webcfg.TrackerServer, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Errorf("Rpc dial failed: %s", err.Error())
		return nil, err
	}
	log.Infof("Tracker server %s", webcfg.TrackerServer)

	rsaPubkey, pubkeyHash, err := register.GetPublicKey(webcfg.TrackerServer)
	if err != nil {
		return nil, err
	}

	om := order.NewOrderManager(webcfg.TrackerServer, log, cfg.Node.PriKey, cfg.Node.NodeId)

	spaceM := NewSpaceManager()
	for _, sp := range cfg.Space {
		log.Infof("Space %d name %s home %s", sp.SpaceNo, sp.Name, sp.Home)
		spaceM.AddSpace(sp.SpaceNo, sp.Password, sp.Home)
	}

	dbPath := filepath.Join(webcfg.ConfigDir, ClientDBName)
	db, err := bolt.Open(dbPath, 0700, &bolt.Options{
		Timeout: 1 * time.Second,
	})
	if err != nil {
		log.WithError(err).Error("Open db failed")
		return nil, err
	}
	store, err := newStore(db, log)
	if err != nil {
		log.WithError(err).Error("New store failed")
		return nil, err
	}

	c := &ClientManager{
		OM:            om,
		Log:           log,
		cfg:           cfg,
		serverConn:    conn,
		store:         store,
		SpaceM:        spaceM,
		webcfg:        webcfg,
		TrackerPubkey: rsaPubkey,
		PubkeyHash:    pubkeyHash,
		TempDir:       os.TempDir(),
		NodeId:        cfg.Node.NodeId,
		done:          make(chan struct{}),
		quit:          make(chan struct{}),
		FileTypeMap:   filetype.SupportTypes(),
		PM:            progress.NewProgressManager(),
		mclient:       mpb.NewMatadataServiceClient(conn),
		MsgChan:       make(chan string, common.MsgQueueLen),
		TaskChan:      make(chan TaskInfo, common.TaskQuqueLen),
	}

	collectClient.NodePtr = cfg.Node

	log.Infof("Temp dir is %s", c.TempDir)
	if _, err := os.Stat(c.TempDir); os.IsNotExist(err) {
		//create the dir.
		if err := os.MkdirAll(c.TempDir, 0744); err != nil {
			panic(err)
		}
	}

	collectClient.Start(webcfg.CollectServer)

	go c.ExecuteTask()
	go c.SendProgressMsg()

	return c, nil
}

func (c *ClientManager) GetMsgCount() uint32 {
	return c.MsgCount
}

func (c *ClientManager) DecreaseMsgCount() {
	c.MsgCount--
}

func (c *ClientManager) GetMsgChan() <-chan string {
	return c.MsgChan
}

func (c *ClientManager) AddDoneMsg(msg string) {
	c.MsgChan <- msg
	c.MsgCount++
}

// Shutdown shutdown tracker connection
func (c *ClientManager) Shutdown() {
	c.serverConn.Close()
	collectClient.Stop()
	close(c.quit)
	<-c.done
	close(c.TaskChan)
	close(c.MsgChan)
}

func map2Req(taskInfo TaskInfo) (TaskInfo, error) {
	var req interface{}
	switch taskInfo.Task.Type {
	case common.TaskUploadFileType:
		req = &common.UploadReq{}
	case common.TaskUploadDirType:
		req = &common.UploadDirReq{}
	case common.TaskDownloadFileType:
		req = &common.DownloadReq{}
	case common.TaskDownloadDirType:
		req = &common.DownloadDirReq{}
	default:
		return taskInfo, errors.New("unknown task type")
	}
	data, err := json.Marshal(taskInfo.Task.Payload)
	if err != nil {
		return taskInfo, err
	}
	err = json.Unmarshal(data, req)
	if err != nil {
		return taskInfo, err
	}
	taskInfo.Task.Payload = req
	return taskInfo, nil
}

// ExecuteTask start handle task
func (c *ClientManager) ExecuteTask() error {
	log := c.Log
	log.Info("Start task goroutine, handle unfinished task first")
	// unfinished task
	unhandleTasks, err := c.store.GetTaskArray(func(rwd TaskInfo) bool {
		return rwd.Status == StatusGotTask
	})
	if err != nil {
		log.WithError(err).Error("Get unhandle task failed")
		return err
	}
	log.Infof("Unhandle task number %d", len(unhandleTasks))
	for _, taskInfo := range unhandleTasks {
		log := log.WithField("taskkey", taskInfo.Key)
		taskInfo, err = map2Req(taskInfo)
		if err != nil {
			log.WithError(err).Error("task cannot deserialization")
			continue
		}
		c.TaskChan <- taskInfo
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		TaskControlChan := make(chan struct{}, common.CCTaskHandleNum)
		defer func() {
			wg.Done()
			close(TaskControlChan)
			close(c.done)
			log.Infof("Shutdown task goroutine")
		}()
		for {
			select {
			case <-c.quit:
				return
			case taskInfo := <-c.TaskChan:
				TaskControlChan <- struct{}{}
				go func(taskInfo TaskInfo) {
					done := make(chan struct{})
					go func() {
						select {
						case <-c.quit:
							<-TaskControlChan
						case <-done:
							return
						}
					}()
					defer func() {
						close(done)
						<-TaskControlChan
					}()
					var err error
					task := taskInfo.Task
					if taskInfo.Key == "" {
						return
					}
					log := log.WithField("task key", taskInfo.Key)
					log.Infof("Handle task %+v", taskInfo)
					doneMsg := common.MakeSuccDoneMsg(task.Type, "")
					switch task.Type {
					case common.TaskUploadFileType:
						req := task.Payload.(*common.UploadReq)
						err = c.UploadFile(req.Filename, req.Dest, req.Interactive, req.NewVersion, req.IsEncrypt, req.Sno)
						doneMsg.Source = req.Filename
					case common.TaskUploadDirType:
						req := task.Payload.(*common.UploadDirReq)
						err = c.UploadDir(req.Parent, req.Dest, req.Interactive, req.NewVersion, req.IsEncrypt, req.Sno)
						doneMsg.Source = req.Parent
					case common.TaskDownloadFileType:
						req := task.Payload.(*common.DownloadReq)
						err = c.DownloadFile(req.FileName, req.Dest, req.FileHash, req.FileSize, req.Sno)
						doneMsg.Source = req.FileName
					case common.TaskDownloadDirType:
						req := task.Payload.(*common.DownloadDirReq)
						err = c.DownloadDir(req.Parent, req.Dest, req.Sno)
						doneMsg.Source = req.Parent
					default:
						err = errors.New("unknown")
					}
					errStr := ""
					if err != nil {
						errStr = err.Error()
						doneMsg.SetError(1, err)
						log.WithError(err).Error("Execute task failed")
					} else {
						log.Infof("Execute task success")
					}
					c.AddDoneMsg(doneMsg.Serialize())
					_, err = c.store.UpdateTaskInfo(taskInfo.Key, func(rs TaskInfo) TaskInfo {
						rs.Status = StatusDone
						rs.UpdatedAt = common.Now()
						rs.Err = errStr
						return rs
					})
					if err != nil {
						log.WithError(err).Error("Update task failed")
					} else {
						log.Infof("Update task success")
					}
				}(taskInfo)
			}
		}
	}()
	wg.Wait()
	return nil
}

// SendProgressMsg send message for websocket
func (c *ClientManager) SendProgressMsg() error {
	log := c.Log
	log.Info("Start send progress")
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			log.Infof("Shutdown progress goroutine")
		}()
		for {
			select {
			case <-c.quit:
				return
			case <-time.After(1 * time.Second):
				msgList, err := c.PM.GetProgressingMsg([]string{})
				if err != nil {
					log.WithError(err).Error("Get progress message failed")
					continue
				}
				for _, msg := range msgList {
					c.AddDoneMsg(msg)
				}

			}
		}
	}()
	wg.Wait()
	return nil
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
	var err error
	log := c.Log
	password, err = passwordPadding(password, sno)
	if err != nil {
		return err
	}
	data, err := c.GetSpaceSysFileData(sno)
	if err == nil {
		if len(data) != 0 {
			log.Infof("Space %d password has been set", sno)
			if verifyPassword(sno, password, data) {
				log.Infof("Space %d password verified success", sno)
				return c.SpaceM.SetSpacePasswd(sno, password)
			}
			return fmt.Errorf("Password incorrect")
		}
	}

	log.Infof("Get space %d sys file error %v", sno, err)
	err = c.SpaceM.SetSpacePasswd(sno, password)
	if err != nil {
		return err
	}

	encryDir := filepath.Join(c.webcfg.ConfigDir, fmt.Sprintf("space%d", sno))
	if !util_file.Exists(encryDir) {
		if err := os.MkdirAll(encryDir, 0700); err != nil {
			return fmt.Errorf("mkdir space %d nebula folder %s failed:%s", sno, encryDir, err)
		}
	}

	shaData, err := genEncryptKey(sno, password)
	if err != nil {
		return err
	}
	encryFile := filepath.Join(encryDir, SysFile)
	if err = ioutil.WriteFile(encryFile, shaData, 0600); err != nil {
		return err
	}

	return c.UploadFile(encryFile, "/", false, false, false, sno)
}

// VerifyPassword set user privacy space password
func (c *ClientManager) VerifyPassword(sno uint32, password string) error {
	var err error
	password, err = passwordPadding(password, sno)
	if err != nil {
		return err
	}
	data, err := c.GetSpaceSysFileData(sno)
	if err == nil {
		if len(data) != 0 {
			if verifyPassword(sno, password, data) {
				return c.SpaceM.SetSpacePasswd(sno, password)
			}
			return fmt.Errorf("Password incorrect")
		}
	}
	return fmt.Errorf("space %d password not set", sno)
}

// CheckSpaceStatus check space status
func (c *ClientManager) CheckSpaceStatus(sno uint32) error {
	if sno == 0 {
		return nil
	}
	password, err := c.SpaceM.GetSpacePasswd(sno)
	if err != nil {
		return err
	}
	if len(password) > 0 {
		return nil
	}
	data, err := c.GetSpaceSysFileData(sno)
	if err == nil && len(data) >= 0 {
		return nil
	}
	return fmt.Errorf("space %d password not set", sno)
}

// UploadDir upload all files in dir to provider
func (c *ClientManager) UploadDir(parent, dest string, interactive, newVersion, isEncrypt bool, sno uint32) error {
	log := c.Log
	if !filepath.IsAbs(parent) {
		return fmt.Errorf("path %s must absolute", parent)
	}
	dirs, err := GetDirsAndFiles(parent)
	if err != nil {
		return err
	}
	if isEncrypt {
		password, err := c.getSpacePassword(sno)
		if err != nil {
			return err
		}
		if len(password) == 0 {
			return fmt.Errorf("Password not set")
		}
	}
	log.Debugf("Upload dirs %+v", dirs)
	newDirs := dirAdjust(dirs, parent, dest, runtime.GOOS)
	log.Debugf("New upload dirs %+v", newDirs)
	errArr := []error{}
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCUploadFileNum)
	for _, dpair := range newDirs {
		if dpair.Folder {
			log.Debugf("Mkfolder %+v", dpair)
			_, err := c.MkFolder(dpair.Parent, []string{dpair.Name}, interactive, sno)
			if err != nil {
				return err
			}
		} else {
			log.Debugf("Upload file %+v", dpair)
			ccControl.Add()
			go func(dpair DirPair) {
				done := make(chan struct{})
				go HandleQuit(c.quit, done, ccControl)
				defer func() {
					close(done)
					ccControl.Done()
				}()
				doneMsg := common.MakeSuccDoneMsg(common.TaskUploadFileType, dpair.Name)
				err := c.UploadFile(dpair.Name, dpair.Parent, interactive, newVersion, isEncrypt, sno)
				if err != nil {
					doneMsg.SetError(1, err)
				}

				c.AddDoneMsg(doneMsg.Serialize())
				if err != nil {
					mutex.Lock()
					errArr = append(errArr, err)
					mutex.Unlock()
				}
			}(dpair)
		}
	}
	ccControl.Wait()
	if len(errArr) > 0 {
		return errArr[0]
	}
	return nil
}

func (c *ClientManager) getSpacePassword(sno uint32) ([]byte, error) {
	log := c.Log
	if sno == 0 {
		return []byte(aes.RandStr(16)), nil
	}
	password, err := c.SpaceM.GetSpacePasswd(sno)
	if err != nil {
		log.Errorf("Get password of space no %d error %v", sno, err)
		return nil, err
	}
	if len(password) == 0 {
		log.Errorf("Please set space %d password first", sno)
		return nil, fmt.Errorf("please set space %d password first", sno)
	}
	return password, nil
}

// AddTask add a task into db and queue
func (c *ClientManager) AddTask(tp string, req interface{}) (string, error) {
	log := c.Log.WithField("task", "add")
	task := NewTask(tp, req)
	taskInfo, err := c.store.StoreTask(task)
	if err != nil {
		log.WithError(err).Error("store task failed")
		return "", err
	}
	c.TaskChan <- taskInfo
	return taskInfo.Key, nil
}

// UploadFile upload file to provider
func (c *ClientManager) UploadFile(fileName, dest string, interactive, newVersion, isEncrypt bool, sno uint32) error {
	var err error
	var password, encryptKey []byte
	log := c.Log.WithField("upload file", fileName)
	if isEncrypt {
		password, err = c.getSpacePassword(sno)
		if err != nil {
			log.WithError(err).Info("Get space password")
			return err
		}
		fmt.Printf("password %s\n", string(password))
		if len(password) == 0 {
			log.Info("Space password not set")
			return fmt.Errorf("Password not set")
		}
		encryptKey, err = rsalong.EncryptLong(c.TrackerPubkey, password, 256)
		if err != nil {
			log.WithError(err).Info("Encrypt password")
			return err
		}
	}
	req, rsp, err := c.CheckFileExists(fileName, dest, interactive, newVersion, password, encryptKey, sno)
	if err != nil {
		return common.StatusErrFromError(err)
	}

	log.Infof("Check file exists resp code %d", rsp.GetCode())
	if rsp.GetCode() == 0 {
		log.Infof("Upload %s success", fileName)
		return nil
	}
	// 1 can upload
	if rsp.GetCode() != 1 {
		return fmt.Errorf("%d:%s", rsp.GetCode(), rsp.GetErrMsg())
	}

	switch rsp.GetStoreType() {
	case mpb.FileStoreType_MultiReplica:
		log.Infof("Upload manner is multi-replication")
		// encrypt file
		originFileName := fileName
		if isEncrypt {
			_, onlyFileName := filepath.Split(fileName)
			// change fileName to encypted file avoid origin file modified
			fileName = filepath.Join(c.TempDir, onlyFileName)
			err := aes.EncryptFile(originFileName, password, fileName)
			if err != nil {
				log.Errorf("Encrypt error %v", err)
				return err
			}
			defer func() {
				deleteTemporaryFile(log, fileName)
			}()
		}
		partitions, err := c.uploadFileByMultiReplica(originFileName, fileName, req, rsp)
		if err != nil {
			return err
		}
		return c.UploadFileDone(req, partitions, encryptKey)
	case mpb.FileStoreType_ErasureCode:
		log.Infof("Upload manner is erasure")
		partFiles := []string{}
		var err error
		fileSize := int64(req.GetFileSize())
		if fileSize > PartitionMaxSize {
			chunkSize, chunkNum := GetChunkSizeAndNum(fileSize, PartitionMaxSize)
			partFiles, err = FileSplit(c.TempDir, fileName, fileSize, chunkSize, int64(chunkNum))
			if err != nil {
				return err
			}
		} else {
			partFiles = append(partFiles, fileName)
		}

		log.Infof("File %s need split to %d partitions", req.GetFileName(), len(partFiles))

		dataShards := int(rsp.GetDataPieceCount())
		verifyShards := int(rsp.GetVerifyPieceCount())

		log.Infof("Prepare response gave %d dataShards, %d verifyShards", dataShards, verifyShards)

		fileInfos := []common.PartitionFile{}

		realSizeAfterRS := int64(0)
		for _, fname := range partFiles {
			fileSlices, err := c.onlyFileSplit(fname, dataShards, verifyShards, isEncrypt, password, sno)
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
			log.Infof("File %s need split to %d blocks", fname, len(fileSlices))
			for _, fs := range fileSlices {
				log.Debugf("Erasure block files %s index %d", fs.FileName, fs.SliceIndex)
				c.PM.SetPartitionMap(fs.FileName, fileName)
				realSizeAfterRS += fs.FileSize
			}
			c.PM.SetPartitionMap(fname, fileName)
		}

		c.PM.SetProgress(common.TaskUploadProgressType, fileName, 0, uint64(realSizeAfterRS))

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

		ufpr, err := c.createUploadPrepareRequest(req, len(partFiles), fileInfos)
		if err != nil {
			return err
		}

		ctx := context.Background()
		log.Info("Send prepare reques")
		ufprsp, err := c.mclient.UploadFilePrepare(ctx, ufpr)
		if err != nil {
			log.Errorf("UploadFilePrepare error %v", err)
			return common.StatusErrFromError(err)
		}

		rspPartitions := ufprsp.GetPartition()
		log.Infof("Upload prepare response partitions count:%d", len(rspPartitions))

		if len(rspPartitions) == 0 {
			return fmt.Errorf("only 0 partitions, not correct")
		}

		for i, part := range rspPartitions {
			auth := part.GetProviderAuth()
			for _, pa := range auth {
				log.Debugf("Partition %d, %s:%d %v hashauth %d", i, pa.Server, pa.Port, pa.Spare, len(pa.HashAuth))
			}
		}

		partitions := []*mpb.StorePartition{}
		for i, partInfo := range fileInfos {
			partition, err := c.uploadFileBatchByErasure(ufpr, rspPartitions[i], partInfo, dataShards)
			if err != nil {
				return err
			}
			log.Debugf("Partition %d has %d store blocks", i, len(partition.GetBlock()))
			partitions = append(partitions, partition)
		}
		log.Infof("There are %d store partitions", len(partitions))

		return c.UploadFileDone(req, partitions, encryptKey)

	}
	return nil
}

func (c *ClientManager) createUploadPrepareRequest(req *mpb.CheckFileExistReq, partFileCount int, fileInfos []common.PartitionFile) (*mpb.UploadFilePrepareReq, error) {
	log := c.Log
	ufpr := &mpb.UploadFilePrepareReq{
		Version:   common.Version,
		NodeId:    req.NodeId,
		FileHash:  req.FileHash,
		FileSize:  req.FileSize,
		Timestamp: common.Now(),
		Partition: make([]*mpb.SplitPartition, partFileCount),
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
	if err := ufpr.SignReq(c.cfg.Node.PriKey); err != nil {
		return nil, err
	}

	return ufpr, nil
}

func deleteTemporaryFile(log logrus.FieldLogger, fileName string) {
	log.Debugf("delete file %s", fileName)
	if err := os.Remove(fileName); err != nil {
		log.Errorf("delete %s failed, error %v", fileName, err)
	}
}

func (c *ClientManager) CheckFileExists(fileName, dest string, interactive, newVersion bool, password, encryptKey []byte, sno uint32) (*mpb.CheckFileExistReq, *mpb.CheckFileExistResp, error) {
	log := c.Log.WithField("filename", fileName)
	hash, err := util_hash.Sha1File(fileName)
	if err != nil {
		return nil, nil, err
	}
	fileSize, err := GetFileSize(fileName)
	if err != nil {
		return nil, nil, err
	}
	fileType := filetype.FileType(fileName)
	_, fname := filepath.Split(fileName)
	ctx := context.Background()
	req := &mpb.CheckFileExistReq{
		FileHash:      hash,
		FileName:      fname,
		NodeId:        c.NodeId,
		EncryptKey:    encryptKey,
		NewVersion:    newVersion,
		Interactive:   interactive,
		PublicKeyHash: c.PubkeyHash,
		Timestamp:     common.Now(),
		FileType:      fileType.Value,
		Version:       common.Version,
		FileSize:      uint64(fileSize),
		Parent:        &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{dest}, SpaceNo: sno},
	}
	mtime, err := GetFileModTime(fileName)
	if err != nil {
		return nil, nil, err
	}
	req.FileModTime = uint64(mtime)
	if fileSize < ReplicaFileSize {
		fileData, err := util_hash.GetFileData(fileName)
		if err != nil {
			log.Errorf("Read file data error %v", err)
			return nil, nil, err
		}
		if len(password) != 0 {
			fileData, err = aes.Encrypt(fileData, password)
			if err != nil {
				log.Errorf("Encrypt file error %v", err)
				return nil, nil, err
			}
		}
		req.FileData = fileData
		c.PM.SetProgress(common.TaskUploadProgressType, fileName, 0, uint64(len(req.FileData)))
		c.PM.SetPartitionMap(fileName, fileName)
		c.PM.SetIncrement(fileName, uint64(len(req.FileData)))
		fmt.Printf("Origin filesize %d, encrypted size %d\n", req.FileSize, len(req.FileData))
	}
	err = req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return nil, nil, err
	}
	log.Info("Check file exist request")
	rsp, err := c.mclient.CheckFileExist(ctx, req)
	if err != nil {
		// tracker public key expired
		st, ok := status.FromError(err)
		if !ok {
			return nil, nil, fmt.Errorf("get status error failed")
		}
		if st.Code() == 500 || st.Message() == "tracker public key expired" {
			rsaPubkey, pubkeyHash, err := register.GetPublicKey(c.webcfg.TrackerServer)
			if err != nil {
				return nil, nil, err
			}
			c.PubkeyHash = pubkeyHash
			c.TrackerPubkey = rsaPubkey
			req.Timestamp = common.Now()
			req.PublicKeyHash = pubkeyHash
			err = req.SignReq(c.cfg.Node.PriKey)
			if err != nil {
				return nil, nil, err
			}
			log.Info("Check file exist request")
			rsp, err := c.mclient.CheckFileExist(ctx, req)
			return req, rsp, err
		}
	}
	return req, rsp, err
}

// MkFolder create folder
func (c *ClientManager) MkFolder(filepath string, folders []string, interactive bool, sno uint32) (bool, error) {
	log := c.Log.WithField("folder parent", filepath)
	ctx := context.Background()
	req := &mpb.MkFolderReq{
		Folder:      folders,
		NodeId:      c.NodeId,
		Interactive: interactive,
		Timestamp:   common.Now(),
		Version:     common.Version,
		Parent:      &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{filepath}, SpaceNo: sno},
	}
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return false, err
	}
	log.Infof("Make folder %+v", req.GetFolder())
	rsp, err := c.mclient.MkFolder(ctx, req)
	if err != nil {
		return false, common.StatusErrFromError(err)
	}
	if rsp.GetCode() != 0 {
		if strings.Contains(rsp.GetErrMsg(), "System error: pq: duplicate key value") {
			log.Warning("Folder exists %s", rsp.GetErrMsg())
			return true, nil
		}
		return false, fmt.Errorf("%s", rsp.GetErrMsg())
	}
	log.Infof("Make folder response code %d", rsp.GetCode())
	return true, nil
}

func (c *ClientManager) onlyFileSplit(fileName string, dataNum, verifyNum int, isEncrypt bool, password []byte, sno uint32) ([]common.HashFile, error) {
	log := c.Log.WithField("filesplit", fileName)
	fileSlices, err := RsEncoder(c.Log, c.TempDir, fileName, dataNum, verifyNum)
	if err != nil {
		log.Errorf("Reedsolomon encoder error %v", err)
		return nil, err
	}
	if isEncrypt {
		for i := range fileSlices {
			err := aes.EncryptFile(fileSlices[i].FileName, password, fileSlices[i].FileName)
			if err != nil {
				log.Errorf("Encrypt error %v", err)
				return nil, err
			}
			hash, err := util_hash.Sha1File(fileSlices[i].FileName)
			if err != nil {
				return nil, err
			}
			fileSlices[i].FileHash = hash
			fileSize, err := GetFileSize(fileSlices[i].FileName)
			if err != nil {
				return nil, err
			}
			fileSlices[i].FileSize = fileSize
		}
	}

	return fileSlices, nil

}

func (c *ClientManager) uploadFileBatchByErasure(req *mpb.UploadFilePrepareReq, rspPartition *mpb.ErasureCodePartition, partFile common.PartitionFile, dataShards int) (*mpb.StorePartition, error) {
	log := c.Log
	partition := &mpb.StorePartition{}
	partition.Block = []*mpb.StoreBlock{}
	phas := rspPartition.GetProviderAuth()
	providers, err := UsingBestProvider(phas)
	if err != nil {
		return nil, err
	}

	var errResult []error
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCUploadGoNum)
	for i, pro := range providers {
		checksum := i >= dataShards
		uploadParas := &common.UploadParameter{
			Checksum:       checksum,
			HF:             partFile.Pieces[i],
			OriginFileHash: partFile.OriginFileHash,
			OriginFileSize: partFile.OriginFileSize,
		}
		ccControl.Add()
		go func(pro *mpb.BlockProviderAuth, tm uint64, uploadPara *common.UploadParameter) {
			done := make(chan struct{})
			go HandleQuit(c.quit, done, ccControl)
			defer func() {
				close(done)
				ccControl.Done()
			}()
			block, err := c.uploadFileToErasureProvider(pro, tm, uploadParas)
			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				log.Errorf("Upload file %s error %v", uploadPara.HF.FileName, err)
				errResult = append(errResult, err)
				return
			}
			partition.Block = append(partition.Block, block)
		}(pro, rspPartition.GetTimestamp(), uploadParas)
	}
	ccControl.Wait()
	if len(errResult) != 0 {
		return partition, errResult[0]
	}
	return partition, nil
}

func (c *ClientManager) uploadFileToErasureProvider(pro *mpb.BlockProviderAuth, tm uint64, uploadPara *common.UploadParameter) (*mpb.StoreBlock, error) {
	log := c.Log
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	uploadPara.Provider = server
	conn, err := grpc.Dial(server, grpc.WithInsecure())
	if err != nil {
		log.Errorf("Rpc dial failed: %s", err.Error())
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
		StoreNodeId: [][]byte{},
		Checksum:    uploadPara.Checksum,
		Hash:        uploadPara.HF.FileHash,
		Size:        uint64(uploadPara.HF.FileSize),
		BlockSeq:    uint32(uploadPara.HF.SliceIndex),
	}
	block.StoreNodeId = append(block.StoreNodeId, []byte(pro.GetNodeId()))
	log.Debugf("Upload %s to provider %s success", uploadPara.HF.FileName, server)

	return block, nil
}

func (c *ClientManager) uploadFileToReplicaProvider(conn *grpc.ClientConn, pro *mpb.ReplicaProvider, uploadPara *common.UploadParameter) ([]byte, error) {
	fileInfo := uploadPara.HF
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	log := c.Log.WithField("uploading", fileInfo.FileName).WithField("provider", server)
	uploadPara.Provider = server
	pclient := pb.NewProviderServiceClient(conn)
	log.Debugf("Upload file hash %x size %d", fileInfo.FileHash, fileInfo.FileSize)

	err := client.StorePiece(log, pclient, uploadPara, pro.Auth, pro.Ticket, pro.Timestamp, c.PM)
	if err != nil {
		log.WithError(err).Error("Upload error")
		return nil, err
	}

	log.Info("Upload file success")

	return pro.GetNodeId(), nil
}

func (c *ClientManager) uploadFileByMultiReplica(originFileName, fileName string, req *mpb.CheckFileExistReq, rsp *mpb.CheckFileExistResp) ([]*mpb.StorePartition, error) {
	log := c.Log
	hash, err := util_hash.Sha1File(fileName)
	if err != nil {
		return nil, err
	}
	fileSize, err := GetFileSize(fileName)
	if err != nil {
		return nil, err
	}
	fileSlices := []common.HashFile{
		common.HashFile{
			SliceIndex: 0,
			FileHash:   hash,
			FileSize:   fileSize,
			FileName:   fileName,
		},
	}
	fileInfos := []common.PartitionFile{
		common.PartitionFile{FileName: fileName,
			Pieces:         fileSlices,
			OriginFileName: req.FileName,
			OriginFileHash: req.FileHash,
			OriginFileSize: req.FileSize,
		}}

	ufpr, err := c.createUploadPrepareRequest(req, 1, fileInfos)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	log = log.WithField("filename", req.GetFileName())
	log.Infof("Send prepare request")
	ufprsp, err := c.mclient.UploadFilePrepare(ctx, ufpr)
	if err != nil {
		log.Errorf("UploadFilePrepare error %v", err)
		return nil, common.StatusErrFromError(err)
	}
	///

	uploadPara := &common.UploadParameter{
		OriginFileHash: req.FileHash,
		OriginFileSize: req.FileSize,
		HF:             fileSlices[0],
	}

	block := &mpb.StoreBlock{
		Checksum:    false,
		StoreNodeId: [][]byte{},
		Hash:        fileSlices[0].FileHash,
		Size:        uint64(fileSlices[0].FileSize),
		BlockSeq:    uint32(fileSlices[0].SliceIndex),
	}

	c.PM.SetPartitionMap(fileName, originFileName)

	providers, err := GetBestReplicaProvider(ufprsp.GetProvider(), ReplicaNum)
	if err != nil {
		return nil, err
	}
	if fileName == originFileName {
		c.PM.SetProgress(common.TaskUploadProgressType, originFileName, 0, uint64(len(providers))*req.FileSize)
	} else {
		c.PM.SetProgress(common.TaskUploadProgressType, originFileName, 0, uint64(int64(len(providers))*fileSize))
	}

	errArr := []error{}
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCUploadFileNum)
	for _, pro := range providers {
		ccControl.Add()
		go func(pro *mpb.ReplicaProvider) {
			server := fmt.Sprintf("%s:%d", pro.Server, pro.Port)
			conn, err := grpc.Dial(server, grpc.WithInsecure())
			if err != nil {
				log.Errorf("Rpc dail failed: %v", err)
				mutex.Lock()
				errArr = append(errArr, err)
				mutex.Unlock()
				return
			}
			done := make(chan struct{})
			go HandleQuit(c.quit, done, ccControl, func() { conn.Close() })
			defer func() {
				conn.Close()
				close(done)
				ccControl.Done()
			}()
			proID, err := c.uploadFileToReplicaProvider(conn, pro, uploadPara)
			if err != nil {
				mutex.Lock()
				errArr = append(errArr, err)
				mutex.Unlock()
			}
			mutex.Lock()
			block.StoreNodeId = append(block.StoreNodeId, proID)
			mutex.Unlock()
		}(pro)
	}

	ccControl.Wait()
	if len(errArr) > 0 {
		return nil, errArr[0]
	}

	partition := &mpb.StorePartition{}
	partition.Block = append(partition.Block, block)
	partitions := []*mpb.StorePartition{partition}
	return partitions, nil
}

func (c *ClientManager) UploadFileDone(reqCheck *mpb.CheckFileExistReq, partitions []*mpb.StorePartition, encryptKey []byte) error {
	req := &mpb.UploadFileDoneReq{
		NodeId:        c.NodeId,
		Partition:     partitions,
		EncryptKey:    encryptKey,
		Timestamp:     common.Now(),
		PublicKeyHash: c.PubkeyHash,
		Version:       common.Version,
		FileHash:      reqCheck.GetFileHash(),
		FileSize:      reqCheck.GetFileSize(),
		FileName:      reqCheck.GetFileName(),
		FileType:      reqCheck.GetFileType(),
		FileModTime:   reqCheck.GetFileModTime(),
		Parent:        reqCheck.GetParent(),
		Interactive:   reqCheck.GetInteractive(),
		NewVersion:    reqCheck.GetNewVersion(),
	}
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}
	ctx := context.Background()
	log := c.Log.WithField("filename", req.GetFileName())
	log.Info("Upload file done request")
	ufdrsp, err := c.mclient.UploadFileDone(ctx, req)
	if err != nil {
		// tracker public key expired
		st, ok := status.FromError(err)
		if !ok {
			return err
		}
		if st.Code() == 500 || st.Message() == "tracker public key expired" {
			rsaPubkey, pubkeyHash, err := register.GetPublicKey(c.webcfg.TrackerServer)
			if err != nil {
				return err
			}
			c.PubkeyHash = pubkeyHash
			c.TrackerPubkey = rsaPubkey
			req.Timestamp = common.Now()
			req.PublicKeyHash = pubkeyHash
			err = req.SignReq(c.cfg.Node.PriKey)
			if err != nil {
				return err
			}
			log.Info("Check file exist request")
			ufdrsp, err = c.mclient.UploadFileDone(ctx, req)
			if err != nil {
				return common.StatusErrFromError(err)
			}
		}
	}
	log.Infof("Upload done code %d", ufdrsp.GetCode())
	if ufdrsp.GetCode() != 0 {
		return fmt.Errorf("%s", ufdrsp.GetErrMsg())
	}

	return nil
}

// ListFiles list files on dir
func (c *ClientManager) ListFiles(path string, pageSize, pageNum uint32, sortType string, ascOrder bool, sno uint32) (*FilePages, error) {
	log := c.Log.WithField("list path", path)
	log.Infof("Parameter size %d, num %d, sortype %s, asc %v", pageSize, pageNum, sortType, ascOrder)
	req := &mpb.ListFilesReq{
		PageNum:   pageNum,
		NodeId:    c.NodeId,
		PageSize:  pageSize,
		Timestamp: common.Now(),
		Version:   common.Version,
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
	req.AscOrder = ascOrder
	req.Parent = &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{path}, SpaceNo: sno}
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	log.Info("List path request")
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
		storedFileType := info.GetFileType()
		fileType, extension := c.FileTypeMap.GetTypeAndExtension(storedFileType)
		df := &DownFile{
			ID:        id,
			FileHash:  hash,
			FileType:  fileType,
			Extension: extension,
			FileName:  info.GetName(),
			Folder:    info.GetFolder(),
			ModTime:   info.GetModTime(),
			FileSize:  info.GetFileSize(),
		}
		fileLists = append(fileLists, df)
	}

	return &FilePages{
		Files: fileLists,
		Total: rsp.GetTotalRecord(),
	}, nil
}

// DownloadDir download dir
func (c *ClientManager) DownloadDir(path, destDir string, sno uint32) error {
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path %s must absolute", path)
	}
	if !util_file.Exists(destDir) {
		return fmt.Errorf("dir %s not exists", destDir)
	}
	downFolders := strings.Split(path, "/")
	// make sure path is not "/"
	if len(downFolders) > 1 {
		downloadFolder := downFolders[len(downFolders)-1]
		destDir = filepath.Join(destDir, downloadFolder)
		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			os.Mkdir(destDir, 0744)
		}
	}
	return c.startDownloadDir(path, destDir, sno)
}

func (c *ClientManager) startDownloadDir(path, destDir string, sno uint32) error {
	log := c.Log.WithField("download dir", path)
	errResult := []error{}
	page := uint32(1)
	for {
		// list 1 page 100 items order by name
		downFiles, err := c.ListFiles(path, 100, page, "name", true, sno)
		if err != nil {
			return err
		}
		if len(downFiles.Files) == 0 {
			break
		}
		log.Infof("Page %d has %d files", page, len(downFiles.Files))
		// next page
		page++
		for _, fileInfo := range downFiles.Files {
			currentFile := strings.Join([]string{path, fileInfo.FileName}, "/")
			destFile := filepath.Join(destDir, fileInfo.FileName)
			if fileInfo.Folder {
				log.Infof("Create folder %s", currentFile)
				if _, err := os.Stat(destFile); os.IsNotExist(err) {
					os.Mkdir(destFile, 0744)
				}
				err = c.startDownloadDir(currentFile, destFile, sno)
				if err != nil {
					log.Errorf("Recursive download %s failed %v", currentFile, err)
					return err
				}
			} else {
				log.Infof("Start download %s", currentFile)
				if fileInfo.FileSize == 0 {
					log.Infof("Only create %s because file size is 0", fileInfo.FileName)
					SaveFile(destFile, []byte{})
				} else {
					doneMsg := common.MakeSuccDoneMsg(common.TaskDownloadFileType, currentFile)
					err = c.DownloadFile(currentFile, destDir, fileInfo.FileHash, fileInfo.FileSize, sno)
					if err != nil {
						doneMsg.SetError(1, err)
					}

					c.AddDoneMsg(doneMsg.Serialize())
					if err != nil {
						log.Errorf("Download file %s error %v", currentFile, err)
						errResult = append(errResult, fmt.Errorf("%s %v", currentFile, common.StatusErrFromError(err)))
						//return err
					}
				}
			}
		}
	}
	if len(errResult) > 0 {
		for _, err := range errResult {
			log.Errorf("Download dir error: %v", err)
		}
		return errResult[0]
	}
	return nil
}

// DownloadFile download file
func (c *ClientManager) DownloadFile(downFileName, destDir, filehash string, fileSize uint64, sno uint32) error {
	log := c.Log.WithField("download file", downFileName)
	defer func() {
		if r := recover(); r != nil {
			log.Error("!!!!!get panic info, recover it %s", r)
		}
	}()
	fileHash, err := hex.DecodeString(filehash)
	if err != nil {
		return err
	}
	req := &mpb.RetrieveFileReq{
		SpaceNo:   sno,
		NodeId:    c.NodeId,
		FileHash:  fileHash,
		FileSize:  fileSize,
		Timestamp: common.Now(),
		Version:   common.Version,
	}
	err = req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}
	_, fileName := filepath.Split(downFileName)
	downFileName = filepath.Join(destDir, fileName)
	c.PM.SetProgress(common.TaskDownloadProgressType, downFileName, 0, req.FileSize)

	ctx := context.Background()
	log.Infof("Download request file hash %x, size %d", fileHash, fileSize)
	rsp, err := c.mclient.RetrieveFile(ctx, req)
	if err != nil {
		return common.StatusErrFromError(err)
	}
	if rsp.GetCode() != 0 {
		return fmt.Errorf("%s", rsp.GetErrMsg())
	}

	password := []byte{}
	encryptKey := rsp.GetEncryptKey()
	if len(encryptKey) > 0 {
		password, err = rsalong.DecryptLong(c.cfg.Node.PriKey, encryptKey, 256)
		if err != nil {
			return err
		}
	}
	// tiny file
	if filedata := rsp.GetFileData(); filedata != nil {
		if len(password) != 0 {
			filedata, err = aes.Decrypt(filedata, password)
			if err != nil {
				log.Errorf("Decrypted error %v", err)
				return err
			}
		}
		SaveFile(downFileName, filedata)
		c.PM.SetProgress(common.TaskDownloadProgressType, downFileName, req.FileSize, req.FileSize)
		log.Info("Download tiny file")
		return nil
	}

	partitions := rsp.GetPartition()
	log.Infof("There is %d partitions", len(partitions))
	partitionCount := len(partitions)
	if partitionCount == 1 {
		blockCount := len(partitions[0].GetBlock())
		if blockCount == 1 { // 1 partition 1 block is multiReplica
			log.Info("File store as multi-replication")
			// for progress stats
			for _, block := range partitions[0].GetBlock() {
				c.PM.SetPartitionMap(hex.EncodeToString(block.GetHash()), downFileName)
			}
			_, _, _, _, err := c.saveFileByPartition(downFileName, partitions[0], rsp.GetTimestamp(), req.FileHash, req.FileSize, true)
			if err != nil {
				return err
			}
			if len(password) != 0 {
				fmt.Printf("password:%s\n", string(password))
				if err := aes.DecryptFile(downFileName, password, downFileName); err != nil {
					log.Errorf("Maybe")
				}
			}
			return nil
		}
	}

	// erasure files handle by below codes

	log.Info("This is erasure file")
	// for progress stats
	realSizeAfterRS := uint64(0)
	for i, partition := range partitions {
		for j, block := range partition.GetBlock() {
			c.PM.SetPartitionMap(hex.EncodeToString(block.GetHash()), downFileName)
			realSizeAfterRS += block.GetSize()
			log.Infof("Partition %d block %d hash %x size %d checksum %v seq %d", i, j, block.Hash, block.Size, block.Checksum, block.BlockSeq)
		}
	}
	c.PM.SetProgress(common.TaskDownloadProgressType, downFileName, 0, realSizeAfterRS)

	if len(partitions) == 1 {
		partition := partitions[0]
		datas, paritys, failedCount, middleFiles, err := c.saveFileByPartition(downFileName, partition, rsp.GetTimestamp(), req.FileHash, req.FileSize, false)
		if failedCount > paritys {
			log.Error("File cannot be recoved!!!")
			return err
		}
		if err != nil {
			log.Errorf("Save file by partition error %v, but file still can be recoverd", err)
		}

		if len(password) != 0 {
			for _, file := range middleFiles {
				fmt.Printf("password:%s\n", string(password))
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

		log.Infof("DataShards %d, parityShards %d, failedCount %d", datas, paritys, failedCount)

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

		return RenameCrossOS(tempDownFileName, downFileName)
	}
	partFiles := []string{}
	for i, partition := range partitions {
		partFileName := fmt.Sprintf("%s.%s.%d", downFileName, TEMP_NAMESPACE, i)
		datas, paritys, failedCount, middleFiles, err := c.saveFileByPartition(partFileName, partition, rsp.GetTimestamp(), req.FileHash, req.FileSize, false)
		if failedCount > paritys {
			log.Errorf("Middle file %s cannot be recoved!!!", partFileName)
			return err
		}
		if err != nil {
			log.Errorf("Save file by partition error %v, but file still can be recoverd", err)
		}
		if len(password) != 0 {
			for _, file := range middleFiles {
				if err := aes.DecryptFile(file, password, file); err != nil {
					return err
				}
			}
		}
		log.Infof("DataShards %d, parityShards %d, failedCount %d", datas, paritys, failedCount)
		_, onlyFileName := filepath.Split(partFileName)
		tempDownFileName := filepath.Join(c.TempDir, onlyFileName)
		// file real size can be calcauted by filesize and partition number
		partitionFileSize := ReverseCalcuatePartFileSize(int64(req.FileSize), len(partitions), i)
		log.Infof("Partition %d, size %d", i, partitionFileSize)
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

	if err := FileJoin(downFileName, partFiles); err != nil {
		log.Errorf("File %s join failed, part files %+v", downFileName, partFiles)
		return err
	}
	return nil
}

func (c *ClientManager) saveFileByPartition(fileName string, partition *mpb.RetrievePartition, tm uint64, fileHash []byte, fileSize uint64, multiReplica bool) (int, int, int, []string, error) {
	log := c.Log.WithField("filename", fileName)
	log.Infof("There is %d blocks", len(partition.GetBlock()))
	dataShards := 0
	parityShards := 0
	failedCount := 0
	middleFiles := []string{}
	errArray := []string{}
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCDownloadGoNum)

	for _, block := range partition.GetBlock() {
		if block.GetChecksum() {
			parityShards++
		} else {
			dataShards++
		}

		ccControl.Add()
		go func(log logrus.FieldLogger, block *mpb.RetrieveBlock, fileName string) {
			node := BestRetrieveNode(block.GetStoreNode())
			server := fmt.Sprintf("%s:%d", node.GetServer(), node.GetPort())
			conn, err := grpc.Dial(server, grpc.WithInsecure(), grpc.WithTimeout(3*time.Second), grpc.WithBlock())
			if err != nil {
				log.Errorf("Rpc dial %s failed, error %v", server, err)
				mutex.Lock()
				failedCount++
				errArray = append(errArray, err.Error())
				mutex.Unlock()
				return
			}
			done := make(chan struct{})
			go HandleQuit(c.quit, done, ccControl, func() { conn.Close() })
			defer func() {
				close(done)
				ccControl.Done()
			}()
			tempFileName := fileName
			if !multiReplica {
				_, onlyFileName := filepath.Split(fileName)
				tempFileName = filepath.Join(c.TempDir, fmt.Sprintf("%s.%d", onlyFileName, block.GetBlockSeq()))
			}
			log = log.WithField("part file", tempFileName).WithField("provider", server)
			log.Infof("Retrieve Hash %x", block.GetHash())
			pclient := pb.NewProviderServiceClient(conn)

			err = client.Retrieve(log, pclient, tempFileName, node.GetAuth(), node.GetTicket(), tm, fileHash, block.GetHash(), fileSize, block.GetSize(), c.PM, server)
			if err != nil {
				conn.Close()
				log.Error("Retrieve failed")
				mutex.Lock()
				failedCount++
				middleFiles = append(middleFiles, tempFileName)
				errArray = append(errArray, err.Error())
				mutex.Unlock()
				return
			}
			log.Info("Retrieve success")
			mutex.Lock()
			middleFiles = append(middleFiles, tempFileName)
			mutex.Unlock()
			conn.Close()
		}(log, block, fileName)
	}

	ccControl.Wait()

	if len(errArray) > 0 {
		errRtn := fmt.Errorf("%s", strings.Join(errArray, "\n"))
		return dataShards, parityShards, failedCount, middleFiles, errRtn
	}
	return dataShards, parityShards, failedCount, middleFiles, nil
}

// RemoveFile remove file
func (c *ClientManager) RemoveFile(target string, recursive bool, isPath bool, sno uint32) error {
	log := c.Log.WithField("target", target)
	req := &mpb.RemoveReq{
		NodeId:    c.NodeId,
		Recursive: recursive,
		Timestamp: common.Now(),
		Version:   common.Version,
	}
	if isPath {
		req.Target = &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{target}, SpaceNo: sno}
	} else {
		id, err := hex.DecodeString(target)
		if err != nil {
			return err
		}
		log.Infof("Delete file binary id %s", id)
		req.Target = &mpb.FilePath{OneOfPath: &mpb.FilePath_Id{id}, SpaceNo: sno}
	}

	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}

	log.Infof("Remove file request")
	rsp, err := c.mclient.Remove(context.Background(), req)
	if err != nil {
		return common.StatusErrFromError(err)
	}
	log.Infof("Remove file resp code %d msg %s", rsp.GetCode(), rsp.GetErrMsg())
	if rsp.GetCode() != 0 {
		return fmt.Errorf("%s", rsp.GetErrMsg())
	}
	return nil

}

// MoveFile move file
func (c *ClientManager) MoveFile(source, dest string, isPath bool, sno uint32) error {
	log := c.Log.WithField("move source", source)
	req := &mpb.MoveReq{
		Dest:      dest,
		NodeId:    c.NodeId,
		Timestamp: common.Now(),
		Version:   common.Version,
	}
	if isPath {
		req.Source = &mpb.FilePath{OneOfPath: &mpb.FilePath_Path{source}, SpaceNo: sno}
	} else {
		id, err := hex.DecodeString(source)
		if err != nil {
			return err
		}
		log.Infof("move file binary id %s", id)
		req.Source = &mpb.FilePath{OneOfPath: &mpb.FilePath_Id{id}, SpaceNo: sno}
	}

	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return err
	}

	log.Infof("Move file to %s", dest)
	rsp, err := c.mclient.Move(context.Background(), req)
	if err != nil {
		return common.StatusErrFromError(err)
	}
	log.Infof("Move file resp code %d msg %s", rsp.GetCode(), rsp.GetErrMsg())
	if rsp.GetCode() != 0 {
		return fmt.Errorf("%s", rsp.GetErrMsg())
	}
	return nil

}

// GetSpaceSysFileData get space password data
func (c *ClientManager) GetSpaceSysFileData(sno uint32) ([]byte, error) {
	log := c.Log
	req := &mpb.SpaceSysFileReq{
		SpaceNo:   sno,
		NodeId:    c.NodeId,
		Timestamp: common.Now(),
		Version:   common.Version,
	}
	err := req.SignReq(c.cfg.Node.PriKey)
	if err != nil {
		return nil, err
	}
	log.Infof("Get space %d sys file", sno)
	rsp, err := c.mclient.SpaceSysFile(context.Background(), req)
	if err != nil {
		return nil, common.StatusErrFromError(err)
	}
	return rsp.GetData(), nil
}

// GetProgress returns progress rate
func (c *ClientManager) GetProgress(files []string) (map[string]float64, error) {
	return c.PM.GetProgress(files)
}

// ImportConfig import config file
func (c *ClientManager) ImportConfig(fileName, clientConfigFile string) error {
	cfg, err := config.LoadConfig(fileName)
	if err != nil {
		return err
	}
	return config.SaveClientConfig(clientConfigFile, cfg)
}

// ExportConfig export config file
func (c *ClientManager) ExportConfig(fileName string) error {
	return config.SaveClientConfig(fileName, c.cfg)
}

// ExportFile export config file
func (c *ClientManager) ExportFile() string {
	return c.cfg.SelfFileName
}

// TaskStatus get task status
func (c *ClientManager) TaskStatus(taskID string) (string, error) {
	taskInfo, err := c.store.GetTask(taskID)
	if err != nil {
		return "", err
	}
	if taskInfo.Err != "" {
		return "", errors.New(taskInfo.Err)
	}
	return taskInfo.Status.String(), nil
}
