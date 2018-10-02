package impl

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"os"
	"sort"
	"sync"
	"time"

	gosync "github.com/lrita/gosync"
	client "github.com/samoslab/nebula/provider/collector_client"
	"github.com/samoslab/nebula/provider/config"
	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/provider/pb"
	provider_client "github.com/samoslab/nebula/provider/provider_client"
	task_client "github.com/samoslab/nebula/provider/task_client"
	tcppb "github.com/samoslab/nebula/tracker/collector/provider/pb"
	ttpb "github.com/samoslab/nebula/tracker/task/pb"
	util_hash "github.com/samoslab/nebula/util/hash"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	leveldb_errors "github.com/syndtr/goleveldb/leveldb/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const stream_data_size = 32 * 1024
const small_file_limit = 512 * 1024

var skip_check_auth = false

type ProviderService struct {
	node               *node.Node
	nodeIdHash         []byte
	providerDb         *leveldb.DB
	taskGetting        gosync.Mutex
	replicateChan      chan *ttpb.Task
	sendChan           chan *ttpb.Task
	removeAndProveChan chan *ttpb.Task
	closeSignal        chan bool
	waitClose          sync.WaitGroup
	taskConnection     *grpc.ClientConn
	ptsc               ttpb.ProviderTaskServiceClient
}

func NewProviderService(taskServer string, private bool) *ProviderService {
	if os.Getenv("NEBULA_TEST_MODE") == "1" {
		skip_check_auth = true
	}
	ps := &ProviderService{}
	ps.node = node.LoadFormConfig()
	ps.nodeIdHash = util_hash.Sha1(ps.node.NodeId)
	var err error
	ps.providerDb, err = leveldb.OpenFile(config.ProviderDbPath(), nil)
	if err != nil {
		log.Fatalf("open Provider DB failed:%s", err)
	}
	ps.initTaskProcessor(taskServer, private)
	return ps
}

func (self *ProviderService) Close() {
	self.providerDb.Close()
}

func (self *ProviderService) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{NodeIdHash: self.nodeIdHash}, nil
}

func now() uint64 {
	return uint64(time.Now().UnixNano())
}

func logWarnAndSetActionLog(err error, al *tcppb.ActionLog) {
	log.Warnln(err)
	if al != nil {
		al.EndTime, al.Info = now(), err.Error()
	}
}

func newActionLogFromStoreReq(req *pb.StoreReq) *tcppb.ActionLog {
	return &tcppb.ActionLog{Type: 1,
		Ticket:    req.Ticket,
		FileHash:  req.FileKey,
		FileSize:  req.FileSize,
		BlockHash: req.BlockKey,
		BlockSize: req.BlockSize,
		BeginTime: now()}
}

func newActionLogFromRetrieveReq(req *pb.RetrieveReq) *tcppb.ActionLog {
	return &tcppb.ActionLog{Type: 2,
		Ticket:    req.Ticket,
		FileHash:  req.FileKey,
		FileSize:  req.FileSize,
		BlockHash: req.BlockKey,
		BlockSize: req.BlockSize,
		BeginTime: now()}
}

func (self *ProviderService) StoreSmall(ctx context.Context, req *pb.StoreReq) (resp *pb.StoreResp, err error) {
	al := newActionLogFromStoreReq(req)
	al.TransportSize = uint64(len(req.Data))
	defer client.Collect(al)
	if req.BlockSize >= small_file_limit || int(req.BlockSize) != len(req.Data) {
		err = status.Errorf(codes.InvalidArgument, "check data size failed, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	if !bytes.Equal(req.BlockKey, util_hash.Sha1(req.Data)) {
		err = status.Errorf(codes.InvalidArgument, "check data hash failed, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	if !skip_check_auth {
		if err = req.CheckAuth(self.node.PubKeyBytes); err != nil {
			err = status.Errorf(codes.Unauthenticated, "check auth failed, blockKey: %x error: %s", req.BlockKey, err)
			logWarnAndSetActionLog(err, al)
			return
		}
	}
	if found, _, _, _ := self.querySubPath(req.BlockKey); found {
		err = status.Errorf(codes.AlreadyExists, "hash point file exist, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	storage := config.GetWriteStorage(req.BlockSize)
	if storage == nil {
		err = status.Errorf(codes.ResourceExhausted, "available disk space of this provider is not enlough, blockKey: %s blockSize: %d", req.BlockKey, req.BlockSize)
		logWarnAndSetActionLog(err, al)
		al.TransportSize += uint64(len(req.Data))
		return
	}
	if err = storage.SmallFileDb.Put(req.BlockKey, req.Data, nil); err != nil {
		err = status.Errorf(codes.Internal, "save to small file db failed, blockKey: %x error: %s", req.BlockKey, err)
		logWarnAndSetActionLog(err, al)
		return
	}
	if err = self.providerDb.Put(req.BlockKey, []byte{storage.Index}, nil); err != nil {
		err = status.Errorf(codes.Internal, "save to provider db failed, blockKey: %x error: %s", req.BlockKey, err)
		logWarnAndSetActionLog(err, al)
		return
	}
	al.Success, al.EndTime = true, now()
	return &pb.StoreResp{Success: true}, nil
}

func (self *ProviderService) Store(stream pb.ProviderService_StoreServer) (er error) {
	var al *tcppb.ActionLog
	first := true
	var tempFilePath string
	var file *os.File
	var storage *config.Storage
	var blockKey []byte
	var blockSize uint64
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			er = status.Errorf(codes.Unknown, "RPC Recv failed unexpectadely while reading chunks from stream, blockKey: %x error: %s", blockKey, err)
			logWarnAndSetActionLog(er, al)
			return
		}
		if first {
			al = newActionLogFromStoreReq(req)
			defer client.Collect(al)
			first, blockKey, blockSize = false, req.BlockKey, req.BlockSize
			if blockSize < small_file_limit {
				er = status.Errorf(codes.InvalidArgument, "check data size failed, blockKey: %x", blockKey)
				logWarnAndSetActionLog(er, al)
				al.TransportSize += uint64(len(req.Data))
				return
			}
			if !skip_check_auth {
				if err = req.CheckAuth(self.node.PubKeyBytes); err != nil {
					er = status.Errorf(codes.Unauthenticated, "check auth failed, blockKey: %x error: %s", blockKey, err)
					logWarnAndSetActionLog(er, al)
					al.TransportSize += uint64(len(req.Data))
					return
				}
			}
			if found, _, _, _ := self.querySubPath(blockKey); found {
				er = status.Errorf(codes.AlreadyExists, "hash point file exist, blockKey: %x", blockKey)
				logWarnAndSetActionLog(er, al)
				al.TransportSize += uint64(len(req.Data))
				return
			}
			storage = config.GetWriteStorage(blockSize)
			if storage == nil {
				er = status.Errorf(codes.ResourceExhausted, "available disk space of this provider is not enlough, blockKey: %s blockSize: %d", blockKey, blockSize)
				logWarnAndSetActionLog(er, al)
				al.TransportSize += uint64(len(req.Data))
				return
			}
			tempFilePath = storage.TempFilePath(blockKey)
			file, err = os.OpenFile(
				tempFilePath,
				os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
				0600)
			if err != nil {
				er = status.Errorf(codes.Internal, "open temp write file failed, blockKey: %x error: %s", blockKey, err)
				logWarnAndSetActionLog(er, al)
				al.TransportSize += uint64(len(req.Data))
				return
			}
			defer file.Close()
		}
		if len(req.Data) == 0 {
			break
		}
		al.TransportSize += uint64(len(req.Data))
		if al.TransportSize > blockSize {
			er = status.Errorf(codes.InvalidArgument, "transport data size exceed: %d, blockKey: %x blockSize: %d", al.TransportSize, blockKey, blockSize)
			logWarnAndSetActionLog(er, al)
			return
		}
		if _, err = file.Write(req.Data); err != nil {
			er = status.Errorf(codes.Internal, "write file failed, blockKey: %x error: %s", blockKey, err)
			logWarnAndSetActionLog(er, al)
			return
		}
	}
	fileInfo, err := os.Stat(tempFilePath)
	if err != nil {
		er = status.Errorf(codes.Internal, "stat temp file failed, blockKey: %x error: %s", blockKey, err)
		logWarnAndSetActionLog(er, al)
		return
	}
	if blockSize != uint64(fileInfo.Size()) {
		er = status.Errorf(codes.InvalidArgument, "check data size failed, blockKey: %x", blockKey)
		logWarnAndSetActionLog(er, al)
		return
	}
	hash, err := util_hash.Sha1File(tempFilePath)
	if err != nil {
		er = status.Errorf(codes.Internal, "sha1 sum file %s failed, blockKey: %x error: %s", tempFilePath, blockKey, err)
		logWarnAndSetActionLog(er, al)
		return
	}
	if !bytes.Equal(hash, blockKey) {
		er = status.Errorf(codes.InvalidArgument, "hash verify failed, blockKey: %x error: %s", blockKey, err)
		logWarnAndSetActionLog(er, al)
		return
	}
	if err := file.Close(); err != nil {
		er = status.Errorf(codes.Internal, "close temp file failed, tempFilePath: %s blockKey: %x error: %s", tempFilePath, blockKey, err)
		logWarnAndSetActionLog(er, al)
		return
	}
	if err := self.saveFile(blockKey, blockSize, tempFilePath, storage); err != nil {
		er = status.Errorf(codes.Internal, "save file failed, tempFilePath: %s blockKey: %x error: %s", tempFilePath, blockKey, err)
		logWarnAndSetActionLog(er, al)
		return
	}
	if err := stream.SendAndClose(&pb.StoreResp{Success: true}); err != nil {
		er = status.Errorf(codes.Unknown, "RPC SendAndClose failed, blockKey: %x error: %s", blockKey, err)
		logWarnAndSetActionLog(er, al)
		return
	}
	if al != nil {
		al.Success, al.EndTime = true, now()
	}
	return nil
}

func (self *ProviderService) RetrieveSmall(ctx context.Context, req *pb.RetrieveReq) (resp *pb.RetrieveResp, err error) {
	al := newActionLogFromRetrieveReq(req)
	defer client.Collect(al)
	if req.BlockSize >= small_file_limit {
		err = status.Errorf(codes.InvalidArgument, "check data size failed, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	if !skip_check_auth {
		if err = req.CheckAuth(self.node.PubKeyBytes); err != nil {
			err = status.Errorf(codes.Unauthenticated, "check auth failed, blockKey: %x error: %s", req.BlockKey, err)
			logWarnAndSetActionLog(err, al)
			return
		}
	}
	found, smallFile, storageIdx, _ := self.querySubPath(req.BlockKey)
	if !found {
		err = status.Errorf(codes.NotFound, "file not exist, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	if !smallFile {
		err = status.Errorf(codes.FailedPrecondition, "is not small file, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	storage := config.GetStorage(storageIdx)
	data, err := storage.SmallFileDb.Get(req.BlockKey, nil)
	if err != nil {
		err = status.Errorf(codes.Internal, "read small file error, blockKey: %x error: %s", req.BlockKey, err)
		logWarnAndSetActionLog(err, al)
		return
	}
	if len(data) != int(req.BlockSize) {
		err = status.Errorf(codes.InvalidArgument, "check data size failed, read length %d != request length %d, blockKey: %x", len(data), req.BlockSize, req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	if !bytes.Equal(util_hash.Sha1(data), req.BlockKey) {
		err = status.Errorf(codes.DataLoss, "hash verify failed, blockKey: %x error: %s", req.BlockKey, err)
		logWarnAndSetActionLog(err, al)
		return
	}
	al.Success, al.EndTime, al.TransportSize = true, now(), uint64(len(data))
	return &pb.RetrieveResp{Data: data}, nil
}

func (self *ProviderService) Retrieve(req *pb.RetrieveReq, stream pb.ProviderService_RetrieveServer) (err error) {
	al := newActionLogFromRetrieveReq(req)
	defer client.Collect(al)
	if req.BlockSize < small_file_limit {
		err = status.Errorf(codes.InvalidArgument, "check data size failed, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	if !skip_check_auth {
		if err = req.CheckAuth(self.node.PubKeyBytes); err != nil {
			err = status.Errorf(codes.Unauthenticated, "check auth failed, blockKey: %x error: %s", req.BlockKey, err)
			logWarnAndSetActionLog(err, al)
			return
		}
	}
	found, smallFile, storageIdx, subPath := self.querySubPath(req.BlockKey)
	if !found {
		err = status.Errorf(codes.NotFound, "file not exist, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	if smallFile {
		err = status.Errorf(codes.FailedPrecondition, "is small file, blockKey: %x", req.BlockKey)
		logWarnAndSetActionLog(err, al)
		return
	}
	path := config.GetStoragePath(storageIdx, subPath)
	hash, err := util_hash.Sha1File(path)
	if err != nil {
		err = status.Errorf(codes.Internal, "sha1 sum file %s failed, blockKey: %x error: %s", path, req.BlockKey, err)
		logWarnAndSetActionLog(err, al)
		return
	}
	if !bytes.Equal(hash, req.BlockKey) {
		err = status.Errorf(codes.DataLoss, "hash verify failed, blockKey: %x error: %s", req.BlockKey, err)
		logWarnAndSetActionLog(err, al)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		err = status.Errorf(codes.Internal, "open file failed, blockKey: %x error: %s", req.BlockKey, err)
		logWarnAndSetActionLog(err, al)
		return
	}
	defer file.Close()
	if err = sendFileToStream(req.BlockKey, path, file, stream, al); err != nil {
		return err
	}
	al.Success, al.EndTime = true, now()
	return nil
}

func sendFileToStream(key []byte, path string, file *os.File, stream pb.ProviderService_RetrieveServer, al *tcppb.ActionLog) (er error) {
	buf := make([]byte, stream_data_size)
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			er = status.Errorf(codes.Internal, "read file: %s failed, blockKey: %x error: %s", path, key, err)
			logWarnAndSetActionLog(er, al)
			return
		}
		if bytesRead > 0 {
			stream.Send(&pb.RetrieveResp{Data: buf[:bytesRead]})
			al.TransportSize += uint64(bytesRead)
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	return nil
}

func (self *ProviderService) Remove(ctx context.Context, req *pb.RemoveReq) (resp *pb.RemoveResp, err error) {
	if !skip_check_auth {
		if err = req.CheckAuth(self.node.PubKeyBytes); err != nil {
			err = status.Errorf(codes.Unauthenticated, "check auth failed, key: %x error: %s", req.Key, err)
			log.Warnln(err)
			return
		}
	}
	found, smallFile, storageIdx, subPath := self.querySubPath(req.Key)
	if !found {
		err = status.Errorf(codes.NotFound, "file not exist, key: %x", req.Key)
		log.Warnln(err)
		return
	}
	if err = self.providerDb.Delete(req.Key, nil); err != nil {
		err = status.Errorf(codes.Internal, "delete from provider db failed, key: %x error: %s", req.Key, err)
		log.Warnln(err)
		return
	}
	if smallFile {
		storage := config.GetStorage(storageIdx)
		if err = storage.SmallFileDb.Delete(req.Key, nil); err != nil {
			err = status.Errorf(codes.Internal, "delete from small file db failed, key: %x error: %s", req.Key, err)
			log.Warnln(err)
			return
		}
	} else {
		path := config.GetStoragePath(storageIdx, subPath)
		if err = os.Remove(path); err != nil {
			err = status.Errorf(codes.Internal, "remove file failed, key: %x error: %s", req.Key, err)
			log.Warnln(err)
			return
		}
	}
	return &pb.RemoveResp{Success: true}, nil
}

func (self *ProviderService) GetFragment(ctx context.Context, req *pb.GetFragmentReq) (resp *pb.GetFragmentResp, err error) {
	if len(req.Positions) == 0 || req.Size == 0 {
		err = status.Errorf(codes.InvalidArgument, "invalid req, key: %x", req.Key)
		log.Warnln(err)
		return
	}
	for i, b := range req.Positions {
		if b >= 100 {
			err = status.Errorf(codes.InvalidArgument, "posisiton out of bounds, Posistion[%d]=%d , key: %x", i, b, req.Key)
			log.Warnln(err)
			return
		}
	}
	if !skip_check_auth {
		if err = req.CheckAuth(self.node.PubKeyBytes); err != nil {
			err = status.Errorf(codes.Unauthenticated, "check auth failed, key: %x error: %s", req.Key, err)
			log.Warnln(err)
			return
		}
	}
	found, smallFile, storageIdx, subPath := self.querySubPath(req.Key)
	if !found {
		err = status.Errorf(codes.NotFound, "file not exist, key: %x", req.Key)
		log.Warnln(err)
		return
	}
	var res [][]byte
	if smallFile {
		storage := config.GetStorage(storageIdx)
		data, er := storage.SmallFileDb.Get(req.Key, nil)
		if er != nil {
			err = status.Errorf(codes.Internal, "read small file error, blockKey: %x error: %s", req.Key, er)
			log.Warnln(err)
			return
		}
		res, err = getFragmentFromByteSlice(req.Key, data, req.Positions, req.Size)
	} else {
		path := config.GetStoragePath(storageIdx, subPath)
		res, err = getFragmentFromFile(req.Key, path, req.Positions, req.Size)
	}
	if err != nil {
		return
	}
	return &pb.GetFragmentResp{Data: res}, nil
}

func getFragmentFromByteSlice(key []byte, data []byte, positions []byte, size uint32) (fragment [][]byte, err error) {
	res := make([][]byte, 0, len(positions))
	fileSize := uint32(len(data))
	for _, posPercent := range positions {
		pos := uint32(posPercent) * fileSize / 100
		if pos+size > fileSize {
			err = status.Errorf(codes.InvalidArgument, "file position %d%%+%d out of bounds, key: %x", posPercent, size, key)
			log.Warnln(err)
			return
		}
		res = append(res, data[pos:pos+size])
	}
	return res, nil
}

func getFragmentFromFile(key []byte, path string, positions []byte, size uint32) (fragment [][]byte, err error) {
	res := make([][]byte, 0, len(positions))
	fileInfo, er := os.Stat(path)
	if er != nil {
		err = status.Errorf(codes.Internal, "stat file failed, key: %x error: %s", key, er)
		log.Warnln(err)
		return
	}
	fileSize := uint64(fileInfo.Size())
	file, er := os.Open(path)
	if er != nil {
		err = status.Errorf(codes.Internal, "open file failed, key: %x error: %s", key, er)
		log.Warnln(err)
		return
	}
	defer file.Close()
	for _, posPercent := range positions {
		pos := int64(posPercent) * int64(fileSize) / 100
		if pos+int64(size) > int64(fileSize) {
			err = status.Errorf(codes.InvalidArgument, "file position %d%%+%d out of bounds, key: %x", posPercent, size, key)
			log.Warnln(err)
			return
		}
		buf := make([]byte, size)
		_, err = file.Seek(pos, 0)
		if err != nil {
			err = status.Errorf(codes.Internal, "seek file %s to %d failed, key: %x error: %s", path, pos, key, err)
			log.Warnln(err)
			return
		}
		if _, err = file.Read(buf); err != nil {
			err = status.Errorf(codes.Internal, "read file failed, key: %x error: %s", key, err)
			log.Warnln(err)
			return
		}
		res = append(res, buf)
	}
	return res, nil
}

func (self *ProviderService) saveFile(key []byte, fileSize uint64, tmpFilePath string, storage *config.Storage) error {
	fullPath, subPath, err := storage.GetPathPair(key)
	if err != nil {
		return err
	}
	err = os.Rename(tmpFilePath, fullPath)
	if err != nil {
		return err
	}
	path := []byte(subPath)
	pathSlice := make([]byte, len(path)+1)
	copy(pathSlice[1:], path)
	pathSlice[0] = storage.Index
	return self.providerDb.Put(key, pathSlice, nil)
}

func (self *ProviderService) queryByKey(key []byte) []byte {
	val, err := self.providerDb.Get(key, nil)
	if err == nil {
		return val
	} else if err != leveldb_errors.ErrNotFound {
		log.Errorf("get %x from provider db error: %s", key, err)
	}
	return nil
}

func (self *ProviderService) querySubPath(key []byte) (found bool, smallFile bool, storageIdx byte, subPath string) {
	bytes := self.queryByKey(key)
	if len(bytes) == 0 {
		return false, false, 0, ""
	} else if len(bytes) == 1 {
		return true, true, bytes[0], ""
	} else {
		return true, false, bytes[0], string(bytes[1:])
	}
}

func (self *ProviderService) CheckAvailable(ctx context.Context, req *pb.CheckAvailableReq) (resp *pb.CheckAvailableResp, err error) {
	if !skip_check_auth {
		if len(req.NodeIdHash) > 0 && !bytes.Equal(req.NodeIdHash, self.nodeIdHash) {
			return nil, status.Errorf(codes.FailedPrecondition, "not the target node")
		}
		if err = req.CheckAuth(self.node.PubKeyBytes); err != nil {
			err = status.Errorf(codes.Unauthenticated, "check auth failed,  error: %s", err)
			log.Warnln(err)
			return
		}
	}
	total, max := config.AvailableVolume()
	return &pb.CheckAvailableResp{Total: total, MaxFileSize: max, Version: 1}, nil
}

func (self *ProviderService) initTaskProcessor(taskServer string, private bool) {
	self.taskGetting = gosync.NewMutex()
	self.closeSignal = make(chan bool, 1)
	self.replicateChan = make(chan *ttpb.Task, 320)
	self.sendChan = make(chan *ttpb.Task, 320)
	self.removeAndProveChan = make(chan *ttpb.Task, 320)
	replicateThread := 2
	sendTread := 1
	processRemoveAndProve := 1
	if private {
		replicateThread, sendTread, processRemoveAndProve = 8, 2, 2
	}
	for i := 0; i < processRemoveAndProve; i++ {
		go self.processRemoveAndProve()
	}
	for i := 0; i < sendTread; i++ {
		go self.processSend()
	}
	for i := 0; i < replicateThread; i++ {
		go self.processReplicate()
	}
	self.waitClose.Add(replicateThread + sendTread + processRemoveAndProve)
	var err error
	self.taskConnection, err = grpc.Dial(taskServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial taskServer %s failed: %s\n", taskServer, err.Error())
		os.Exit(60)
	}
	self.ptsc = ttpb.NewProviderTaskServiceClient(self.taskConnection)
}

func (self *ProviderService) CloseTaskProcessor() {
	defer self.taskConnection.Close()
	self.closeSignal <- true
	self.waitClose.Wait()
}

func (self *ProviderService) processReplicate() {
	for {
		select {
		case <-self.closeSignal:
			self.waitClose.Done()
			return
		case ta := <-self.replicateChan:
			if len(ta.OppositeId) == 0 {
				fmt.Printf("Task [%x] info error, REPLICATE task haven't opposite id\n", ta.Id)
				continue
			}
			resp, err := task_client.GetOppositeInfo(self.ptsc, ta.Id)
			if err != nil {
				fmt.Printf("Get task [%x] opposite info failed: %s\n", ta.Id, err.Error())
				continue
			}
			if len(resp.Info) == 0 {
				fmt.Printf("Get task [%x] opposite info error, REPLICATE task haven't opposite info\n", ta.Id)
				continue
			}
			var remark string
			success := true
			if err = self.taskReplicate(ta.FileHash, ta.FileSize, ta.BlockHash, ta.BlockSize, resp.Timestamp, resp.Info); err != nil {
				remark = err.Error()
				success = false
				fmt.Printf("taskReplicate failed, blockKey: %x, error: %s\n", ta.BlockHash, remark)
			}
			if err = task_client.FinishTask(self.ptsc, ta.Id, uint64(time.Now().Unix()), success, remark); err != nil {
				fmt.Printf("Finish replicate task [%x] failed: %s\n", ta.Id, err.Error())
			}
		}
	}
}

func (self *ProviderService) processSend() {
	for {
		select {
		case <-self.closeSignal:
			self.waitClose.Done()
			return
		case ta := <-self.sendChan:
			if len(ta.OppositeId) != 1 {
				fmt.Printf("Task [%x] info error, SEND task haven't single opposite id\n", ta.Id)
				continue
			}
			resp, err := task_client.GetOppositeInfo(self.ptsc, ta.Id)
			if err != nil {
				fmt.Printf("Get task [%x] opposite info failed: %s\n", ta.Id, err.Error())
				continue
			}
			if len(resp.Info) != 1 {
				fmt.Printf("Get task [%x] opposite info error, SEND task haven't single opposite info\n", ta.Id)
				continue
			}
			var remark string
			success := true
			if err = self.taskSend(ta.FileHash, ta.FileSize, ta.BlockHash, ta.BlockSize, resp.Timestamp, resp.Info[0]); err != nil {
				remark = err.Error()
				success = false
				fmt.Printf("taskSend failed, blockKey: %x, error: %s\n", ta.BlockHash, remark)
			}
			if err = task_client.FinishTask(self.ptsc, ta.Id, uint64(time.Now().Unix()), success, remark); err != nil {
				fmt.Printf("Finish send task [%x] failed: %s\n", ta.Id, err.Error())
			}
		}
	}
}

func (self *ProviderService) processRemoveAndProve() {
	for {
		select {
		case <-self.closeSignal:
			self.waitClose.Done()
			return
		case ta := <-self.removeAndProveChan:
			if ta.Type == ttpb.TaskType_REMOVE {
				var remark string
				success := true
				if err := self.taskRemove(ta.FileHash, ta.FileSize, ta.BlockHash, ta.BlockSize); err != nil {
					remark = err.Error()
					success = false
					fmt.Printf("taskRemove failed, blockKey: %x, error: %s\n", ta.BlockHash, remark)
				}
				if err := task_client.FinishTask(self.ptsc, ta.Id, uint64(time.Now().Unix()), success, remark); err != nil {
					fmt.Printf("Finish remove task [%x] failed: %s\n", ta.Id, err.Error())
				}
			} else if ta.Type == ttpb.TaskType_PROVE {
				proofId, chunkSize, chunkSeq, err := task_client.GetProveInfo(self.ptsc, ta.Id)
				if err != nil {
					fmt.Printf("Get task [%x] prove info failed: %s\n", ta.Id, err.Error())
					continue
				}
				if len(proofId) == 0 {
					fmt.Printf("none proof id, task id: %x\n", ta.Id)
					continue
				}
				if len(ta.ProofId) > 0 && !bytes.Equal(ta.ProofId, proofId) {
					fmt.Printf("task [%x] prove id not same\n", ta.Id)
					continue
				}
				result, err := self.taskProve(ta.BlockHash, ta.BlockSize, chunkSize, chunkSeq)
				var remark string
				if err != nil {
					remark = err.Error()
				}
				if err = task_client.FinishProve(self.ptsc, ta.Id, proofId, uint64(time.Now().Unix()), result, remark); err != nil {
					fmt.Printf("Finish prove task [%x] failed: %s\n", ta.Id, err.Error())
				}
			}
		}
	}
}

func (self *ProviderService) GetTask() {
	if self.taskGetting.TryLock() {
		defer self.taskGetting.UnLock()
	} else {
		return
	}
	taskList, err := task_client.TaskList(self.ptsc)
	if err != nil {
		fmt.Printf("Get task list failed: %s\n", err.Error())
		return
	}
	if len(taskList) == 0 {
		return
	}
	for _, ta := range taskList {
		switch ta.Type {
		case ttpb.TaskType_REMOVE:
			self.removeAndProveChan <- ta
		case ttpb.TaskType_PROVE:
			self.removeAndProveChan <- ta
		case ttpb.TaskType_SEND:
			self.sendChan <- ta
		case ttpb.TaskType_REPLICATE:
			self.replicateChan <- ta
		}
	}
}

func (self *ProviderService) taskRemove(fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64) (err error) {
	found, smallFile, storageIdx, subPath := self.querySubPath(blockHash)
	if !found {
		return
	}
	if err = self.providerDb.Delete(blockHash, nil); err != nil {
		return fmt.Errorf("delete from provider db failed, error: %s", err)
	}
	if smallFile {
		storage := config.GetStorage(storageIdx)
		if err = storage.SmallFileDb.Delete(blockHash, nil); err != nil {
			return fmt.Errorf("delete from small file db failed, error: %s", err)
		}
	} else {
		path := config.GetStoragePath(storageIdx, subPath)
		if err = os.Remove(path); err != nil {
			return fmt.Errorf("remove file failed, error: %s", err)
		}
	}
	return
}

func (self *ProviderService) taskProve(blockHash []byte, blockSize uint64, chunkSize uint32, chunkSeq map[uint32][]byte) (result []byte, err error) {
	found, smallFile, storageIdx, subPath := self.querySubPath(blockHash)
	if !found {
		return nil, fmt.Errorf("file not exist")
	}
	keys := make([]uint32, 0, len(chunkSeq))
	for k, _ := range chunkSeq {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	res := big.NewInt(0)
	if smallFile {
		storage := config.GetStorage(storageIdx)
		data, er := storage.SmallFileDb.Get(blockHash, nil)
		if er != nil {
			return nil, fmt.Errorf("read small file error, error: %s", er)
		}
		length := int64(len(data))
		for _, k := range keys {
			start := int64(k-1) * int64(chunkSize)
			if k < 1 || start >= length {
				return nil, fmt.Errorf("seq out of range: %d", k)
			}
			end := start + int64(chunkSize)
			if end > length {
				end = length
			}
			bm := new(big.Int)
			bm.SetBytes(data[start:end])
			bv := new(big.Int)
			bv.SetBytes(chunkSeq[k])
			bm.Mul(bm, bv)
			res.Add(res, bm)
		}
	} else {
		path := config.GetStoragePath(storageIdx, subPath)
		fileInfo, er := os.Stat(path)
		if er != nil {
			return nil, fmt.Errorf("stat file failed, error: %s", er)
		}
		fileSize := int64(fileInfo.Size())
		file, er := os.Open(path)
		if er != nil {
			return nil, fmt.Errorf("open file failed, error: %s", er)
		}
		defer file.Close()
		buf := make([]byte, chunkSize)
		for _, k := range keys {
			start := int64(k-1) * int64(chunkSize)
			if start >= fileSize {
				return nil, fmt.Errorf("seq out of range")
			}
			file.Seek(start, 0)
			bytesRead, er := file.Read(buf)
			if er != nil && er != io.EOF {
				return nil, fmt.Errorf("read file failed, error: %s", er)
			}
			if bytesRead > 0 {
				bm := new(big.Int)
				bm.SetBytes(buf[:bytesRead])
				bv := new(big.Int)
				bv.SetBytes(chunkSeq[k])
				bm.Mul(bm, bv)
				res.Add(res, bm)
			} else {
				return nil, fmt.Errorf("read file get nothing, start position: %d", start)
			}
		}
	}
	return res.Bytes(), nil
}

func (self *ProviderService) taskSend(fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64, timestamp uint64, oppositeInfo *ttpb.OppositeInfo) (err error) {
	found, smallFile, storageIdx, subPath := self.querySubPath(blockHash)
	if !found {
		return fmt.Errorf("file not exist")
	}
	providerAddr := fmt.Sprintf("%s:%d", oppositeInfo.Host, oppositeInfo.Port)
	conn, err := grpc.Dial(providerAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("RPC Dial taskServer %s failed: %s", providerAddr, err.Error())
	}
	defer conn.Close()
	psc := pb.NewProviderServiceClient(conn)
	if smallFile {
		storage := config.GetStorage(storageIdx)
		data, er := storage.SmallFileDb.Get(blockHash, nil)
		if er != nil {
			return fmt.Errorf("read small file error, error: %s", er)
		}
		if len(data) != int(blockSize) {
			return fmt.Errorf("file length not same")
		}
		if !bytes.Equal(util_hash.Sha1(data), blockHash) {
			return fmt.Errorf("hash verify failed")
		}
		return provider_client.StoreSmall(psc, data, oppositeInfo.Auth, timestamp, oppositeInfo.Ticket, fileHash, fileSize, blockHash, blockSize)
	} else {
		path := config.GetStoragePath(storageIdx, subPath)
		fileInfo, er := os.Stat(path)
		if er != nil {
			return fmt.Errorf("stat file failed, error: %s", er)
		}
		if fileInfo.Size() != int64(blockSize) {
			return fmt.Errorf("file length not same")
		}
		hash, err := util_hash.Sha1File(path)
		if err != nil {
			return fmt.Errorf("sha1 sum file error: %s", err)
		}
		if !bytes.Equal(hash, blockHash) {
			return fmt.Errorf("hash verify failed")
		}
		return provider_client.Store(psc, path, oppositeInfo.Auth, timestamp, oppositeInfo.Ticket, fileHash, fileSize, blockHash, blockSize)
	}
}

func (self *ProviderService) taskReplicate(fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64, timestamp uint64, oppositeInfo []*ttpb.OppositeInfo) (err error) {
	found, smallFile, storageIdx, subPath := self.querySubPath(blockHash)
	var storage *config.Storage
	if found {
		if smallFile {
			storage := config.GetStorage(storageIdx)
			data, er := storage.SmallFileDb.Get(blockHash, nil)
			if er != nil {
				return fmt.Errorf("read small file error, error: %s", er)
			}
			if len(data) == int(blockSize) && bytes.Equal(util_hash.Sha1(data), blockHash) {
				return nil
			}
		} else {
			path := config.GetStoragePath(storageIdx, subPath)
			fileInfo, er := os.Stat(path)
			if er != nil {
				return fmt.Errorf("stat file failed, error: %s", er)
			}
			if fileInfo.Size() == int64(blockSize) {
				hash, err := util_hash.Sha1File(path)
				if err != nil {
					return fmt.Errorf("sha1 sum file error: %s", err)
				}
				if bytes.Equal(hash, blockHash) {
					return nil
				}
			}
		}
		storage = config.GetStorage(storageIdx)
	} else {
		storage = config.GetWriteStorage(blockSize)
	}
	smallFile = (blockSize < small_file_limit)
	providers := testPing(oppositeInfo)
	errs := make([]error, 0, 8)
	for _, pro := range providers {
		providerAddr := fmt.Sprintf("%s:%d", pro.Host, pro.Port)
		conn, err := grpc.Dial(providerAddr, grpc.WithInsecure())
		if err != nil {
			errs = append(errs, fmt.Errorf("RPC Dial provider %s failed: %s", providerAddr, err.Error()))
			continue
		}
		defer conn.Close()
		psc := pb.NewProviderServiceClient(conn)
		if smallFile {
			data, err := provider_client.RetrieveSmall(psc, pro.Auth, timestamp, pro.Ticket, fileHash, fileSize, blockHash, blockSize)
			if err != nil {
				errs = append(errs, fmt.Errorf("retrieve small file failed, provider id: %s, block key: %x error: %s", pro.NodeId, blockHash, err))
				continue
			}
			if len(data) != int(blockSize) {
				errs = append(errs, fmt.Errorf("file length wrong, provider id: %s, block key: %x", pro.NodeId, blockHash))
				continue
			}
			if !bytes.Equal(blockHash, util_hash.Sha1(data)) {
				errs = append(errs, fmt.Errorf("check data hash failed, provider id: %s, block key: %x", pro.NodeId, blockHash))
				continue
			}
			if err = storage.SmallFileDb.Put(blockHash, data, nil); err != nil {
				return fmt.Errorf("save to small file db failed, error: %s", err)
			}
			if err = self.providerDb.Put(blockHash, []byte{storage.Index}, nil); err != nil {
				return fmt.Errorf("save to provider db failed, error: %s", err)
			}
		} else {
			tempFilePath := storage.TempFilePath(blockHash)
			file, err := os.OpenFile(
				tempFilePath,
				os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
				0600)
			if err != nil {
				return fmt.Errorf("open temp write file failed, error: %s", err)
			}
			defer file.Close()
			if err = provider_client.Retrieve(psc, tempFilePath, pro.Auth, timestamp, pro.Ticket, fileHash, fileSize, blockHash, blockSize); err != nil {
				errs = append(errs, fmt.Errorf("retrieve file failed, provider id: %s, block key: %x error: %s", pro.NodeId, blockHash, err))
				continue
			}
			fileInfo, err := os.Stat(tempFilePath)
			if err != nil {
				return fmt.Errorf("stat temp file failed, error: %s", err)
			}
			if blockSize != uint64(fileInfo.Size()) {
				errs = append(errs, fmt.Errorf("file length wrong, provider id: %s, block key: %x", pro.NodeId, blockHash))
				continue
			}
			hash, err := util_hash.Sha1File(tempFilePath)
			if err != nil {
				return fmt.Errorf("sha1 sum file %s failed, error: %s", tempFilePath, err)
			}
			if !bytes.Equal(hash, blockHash) {
				errs = append(errs, fmt.Errorf("check data hash failed, provider id: %s, block key: %x", pro.NodeId, blockHash))
				continue
			}
			if err = file.Close(); err != nil {
				return fmt.Errorf("close temp file failed, tempFilePath: %s error: %s", tempFilePath, err)
			}
			if found {
				path := config.GetStoragePath(storageIdx, subPath)
				if err = os.Remove(path); err != nil {
					return fmt.Errorf("remove old file failed, path: %s error: %s", path, err)
				}
			}
			if err := self.saveFile(blockHash, blockSize, tempFilePath, storage); err != nil {
				return fmt.Errorf("save file failed, tempFilePath: %s error: %s", tempFilePath, err)
			}
		}
		return nil
	}
	var errMsg string
	for _, err := range errs {
		errMsg += err.Error()
		if len(errMsg) > 255 {
			break
		}
		errMsg += "\n"
	}
	if len(errMsg) == 0 {
		errMsg = "all opposite provider ping timeout"
	}
	return fmt.Errorf(errMsg)
}

type OppositeProvider struct {
	*ttpb.OppositeInfo
	lantency int64
}

func testPing(oppositeInfo []*ttpb.OppositeInfo) []*OppositeProvider {
	result := make([]*OppositeProvider, 0, len(oppositeInfo))
	timeout := 5
	for _, oi := range oppositeInfo {
		nodeIdHash, latency, err := provider_client.Ping(oi.Host, oi.Port, timeout)
		if err != nil {
			fmt.Printf("ping provider %s:%d failed: %s\n", oi.Host, oi.Port, err)
			continue
		}
		nodeId, err := base64.StdEncoding.DecodeString(oi.NodeId)
		if err != nil {
			fmt.Printf("decode provider id %x failed: %s\n", oi.NodeId, err)
			continue
		}
		if len(nodeIdHash) > 0 && !bytes.Equal(util_hash.Sha1(nodeId), nodeIdHash) {
			fmt.Printf("provider %s:%d node id %x not same\n", oi.Host, oi.Port, oi.NodeId)
			continue
		}
		result = append(result, &OppositeProvider{OppositeInfo: oi, lantency: latency})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].lantency < result[j].lantency })
	return result
}
