package config

import (
	"bytes"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/samoslab/nebula/provider/node"
	"github.com/samoslab/nebula/util/file"
)

var (
	// ErrNoConf error for no config file
	ErrNoConf = errors.New("not found config file")

	// ErrConfVerify error for failed verified
	ErrConfVerify = errors.New("verify config file failed")
)

// ReadableSpace user space
type ReadableSpace struct {
	SpaceNo  uint32 `json:"no"`
	Password string `json:"-"`
	Home     string `json:"home"`
	Name     string `json:"name"`
}

// ClientConfig client role config struct json format
type ClientConfig struct {
	NodeId       string          `json:"node_id"`
	PublicKey    string          `json:"public_key"`
	PrivateKey   string          `json:"private_key"`
	Email        string          `json:"email"`
	Node         *node.Node      `json:"-"`
	Root         string          `json:"root"`
	Space        []ReadableSpace `json:"space"`
	SelfFileName string          `json:"self_filename"`
}

// LoadConfig load config from config file
func LoadConfig(configDirPath string) (*ClientConfig, error) {
	if !file.Exists(configDirPath) {
		return nil, ErrNoConf
	}
	cc, err := readConfig(configDirPath)
	if err != nil {
		return nil, err
	}
	if err = verifyConfig(cc); err != nil {
		return nil, ErrConfVerify
	}
	cc.Node = GetNodeFromConfig(cc)
	return cc, nil
}

// GetNodeFromConfig get node from config
func GetNodeFromConfig(conf *ClientConfig) *node.Node {
	pubKeyBytes, err := hex.DecodeString(conf.PublicKey)
	if err != nil {
		log.Fatalf("DecodeString Public Key failed: %s", err)
	}
	if conf.NodeId != hex.EncodeToString(sha1Sum(pubKeyBytes)) {
		log.Fatalln("NodeId is not match PublicKey")
	}
	pubK, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		log.Fatalf("ParsePKCS1PublicKey failed: %s", err)
	}
	priKeyBytes, err := hex.DecodeString(conf.PrivateKey)
	if err != nil {
		log.Fatalf("DecodeString Private Key failed: %s", err)
	}
	priK, err := x509.ParsePKCS1PrivateKey(priKeyBytes)
	if err != nil {
		log.Fatalf("ParsePKCS1PrivateKey failed: %s", err)
	}
	nodeId, err := hex.DecodeString(conf.NodeId)
	if err != nil {
		log.Fatalf("DecodeString node id hex string failed: %s", err)
	}

	return &node.Node{NodeId: nodeId, PubKey: pubK, PriKey: priK, PubKeyBytes: pubKeyBytes}
}

// SaveClientConfig create client config save to disk
func SaveClientConfig(configDirFile string, cc *ClientConfig) error {
	configDir, _ := filepath.Split(configDirFile)
	if !file.Exists(configDir) {
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
	// Open our jsonFile
	jsonFile, err := os.Open(configFilePath)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	cc := new(ClientConfig)
	err = json.Unmarshal(byteValue, cc)
	if err != nil {
		return nil, err
	}

	return cc, nil
}

func saveClientConfig(configPath string, cc *ClientConfig) error {
	cc.SelfFileName = configPath
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

func sha1Sum(content []byte) []byte {
	h := sha1.New()
	h.Write(content)
	return h.Sum(nil)
}

// GetConfigFile get config file
func GetConfigFile() (string, string) {
	defaultConfig := filepath.Join(file.UserHome(), DefaultConfigDir, DefaultConfig)
	defaultAppDir, _ := filepath.Split(defaultConfig)
	return defaultAppDir, defaultConfig
}
