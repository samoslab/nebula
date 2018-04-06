package impl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spolabs/nebula/provider/config"
	"github.com/spolabs/nebula/provider/node"
	pb "github.com/spolabs/nebula/provider/pb"
	util_bytes "github.com/spolabs/nebula/util/bytes"
	util_file "github.com/spolabs/nebula/util/file"
	util_hash "github.com/spolabs/nebula/util/hash"
	util_num "github.com/spolabs/nebula/util/num"
)

const stream_data_size = 32 * 1024
const sys_folder = "nebula"
const tmp_folder = "temp"
const filename_suffix = ".blk"
const slash = "/"

var skip_check_auth = false

const sep = string(os.PathSeparator)
const timestamp_expired = 60

type ProviderService struct {
	Node       *node.Node
	ProviderDb *bolt.DB
}

func initAllStorage(conf *config.ProviderConfig) {
	initStorage(conf.MainStoragePath, 0)
	for _, info := range conf.ExtraStorage {
		initStorage(info.Path, int(info.Index+1))
	}
}

func initStorage(path string, index int) {
	exist, fileInfo := util_file.ExistsWithInfo(path)
	if !exist {
		log.Fatalf("storage path not exists: %s", path)
	}
	if fileInfo != nil && fileInfo.Mode().IsRegular() {
		log.Fatalf("storage path is a file: %s", path)
	}
	p := path + sep + sys_folder
	exist, fileInfo = util_file.ExistsWithInfo(p)
	if !exist {
		err := os.Mkdir(p, 0700)
		if err != nil {
			log.Fatalf("mkdir sys folder: %s failed: %s", p, err)
		}
		newFile, err := os.Create(p + sep + "storage-" + strconv.Itoa(index) + ".nebula")
		if err != nil {
			log.Fatalf("create storage index file failed: %s", err)
		}
		newFile, err = os.Create(p + sep + "do_not_delete.txt")
		if err != nil {
			log.Fatalf("create notice file failed: %s", err)
		}
		newFile.Close()
	}
	if fileInfo != nil && fileInfo.Mode().IsRegular() {
		log.Fatalf("storage sys folder path is a file: %s", p)
	}
}

func NewProviderService() *ProviderService {
	conf := config.GetProviderConfig()
	if os.Getenv("NEBULA_TEST_MODE") == "1" {
		skip_check_auth = true
	}
	initAllStorage(conf)
	ps := &ProviderService{}
	ps.Node = node.LoadFormConfig()
	dbPath := conf.MainStoragePath + sep + sys_folder + sep + "provider.db"
	exists := util_file.Exists(dbPath)
	var err error
	ps.ProviderDb, err = bolt.Open(dbPath,
		0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatalf("open Provider DB failed:%s", err)
	}
	if !exists {
		ps.ProviderDb.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucket(path_bucket)
			return err
		})
		if err != nil {
			log.Fatalf("create bucket: %s", err)
		}
	}
	return ps
}

func (self *ProviderService) Close() {
	self.ProviderDb.Close()
}

func (self *ProviderService) checkAuth(method string, auth []byte, key []byte, fileSize uint64, timestamp uint64) error {
	if skip_check_auth {
		return nil
	}
	if len(key) < 4 {
		return errors.New("empty key")
	}
	nowTs := uint64(time.Now().Unix())
	if nowTs-timestamp > timestamp_expired || timestamp-nowTs > timestamp_expired {
		return errors.New("auth info expired")
	}
	hash := hmac.New(sha256.New, self.Node.PubKeyBytes)
	hash.Write([]byte(method))
	hash.Write(key)
	hash.Write(util_bytes.FromUint64(fileSize))
	hash.Write(util_bytes.FromUint64(timestamp))
	if util_bytes.SameBytes(auth, hash.Sum(nil)) {
		return nil
	}
	return errors.New("auth verify failed")
}

func (self *ProviderService) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{}, nil
}

func (self *ProviderService) Store(stream pb.ProviderService_StoreServer) error {
	first := true
	var ticket string
	var file *os.File
	var tempFilePath string
	var storage *config.Storage
	var key []byte
	var fileSize uint64
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("RPC Recv failed: %s", err.Error())
			return errors.Wrapf(err, "failed unexpectadely while reading chunks from stream")
		}
		if first {
			first = false
			ticket = req.Ticket
			key = req.Key
			fileSize = req.FileSize
			if err = self.checkAuth("Store", req.Auth, req.Key, req.FileSize, req.Timestamp); err != nil {
				log.Warnf("check auth failed: %s", err.Error())
				return err
			}
			filename := hex.EncodeToString(req.Key)
			if self.queryByKey(req.Key) != nil {
				log.Warnf("hash point file exist, hash: %s", filename)
				return os.ErrExist
			}
			storage = config.GetStoragePath()
			tempFolder := storage.Path + sep + sys_folder + sep + tmp_folder
			_, err := os.Stat(tempFolder)
			if err != nil && os.IsNotExist(err) {
				if err = os.MkdirAll(tempFolder, 0700); err != nil {
					log.Errorf("mkdir failed, folder: %s", tempFolder)
					return err
				}
			}
			tempFilePath = tempFolder + sep + filename + filename_suffix
			file, err = os.OpenFile(
				tempFilePath,
				os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
				0600)
			if err != nil {
				fmt.Printf("open file failed: %s", err.Error())
				return errors.Wrapf(err, "open file failed")
			}
			defer file.Close()
		}
		if len(req.Data) == 0 {
			break
		}
		if _, err = file.Write(req.Data); err != nil {
			fmt.Printf("write file %d bytes failed : %s", len(req.Data), err.Error())
			return errors.Wrapf(err, "write file failed")
		}
	}
	hash, err := util_hash.Sha1File(tempFilePath)
	if err != nil {
		return err
	}
	if !util_bytes.SameBytes(hash, key) {
		return errors.New("hash verify failed")
	}
	if err := self.saveFile(key, fileSize, tempFilePath, storage); err != nil {
		log.Errorf("save file failed, tempFilePath: %s", tempFilePath)
		return err
	}
	// TODO send log with ticket
	fmt.Println(ticket)
	if err := stream.SendAndClose(&pb.StoreResp{Success: true}); err != nil {
		fmt.Printf("RPC SendAndClose failed: %s", err.Error())
		return errors.Wrapf(err, "SendAndClose failed")
	}
	return nil
}

func (self *ProviderService) Retrieve(req *pb.RetrieveReq, stream pb.ProviderService_RetrieveServer) error {
	if err := self.checkAuth("Retrieve", req.Auth, req.Key, req.FileSize, req.Timestamp); err != nil {
		return err
	}
	subPath, bigFile, storageIdx, position, fileSize := self.querySubPath(req.Key)
	if subPath == "" {
		return os.ErrNotExist
	}
	path, err := getAbsPathOfSubPath(subPath, int(storageIdx))
	if err != nil {
		return err
	}
	var hash []byte
	if bigFile {
		hash, err = util_hash.Sha1File(path)
	} else {
		hash, err = util_hash.Sha1FilePiece(path, position, fileSize)
	}
	if err != nil {
		return err
	}
	if !util_bytes.SameBytes(hash, req.Key) {
		return errors.New("hash verify failed")
	}
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("open file failed: %s", err.Error())
		return errors.Wrapf(err, "open file failed")
	}
	defer file.Close()
	if bigFile {
		if err = sendFileToStream(path, file, stream); err != nil {
			return err
		}
	} else {
		if err = sendFilePieceToStream(path, file, position, fileSize, stream); err != nil {
			return err
		}
	}
	// TODO send log with ticket
	fmt.Println(req.Ticket)
	return nil
}

func sendFilePieceToStream(path string, file *os.File, start uint32, size uint32, stream pb.ProviderService_RetrieveServer) error {
	newPosition, err := file.Seek(int64(start), 0)
	if err != nil {
		log.Errorf("Seek file: %s to %d failed: %s", path, start, err)
		return err
	}
	if newPosition != int64(start) {
		log.Errorf("Seek file: %s to %d failed", path, start)
		return errors.New("Seek file failed")
	}
	buf := make([]byte, stream_data_size)
	for {
		if size < stream_data_size {
			buf = make([]byte, size)
		}
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("read file: %s failed: %s", path, err.Error())
			return errors.Wrapf(err, "read file failed: "+path)
		}
		if bytesRead < len(buf) {
			return io.EOF
		}
		stream.Send(&pb.RetrieveResp{Data: buf})
		size -= uint32(len(buf))
		if size == 0 {
			break
		}
	}
	return nil
}
func sendFileToStream(path string, file *os.File, stream pb.ProviderService_RetrieveServer) error {
	buf := make([]byte, stream_data_size)
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("read file: %s failed: %s", path, err.Error())
			return errors.Wrapf(err, "read file failed: "+path)
		}
		if bytesRead > 0 {
			stream.Send(&pb.RetrieveResp{Data: buf[:bytesRead]})
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	return nil
}

func (self *ProviderService) GetFragment(ctx context.Context, req *pb.GetFragmentReq) (*pb.GetFragmentResp, error) {
	if len(req.Positions) == 0 || req.Size == 0 {
		return nil, errors.New("invalid req")
	}
	for i, b := range req.Positions {
		if b >= 100 {
			return nil, errors.New("posisiton out of bounds, Posistion " + strconv.Itoa(i) + ": " + strconv.Itoa(int(b)))
		}
	}
	if err := self.checkAuth("GetFragment", req.Auth, req.Key, uint64(req.Size), req.Timestamp); err != nil {
		log.Warnf("check auth failed: %s", err.Error())
		return nil, err
	}
	subPath, bigFile, storageIdx, position, size := self.querySubPath(req.Key)
	if subPath == "" {
		return nil, os.ErrNotExist
	}
	path, err := getAbsPathOfSubPath(subPath, int(storageIdx))
	if err != nil {
		return nil, err
	}
	var hash []byte
	if bigFile {
		hash, err = util_hash.Sha1File(path)
	} else {
		hash, err = util_hash.Sha1FilePiece(path, position, size)
	}
	if err != nil {
		return nil, err
	}
	if !util_bytes.SameBytes(hash, req.Key) {
		return nil, errors.New("hash verify failed")
	}
	var fileSize uint64
	var startPos int64
	if bigFile {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		fileSize = uint64(fileInfo.Size())
	} else {
		fileSize = uint64(size)
		startPos = int64(position)
	}
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("open file failed: %s", err.Error())
		return nil, errors.Wrapf(err, "open file failed")
	}
	defer file.Close()
	res := make([][]byte, 0, len(req.Positions))
	for _, posPercent := range req.Positions {
		pos := int64(posPercent) * int64(fileSize) / 100
		if pos+int64(req.Size) > int64(fileSize) {
			return nil, errors.New("file position out of bounds")
		}
		buf := make([]byte, req.Size)
		_, err := file.Seek(startPos+pos, 0) // TODO optimize: Seek offset
		if err != nil {
			log.Errorf("Seek file: %s to %d failed: %s", path, startPos+pos, err)
			return nil, err
		}
		_, err = file.Read(buf)
		if err != nil {
			return nil, err
		}
		res = append(res, buf)
	}
	return &pb.GetFragmentResp{Data: res}, nil
}

func getAbsPathOfSubPath(subPath string, storageIdx int) (string, error) {
	conf := config.GetProviderConfig()
	parent := conf.MainStoragePath
	if storageIdx != 0 {
		if storageIdx > len(conf.ExtraStorage) {
			return "", errors.New("storage index out of bounds")
		}
		parent = conf.ExtraStorage[strconv.Itoa(storageIdx)].Path
	}
	return parent + strings.Replace(subPath, slash, sep, -1), nil
}

func (self *ProviderService) saveFile(key []byte, fileSize uint64, tmpFilePath string, storage *config.Storage) error {
	filename := hex.EncodeToString(key)
	if fileSize > max_combine_file_size {
		val := util_bytes.ToUint32(key, len(key)-4)
		sub1 := util_num.FixLength(val&(config.ModFactor-1), 4)
		sub2 := util_num.FixLength((val>>config.ModFactorExp)&(config.ModFactor-1), 4)
		subPath := slash + sub1 + slash + sub2 + slash + filename + filename_suffix
		path := []byte(subPath)
		pathSlice := make([]byte, len(path)+1)
		pathSlice[0] = 128 | storage.Index
		for idx, v := range path {
			pathSlice[idx+1] = v
		}
		fullFolder := storage.Path + sep + sub1 + sep + sub2
		fullPath := fullFolder + sep + filename + filename_suffix
		if !util_file.Exists(fullFolder) {
			if err := os.MkdirAll(fullFolder, 0700); err != nil {
				return err
			}
		}
		err := os.Rename(tmpFilePath, fullPath)
		if err != nil {
			return err
		}
		err = self.savePath(key, pathSlice)
		if err != nil {
			return err
		}
		logPath(subPath, true, storage.Index, 0, 0)
	} else {
		return self.saveSmallFile(key, uint32(fileSize), tmpFilePath, storage)
	}
	return nil
}
func (self *ProviderService) saveSmallFile(key []byte, fileSize uint32, tmpFilePath string, storage *config.Storage) error {
	storage.SmallFileMutex.Lock()
	defer storage.SmallFileMutex.Unlock()
	subPath := storage.CurrCombineSubPath
	path := []byte(subPath)
	pathSlice := make([]byte, len(path)+9)
	pathSlice[0] = storage.Index
	currCombineSize := storage.CurrCombineSize()
	util_bytes.FillUint32(pathSlice, 1, currCombineSize)
	util_bytes.FillUint32(pathSlice, 5, fileSize)
	for idx, b := range path {
		pathSlice[9+idx] = b
	}
	var err error
	if err = util_file.ConcatFile(storage.CurrCombinePath, tmpFilePath, true); err != nil {
		return err
	}
	err = self.savePath(key, pathSlice)
	storage.NextCombinePath(currCombineSize, fileSize)
	if err != nil {
		return err
	}
	logPath(subPath, false, storage.Index, currCombineSize, fileSize)
	return nil
}

func logPath(subPath string, bigFile bool, storageIdx byte, position uint32, size uint32) {
	// path := config.GetProviderConfig().MainStoragePath + sep + sys_folder + sep + "log"
	// TODO
}

const max_combine_file_size = 1048576 //1M
var path_bucket = []byte("path")

func (self *ProviderService) queryByKey(key []byte) []byte {
	var res []byte
	self.ProviderDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(path_bucket)
		v := b.Get(key)
		res = v
		return nil
	})
	return res
}

func (self *ProviderService) querySubPath(key []byte) (subPath string, bigFile bool, storageIdx byte, position uint32, size uint32) {
	bytes := self.queryByKey(key)
	if bytes == nil {
		return "", false, 0, 0, 0
	}
	bigFile = bytes[0]&128 == 128
	storageIdx = bytes[0] & 127
	if bigFile {
		return string(bytes[1:]), bigFile, storageIdx, 0, 0
	}
	return string(bytes[9:]), bigFile, storageIdx, util_bytes.ToUint32(bytes, 1), util_bytes.ToUint32(bytes, 5)
}

func (self *ProviderService) savePath(key []byte, pathSlice []byte) error {
	var err error
	self.ProviderDb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(path_bucket)
		err = b.Put(key, pathSlice)
		return err
	})
	return err
}
