package daemon

import (
	"bytes"
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
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	collectClient "github.com/samoslab/nebula/client/collector_client"
	"github.com/samoslab/nebula/client/errcode"
	"github.com/samoslab/nebula/client/progress"
	tcppb "github.com/samoslab/nebula/tracker/collector/client/pb"
	"github.com/samoslab/nebula/util/filetype"

	"github.com/samoslab/nebula/client/common"
	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/order"
	client "github.com/samoslab/nebula/client/provider_client"
	pb "github.com/samoslab/nebula/provider/pb"
	mpb "github.com/samoslab/nebula/tracker/metadata/pb"
	"github.com/samoslab/nebula/util/aes"
	util_file "github.com/samoslab/nebula/util/file"
	"github.com/samoslab/nebula/util/filecheck"
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
	WRONGSNO     = 999
)

var (
	// ReplicaFileSize using replication if file size less than
	ReplicaFileSize = int64(8 * 1024)

	// PartitionMaxSize  max size of one partition
	PartitionMaxSize = int64(256 * 1024 * 1024)

	// DefaultTempDir default temporary file
	DefaultTempDir = "/tmp/nebula_client"

	// ReplicaNum number of muliti-replication
	ReplicaNum    = 4
	MinReplicaNum = 3

	RetryCount = 3

	ErrNoMetaData = errors.New("no metadata")
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

type MetaData struct {
	paraStr   string
	generator []byte
	pubKey    []byte
	random    []byte
	phi       [][]byte
	err       error
}

type MetaDataMap struct {
	md    map[string]MetaData
	mutex sync.Mutex
}

func (m *MetaDataMap) Add(fileName, paraStr string, generator, pubKey, random []byte, phi [][]byte, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.md[fileName] = MetaData{paraStr: paraStr, generator: generator, pubKey: pubKey, random: random, phi: phi, err: err}
}

func (m *MetaDataMap) Get(fileName string) (paraStr string, generator, pubKey, random []byte, phi [][]byte, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if md, ok := m.md[fileName]; ok {
		return md.paraStr, md.generator, md.pubKey, md.random, md.phi, md.err
	}
	return "", nil, nil, nil, nil, ErrNoMetaData
}

type MetaKey struct {
	FileName  string
	ChunkSize uint32
}

// ClientManager client manager
type ClientManager struct {
	NodeId        []byte
	store         *store
	TempDir       string
	PubkeyHash    []byte
	Root          string
	mutex         sync.Mutex
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
	mdm           *MetaDataMap
	MetaChan      chan MetaKey
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
	conn, err := common.GrpcDial(webcfg.TrackerServer)
	if err != nil {
		log.Errorf("Rpc dial failed: %s", err.Error())
		return nil, err
	}
	log.Infof("Tracker server %s", webcfg.TrackerServer)

	rsaPubkey, pubkeyHash, err := register.GetPublicKey(webcfg.TrackerServer)
	if err != nil {
		return nil, err
	}

	om := order.NewOrderManager(conn, log, cfg.Node.PriKey, cfg.Node.NodeId)

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
		log.WithError(err).Errorf("Open db %s failed", dbPath)
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
		MetaChan:      make(chan MetaKey, common.MetaQuqueLen),
		mdm:           &MetaDataMap{md: map[string]MetaData{}},
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
	go c.GenMetadataInOrder()

	return c, nil
}

func (c *ClientManager) GetMsgCount() uint32 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.MsgCount
}

func (c *ClientManager) DecreaseMsgCount() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.MsgCount--
}

func (c *ClientManager) GetMsgChan() <-chan string {
	return c.MsgChan
}

func (c *ClientManager) AddDoneMsg(msg string) {
	c.MsgChan <- msg
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
		return rwd.Status == StatusGotTask && rwd.Deleted == false
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
					retry := 0
				RETRY:
					doneMsg := common.MakeSuccDoneMsg(task.Type, "", WRONGSNO)
					retry++
					switch task.Type {
					case common.TaskUploadFileType:
						req := task.Payload.(*common.UploadReq)
						err = c.UploadFile(req.Filename, req.Dest, req.Interactive, req.NewVersion, req.IsEncrypt, req.Sno)
						log.Infof("UploadFile result %v", err)
						doneMsg.Key = common.ProgressKey(serverPath(req.Dest, req.Filename), req.Sno)
						doneMsg.SpaceNo = req.Sno
						doneMsg.Local = req.Filename
					case common.TaskUploadDirType:
						req := task.Payload.(*common.UploadDirReq)
						err = c.UploadDir(req.Parent, req.Dest, req.Interactive, req.NewVersion, req.IsEncrypt, req.Sno)
						doneMsg.Key = req.Dest
						doneMsg.SpaceNo = req.Sno
						doneMsg.Local = req.Parent
					case common.TaskDownloadFileType:
						req := task.Payload.(*common.DownloadReq)
						err = c.DownloadFile(req.FileName, req.Dest, req.FileHash, req.FileSize, req.Sno)
						doneMsg.Key = common.ProgressKey(req.FileName, req.Sno)
						doneMsg.SpaceNo = req.Sno
						doneMsg.Local = localPath(req.Dest, req.FileName)
					case common.TaskDownloadDirType:
						req := task.Payload.(*common.DownloadDirReq)
						err = c.DownloadDir(req.Parent, req.Dest, req.Sno)
						doneMsg.Key = req.Parent
						doneMsg.SpaceNo = req.Sno
						doneMsg.Local = req.Dest
					default:
						err = errors.New("unknown")
					}
					errStr := ""
					if err != nil && !strings.Contains(err.Error(), "path is not exists") {
						log.WithError(err).Error("Execute task failed")
						code, errMsg := common.StatusErrFromError(err)
						if code == 300 && retry < RetryCount {
							log.WithError(err).Error("Execute task failed, retrying")
							goto RETRY
						}
						if strings.Contains(errMsg, "context deadline exceeded") && retry < RetryCount-1 {
							log.WithError(err).Error("Execute task failed, retrying")
							goto RETRY
						}
						//if retry < RetryCount {
						//	log.WithError(err).Error("Execute task failed, retrying")
						//	goto RETRY
						//}
						errStr = err.Error()
						doneMsg.SetError(1, err)
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
		pingTicker := time.NewTicker(5 * time.Second)
		for {
			select {
			case <-c.quit:
				return
			case <-pingTicker.C:
				c.SendPingMsg()
			case <-time.After(time.Second):
				msgList, err := c.PM.GetProgressingMsg([]string{})
				if err != nil {
					log.WithError(err).Error("Get progress message failed")
					continue
				}
				for _, msg := range msgList {
					log.Infof("progress message %+v", msg)
					c.AddDoneMsg(msg)
				}
			}
		}
	}()
	wg.Wait()
	return nil
}

// GenMetadataInOrder gen metadata by single goroutine due to crash if runing by multi-goroutine
func (c *ClientManager) GenMetadataInOrder() error {
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
			case mk := <-c.MetaChan:
				t1 := time.Now()
				paraStr, generator, pubKey, random, phi, err := filecheck.GenMetadata(mk.FileName, mk.ChunkSize)
				t2 := time.Now()
				log.Infof("gen %s metadata time elapased %+v", mk.FileName, t2.Sub(t1).Seconds())
				c.mdm.Add(mk.FileName, paraStr, generator, pubKey, random, phi, err)
			}
		}
	}()
	wg.Wait()
	return nil
}

func (c *ClientManager) SendPingMsg() error {
	req := &mpb.PingReq{
		Version: common.Version,
	}

	_, err := c.mclient.Ping(context.Background(), req)
	if err != nil {
		return err
	}

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
	log.Infof("Upload dirs %+v", dirs)
	newDirs := dirAdjust(dirs, parent, dest, runtime.GOOS)
	log.Infof("New upload dirs %+v", newDirs)
	errArr := []error{}
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCUploadFileNum)
	for _, dpair := range newDirs {
		if dpair.Folder {
			log.Infof("Mkfolder %+v", dpair)
			_, err := c.MkFolder(dpair.Parent, []string{dpair.Name}, interactive, sno)
			if err != nil {
				return err
			}
		} else {
			log.Infof("Upload file %+v", dpair)
			ccControl.Add()
			go func(dpair DirPair) {
				done := make(chan struct{})
				go HandleQuit(c.quit, done, ccControl)
				defer func() {
					close(done)
					ccControl.Done()
				}()
				doneMsg := common.MakeSuccDoneMsg(common.TaskUploadFileType, dpair.Name, sno)
				doneMsg.Local = dpair.Name
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

func serverPath(dest, fileName string) string {
	_, onlyFileName := filepath.Split(fileName)
	if strings.HasSuffix(dest, "/") {
		return dest + onlyFileName
	}
	return dest + "/" + onlyFileName
}

func localPath(dest, fileName string) string {
	_, onlyFileName := filepath.Split(fileName)
	if runtime.GOOS == "windows" {
		return dest + "\\" + onlyFileName
	}
	return filepath.Join(dest, onlyFileName)
}

// UploadFile upload file to provider
func (c *ClientManager) UploadFile(fileName, dest string, interactive, newVersion, isEncrypt bool, sno uint32) error {
	var err error
	var password, encryptKey []byte
	log := c.Log.WithField("upload file", fileName)
	defer func() {
		if r := recover(); r != nil {
			log.Error("!!!!!get panic info, recover it %s", r)
			debug.PrintStack()
		}
	}()
	if isEncrypt {
		password, err = c.getSpacePassword(sno)
		if err != nil {
			log.WithError(err).Info("Get space password")
			return err
		}
		if len(password) == 0 {
			log.Info("Space password not set")
			return fmt.Errorf("Password not set")
		}
		if sno == 0 {
			encryptKey, err = rsalong.EncryptLong(c.TrackerPubkey, password, 256)
			if err != nil {
				log.WithError(err).Info("Encrypt password")
				return err
			}
		}
	}

	fileType := filetype.FileType(fileName)
	// privacy space need encryp file whole
	if sno > 0 && isEncrypt {
		if len(password) == 0 {
			return errors.New("privacy no password")
		}
		log.Infof("privacy space %d , so encrypt file first", sno)
		originFileName := fileName
		// change fileName to encypted file avoid origin file modified
		_, onlyFileName := filepath.Split(fileName)
		fileName = filepath.Join(c.TempDir, onlyFileName)
		err := aes.EncryptFile(originFileName, password, fileName)
		if err != nil {
			log.Errorf("Encrypt error %v", err)
			return err
		}
		log = log.WithField("encrypted file", fileName)
	}

	req, rsp, err := c.CheckFileExists(fileName, dest, interactive, newVersion, password, encryptKey, sno, fileType)
	if err != nil {
		return err
	}

	log.Infof("Check file exists resp code %d", rsp.GetCode())
	if rsp.GetCode() == 0 {
		sp := serverPath(dest, fileName)
		c.PM.SetProgress(common.TaskUploadProgressType, common.ProgressKey(sp, sno), req.FileSize, req.FileSize, sno, fileName)
		log.Infof("Upload %s success", fileName)
		return nil
	}
	// 1 can upload
	if rsp.GetCode() != 1 {
		return common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}

	switch rsp.GetStoreType() {
	case mpb.FileStoreType_MultiReplica:
		log.Infof("Upload manner is multi-replication")
		// encrypt file
		originFileName := fileName
		if sno == 0 && isEncrypt {
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
		partitions, err := c.uploadFileByMultiReplica(originFileName, fileName, req, rsp, sno)
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
		sp := serverPath(dest, fileName)
		uniqKey := common.ProgressKey(sp, sno)
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
				log.Infof("Erasure block files %s index %d", fs.FileName, fs.SliceIndex)
				c.PM.SetPartitionMap(fs.FileName, uniqKey)
				realSizeAfterRS += fs.FileSize
			}
			c.PM.SetPartitionMap(fname, uniqKey)
		}

		c.PM.SetProgress(common.TaskUploadProgressType, uniqKey, 0, uint64(realSizeAfterRS), sno, fileName)

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
			return err
		}

		rspPartitions := ufprsp.GetPartition()
		log.Infof("Upload prepare response partitions count:%d", len(rspPartitions))

		if len(rspPartitions) == 0 {
			return fmt.Errorf("only 0 partitions, not correct")
		}

		for i, part := range rspPartitions {
			auth := part.GetProviderAuth()
			for _, pa := range auth {
				log.Infof("Partition %d, %s:%d %v hashauth %d", i, pa.Server, pa.Port, pa.Spare, len(pa.HashAuth))
			}
		}

		partitions := []*mpb.StorePartition{}
		for i, partInfo := range fileInfos {
			partition, err := c.uploadFileBatchByErasure(ufpr, rspPartitions[i], partInfo, dataShards, rsp.GetChunkSize())
			if err != nil {
				return err
			}
			log.Infof("Partition %d has %d store blocks", i, len(partition.GetBlock()))
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
			log.Infof("%s %dth piece", slice.FileName, j)
			block++
		}
		ufpr.Partition[i] = &mpb.SplitPartition{phslist}
		log.Infof("%s is %dth partitions", partInfo.FileName, i)
	}
	log.Infof("upload request has %d partitions, %d pieces", len(ufpr.Partition), block)
	if err := ufpr.SignReq(c.cfg.Node.PriKey); err != nil {
		return nil, err
	}

	return ufpr, nil
}

func deleteTemporaryFile(log logrus.FieldLogger, fileName string) {
	log.Infof("delete file %s", fileName)
	if err := os.Remove(fileName); err != nil {
		log.Errorf("delete %s failed, error %v", fileName, err)
	}
}

// CheckFileExists check file exists or not in tracker
func (c *ClientManager) CheckFileExists(fileName, dest string, interactive, newVersion bool, password, encryptKey []byte, sno uint32, fileType filetype.MIME) (*mpb.CheckFileExistReq, *mpb.CheckFileExistResp, error) {
	log := c.Log.WithField("filename", fileName)
	// privacy space
	hash, err := util_hash.Sha1File(fileName)
	if err != nil {
		return nil, nil, err
	}
	fileSize, err := GetFileSize(fileName)
	if err != nil {
		return nil, nil, err
	}
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
		if sno == 0 && len(password) != 0 {
			fileData, err = aes.Encrypt(fileData, password)
			if err != nil {
				log.Errorf("Encrypt file error %v", err)
				return nil, nil, err
			}
		}
		req.FileData = fileData
		sp := serverPath(dest, fileName)
		c.PM.SetProgress(common.TaskUploadProgressType, common.ProgressKey(sp, sno), req.FileSize, req.FileSize, sno, fileName)
		log.Infof("Origin filesize %d, encrypted size %d", req.FileSize, len(req.FileData))
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
		return false, common.NewStatus(errcode.RetSignFailed, err)
	}
	log.Infof("Make folder %+v", req.GetFolder())
	rsp, err := c.mclient.MkFolder(ctx, req)
	if err != nil {
		return false, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}
	if rsp.GetCode() != 0 {
		if strings.Contains(rsp.GetErrMsg(), "System error: pq: duplicate key value") {
			log.Warning("Folder exists %s", rsp.GetErrMsg())
			return true, nil
		}
		return false, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}
	log.Info("Make folder success")
	return true, nil
}

func (c *ClientManager) onlyFileSplit(fileName string, dataNum, verifyNum int, isEncrypt bool, password []byte, sno uint32) ([]common.HashFile, error) {
	log := c.Log.WithField("filesplit", fileName)
	fileSlices, err := RsEncoder(c.Log, c.TempDir, fileName, dataNum, verifyNum)
	if err != nil {
		log.Errorf("Reedsolomon encoder error %v", err)
		return nil, err
	}
	if sno == 0 && isEncrypt {
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

func (c *ClientManager) uploadFileBatchByErasure(req *mpb.UploadFilePrepareReq, rspPartition *mpb.ErasureCodePartition, partFile common.PartitionFile, dataShards int, chunkSize uint32) (*mpb.StorePartition, error) {
	log := c.Log
	partition := &mpb.StorePartition{}
	partition.Block = []*mpb.StoreBlock{}
	phas := rspPartition.GetProviderAuth()
	providers, backupPros := UsingBestProvider(phas)

	backupProMap := CreateBackupProvicer(backupPros)

	var errResult []error
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCUploadGoNum)
	for i, sortPro := range providers {
		checksum := i >= dataShards
		uploadParas := &common.UploadParameter{
			Checksum:       checksum,
			HF:             partFile.Pieces[i],
			OriginFileHash: partFile.OriginFileHash,
			OriginFileSize: partFile.OriginFileSize,
		}
		ccControl.Add()
		go func(pro *mpb.BlockProviderAuth, tm uint64, uploadPara *common.UploadParameter, chunkSize uint32) {
			log := log.WithField("provider", fmt.Sprintf("%s:%d", pro.Server, pro.Port))
			done := make(chan struct{})
			go HandleQuit(c.quit, done, ccControl)
			defer func() {
				close(done)
				ccControl.Done()
			}()
			var block *mpb.StoreBlock
			var err error
			for {
				block, err = c.uploadFileToErasureProvider(pro, tm, uploadParas, chunkSize)
				if err != nil {
					mutex.Lock()
					newPro := ChooseBackupProvicer(pro.HashAuth[0].Hash, backupProMap)
					mutex.Unlock()
					if newPro != nil {
						log.Infof("Choose provider success new %s:%d", newPro.Pro.Server, newPro.Pro.Port)
						pro = newPro.Pro
					} else {
						log.Info("Choose provider failed")
						break
					}
				} else {
					break
				}
			}

			mutex.Lock()
			defer mutex.Unlock()
			if err != nil {
				log.Errorf("Upload file %s error %v", uploadPara.HF.FileName, err)
				errResult = append(errResult, err)
				return
			}
			partition.Block = append(partition.Block, block)
		}(sortPro.Pro, rspPartition.GetTimestamp(), uploadParas, chunkSize)
	}
	ccControl.Wait()
	if len(errResult) != 0 {
		return partition, errResult[0]
	}
	return partition, nil
}

func (c *ClientManager) uploadFileToErasureProvider(pro *mpb.BlockProviderAuth, tm uint64, uploadPara *common.UploadParameter, chunkSize uint32) (*mpb.StoreBlock, error) {
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	log := c.Log.WithField("server", server).WithField("erasurefile", uploadPara.HF.FileName)
	uploadPara.Provider = server
	conn, err := common.GrpcDial(server)
	if err != nil {
		log.Errorf("Rpc dial failed: %s", err.Error())
		return nil, err
	}
	defer conn.Close()
	pclient := pb.NewProviderServiceClient(conn)

	if chunkSize <= 0 {
		log.Errorf("chunksize[%d] can not less than 0", chunkSize)
		return nil, fmt.Errorf("chunksize[%d] can not less than 0", chunkSize)
	}

	t1 := time.Now()
	c.MetaChan <- MetaKey{FileName: uploadPara.HF.FileName, ChunkSize: chunkSize}

	ha := pro.GetHashAuth()[0]
	err = client.StorePiece(log, pclient, uploadPara, ha.GetAuth(), ha.GetTicket(), tm, c.PM)
	if err != nil {
		return nil, err
	}

	var paraStr string
	var generator, pubKey, random []byte
	var phi [][]byte
	var waitCount int
	paraStr, generator, pubKey, random, phi, err = c.mdm.Get(uploadPara.HF.FileName)
	for err != nil && err == ErrNoMetaData && waitCount < 60*60 {
		time.Sleep(1)
		paraStr, generator, pubKey, random, phi, err = c.mdm.Get(uploadPara.HF.FileName)
		waitCount++
	}
	t2 := time.Now()
	log.Infof("upload %s time elapased %+v", uploadPara.HF.FileName, t2.Sub(t1).Seconds())

	if err != nil {
		return nil, err
	}

	block := &mpb.StoreBlock{
		StoreNodeId: [][]byte{},
		Checksum:    uploadPara.Checksum,
		Hash:        uploadPara.HF.FileHash,
		Size:        uint64(uploadPara.HF.FileSize),
		BlockSeq:    uint32(uploadPara.HF.SliceIndex),
		ChunkSize:   chunkSize,
		ParamStr:    paraStr,
		Generator:   generator,
		PubKey:      pubKey,
		Random:      random,
		Phi:         phi,
	}
	block.StoreNodeId = append(block.StoreNodeId, []byte(pro.GetNodeId()))
	log.Infof("Upload to provider success")

	return block, nil
}

func (c *ClientManager) uploadFileToReplicaProvider(conn *grpc.ClientConn, pro *mpb.ReplicaProvider, uploadPara *common.UploadParameter) ([]byte, error) {
	fileInfo := uploadPara.HF
	server := fmt.Sprintf("%s:%d", pro.GetServer(), pro.GetPort())
	log := c.Log.WithField("uploading", fileInfo.FileName).WithField("provider", server)
	uploadPara.Provider = server
	pclient := pb.NewProviderServiceClient(conn)
	log.Infof("Upload file hash %x size %d", fileInfo.FileHash, fileInfo.FileSize)

	err := client.StorePiece(log, pclient, uploadPara, pro.Auth, pro.Ticket, pro.Timestamp, c.PM)
	if err != nil {
		log.WithError(err).Error("Upload error")
		return nil, err
	}

	log.Info("Upload file success")

	return pro.GetNodeId(), nil
}

func newActionLogFromUpload(fileName string) *tcppb.ActionLog {
	return &tcppb.ActionLog{
		Type:      3,
		BeginTime: common.Now(),
		Info:      fileName,
	}
}

func (c *ClientManager) uploadFileByMultiReplica(originFileName, fileName string, req *mpb.CheckFileExistReq, rsp *mpb.CheckFileExistResp, sno uint32) ([]*mpb.StorePartition, error) {
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
		return nil, err
	}
	///

	uploadPara := &common.UploadParameter{
		OriginFileHash: req.FileHash,
		OriginFileSize: req.FileSize,
		HF:             fileSlices[0],
	}

	t1 := time.Now()
	fmt.Printf("chunksize %d, filename %s\n", rsp.GetChunkSize(), fileName)

	if rsp.GetChunkSize() <= 0 {
		log.Errorf("chunksize[%d] can not less than 0", rsp.GetChunkSize())
		return nil, fmt.Errorf("chunksize[%d] can not less than 0", rsp.GetChunkSize())
	}

	c.MetaChan <- MetaKey{FileName: fileName, ChunkSize: rsp.GetChunkSize()}
	var paraStr string
	var generator, pubKey, random []byte
	var phi [][]byte

	var waitCount int
	paraStr, generator, pubKey, random, phi, err = c.mdm.Get(fileName)
	for err != nil && err == ErrNoMetaData && waitCount < 60*60 {
		time.Sleep(1)
		paraStr, generator, pubKey, random, phi, err = c.mdm.Get(fileName)
		waitCount++
	}

	if err != nil {
		return nil, err
	}
	t2 := time.Now()
	log.Infof("upload %s time elapased %+v", fileName, t2.Sub(t1).Seconds())

	block := &mpb.StoreBlock{
		Checksum:    false,
		StoreNodeId: [][]byte{},
		Hash:        fileSlices[0].FileHash,
		Size:        uint64(fileSlices[0].FileSize),
		BlockSeq:    uint32(fileSlices[0].SliceIndex),
		ChunkSize:   rsp.GetChunkSize(),
		ParamStr:    paraStr,
		Generator:   generator,
		PubKey:      pubKey,
		Random:      random,
		Phi:         phi,
	}

	sp := serverPath(req.Parent.GetPath(), fileName)
	uniqKey := common.ProgressKey(sp, sno)
	c.PM.SetPartitionMap(fileName, uniqKey)

	providers, backupPros, err := GetBestReplicaProvider(ufprsp.GetProvider(), MinReplicaNum)
	if err != nil {
		return nil, err
	}
	if fileName == originFileName {
		c.PM.SetProgress(common.TaskUploadProgressType, uniqKey, 0, uint64(len(providers))*req.FileSize, sno, originFileName)
	} else {
		c.PM.SetProgress(common.TaskUploadProgressType, uniqKey, 0, uint64(int64(len(providers))*fileSize), sno, originFileName)
	}

	HandlerUpload := func(providers []*mpb.ReplicaProvider, block *mpb.StoreBlock, uploadPara *common.UploadParameter) []error {
		errArr := []error{}
		var mutex sync.Mutex
		ccControl := NewCCController(common.CCUploadFileNum)
		for _, pro := range providers {
			ccControl.Add()
			go func(pro *mpb.ReplicaProvider) {
				server := fmt.Sprintf("%s:%d", pro.Server, pro.Port)
				conn, err := common.GrpcDial(server)
				if err != nil {
					log.Errorf("Rpc dail failed: %v", err)
					mutex.Lock()
					errArr = append(errArr, err)
					mutex.Unlock()
					ccControl.Done()
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
		return errArr
	}
	//al := newActionLogFromUpload(fileName)
	//defer collectClient.Collect(al)
	errArr := HandlerUpload(providers, block, uploadPara)
	if len(errArr) > 0 {
		log.Errorf("Upload error %v, total %d", errArr[0], len(errArr))
		if len(backupPros) < len(errArr) {
			log.Errorf("Not enough bakcup providers, backup provider %d", len(backupPros))
			errInfo := ""
			for _, err := range errArr {
				errInfo += err.Error() + "\n"
			}
			//client.SetActionLog(err, al)
			return nil, errors.New(errInfo)
		}
		errArr = HandlerUpload(backupPros[0:len(errArr)], block, uploadPara)
	}
	if len(errArr) > 0 {
		log.Errorf("No solution for upload")
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
		return common.NewStatus(errcode.RetSignFailed, err)
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
				return common.NewStatus(errcode.RetSignFailed, err)
			}
			log.Info("Check file exist request")
			ufdrsp, err = c.mclient.UploadFileDone(ctx, req)
			if err != nil {
				return err
			}
		}
	}
	log.Infof("Upload done code %d", ufdrsp.GetCode())
	if ufdrsp.GetCode() != 0 {
		return common.NewStatusErr(ufdrsp.Code, ufdrsp.ErrMsg)
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
		return nil, common.NewStatus(errcode.RetSignFailed, err)
	}
	ctx := context.Background()
	log.Info("List path request")
	rsp, err := c.mclient.ListFiles(ctx, req)

	if err != nil {
		return nil, common.NewStatus(errcode.RetTrackerFailed, err)
	}

	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
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
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCDownloadGoNum)
	for {
		// list 1 page 100 items order by name
		retry := 0
	RETRY:
		downFiles, err := c.ListFiles(path, 100, page, "name", true, sno)
		retry++
		if err != nil {
			code, _ := common.StatusErrFromError(err)
			if code == 300 && retry < RetryCount {
				goto RETRY
			}
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
					ccControl.Add()
					go func(currentFile, destDir string, fileInfo *DownFile, sno uint32) {
						done := make(chan struct{})
						go HandleQuit(c.quit, done, ccControl)
						defer func() {
							close(done)
							ccControl.Done()
						}()
						doneMsg := common.MakeSuccDoneMsg(common.TaskDownloadFileType, currentFile, sno)
						err = c.DownloadFile(currentFile, destDir, fileInfo.FileHash, fileInfo.FileSize, sno)
						if err != nil {
							doneMsg.SetError(1, err)
						}

						c.AddDoneMsg(doneMsg.Serialize())
						if err != nil {
							log.Errorf("Download file %s error %v", currentFile, err)
							//errResult = append(errResult, fmt.Errorf("%s %v", currentFile, err))
							mutex.Lock()
							errResult = append(errResult, err)
							mutex.Unlock()
							//return err
						}
					}(currentFile, destDir, fileInfo, sno)
				}
			}
		}
	}
	ccControl.Wait()
	if len(errResult) > 0 {
		for _, err := range errResult {
			log.Errorf("Download dir error: %v", err)
		}
		return errResult[0]
	}
	return nil
}

func (c *ClientManager) CheckFileExistsInLocal(downFileName string, filehash []byte) bool {
	if !util_file.Exists(downFileName) {
		return false
	}
	localHash, err := util_hash.Sha1File(downFileName)
	if err != nil {
		return false
	}
	if bytes.Compare(localHash, filehash) == 0 {
		return true
	}
	return false
}

// DownloadFile download file
func (c *ClientManager) DownloadFile(downFileName, destDir, filehash string, fileSize uint64, sno uint32) error {
	serverFile := downFileName
	_, fileName := filepath.Split(downFileName)
	downFileName = filepath.Join(destDir, fileName)
	log := c.Log.WithField("download file", downFileName)
	defer func() {
		if r := recover(); r != nil {
			log.Error("!!!!!get panic info, recover it %s", r)
			debug.PrintStack()
		}
	}()
	fileHash, err := hex.DecodeString(filehash)
	if err != nil {
		return err
	}
	exists := c.CheckFileExistsInLocal(downFileName, fileHash)
	if exists {
		log.Info("File exists in local")
		c.PM.SetProgress(common.TaskDownloadProgressType, common.ProgressKey(serverFile, sno), fileSize, fileSize, sno, downFileName)
		return nil
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
	c.PM.SetProgress(common.TaskDownloadProgressType, common.ProgressKey(serverFile, sno), 0, req.FileSize, sno, downFileName)

	ctx := context.Background()
	log.Infof("Download request file hash %x, size %d", fileHash, fileSize)
	rsp, err := c.mclient.RetrieveFile(ctx, req)
	if err != nil {
		return err
	}
	if rsp.GetCode() != 0 {
		return common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}

	password := []byte{}
	if sno > 0 {
		password, err = c.getSpacePassword(sno)
		if err != nil {
			log.WithError(err).Info("Get space password")
			return err
		}
		if len(password) == 0 {
			log.Info("Space %d password not set", sno)
			return fmt.Errorf("sno %d password not set", sno)
		}
	} else {
		encryptKey := rsp.GetEncryptKey()
		if len(encryptKey) > 0 {
			password, err = rsalong.DecryptLong(c.cfg.Node.PriKey, encryptKey, 256)
			if err != nil {
				return err
			}
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
		c.PM.SetProgress(common.TaskDownloadProgressType, common.ProgressKey(serverFile, sno), req.FileSize, req.FileSize, sno, downFileName)
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
				c.PM.SetPartitionMap(hex.EncodeToString(block.GetHash()), common.ProgressKey(serverFile, sno))
			}
			_, _, _, _, _, err := c.saveFileByPartition(downFileName, partitions[0], rsp.GetTimestamp(), req.FileHash, req.FileSize, true)
			if err != nil {
				return err
			}
			if len(password) != 0 {
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
	for _, partition := range partitions {
		for j, block := range partition.GetBlock() {
			for _, sn := range block.GetStoreNode() {
				fmt.Printf("block %d hash %x seq %d provider %s:%d\n", j, block.Hash, block.BlockSeq, sn.Server, sn.Port)
			}
		}
	}
	for i, partition := range partitions {
		for j, block := range partition.GetBlock() {
			c.PM.SetPartitionMap(hex.EncodeToString(block.GetHash()), common.ProgressKey(serverFile, sno))
			realSizeAfterRS += block.GetSize()
			log.Infof("Partition %d block %d hash %x size %d checksum %v seq %d", i, j, block.Hash, block.Size, block.Checksum, block.BlockSeq)
			for _, sn := range block.GetStoreNode() {
				log.Infof("block %d hash %x seq %d provider %s:%d", j, block.Hash, block.BlockSeq, sn.Server, sn.Port)
			}
		}
	}
	c.PM.SetProgress(common.TaskDownloadProgressType, common.ProgressKey(serverFile, sno), 0, realSizeAfterRS, sno, downFileName)

	if len(partitions) == 1 {
		partition := partitions[0]
		datas, paritys, failedCount, middleFiles, allMiddleFiles, err := c.saveFileByPartition(downFileName, partition, rsp.GetTimestamp(), req.FileHash, req.FileSize, false)
		if failedCount > paritys {
			log.Error("File cannot be recoved!!!")
			return err
		}
		log.Infof("DataShards %d, parityShards %d, failedCount %d, middlefile %d", datas, paritys, failedCount, len(middleFiles))
		if err != nil {
			log.Errorf("Save file by partition error %v, but file still can be recoverd", err)
			for _, file := range allMiddleFiles {
				deleted := true
				for _, wellFile := range middleFiles {
					if file == wellFile {
						deleted = false
					}
				}
				if deleted {
					deleteTemporaryFile(log, file)
				}
			}
		}
		if len(middleFiles) < datas {
			err := fmt.Errorf("need %d shards, but only download %d, so cannot reconstrct", datas, len(middleFiles))
			log.Error(err)
			return err
		}

		if sno == 0 && len(password) != 0 {
			for _, file := range middleFiles {
				log.Infof("middle file %s", file)
				if err := aes.DecryptFile(file, password, file); err != nil {
					return err
				}
			}
		}

		// delete middle files
		defer func() {
			for _, file := range allMiddleFiles {
				deleteTemporaryFile(log, file)
			}
		}()

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

		if sno > 0 && len(password) > 0 {
			if err := aes.DecryptFile(tempDownFileName, password, tempDownFileName); err != nil {
				return err
			}
		}

		return RenameCrossOS(tempDownFileName, downFileName)
	}
	partFiles := []string{}
	for i, partition := range partitions {
		partFileName := fmt.Sprintf("%s.%s.%d", downFileName, TEMP_NAMESPACE, i)
		log := log.WithField("middle file", partFileName)
		datas, paritys, failedCount, middleFiles, allMiddleFiles, err := c.saveFileByPartition(partFileName, partition, rsp.GetTimestamp(), req.FileHash, req.FileSize, false)
		if failedCount > paritys {
			log.Errorf("Middle file %s cannot be recoved!!!", partFileName)
			return err
		}
		if err != nil {
			log.WithError(err).Error("Save file by partition error, but file still can be recoverd")
			for _, file := range allMiddleFiles {
				deleted := true
				for _, wellFile := range middleFiles {
					if file == wellFile {
						deleted = false
					}
				}
				if deleted {
					deleteTemporaryFile(log, file)
				}
			}
		}
		if sno == 0 && len(password) != 0 {
			for _, file := range middleFiles {
				log.Infof("middle file %s", file)
				if err := aes.DecryptFile(file, password, file); err != nil {
					log.WithError(err).Error("decrypt file failed")
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

		if sno > 0 && len(password) > 0 {
			if err := aes.DecryptFile(tempDownFileName, password, tempDownFileName); err != nil {
				return err
			}
		}

		partFiles = append(partFiles, tempDownFileName)
		// delete middle files
		for _, file := range allMiddleFiles {
			deleteTemporaryFile(log, file)
		}

	}

	defer func() {
		for _, file := range partFiles {
			deleteTemporaryFile(log, file)
		}
	}()

	for _, f := range partFiles {
		log.Infof("Part file %s for join", f)
	}

	if err := FileJoin(downFileName, partFiles); err != nil {
		log.Errorf("File %s join failed, part files %+v", downFileName, partFiles)
		return err
	}
	return nil
}

func (c *ClientManager) saveFileByPartition(fileName string, partition *mpb.RetrievePartition, tm uint64, fileHash []byte, fileSize uint64, multiReplica bool) (int, int, int, []string, []string, error) {
	log := c.Log.WithField("filename", fileName)
	log.Infof("There is %d blocks", len(partition.GetBlock()))
	dataShards := 0
	parityShards := 0
	failedCount := 0
	successCount := 0
	middleFiles := []string{}
	allMiddleFiles := []string{}
	errArray := []string{}
	var mutex sync.Mutex
	ccControl := NewCCController(common.CCDownloadGoNum)
	for _, block := range partition.GetBlock() {
		if block.GetChecksum() {
			parityShards++
		} else {
			dataShards++
		}
		tempFileName := fileName
		if !multiReplica {
			_, onlyFileName := filepath.Split(fileName)
			tempFileName = filepath.Join(c.TempDir, fmt.Sprintf("%s.%d", onlyFileName, block.GetBlockSeq()))
		}
		allMiddleFiles = append(allMiddleFiles, tempFileName)
	}

	normalFlow := true
	currQuit := make(chan struct{})
	go func() {
		for {
			select {
			case <-time.After(30 * time.Second):
				log.Infof("success %d datashars %d", successCount, dataShards)
				if successCount >= dataShards {
					normalFlow = false
					close(currQuit)
					return
				}
			case <-c.quit:
				return
			case <-currQuit:
				return
			}
		}
	}()

	//al := newActionLogFromUpload(fileName)
	//defer collectClient.Collect(al)

	for _, block := range partition.GetBlock() {
		ccControl.Add()
		go func(log logrus.FieldLogger, block *mpb.RetrieveBlock, fileName string) {
			node := BestRetrieveNode(block.GetStoreNode())
			server := fmt.Sprintf("%s:%d", node.GetServer(), node.GetPort())
			conn, err := common.GrpcDial(server)
			if err != nil {
				log.Errorf("Rpc dial %s failed, error %v", server, err)
				mutex.Lock()
				failedCount++
				errArray = append(errArray, err.Error())
				mutex.Unlock()
				ccControl.Done()
				//client.SetActionLog(err, al)
				return
			}
			done := make(chan struct{})
			go HandleQuit(currQuit, done, ccControl, func() { conn.Close() })
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
				errArray = append(errArray, err.Error())
				mutex.Unlock()
				//client.SetActionLog(err, al)
				return
			}
			log.Info("Retrieve success")
			mutex.Lock()
			middleFiles = append(middleFiles, tempFileName)
			successCount++
			mutex.Unlock()
			conn.Close()

		}(log, block, fileName)
	}

	ccControl.Wait()
	if normalFlow {
		close(currQuit)
	}

	if len(errArray) > 0 {
		fmt.Printf("download goroutine failed %d\n", len(errArray))
		errRtn := fmt.Errorf("%s", strings.Join(errArray, "\n"))
		return dataShards, parityShards, failedCount, middleFiles, allMiddleFiles, errRtn
	}
	return dataShards, parityShards, failedCount, middleFiles, allMiddleFiles, nil
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
		return err
	}
	log.Infof("Remove file resp code %d msg %s", rsp.GetCode(), rsp.GetErrMsg())
	if rsp.GetCode() != 0 {
		return common.NewStatusErr(rsp.Code, rsp.ErrMsg)
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
		return err
	}
	log.Infof("Move file resp code %d msg %s", rsp.GetCode(), rsp.GetErrMsg())
	if rsp.GetCode() != 0 {
		return common.NewStatusErr(rsp.Code, rsp.ErrMsg)
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
		return nil, err
	}
	return rsp.GetData(), nil
}

// GetProgress returns progress rate
func (c *ClientManager) GetProgress(files []string) (progress.ProgressReadable, error) {
	return c.PM.GetProgress(files)
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

// TaskDelete get task status
func (c *ClientManager) TaskDelete(taskID string) (TaskInfo, error) {
	return c.store.UpdateTaskInfo(taskID, func(rs TaskInfo) TaskInfo {
		rs.UpdatedAt = common.Now()
		rs.Deleted = true
		return rs
	})
}
