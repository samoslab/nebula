package main

import (
	"fmt"
	"os"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/daemon"
	"github.com/samoslab/nebula/client/service"
	"github.com/spf13/pflag"
)

func main() {
	configFile := pflag.StringP("conffile", "", "config.json", "config file")
	serverAddr := pflag.StringP("server", "", "127.0.0.1:7788", "listen address ip:port")
	pflag.Parse()
	defaultAppDir, _ := daemon.GetConfigFile()
	if _, err := os.Stat(defaultAppDir); os.IsNotExist(err) {
		//create the dir.
		if err := os.MkdirAll(defaultAppDir, 0744); err != nil {
			panic(err)
		}
	}
	log, err := daemon.NewLogger("", true)
	if err != nil {
		return
	}
	webcfg, err := config.LoadWebConfig(*configFile)
	if err != nil {
		fmt.Printf("load config error  %v\n", err)
		//return
	}

	webcfg = &config.Config{}
	webcfg.SetDefault()
	webcfg.HTTPAddr = *serverAddr

	fmt.Printf("webcfg %+v\n", webcfg)
	server := service.NewHTTPServer(log, *webcfg)

	defer server.Shutdown()
	fmt.Printf("start http listen\n")
	server.Run()

}
