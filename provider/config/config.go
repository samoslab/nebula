package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/koding/multiconfig"
	"github.com/robfig/cron"
)

type ProviderConfig struct {
	NodeId            string
	WalletAddress     string
	BillEmail         string
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

func LoadConfig(configDir *string) {
	configFilePath = *configDir + string(os.PathSeparator) + config_filename
	providerConfig = readConfig()
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
		fmt.Printf("getConfigFileModTime Error: %s\n", err)
		return
	}
	if modTs != configFileModTs {
		providerConfig = readConfig()
	}
}

func getConfigFileModTime() (int64, error) {
	fileInfo, err := os.Stat(configFilePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}

func readConfig() *ProviderConfig {
	m := multiconfig.NewWithPath(configFilePath) // supports TOML, JSON and YAML
	pc := new(ProviderConfig)
	err := m.Load(pc) // Check for error
	if err != nil {
		fmt.Printf("readConfig Error: %s\n", err)
		panic(err)
	}
	m.MustLoad(pc) // Panic's if there is any error
	configFileModTs, err = getConfigFileModTime()
	if err != nil {
		panic(err)
	}
	return pc
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
