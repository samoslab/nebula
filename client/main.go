package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"google.golang.org/grpc"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/client/daemon"
	client "github.com/samoslab/nebula/client/register"
	"github.com/samoslab/nebula/provider/node"
	regpb "github.com/samoslab/nebula/tracker/register/client/pb"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

// NewLogger create logger instance
func NewLogger(logFilename string, debug bool) (*logrus.Logger, error) {
	log := logrus.New()
	log.Out = os.Stdout
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp:    true,
		QuoteEmptyFields: true,
	}
	log.Level = logrus.InfoLevel

	if debug {
		log.Level = logrus.DebugLevel
	}

	return log, nil
}

func verifyEmail(configDir string, trackerServer string, verifyCode string) {
	cc, err := config.LoadConfig(configDir)
	if err != nil {
		if err == config.ErrNoConf {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return
		} else if err == config.ErrConfVerify {
			fmt.Println("Config file wrong, can not verify email.")
			return
		}
		fmt.Println("failed to load config, can not verify email: " + err.Error())
		return
	}
	if verifyCode == "" {
		fmt.Printf("verifyCode is required.\n")
		os.Exit(9)
	}
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return
	}
	defer conn.Close()
	registerClient := regpb.NewClientRegisterServiceClient(conn)
	code, errMsg, err := client.VerifyContactEmail(registerClient, verifyCode, cc.Node)
	if err != nil {
		fmt.Printf("verifyEmail failed: %s\n", err.Error())
		return
	}
	if code != 0 {
		fmt.Println(errMsg)
		return
	}
	fmt.Println("verifyEmail success, you can start daemon now.")
}

// RegisterClient register client info to tracker
func RegisterClient(log *logrus.Logger, configDir, trackerServer, emailAddress string) error {
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("RPC Dial failed: %s", err.Error())
		return err
	}
	defer conn.Close()

	registerClient := regpb.NewClientRegisterServiceClient(conn)

	cc, err := config.LoadConfig(configDir)
	if err != nil {
		fmt.Printf("load config error %v\n", err)
		fmt.Printf("generate config\n")
		no := node.NewNode(10)
		cc = &config.ClientConfig{}
		cc.PublicKey = no.PublicKeyStr()
		cc.PrivateKey = no.PrivateKeyStr()
		cc.NodeId = no.NodeIdStr()
		cc.Email = emailAddress
		cc.TrackerServer = trackerServer
		cc.Node = no
		err = config.SaveClientConfig(configDir, cc)
		if err != nil {
			log.Infof("create config failed %v\n", err)
			return err
		}
	}

	rsp, err := client.DoRegister(registerClient, cc)
	if err != nil {
		log.Infof("register error %v", err)
		return err
	}

	if rsp.GetCode() != 0 {
		log.Infof("register failed: %+v\n", rsp.GetErrMsg())
	}

	log.Infof("register success")
	return nil
}

func resendVerifyCode(configDir string, trackerServer string) {
	cc, err := config.LoadConfig(configDir)
	if err != nil {
		if err == config.ErrNoConf {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return
		} else if err == config.ErrConfVerify {
			fmt.Println("Config file wrong, can not resend verify code email.")
			return
		}
		fmt.Println("failed to load config, can not resend verify code email: " + err.Error())
		return
	}
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return
	}
	defer conn.Close()
	crsc := regpb.NewClientRegisterServiceClient(conn)
	success, err := client.ResendVerifyCode(crsc, cc.Node)
	if err != nil {
		fmt.Printf("resendVerifyCode failed: %s\n", err.Error())
		return
	}
	if !success {
		fmt.Println("resendVerifyCode failed, please retry")
		return
	}
	fmt.Println("resendVerifyCode success, you can verify bill email.")
}
func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Get OS current user failed: %s", err)
	}
	defaultAppDir := filepath.Join(usr.HomeDir, ".spo-nebula-client")
	defaultConfig := filepath.Join(defaultAppDir, "config.json")
	configDirOpt := pflag.StringP("configfile", "f", defaultConfig, "config data file")
	trackerServer := pflag.StringP("tracker", "s", "127.0.0.1:6677", "tracker server, 127.0.0.1:6677")
	emailAddress := pflag.StringP("email", "e", "zhiyuan_06@foxmail.com", "email address")
	upfile := pflag.StringP("upfile", "u", "/tmp/test.zip", "upload file")
	code := pflag.StringP("code", "c", "", "email verify code")
	operation := pflag.StringP("operation", "o", "upload", "client operation, can be register|verify|upload|download|list, default is upload ")

	pflag.Parse()

	if *trackerServer == "" {
		log.Fatal("need tracker server -s")
	}
	log, err := NewLogger("", true)
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
		RegisterClient(log, *configDirOpt, *trackerServer, *emailAddress)
	case "verify":
		if *code == "" {
			fmt.Printf("code can not empy")
			return
		}
		fmt.Printf("config %v\n", *configDirOpt)
		fmt.Printf("tracker %v\n", *trackerServer)
		fmt.Printf("code %v\n", *code)
		verifyEmail(*configDirOpt, *trackerServer, *code)
	case "resend":
		resendVerifyCode(*configDirOpt, *trackerServer)
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
	log.Infof("start client")
	path := "/"
	switch *operation {
	case "mkfolder":
		folders := []string{"tmp"}
		log.Infof("create folder %+v", folders)
		success, err := cm.MkFolder(path, folders, clientConfig.Node)
		if err != nil {
			log.Fatalf("mkdir folder error %v", err)
		}
		log.Infof("create folder rsp:%v", success)
		if !success {
			log.Fatalf("create folder failed")
		}
	case "upload":
		tempFile := *upfile
		log.Infof("upload file %s", tempFile)
		err = cm.UploadFile(tempFile)
		if err != nil {
			log.Fatalf("upload file error %v", err)
		}
		log.Infof("file %s upload success", tempFile)
	case "list":
		rsp, err := cm.ListFiles(path + "tmp")
		if err != nil {
			log.Fatalf("list files error %v", err)
		}
		log.Infof("list files records %d", rsp.GetTotalRecord())
		for _, info := range rsp.GetFof() {
			log.Infof("%s %s %v %d %x", info.GetId(), info.GetName(), info.GetFolder(), info.GetFileSize(), info.GetFileHash())
		}

	case "download":
		fileHash := []byte("")
		fileSize := uint64(0)
		fileName := ""
		folder := false
		err := cm.DownloadFile(fileName, fileHash, fileSize, folder)
		if err != nil {
			log.Errorf("download failed %s", fileName)
		}
	case "remove":
		path := "/tmp"
		recursive := false
		err := cm.RemoveFile(path, recursive)
		if err != nil {
			log.Fatalf("remove files error %v", err)
		}
		log.Infof("remote %s success", path)
	}

}
