package impl

import (
	"bytes"
	"io"
	"os"
	"time"

	"golang.org/x/net/context"

	client "github.com/samoslab/nebula/provider/collector_client"
	"github.com/samoslab/nebula/provider/config"
	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/provider/pb"
	tcppb "github.com/samoslab/nebula/tracker/collector/provider/pb"
	util_hash "github.com/samoslab/nebula/util/hash"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	leveldb_errors "github.com/syndtr/goleveldb/leveldb/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const stream_data_size = 32 * 1024
const small_file_limit = 512 * 1024

var skip_check_auth = false

type ProviderService struct {
	node       *node.Node
	providerDb *leveldb.DB
}

func NewProviderService() *ProviderService {
	if os.Getenv("NEBULA_TEST_MODE") == "1" {
		skip_check_auth = true
	}
	ps := &ProviderService{}
	ps.node = node.LoadFormConfig()
	var err error
	ps.providerDb, err = leveldb.OpenFile(config.ProviderDbPath(), nil)
	if err != nil {
		log.Fatalf("open Provider DB failed:%s", err)
	}
	return ps
}

func (self *ProviderService) Close() {
	self.providerDb.Close()
}

func (self *ProviderService) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{}, nil
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
