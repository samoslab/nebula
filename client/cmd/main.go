package main

import (
	"context"
	"crypto/x509"
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
	rsalong "github.com/samoslab/nebula/util/rsa"
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

func doGetPubkey(registClient regpb.ClientRegisterServiceClient) ([]byte, error) {
	ctx := context.Background()
	getPublicKeyReq := regpb.GetPublicKeyReq{}
	getPublicKeyReq.Version = 1
	pubKey, err := registClient.GetPublicKey(ctx, &getPublicKeyReq)
	if err != nil {
		fmt.Printf("pubkey get failed\n")
		return nil, err
	}
	return pubKey.GetPublicKey(), nil
}

func doRegister(registClient regpb.ClientRegisterServiceClient, cfg *config.ClientConfig) (*regpb.RegisterResp, error) {
	ctx := context.Background()
	pubkey, err := doGetPubkey(registClient)
	if err != nil {
		return nil, err
	}
	rsaPubkey, err := x509.ParsePKCS1PublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	fmt.Printf("cfg  node id %v\n", cfg.NodeId)
	pubkeyEnc, err := rsalong.EncryptLong(rsaPubkey, cfg.Node.PubKeyBytes, 256)
	if err != nil {
		return nil, err
	}
	fmt.Printf("2\n")
	contactEmailEnc, err := rsalong.EncryptLong(rsaPubkey, []byte(cfg.Email), 256)
	if err != nil {
		return nil, err
	}
	registerReq := regpb.RegisterReq{}
	registerReq.Version = 1
	registerReq.NodeId = cfg.Node.NodeId
	registerReq.PublicKeyEnc = pubkeyEnc
	registerReq.ContactEmailEnc = contactEmailEnc

	rsp, err := registClient.Register(ctx, &registerReq)
	if err != nil {
		fmt.Printf("register failed\n")
		return nil, err
	}
	fmt.Printf("rsp %+v\n", rsp)
	return rsp, nil
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

	rsp, err := doRegister(registerClient, cc)
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
	regAction := pflag.StringP("register", "r", "no", "register first")
	verifyAction := pflag.StringP("verify", "v", "no", "verify first")
	code := pflag.StringP("code", "c", "", "email verify code")
	resendAction := pflag.StringP("resend", "a", "no", "resend again")

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
	if *regAction == "yes" {
		RegisterClient(log, *configDirOpt, *trackerServer, *emailAddress)
		return
	}

	if *verifyAction == "yes" {
		if *code == "" {
			fmt.Printf("code can not empy")
			return
		}
		fmt.Printf("config %v\n", *configDirOpt)
		fmt.Printf("tracker %v\n", *trackerServer)
		fmt.Printf("code %v\n", *code)
		verifyEmail(*configDirOpt, *trackerServer, *code)
		return
	}

	if *resendAction == "yes" {
		resendVerifyCode(*configDirOpt, *trackerServer)
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
	log.Infof("start client")
	tempFile := "/tmp/test.zip"
	err = cm.UploadFile(tempFile)
	if err != nil {
		log.Fatalf("upload file error %v", err)
	}
	cm.Shutdown()
}
