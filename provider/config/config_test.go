package config

import (
	"os"
	"testing"
)

func TestSaveProviderConfig(t *testing.T) {
	configFilePath = "/tmp/config-test.json"
	removeConfigFile()
	providerConfig = &ProviderConfig{
		NodeId:            "test-node-id",
		WalletAddress:     "test-wallet-address",
		BillEmail:         "test-bill-email",
		Availability:      0.98,
		MainStoragePath:   "/main/storage/path",
		MainStorageVolume: 200000000000,
		UpBandwidth:       4,
		DownBandwidth:     100,
		ExtraStorage: map[string]ExtraStorageInfo{
			"1": ExtraStorageInfo{Path: "/extra/storage/path1",
				Volume: 200000000000,
				Index:  1},
			"2": ExtraStorageInfo{Path: "/extra/storage/path2",
				Volume: 100000000000,
				Index:  2},
		},
	}
	SaveProviderConfig()
	if getConfigFileSize() < 300 {
		t.Errorf("Failed. ")
	}

}

func removeConfigFile() {
	_, err := os.Stat(configFilePath)
	if err != nil && os.IsNotExist(err) {
		return
	}
	err = os.Remove(configFilePath)
	if err != nil {
		panic(err)
	}
}

func getConfigFileSize() int64 {
	fileInfo, err := os.Stat(configFilePath)
	if err != nil {
		panic(err)
	}
	return fileInfo.Size()
}
