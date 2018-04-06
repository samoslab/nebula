package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spolabs/nebula/client/config"
	"github.com/spolabs/nebula/client/daemon"
	"github.com/spolabs/nebula/provider/node"
	regpb "github.com/spolabs/nebula/tracker/register/client/pb"
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
	fmt.Printf("pubkey %x\n", pubKey)
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
	rng := rand.Reader
	pubkeyEnc, err := rsa.EncryptPKCS1v15(rng, rsaPubkey, []byte(cfg.PublicKey))
	if err != nil {
		return nil, err
	}
	contactEmailEnc, err := rsa.EncryptPKCS1v15(rng, rsaPubkey, []byte(cfg.Email))
	if err != nil {
		return nil, err
	}
	registerReq := regpb.RegisterReq{}
	registerReq.Version = 1
	registerReq.NodeId = []byte(cfg.NodeId)
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

// RegisterClient register client info to tracker
func RegisterClient(log *logrus.Logger, configDir, trackerServer string) error {
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("RPC Dial failed: %s", err.Error())
		return err
	}
	defer conn.Close()

	registerClient := regpb.NewClientRegisterServiceClient(conn)

	no := node.NewNode(10)
	cc := config.ClientConfig{}
	cc.PublicKey = no.PublicKeyStr()
	cc.PrivateKey = no.PrivateKeyStr()
	cc.NodeId = no.NodeIdStr()
	config.CreateClientConfig(configDir, &cc)

	//clientConfig := config.GetClientConfig()
	rsp, err := doRegister(registerClient, &cc)
	if err != nil {
		log.Infof("register error %v", err)
		return err
	}

	log.Infof("rsp: %+v\n", rsp)
	log.Infof("register success")
	return nil
}

func main() {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Get OS current user failed: %s", err)
	}
	defaultAppDir := filepath.Join(usr.HomeDir, ".spo-nebula-client")
	configDirOpt := pflag.StringP("configdir", "d", defaultAppDir, "config data directory")
	trackerServer := pflag.StringP("tracker", "s", "", "tracker server, 1.1.1.1:666")
	regAction := pflag.StringP("register", "r", "no", "register first")

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
		RegisterClient(log, *configDirOpt, *trackerServer)
		return
	}

	err = config.LoadConfig(*configDirOpt)
	if err != nil {
		if err == config.ErrNoConf {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return
		} else if err == config.ErrConfVerify {
			fmt.Println("Config file wrong, can not start daemon.")
			return
		}
	}

	clientConfig := config.GetClientConfig()
	cm, err := daemon.NewClientManager(log, *trackerServer, clientConfig)
	if err != nil {
		return
	}
	log.Infof("start client")
	tempFile := "test.zip"
	err = cm.UploadFile(tempFile)
	if err != nil {
		log.Fatalf("upload file error %v", err)
	}
}
