package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveClientConfig(t *testing.T) {
	configFilePath := "/tmp/config-test.json"
	removeConfigFile(configFilePath)
	clientConfig = &ClientConfig{
		NodeId:        "test-node-id",
		TrackerServer: "127.0.0.1:8080",
		PublicKey:     "test-publickey",
		PrivateKey:    "test-privatekey",
		TempDir:       "/tmp",
		Email:         "test-email",
	}
	err := SaveClientConfig(configFilePath, clientConfig)
	require.NoError(t, err)

	size := getConfigFileSize(configFilePath)
	fmt.Printf("size %d\n", size)
	require.True(t, size > 100)

	_, err = LoadConfig(configFilePath)
	require.Error(t, err)
	//require.NotNil(t, cc)
	//require.Equal(t, cc.NodeId, clientConfig.NodeId)
	//require.Equal(t, cc.PublicKey, clientConfig.PublicKey)
	//require.Equal(t, cc.TempDir, clientConfig.TempDir)
	removeConfigFile(configFilePath)
}

func removeConfigFile(filename string) {
	_, err := os.Stat(filename)
	if err != nil && os.IsNotExist(err) {
		return
	}
	err = os.Remove(filename)
	if err != nil {
		panic(err)
	}
}

func getConfigFileSize(filename string) int64 {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		panic(err)
	}
	return fileInfo.Size()
}
