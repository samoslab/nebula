package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/daemon"
	client "github.com/samoslab/nebula/client/register"
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
	emailAddress := pflag.StringP("email", "e", "zhiyuan_06@foxmail.com", "email address")
	opfile := pflag.StringP("opfile", "f", "", "operation file")
	code := pflag.StringP("code", "i", "", "email verify code")
	operation := pflag.StringP("operation", "o", "", "client operation, can be register|verify|mkfolder|upload|download|list|remove")
	rootpath := pflag.StringP("rootpath", "p", "", "root path for create folder")
	newfolder := pflag.StringP("newfolder", "n", "", "newfolder for create, join by ,")
	downsize := pflag.Uint64P("downsize", "", 0, "downfile size")
	downhash := pflag.StringP("downhash", "", "", "downhash string")
	interactive := pflag.BoolP("interactive", "", false, "interactive or not")
	newVersion := pflag.BoolP("newversion", "", false, "newversion or not")
	ispath := pflag.BoolP("ispath", "", true, "is path or fileid, true is path")
	recursive := pflag.BoolP("recursive", "", false, "recursive delete or not")
	pageSize := pflag.Uint32P("pagesize", "", 10, "page size")
	pageNum := pflag.Uint32P("pagenum", "", 3, "page number")
	sortType := pflag.Int32P("sorttype", "", 1, "list files sort type")
	ascOrder := pflag.BoolP("ascorder", "", true, "asc order or not")

	pflag.Parse()
	if *operation == "" {
		fmt.Printf("need -o or --operation argument\n")
		pflag.PrintDefaults()
		return
	}

	if *trackerServer == "" {
		pflag.PrintDefaults()
		log.Fatal("need tracker server -s")
	}
	log, err := daemon.NewLogger("", true)
	if err != nil {
		return
	}
	log.Infof("config dir %s", *configDirOpt)
	log.Infof("tracker server %s", *trackerServer)
	dir, _ := filepath.Split(*configDirOpt)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		//create the dir.
		if err := os.MkdirAll(dir, 0744); err != nil {
			panic(err)
		}
	}
	switch *operation {
	case "register":
		client.RegisterClient(log, *configDirOpt, *trackerServer, *emailAddress)
	case "verify":
		if *code == "" {
			fmt.Printf("code can not empy")
			return
		}
		fmt.Printf("config %v\n", *configDirOpt)
		fmt.Printf("tracker %v\n", *trackerServer)
		fmt.Printf("code %v\n", *code)
		client.VerifyEmail(*configDirOpt, *trackerServer, *code)
	case "resend":
		client.ResendVerifyCode(*configDirOpt, *trackerServer)
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
	log.Infof("start client id:%s", clientConfig.NodeId)
	switch *operation {
	case "mkfolder":
		if *rootpath == "" {
			log.Fatal("need --rootpath argument")
		}
		if *newfolder == "" {
			log.Fatal("need --newfolder argument")
		}
		folders := strings.Split(*newfolder, ",")
		log.Infof("create folder %+v", folders)
		success, err := cm.MkFolder(*rootpath, folders, *interactive)
		if err != nil {
			log.Fatalf("mkdir folder error %v", err)
		}
		log.Infof("create folder rsp:%v", success)
		if !success {
			log.Fatalf("create folder failed")
		}
	case "upload":
		if *rootpath == "" {
			log.Fatal("need --rootpath argument")
		}
		if *opfile == "" {
			log.Fatal("need -p argument")
		}
		log.Infof("upload file %s", *opfile)
		err = cm.UploadFile(*rootpath, *opfile, *interactive, *newVersion)
		if err != nil {
			log.Fatalf("upload file error %v", err)
		}
		log.Infof("file %s upload success", *opfile)
	case "list":
		if *rootpath == "" {
			log.Fatal("need --rootpath argument")
		}
		log.Infof("rootpath %s", *rootpath)
		rsp, err := cm.ListFiles(*rootpath, *pageSize, *pageNum, *sortType, *ascOrder)
		if err != nil {
			log.Fatalf("list files error %v", err)
		}
		log.Infof("list files records %d", len(rsp))
		for _, info := range rsp {
			log.Infof("%s %s %v %d %s", info.FileHash, info.FileName, info.Folder, info.FileSize, info.ID)
		}

	case "download":
		if *downhash == "" || *opfile == "" || *downsize == 0 {
			log.Fatalf("need downhash downname downsize")
		}
		fileHash := *downhash
		fileSize := *downsize
		fileName := *opfile
		folder := false
		//bc6bfe7d-7407-4b56-aae9-785b1dd77f67 /tmp/big1 false 181529811 07d0cf85ed032f73c91726e1e5063a620a9f23d4
		err := cm.DownloadFile(fileName, fileHash, fileSize, folder)
		if err != nil {
			log.Fatalf("download failed %s, err %v", fileName, err)
		}
		log.Infof("down success %s", fileSize)
	case "remove":
		if *rootpath == "" {
			log.Fatal("need --rootpath argument")
		}
		recursive := *recursive
		isPath := *ispath
		fmt.Printf("ispath %v, recu %v\n", isPath, recursive)
		err := cm.RemoveFile(*rootpath, recursive, isPath)
		if err != nil {
			log.Fatalf("remove files error %v", err)
		}
		log.Infof("remove %s success", *rootpath)
	}

}
