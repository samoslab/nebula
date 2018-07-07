package config

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/samoslab/nebula/client/util/file"
)

const (
	// DefaultConfig default store filename of config
	DefaultConfig = ".samos-nebula-client/config.json"
	// DefaultTracker default tracker server
	DefaultTracker = "127.0.0.1:6677"
	// DefaultCollect default collect server
	DefaultCollect = "127.0.0.1:6688"

	// DefaultServer default listen addr
	DefaultServer = "127.0.0.1:7788"

	// DefaultWebDir default web dir
	DefaultWebDir = "./web/build"
)

// Config config for web
type Config struct {
	TrackerServer    string        `json:"tracker_server"`
	CollectServer    string        `json:"collect_server"`
	ConfigDir        string        `json:"config_dir"`
	ConfigFile       string        `json:"config_file"`
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

// SetDefault set default value
func (cfg *Config) SetDefault() {
	if cfg.TrackerServer == "" {
		cfg.TrackerServer = DefaultTracker
	}
	if cfg.CollectServer == "" {
		cfg.CollectServer = DefaultCollect
	}
	if cfg.ConfigDir == "" {
		cfg.ConfigDir = file.UserHome()
	}
	if cfg.ConfigFile == "" {
		cfg.ConfigFile = filepath.Join(file.UserHome(), DefaultConfig)
	}
	if cfg.HTTPAddr == "" {
		cfg.HTTPAddr = DefaultServer
	}

	if cfg.StaticDir == "" {
		cfg.StaticDir = DefaultWebDir
	}
}

// Validate validate config correctness
func (cfg *Config) Validate() error {
	if cfg.TrackerServer == "" {
		return errors.New("need tracker server")
	}
	if cfg.ConfigDir == "" {
		return errors.New("need config dir")
	}
	return nil
}

// LoadWebConfig load web config
func LoadWebConfig(configFilePath string) (*Config, error) {
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
