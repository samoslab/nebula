package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/koding/multiconfig"
	util_file "github.com/spolabs/nebula/util/file"
)

var (
	// ErrNoConf error for no config file
	ErrNoConf = errors.New("not found config file")

	// ErrConfVerify error for failed verified
	ErrConfVerify = errors.New("verify config file failed")
)

// ClientConfig client role config struct json format
type ClientConfig struct {
	TempDir       string
	TrackerServer string
	NodeId        string
	PublicKey     string
	PrivateKey    string
	Email         string
	EncryptKey    map[string]string // key: version, eg: 0, 1, 2
}

var clientConfig *ClientConfig

// LoadConfig load config from config file
func LoadConfig(configDirPath string) (*ClientConfig, error) {
	if !util_file.Exists(configDirPath) {
		return nil, ErrNoConf
	}
	cc, err := readConfig(configDirPath)
	if err != nil {
		return nil, err
	}
	if err = verifyConfig(cc); err != nil {
		return nil, ErrConfVerify
	}
	return cc, nil
}

// SaveClientConfig create client config save to disk
func SaveClientConfig(configDirFile string, cc *ClientConfig) error {
	configDir, _ := filepath.Split(configDirFile)
	if !util_file.Exists(configDir) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("mkdir config folder %s failed:%s", configDir, err)
		}
	}
	return saveClientConfig(configDirFile, cc)
}

func verifyConfig(cc *ClientConfig) error {
	return nil
}

func readConfig(configFilePath string) (*ClientConfig, error) {
	m := multiconfig.NewWithPath(configFilePath) // supports TOML, JSON and YAML
	cc := new(ClientConfig)
	err := m.Load(cc) // Check for error
	if err != nil {
		return nil, err
	}
	m.MustLoad(cc) // Panic's if there is any error
	return cc, nil
}

func saveClientConfig(configPath string, cc *ClientConfig) error {
	b, err := json.Marshal(cc)
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
