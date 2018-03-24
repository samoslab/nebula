package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"

	"github.com/koding/multiconfig"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
	util_file "github.com/spolabs/nebula/util/file"
	util_num "github.com/spolabs/nebula/util/num"
)

var NoConfErr = errors.New("not found config file")

var ConfVerifyErr = errors.New("verify config file failed")

type ExtraStorageInfo struct {
	Path   string
	Volume uint64
	Index  byte // 1-based
}

type ProviderConfig struct {
	NodeId            string
	WalletAddress     string
	BillEmail         string
	PublicKey         string
	PrivateKey        string
	Availability      float64
	MainStoragePath   string
	MainStorageVolume uint64
	UpBandwidth       uint64
	DownBandwidth     uint64
	EncryptKey        map[string]string           // key: version, eg: 0, 1, 2
	ExtraStorage      map[string]ExtraStorageInfo `json:",omitempty"` //key:storage index, 1-based eg: 1, 2, 3
}

var providerConfig *ProviderConfig

const config_filename = "config.json"

var configFilePath string
var configFileModTs int64

var cronRunner *cron.Cron

func LoadConfig(configDir string) error {
	configFilePath = configDir + string(os.PathSeparator) + config_filename
	if !util_file.Exists(configFilePath) {
		return NoConfErr
	}
	pc, err := readConfig()
	if err != nil {
		return err
	}
	if err = verifyConfig(pc); err != nil {
		return ConfVerifyErr
	}
	providerConfig = pc
	checkAllStorageAvailableSpace()
	return nil
}

func verifyConfig(pc *ProviderConfig) error {
	//TODO
	// TODO verify ExtraStorage map key equals Index
	return nil
}

func StartAutoCheck() {
	cronRunner := cron.New()
	cronRunner.AddFunc("0,15,30,45 * * * * *", checkAndReload)
	cronRunner.AddFunc("7 */3 * * * *", checkAllStorageAvailableSpace)
	cronRunner.Start()
}

func StopAutoCheck() {
	cronRunner.Stop()
}

func checkAndReload() {
	modTs, err := getConfigFileModTime()
	if err != nil {
		log.Errorf("getConfigFileModTime Error: %s", err)
		return
	}
	if modTs != configFileModTs {
		pc, err := readConfig()
		if err != nil {
			log.Errorf("readConfig Error: %s", err)
		} else if verifyConfig(pc) == nil {
			providerConfig = pc
		}

	}
}

func getConfigFileModTime() (int64, error) {
	fileInfo, err := os.Stat(configFilePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}

func readConfig() (*ProviderConfig, error) {
	m := multiconfig.NewWithPath(configFilePath) // supports TOML, JSON and YAML
	pc := new(ProviderConfig)
	err := m.Load(pc) // Check for error
	if err != nil {
		return nil, err
	}
	m.MustLoad(pc) // Panic's if there is any error
	configFileModTs, err = getConfigFileModTime()
	if err != nil {
		return nil, err
	}
	return pc, nil
}

func GetProviderConfig() *ProviderConfig {
	return providerConfig
}

func CreateProviderConfig(configDir string, pc *ProviderConfig) {
	if !util_file.Exists(configDir) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			log.Fatalf("mkdir config folder %s failed:%s", configDir, err)
		}
	}
	path := configDir + string(os.PathSeparator) + config_filename
	if util_file.Exists(path) {
		log.Fatalf("config file is adready exsits:%s", configDir)
	}
	saveProviderConfig(path, pc)
}

func saveProviderConfig(configPath string, pc *ProviderConfig) error {
	b, err := json.Marshal(pc)
	if err != nil {
		return err
	}
	var out bytes.Buffer
	if err = json.Indent(&out, b, "", "  "); err != nil {
		return err
	}
	if err = ioutil.WriteFile(configPath, out.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func SaveProviderConfig() {
	if err := saveProviderConfig(configFilePath, providerConfig); err != nil {
		log.Errorf("save config file err: %s", err)
	}
}

type Storage struct {
	Path               string
	Index              byte // 0 as Main Storage
	Volume             uint64
	SmallFileMutex     sync.Mutex
	CurrCombinePath    string
	CurrCombineSubPath string
	CurrCombineIdx     uint32
}

var storageSlice []*Storage

const max_combine_file_size = 1024 * 1024 * 1024 // cannot more than max uint32 value: 4294967295
const combine_filename_suffix = ".blks"
const combine_folder = "combine"
const ModFactorExp = 13
const ModFactor = 1 << ModFactorExp
const slash = "/"

func (self *Storage) getFolderAndFileName(idx uint32) (string, string) {
	return util_num.FixLength(idx&(ModFactor-1), 4), util_num.FixLength(idx, 8) + combine_filename_suffix
}
func (self *Storage) findOrTouchCombinePath() {
	self.findOrTouchCombinePathFormIdx(0)
}
func (self *Storage) findOrTouchCombinePathFormIdx(i uint32) {
	errTimes := 0
	for ; i < 4294967295; i++ {
		folder, filename := self.getFolderAndFileName(i)
		folderPath := self.Path + string(os.PathSeparator) + combine_folder + string(os.PathSeparator) + folder
		filePath := folderPath + string(os.PathSeparator) + filename
		fileInfo, err := os.Stat(filePath)
		if err != nil && os.IsNotExist(err) {
			self.CurrCombineIdx = i
			self.CurrCombinePath = filePath
			self.CurrCombineSubPath = slash + folder + slash + filename
			if err = self.touchCombineFile(folderPath, filePath); err != nil {
				if errTimes > 10 {
					panic(err)
				}
				errTimes++
				continue
			}
			break
		}
		if fileInfo != nil && fileInfo.Size() < max_combine_file_size {
			self.CurrCombineIdx = i
			self.CurrCombinePath = filePath
			self.CurrCombineSubPath = slash + folder + slash + filename
			break
		}
	}
}

func (self *Storage) touchCombineFile(folderPath string, filePath string) error {
	if err := os.MkdirAll(folderPath, 0700); err != nil {
		log.Errorf("mkdir folder: %s failed: %s", folderPath, err)
		return err
	}
	_, err := os.Create(filePath)
	if err != nil {
		log.Errorf("create file: %s failed: %s", filePath, err)
		return err
	}
	return nil
}

func (self *Storage) CurrCombineSize() uint32 {
	fileInfo, err := os.Stat(self.CurrCombinePath)
	if err == nil {
		log.Fatalf("Stat file: %s failed: %s", self.CurrCombinePath, err)
	}
	return uint32(fileInfo.Size())
}

func (self *Storage) NextCombinePath(currCombineSize uint32, fileSize uint32) {
	if currCombineSize+fileSize > max_combine_file_size {
		self.findOrTouchCombinePathFormIdx(self.CurrCombineIdx + 1)
	}
}

func checkAllStorageAvailableSpace() {
	if storageSlice == nil {
		sl := make([]*Storage, 0, 1)
		s := &Storage{Path: providerConfig.MainStoragePath, Index: 0}
		s.findOrTouchCombinePath()
		sl = append(sl, s)
		storageSlice = sl
	}
	// TODO  Available Space less than 1G will not to store.
}

func GetStoragePath() *Storage {
	// TODO rotate store all Available storage path
	return storageSlice[0]
}
