package main

import (
	"fmt"
	"os"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/daemon"
	"github.com/samoslab/nebula/client/service"
)

func main() {
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
	webcfg, err := config.LoadWebConfig("./config.json")
	if err != nil {
		fmt.Printf("load config error  %v\n", err)
		//return
	}

	webcfg = &config.Config{}
	webcfg.SetDefault()
	webcfg.HTTPAddr = "127.0.0.1:7788"

	fmt.Printf("webcfg %+v\n", webcfg)
	server := service.NewHTTPServer(log, *webcfg)

	defer server.Shutdown()
	fmt.Printf("start http port listen\n")
	server.Run()

}
