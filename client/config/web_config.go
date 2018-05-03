package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

type Config struct {
	TrackerServer    string        `json:"tracker_server"`
	ConfigDir        string        `json:"config_dir"`
	HTTPAddr         string        `json:"http_addr"`
	HTTPPort         uint32        `json:"http_port"`
	HTTPSAddr        string        `json:"https_addr"`
	StaticDir        string        `json:"static_dir"`
	AutoTLSHost      string        `json:"auto_tls_host"`
	TLSCert          string        `json:"tls_cert"`
	TLSKey           string        `json:"tls_key"`
	ThrottleMax      int64         `json:"throttle_max"` // Maximum number of requests per duration
	ThrottleDuration time.Duration `json:"throttle_duration"`
	BehindProxy      bool          `json:"behind_proxy"`
	APIEnabled       bool          `json:"api_enabled"`
}

func (cfg *Config) SetDefault() {
	if cfg.TrackerServer == "" {
		cfg.TrackerServer = "127.0.0.1:6677"
	}
	if cfg.ConfigDir == "" {
		usr, err := user.Current()
		if err != nil {
			log.Fatalf("Get OS current user failed: %s", err)
		}
		defaultAppDir := filepath.Join(usr.HomeDir, ".spo-nebula-client/config.json")
		cfg.ConfigDir = defaultAppDir
	}
}

func (cfg *Config) Validate() error {
	if cfg.TrackerServer == "" {
		return errors.New("need tracker server")
	}
	if cfg.ConfigDir == "" {
		return errors.New("need config dir")
	}
	return nil
}

func LoadWebConfig(configFilePath string) (*Config, error) {
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
	cc := new(Config)
	err = json.Unmarshal(byteValue, cc)
	if err != nil {
		return nil, err
	}
	cc.SetDefault()
	if err := cc.Validate(); err != nil {
		return nil, err
	}
	return cc, nil
}
