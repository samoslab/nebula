package server

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"golang.org/x/net/context"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spolabs/nebula/provider/config"
	"github.com/spolabs/nebula/provider/node"
	pb "github.com/spolabs/nebula/provider/pb"
)

const stream_data_size = 32 * 1024
const sys_folder = "nebula"
const tmp_folder = "temp"
const filename_suffix = ".blk"

const sep = string(os.PathSeparator)
const timestamp_expired = 60

type ProviderServer struct {
	Node       *node.Node
	ProviderDb *bolt.DB
}

func initAllStorage(conf *config.ProviderConfig) {
	initStorage(conf.MainStoragePath, 0)
	for k, i := range conf.ExtraStorage {
		initStorage(k, int(i+1))
	}
}

func initStorage(path string, index int) {
	fileInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		p := path + sep + sys_folder
		err = os.Mkdir(p, 0700)
		if err != nil {
			log.Fatalf("mkdir sys folder failed:%s", err)
		}
		newFile, err := os.Create(p + sep + "storage-" + strconv.Itoa(index) + ".nebula")
		if err != nil {
			log.Fatalf("create storage index file failed:%s", err)
		}
		newFile, err = os.Create(p + sep + "do_not_delete.txt")
		if err != nil {
			log.Fatalf("create notice file failed:%s", err)
		}
		newFile.Close()
	} else if fileInfo.Mode().IsRegular() {
		log.Fatalf("%s is regular file", path)
	}
}

func NewProviderServer() *ProviderServer {
	conf := config.GetProviderConfig()
	initAllStorage(conf)
	ps := &ProviderServer{}
	ps.Node = node.LoadFormConfig()
	var err error
	ps.ProviderDb, err = bolt.Open(conf.MainStoragePath+sep+sys_folder+sep+"provider.db",
		0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}
	return ps
}

func (self *ProviderServer) Close() {
	self.ProviderDb.Close()
}

func (self *ProviderServer) checkAuth(method string, auth []byte, key []byte, fileSize uint64, timestamp uint64) error {
	if len(key) < 4 {
		return errors.New("empty key")
	}
	nowTs := uint64(time.Now().Unix())
	if nowTs-timestamp > timestamp_expired || timestamp-nowTs > timestamp_expired {
		return errors.New("auth info expired")
	}
	// TODO check auth
	return nil
}

func (self *ProviderServer) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{}, nil
}

func (self *ProviderServer) Store(stream pb.ProviderService_StoreServer) error {
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

func (self *ProviderServer) Retrieve(req *pb.RetrieveReq, stream pb.ProviderService_RetrieveServer) error {
	if err := self.checkAuth("Retrieve", req.Auth, req.Key, req.FileSize, req.Timestamp); err != nil {
		return err
	}
	path, bigFile, position, fileSize := self.queryPath(req.Key)
	if path == "" {
		return os.ErrNotExist
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

func (self *ProviderServer) GetFragment(ctx context.Context, req *pb.GetFragmentReq) (*pb.GetFragmentResp, error) {
	if err := self.checkAuth("GetFragment", req.Auth, req.Key, uint64(req.Size), req.Timestamp); err != nil {
		log.Warnf("check auth failed: %s", err.Error())
		return nil, err
	}
	path, bigFile, position, fileSize := self.queryPath(req.Key)
	if path == "" {
		return nil, os.ErrNotExist
	}
	file, err := os.Open(path)
	if err != nil {
		log.Errorf("open file failed: %s", err.Error())
		return nil, errors.Wrapf(err, "open file failed")
	}
	defer file.Close()
	if bigFile {
		//TODO
	} else {
		newPosition, err := file.Seek(int64(position), 0)
		if err != nil {
			log.Errorf("Seek file: %s to %d failed: %s", path, position, err)
			return nil, err
		}
		if newPosition != int64(position) {
			log.Errorf("Seek file: %s to %d failed", path, position)
			return nil, errors.New("Seek file failed")
		}
		fmt.Println(fileSize)
		// TODO
	}
	return nil, nil
}

func (self *ProviderServer) saveFile(key []byte, fileSize uint64, tmpFilePath string, storage *config.Storage) error {
	filename := hex.EncodeToString(key)
	if fileSize > max_combine_file_size {
		le := len(key)
		var val uint32
		val = uint32(key[le-1]) + 256*uint32(key[le-2]) + 256*256*uint32(key[le-3]) + 256*256*256*uint32(key[le-4])
		sub1 := config.FixLength(val%config.ModFactor, 4)
		sub2 := config.FixLength((val/config.ModFactor)%config.ModFactor, 4)
		path := []byte("/" + sub1 + "/" + sub2 + "/" + filename + filename_suffix)
		pathSlice := make([]byte, 0, len(path)+1)
		pathSlice[0] = 128 | storage.Index
		for v, idx := range path {
			pathSlice[idx+1] = byte(v)
		}
		fullPath := storage.Path + sep + sub1 + sep + sub2 + sep + filename + filename_suffix
		_, err := os.Stat(fullPath)
		if err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(fullPath, 0700); err != nil {
				return err
			}
		}
		err = os.Rename(tmpFilePath, fullPath)
		if err != nil {
			return err
		}
		err = self.savePath(key, pathSlice)
		if err != nil {
			return err
		}
	} else {
		return self.saveSmallFile(key, fileSize, tmpFilePath, storage)
	}
	return nil
}
func (self *ProviderServer) saveSmallFile(key []byte, fileSize uint64, tmpFilePath string, storage *config.Storage) error {
	storage.SmallFileMutex.Lock()
	defer storage.SmallFileMutex.Unlock()
	subPath := []byte(storage.CurrCombineSubPath)
	pathSlice := make([]byte, 0, len(subPath)+9)
	pathSlice[0] = storage.Index
	currCombineSize := storage.CurrCombineSize()
	fillByteSlice(pathSlice, 1, currCombineSize)
	fillByteSlice(pathSlice, 5, uint32(fileSize))
	for b, idx := range subPath {
		pathSlice[9+idx] = byte(b)
	}
	var err error
	if err = concatFile(storage.CurrCombinePath, tmpFilePath); err != nil {
		return err
	}
	err = self.savePath(key, pathSlice)
	storage.NextCombinePath(currCombineSize, uint32(fileSize))
	if err != nil {
		return err
	}
	return nil
}

func fillByteSlice(bytes []byte, startIdx int, val uint32) {
	// TODO optimize
	for i := 3; i >= 0; i-- {
		bytes[startIdx+i] = byte(val % 256)
		val = val / 256
		if val == 0 {
			break
		}
	}
}

func concatFile(writeFile string, readFile string) error {
	fi, err := os.Open(readFile)
	if err != nil {
		return err
	}
	defer fi.Close()
	fo, err := os.OpenFile(writeFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer fo.Close()
	buf := make([]byte, 8192)
	for {
		// read a chunk
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			return err
		}
	}
	err = os.Remove(readFile)
	if err != nil {
		log.Errorf("Remove file: %s failed:%s", readFile, err)
	}
	return nil
}

const max_combine_file_size = 1048576 //1M
var path_bucket = []byte("path")

func (self *ProviderServer) queryByKey(key []byte) []byte {
	var res []byte
	self.ProviderDb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(path_bucket)
		v := b.Get(key)
		res = v
		return nil
	})
	return res
}

func (self *ProviderServer) queryPath(key []byte) (string, bool, uint32, uint32) {
	bytes := self.queryByKey(key)
	if bytes == nil {
		return "", false, 0, 0
	}
	bigFile := bytes[0]&128 == 128
	if bigFile {
		return string(bytes[1:]), bigFile, 0, 0
	}
	return string(bytes[9:]), bigFile, byteSliceToUint32(bytes, 1), byteSliceToUint32(bytes, 5)
}

func byteSliceToUint32(b []byte, startIdx int) uint32 {
	return uint32(b[startIdx+3]) | uint32(b[startIdx+2])<<8 | uint32(b[startIdx+1])<<16 | uint32(b[startIdx])<<24
}

func (self *ProviderServer) savePath(key []byte, pathSlice []byte) error {
	var err error
	self.ProviderDb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(path_bucket)
		err = b.Put(key, pathSlice)
		return err
	})
	return err
}
