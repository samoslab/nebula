package config

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/samoslab/nebula/provider/disk"
	util_bytes "github.com/samoslab/nebula/util/bytes"
	util_file "github.com/samoslab/nebula/util/file"
	util_num "github.com/samoslab/nebula/util/num"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

const sys_folder = "nebula"
const tmp_folder = "temp"
const sep = string(os.PathSeparator)
const filename_suffix = ".blk"

const ModFactorExp = 13
const ModFactor = 1 << ModFactorExp
const slash = "/"

type Storage struct {
	Path        string
	Index       byte // 0 as Main Storage
	Volume      uint64
	SmallFileDb *leveldb.DB
}

func (self *Storage) initStorage() error {
	exist, fileInfo := util_file.ExistsWithInfo(self.Path)
	if !exist {
		return fmt.Errorf("storage path not exists: %s", self.Path)
	}
	if fileInfo != nil && fileInfo.Mode().IsRegular() {
		return fmt.Errorf("storage path is a file: %s", self.Path)
	}
	p := self.Path + sep + sys_folder
	exist, fileInfo = util_file.ExistsWithInfo(p)
	if !exist {
		err := os.Mkdir(p, 0700)
		if err != nil {
			return fmt.Errorf("mkdir sys folder: %s failed: %s", p, err)
		}
		newFile, err := os.Create(p + sep + "storage-" + strconv.FormatInt(int64(self.Index), 10) + ".nebula")
		if err != nil {
			return fmt.Errorf("create storage index file failed: %s", err)
		}
		newFile.Close()
		newFile, err = os.Create(p + sep + "do-not-change.txt")
		if err != nil {
			return fmt.Errorf("create notice file failed: %s", err)
		}
		newFile.Close()
	}
	if fileInfo != nil && fileInfo.Mode().IsRegular() {
		return fmt.Errorf("storage sys folder path is a file: %s", p)
	}
	var err error
	self.SmallFileDb, err = leveldb.OpenFile(p+sep+"small-file", nil)
	if err != nil {
		return fmt.Errorf("open small file db failed: %s", err)
	}
	tempPath := self.TempPath()
	if !util_file.Exists(tempPath) {
		if err = os.MkdirAll(tempPath, 0700); err != nil {
			return err
		}
	}
	return nil
}

func (self *Storage) TempPath() string {
	return self.Path + sep + sys_folder + sep + tmp_folder
}

func (self *Storage) cleanTemp() {
	tempPath := self.TempPath()
	files, err := ioutil.ReadDir(tempPath)
	if err != nil {
		log.Warnf("read files of storage %s temp path failed, error: %s", self.Path, err)
		return
	}
	currentTs := time.Now().Unix()
	for _, f := range files {
		if currentTs-f.ModTime().Unix() > 7200 {
			if err = os.Remove(tempPath + sep + f.Name()); err != nil {
				log.Warnf("delete obsolete file %s of storage %s temp path failed, error: %s", f.Name(), self.Path, err)
				continue
			}
		}
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func (self *Storage) TempFilePath(key []byte) string {
	return self.TempPath() + sep + hex.EncodeToString(key) + "-" + randStr(8) + filename_suffix
}

func (self *Storage) GetPathPair(key []byte) (fullPath string, subPath string, err error) {
	val := util_bytes.ToUint32(key, len(key)-4)
	sub1 := util_num.FixLength(val&(ModFactor-1), 4)
	sub2 := util_num.FixLength((val>>ModFactorExp)&(ModFactor-1), 4)
	filename := hex.EncodeToString(key)
	subPath = slash + sub1 + slash + sub2 + slash + filename + filename_suffix
	fullFolder := self.Path + sep + sub1 + sep + sub2
	if !util_file.Exists(fullFolder) {
		if err = os.MkdirAll(fullFolder, 0700); err != nil {
			return
		}
	}
	fullPath = fullFolder + sep + filename + filename_suffix
	return
}

func NewStorage(path string, index byte) (storage *Storage, err error) {
	path = cleanPath(path)
	_, free, err := disk.Space(path)
	if err != nil {
		return nil, err
	}
	s := &Storage{Path: path, Index: index, Volume: free}
	if err = s.initStorage(); err != nil {
		return nil, err
	}
	return s, nil
}

func cleanPath(path string) string {
	l := len(path)
	ls := len(sep)
	if path[l-ls:] == sep {
		return path[:l-ls]
	}
	return path
}

func GetWriteStorage(size uint64) *Storage {
	sl := storageSlice
	l := len(sl)
	if l == 0 {
		return nil
	}
	defer incrementStorageIdx()
	//104857600 = 100M
	if size < 104857600 {
		return sl[currentStorageIdx%uint64(l)]
	} else {
		first := int(currentStorageIdx % uint64(l))
		for i := first; i < first+l; i++ {
			s := sl[i%l]
			_, free, err := disk.Space(s.Path)
			if err != nil {
				log.Warnf("get storage %s free space error:%s", s.Path, err)
				continue
			} else if free > min_available_volume+size {
				return s
			} else {
				continue
			}
		}
		return nil
	}
}

func GetStoragePath(index byte, subPath string) string {
	return storageMap[strconv.FormatInt(int64(index), 10)].Path + strings.Replace(subPath, slash, sep, -1)
}

func GetStorage(index byte) *Storage {
	idx := strconv.FormatInt(int64(index), 10)
	return storageMap[idx]
}

func ProviderDbPath() string {
	return GetStorage(0).Path + sep + sys_folder + sep + "provider-db"
}

var storageSlice []*Storage
var storageMap map[string]*Storage

var currentStorageIdx uint64 = 0

func incrementStorageIdx() {
	atomic.AddUint64(&currentStorageIdx, 1)
}

func checkStorageAvailableSpace() {
	if storageSlice == nil {
		checkStorageAvailableSpaceOfConf()
	}
	sl := make([]*Storage, 0, len(storageSlice))
	for _, s := range storageSlice {
		_, free, err := disk.Space(s.Path)
		if err != nil {
			log.Warnf("get storage %s free space error:%s", s.Path, err)
			continue
		}
		if (s.Index == 0 && free <= 3*min_available_volume) || (s.Index != 0 && free <= min_available_volume) {
			log.Warnf("storage path %s available space less than 1GB", s.Path)
			continue
		}
		sl = append(sl, s)
	}
	storageSlice = sl
}

var min_available_volume uint64 = 1024 * 1024 * 1024

var checkStorageOfConf *sync.Mutex = &sync.Mutex{}
var checkStorageOfConfFirst = true

func checkStorageAvailableSpaceOfConf() {
	checkStorageOfConf.Lock()
	defer checkStorageOfConf.Unlock()
	if storageMap == nil {
		storageMap = make(map[string]*Storage, 1+len(providerConfig.ExtraStorage))
	}
	sl := make([]*Storage, 0, 1+len(providerConfig.ExtraStorage))
	var ok bool
	var s *Storage
	var err error
	if s, ok = storageMap["0"]; !ok {
		s, err = NewStorage(providerConfig.MainStoragePath, 0)
		if err != nil {
			log.Fatalf("main storage error: %s", err)
		}
		storageMap["0"] = s
	}
	s.cleanTemp()
	if s.Volume > 3*min_available_volume {
		sl = append(sl, s)
	} else {
		log.Errorf("main storage available space less than 1GB")
	}
	if len(providerConfig.ExtraStorage) > 0 {
		for _, v := range providerConfig.ExtraStorage {
			idx := strconv.FormatInt(int64(v.Index), 10)
			if s, ok = storageMap[idx]; !ok {
				s, err = NewStorage(v.Path, v.Index)
				if err != nil {
					if checkStorageOfConfFirst {
						log.Fatalf("extra storage %s init error: %s", v.Path, err)
					} else {
						log.Warnf("extra storage %s init error: %s", v.Path, err)
					}
					continue
				}
				storageMap[idx] = s
			}
			if s.Volume > min_available_volume {
				sl = append(sl, s)
			} else {
				log.Warnf("extra storage %s available space less than 1GB", v.Path)
			}
		}
	}
	storageSlice = sl
	checkStorageOfConfFirst = false
}

func stopStorage() {
	if storageMap != nil {
		for _, v := range storageMap {
			v.SmallFileDb.Close()
		}
	}
}
