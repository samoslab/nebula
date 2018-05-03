package config

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/koding/multiconfig"
	"github.com/robfig/cron"
	util_file "github.com/samoslab/nebula/util/file"
	util_hash "github.com/samoslab/nebula/util/hash"
	log "github.com/sirupsen/logrus"
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
	checkStorageAvailableSpaceOfConf()
	return nil
}

func verifyConfig(pc *ProviderConfig) (err error) {
	if len(pc.ExtraStorage) > 0 {
		for k, v := range pc.ExtraStorage {
			if k == "0" || k != strconv.Itoa(int(v.Index)) {
				return errors.New("extraStorage index error")
			}
		}
	}
	_, _, _, _, _, err = parseNodeFromConf(pc)
	return err
}

func ParseNode() (nodeId []byte, pubKey *rsa.PublicKey, priKey *rsa.PrivateKey, pubKeyBytes []byte, encryptKey map[string][]byte, err error) {
	return parseNodeFromConf(GetProviderConfig())
}

func parseNodeFromConf(conf *ProviderConfig) (nodeId []byte, pubKey *rsa.PublicKey, priKey *rsa.PrivateKey, pubKeyBytes []byte, encryptKey map[string][]byte, er error) {
	var err error
	pubKeyBytes, err = hex.DecodeString(conf.PublicKey)
	if err != nil {
		er = fmt.Errorf("DecodeString Public Key failed: %s", err)
		return
	}
	if conf.NodeId != hex.EncodeToString(util_hash.Sha1(pubKeyBytes)) {
		er = fmt.Errorf("NodeId is not match PublicKey")
		return
	}
	pubKey, err = x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		er = fmt.Errorf("ParsePKCS1PublicKey failed: %s", err)
		return
	}
	priKeyBytes, err := hex.DecodeString(conf.PrivateKey)
	if err != nil {
		er = fmt.Errorf("DecodeString Private Key failed: %s", err)
		return
	}
	priKey, err = x509.ParsePKCS1PrivateKey(priKeyBytes)
	if err != nil {
		er = fmt.Errorf("ParsePKCS1PrivateKey failed: %s", err)
		return
	}
	encryptKey = make(map[string][]byte, len(conf.EncryptKey))
	for k, v := range conf.EncryptKey {
		encryptKey[k], err = hex.DecodeString(v)
		if err != nil {
			er = fmt.Errorf("DecodeString EncryptKey %s failed: %s", v, err)
			return
		}
	}
	nodeId, err = hex.DecodeString(conf.NodeId)
	if err != nil {
		er = fmt.Errorf("DecodeString node id hex string failed: %s", err)
		return
	}
	return
}

func StartAutoCheck() {
	cronRunner = cron.New()
	cronRunner.AddFunc("0,15,30,45 * * * * *", checkAndReload)
	cronRunner.AddFunc("7 */3 * * * *", checkStorageAvailableSpace)
	cronRunner.AddFunc("37 1,31 * * * *", checkStorageAvailableSpaceOfConf)
	cronRunner.Start()
}

func StopAutoCheck() {
	cronRunner.Stop()
	stopStorage()
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
		} else {
			err = verifyConfig(pc)
			if err == nil {
				providerConfig = pc
			} else {
				log.Warnln(err)
			}
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
	m := multiconfig.NewWithPath(configFilePath)
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

func CreateProviderConfig(configDir string, pc *ProviderConfig) string {
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
	return path
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
