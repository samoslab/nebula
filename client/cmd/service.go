package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/daemon"
	"github.com/samoslab/nebula/client/service"
	"github.com/spf13/pflag"
)

func main() {

	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Get OS current user failed: %s", err)
	}
	defaultAppDir := filepath.Join(usr.HomeDir, ".spo-nebula-client")
	defaultConfig := filepath.Join(defaultAppDir, "config.json")
	configDirOpt := pflag.StringP("configfile", "c", defaultConfig, "config data file")
	trackerServer := pflag.StringP("tracker", "s", "127.0.0.1:6677", "tracker server, 127.0.0.1:6677")
	pflag.Parse()
	log, err := daemon.NewLogger("", true)
	if err != nil {
		return
	}
	clientConfig, err := config.LoadConfig(*configDirOpt)

	if err != nil {
		if err == config.ErrNoConf {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return
		} else if err == config.ErrConfVerify {
			fmt.Println("Config file wrong, can not start daemon.")
			return
		}
	}
	clientConfig.TempDir = "/tmp"

	cm, err := daemon.NewClientManager(log, clientConfig.TrackerServer, clientConfig)
	if err != nil {
		fmt.Printf("new client manager failed %v\n", err)
		return
	}
	defer cm.Shutdown()

	webcfg, err := config.LoadWebConfig("./config.json")
	if err != nil {
		fmt.Printf("load config error %v\n", err)
		return
	}
	fmt.Printf("webcfg %+v", webcfg)
	server := service.NewHTTPServer(log, cm, *webcfg)

	defer server.Shutdown()
	fmt.Printf("start http port listen\n")
	server.Run()

}
