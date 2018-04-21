package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
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
	return cc, nil
}
