package register

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"time"

	"google.golang.org/grpc"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/tracker/register/client/pb"
	regpb "github.com/samoslab/nebula/tracker/register/client/pb"
	util_bytes "github.com/samoslab/nebula/util/bytes"
	rsalong "github.com/samoslab/nebula/util/rsa"
	"github.com/sirupsen/logrus"
)

func doGetPubkey(registClient pb.ClientRegisterServiceClient) ([]byte, error) {
	ctx := context.Background()
	getPublicKeyReq := pb.GetPublicKeyReq{}
	getPublicKeyReq.Version = 1
	pubKey, err := registClient.GetPublicKey(ctx, &getPublicKeyReq)
	if err != nil {
		fmt.Printf("pubkey get failed\n")
		return nil, err
	}
	return pubKey.GetPublicKey(), nil
}

func DoRegister(registClient pb.ClientRegisterServiceClient, cfg *config.ClientConfig) (*pb.RegisterResp, error) {
	ctx := context.Background()
	pubkey, err := doGetPubkey(registClient)
	if err != nil {
		return nil, err
	}
	rsaPubkey, err := x509.ParsePKCS1PublicKey(pubkey)
	if err != nil {
		return nil, err
	}
	pubkeyEnc, err := rsalong.EncryptLong(rsaPubkey, cfg.Node.PubKeyBytes, 256)
	if err != nil {
		return nil, err
	}
	contactEmailEnc, err := rsalong.EncryptLong(rsaPubkey, []byte(cfg.Email), 256)
	if err != nil {
		return nil, err
	}
	registerReq := pb.RegisterReq{}
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

func signVerifyContactEmailReq(req *pb.VerifyContactEmailReq, priKey *rsa.PrivateKey) {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write([]byte(req.VerifyCode))
	sign, err := rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		log.Fatal("sign VerifyContactEmail error: " + err.Error())
	}
	req.Sign = sign
}

func VerifyContactEmail(client pb.ClientRegisterServiceClient, verifyCode string, node *node.Node) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &pb.VerifyContactEmailReq{NodeId: node.NodeId,
		Timestamp:  uint64(time.Now().Unix()),
		VerifyCode: verifyCode}

	fmt.Printf("nodeid:%x\n", node.NodeId)
	fmt.Printf("prikey:%v\n", node.PriKey)
	err = req.SignReq(node.PriKey)
	if err != nil {
		return 0, "", err
	}
	//signVerifyContactEmailReq(req, node.PriKey)
	resp, err := client.VerifyContactEmail(ctx, req)
	if err != nil {
		return 0, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func resendVerifyCode(client pb.ClientRegisterServiceClient, node *node.Node) (success bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &pb.ResendVerifyCodeReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	err = req.SignReq(node.PriKey)
	if err != nil {
		return false, err
	}
	resp, err := client.ResendVerifyCode(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Success, nil
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

	rsp, err := DoRegister(registerClient, cc)
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

func VerifyEmail(configDir string, trackerServer string, verifyCode string) error {
	cc, err := config.LoadConfig(configDir)
	if err != nil {
		if err == config.ErrNoConf {
			fmt.Printf("Config file is not ready, please run \"%s register\" to register first\n", os.Args[0])
			return err
		} else if err == config.ErrConfVerify {
			fmt.Println("Config file wrong, can not verify email.")
			return err
		}
		fmt.Println("failed to load config, can not verify email: " + err.Error())
		return err
	}
	if verifyCode == "" {
		fmt.Printf("verifyCode is required.\n")
		os.Exit(9)
	}
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return err
	}
	defer conn.Close()
	registerClient := regpb.NewClientRegisterServiceClient(conn)
	code, errMsg, err := VerifyContactEmail(registerClient, verifyCode, cc.Node)
	if err != nil {
		fmt.Printf("verifyEmail failed: %s\n", err.Error())
		return err
	}
	if code != 0 {
		fmt.Println(errMsg)
		return fmt.Errorf(errMsg)
	}
	fmt.Println("verifyEmail success, you can start upload now.")
	return nil
}

func ResendVerifyCode(configDir string, trackerServer string) {
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
	success, err := resendVerifyCode(crsc, cc.Node)
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
