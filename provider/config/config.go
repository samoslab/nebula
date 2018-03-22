package node

import (
	"fmt"
	"os"

	"github.com/koding/multiconfig"
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

var initProviderConfig = false
var providerConfig *ProviderConfig

func GetProviderConfig(configDir *string) *ProviderConfig {
	if initProviderConfig {
		return providerConfig
	}
	m := multiconfig.NewWithPath(*configDir + string(os.PathSeparator) + "config.toml") // supports TOML, JSON and YAML
	pc := new(ProviderConfig)
	err := m.Load(pc) // Check for error
	if err != nil {
		fmt.Printf("GetProviderConfig Error: %s", err)
	}
	m.MustLoad(pc) // Panic's if there is any error
	//	fmt.Printf("%+v\n", config)
	providerConfig = pc
	initProviderConfig = true
	return providerConfig
}

func SaveProviderConfig(configDir *string) {

}
