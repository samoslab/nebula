package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/koding/multiconfig"
	"github.com/robfig/cron"
	log "github.com/sirupsen/logrus"
)

type ProviderConfig struct {
	NodeId            string
	WalletAddress     string
	BillEmail         string
	EncryptKey        string
	Availability      float64
	MainStoragePath   string
	MainStorageVolume uint64
	UpBandwidth       uint64
	DownBandwidth     uint64
	ExtraStorage      map[string]uint64
}

var providerConfig *ProviderConfig

const config_filename = "config.json"

var configFilePath string
var configFileModTs int64

var cronRunner *cron.Cron

func LoadConfig(configDir *string) error {
	configFilePath = *configDir + string(os.PathSeparator) + config_filename
	pc, err := readConfig()
	if err != nil {
		return err
	}
	providerConfig = pc
	return nil
}

func StartAutoReload() {
	cronRunner := cron.New()
	cronRunner.AddFunc("0,15,30,45 * * * * *", checkAndReload)
	cronRunner.Start()
}

func StopAutoReload() {
	cronRunner.Stop()
}

func checkAndReload() {
	modTs, err := getConfigFileModTime()
	if err != nil {
		log.Errorf("getConfigFileModTime Error: %s\n", err)
		return
	}
	if modTs != configFileModTs {
		pc, err := readConfig()
		if err != nil {
			log.Errorf("readConfig Error: %s\n", err)
		} else {
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

func SaveProviderConfig() {
	b, err := json.Marshal(providerConfig)
	if err != nil {
		fmt.Println("json Marshal err:", err)
		return
	}
	var out bytes.Buffer
	if err = json.Indent(&out, b, "", "  "); err != nil {
		fmt.Println("json Indent err:", err)
		return
	}
	if err = ioutil.WriteFile(configFilePath, out.Bytes(), 0644); err != nil {
		fmt.Println("write err:", err)
	}
}
