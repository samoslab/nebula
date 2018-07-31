package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/service"
	"github.com/samoslab/nebula/client/wsservice"
	"github.com/samoslab/nebula/util/apputil"
	"github.com/samoslab/nebula/util/browser"
	"github.com/samoslab/nebula/util/file"
	"github.com/samoslab/nebula/util/logger"
	"github.com/spf13/pflag"
)

func main() {
	configFile := pflag.StringP("conf", "c", "config.json", "config file")
	serverAddr := pflag.StringP("server", "s", "127.0.0.1:7788", "listen address ip:port")
	wsAddr := pflag.StringP("wsaddr", "w", "127.0.0.1:7799", "websocket listen address ip:port")
	collectAddr := pflag.StringP("collect", "", "", "collect server format is ip:port")
	trackerAddr := pflag.StringP("tracker", "", "", "tracker server format is ip:port")
	webDir := pflag.StringP("webdir", "d", "./web/build", "web static directory")
	launchBrowser := pflag.BoolP("launch-browser", "l", false, "launch system default webbrowser at client startup")
	pflag.Parse()
	defaultAppDir, _ := config.GetConfigFile()
	if _, err := os.Stat(defaultAppDir); os.IsNotExist(err) {
		//create the dir.
		if err := os.MkdirAll(defaultAppDir, 0744); err != nil {
			panic(err)
		}
	}
	log, err := logger.NewLogger("", true)
	if err != nil {
		return
	}
	fmt.Printf("configFile %s\n", *configFile)
	webcfg, err := config.LoadWebConfig(*configFile)
	if err != nil {
		log.Errorf("load config error  %v\n", err)
		webcfg = &config.Config{}
		webcfg.SetDefault()
	}
	if *serverAddr != "" {
		webcfg.HTTPAddr = *serverAddr
	}
	if *wsAddr != "" {
		webcfg.WSAddr = *wsAddr
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

	quit := make(chan struct{})
	go apputil.CatchInterrupt(quit)
	go apputil.CatchDebug()

	log.Infof("webcfg %+v", webcfg)
	server := service.NewHTTPServer(log, *webcfg)

	log.Infof("start http listen on %s", webcfg.HTTPAddr)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		server.Run()
	}()

	ws := wsservice.NewWSController(log, server.GetClientManager())
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Infof("start websocket listen at %s", webcfg.WSAddr)
		if err := ws.Run(webcfg.WSAddr); err != nil {
			log.Error("websocket run failed")
		}
	}()

	if *launchBrowser {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Wait a moment just to make sure the http interface is up
			time.Sleep(time.Millisecond * 100)

			fullAddress := "http://" + *serverAddr + "/index.html"
			log.Infof("Launching System Browser with %s\n", fullAddress)
			if err := browser.Open(fullAddress); err != nil {
				log.Errorf("%v", err)
				return
			}
		}()
	}

	select {
	case <-quit:
		log.Info("Get quit signal, exit")
	}
	server.Shutdown()
	ws.Shutdown()
	wg.Wait()
}
