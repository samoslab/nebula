package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/daemon"
	"github.com/samoslab/nebula/client/service"
	"github.com/samoslab/nebula/client/util/file"
	"github.com/samoslab/nebula/util/browser"
	"github.com/spf13/pflag"
)

func main() {
	configFile := pflag.StringP("conf", "c", "config.json", "config file")
	serverAddr := pflag.StringP("server", "s", "127.0.0.1:7788", "listen address ip:port")
	collectAddr := pflag.StringP("collect", "", "", "collect server format is ip:port")
	trackerAddr := pflag.StringP("tracker", "", "", "tracker server format is ip:port")
	webDir := pflag.StringP("webdir", "d", "./web/build", "web static directory")
	launchBrowser := pflag.BoolP("launch-browser", "l", true, "launch system default webbrowser at client startup")
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
	fmt.Printf("configFile %s\n", *configFile)
	webcfg, err := config.LoadWebConfig(*configFile)
	if err != nil {
		fmt.Printf("load config error  %v\n", err)
		// set default webcfg avoid crash
		webcfg = &config.Config{}
		webcfg.SetDefault()
	}
	if *serverAddr != "" {
		webcfg.HTTPAddr = *serverAddr
	}

	if *webDir != "" {
		path := file.ResolveResourceDirectory(*webDir)
		webcfg.StaticDir = path
	}
	if *collectAddr != "" {
		webcfg.CollectServer = *collectAddr
	}
	if *trackerAddr != "" {
		webcfg.TrackerServer = *trackerAddr
	}

	fmt.Printf("webcfg %+v\n", webcfg)
	server := service.NewHTTPServer(log, *webcfg)

	defer server.Shutdown()
	fmt.Printf("start http listen on %s\n", webcfg.HTTPAddr)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Run()
	}()
	if *launchBrowser {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Wait a moment just to make sure the http interface is up
			time.Sleep(time.Millisecond * 100)

			fullAddress := "http://" + *serverAddr + "/index.html"
			fmt.Printf("Launching System Browser with %s\n", fullAddress)
			if err := browser.Open(fullAddress); err != nil {
				fmt.Printf("%v", err)
				return
			}
		}()
	}
	wg.Wait()
}
