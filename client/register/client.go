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
	"time"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/tracker/register/client/pb"
	util_bytes "github.com/samoslab/nebula/util/bytes"
	rsalong "github.com/samoslab/nebula/util/rsa"
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

func ResendVerifyCode(client pb.ClientRegisterServiceClient, node *node.Node) (success bool, err error) {
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
