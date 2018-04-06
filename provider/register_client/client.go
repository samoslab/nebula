package register_client

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spolabs/nebula/provider/node"
	pb "github.com/spolabs/nebula/tracker/register/provider/pb"
	util_bytes "github.com/spolabs/nebula/util/bytes"
)

func GetPublicKey(client pb.ProviderRegisterServiceClient) (pubKey []byte, ip string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetPublicKey(ctx, &pb.GetPublicKeyReq{})
	if err != nil {
		return nil, "", err
	}
	return resp.PublicKey, resp.Ip, nil
}

func signRegisterReq(req *pb.RegisterReq, priKey *rsa.PrivateKey) {
	hasher := sha256.New()
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write(req.NodeIdEnc)
	hasher.Write(req.PublicKeyEnc)
	hasher.Write(req.EncryptKeyEnc)
	hasher.Write(req.WalletAddressEnc)
	hasher.Write(req.BillEmailEnc)
	hasher.Write(util_bytes.FromUint64(req.MainStorageVolume))
	hasher.Write(util_bytes.FromUint64(req.UpBandwidth))
	hasher.Write(util_bytes.FromUint64(req.DownBandwidth))
	hasher.Write(util_bytes.FromUint64(req.TestUpBandwidth))
	hasher.Write(util_bytes.FromUint64(req.TestDownBandwidth))
	hasher.Write(util_bytes.FromUint64(math.Float64bits(req.Availability)))
	hasher.Write(util_bytes.FromUint32(req.Port))
	hasher.Write(req.HostEnc)
	hasher.Write(req.DynamicDomainEnc)
	for _, val := range req.ExtraStorageVolume {
		hasher.Write(util_bytes.FromUint64(val))
	}
	sign, err := rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		log.Fatal("sign Register error: " + err.Error())
	}
	req.Sign = sign
}

func Register(client pb.ProviderRegisterServiceClient, nodeIdEnc []byte, publicKeyEnc []byte,
	encryptKeyEnc []byte, walletAddressEnc []byte, billEmailEnc []byte, mainStorageVolume uint64,
	upBandwidth uint64, downBandwidth uint64, testUpBandwidth uint64, testDownBandwidth uint64,
	availability float64, port uint32, hostEnc []byte, dynamicDomainEnc []byte,
	extraStorageVolume []uint64, priKey *rsa.PrivateKey) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &pb.RegisterReq{NodeIdEnc: nodeIdEnc,
		PublicKeyEnc:       publicKeyEnc,
		EncryptKeyEnc:      encryptKeyEnc,
		WalletAddressEnc:   walletAddressEnc,
		BillEmailEnc:       billEmailEnc,
		MainStorageVolume:  mainStorageVolume,
		UpBandwidth:        upBandwidth,
		DownBandwidth:      downBandwidth,
		TestUpBandwidth:    testUpBandwidth,
		TestDownBandwidth:  testDownBandwidth,
		Availability:       availability,
		Port:               port,
		HostEnc:            hostEnc,
		DynamicDomainEnc:   dynamicDomainEnc,
		ExtraStorageVolume: extraStorageVolume}
	signRegisterReq(req, priKey)
	resp, err := client.Register(ctx, req)
	if err != nil {
		return 0, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func signVerifyBillEmailReq(req *pb.VerifyBillEmailReq, priKey *rsa.PrivateKey) {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write([]byte(req.VerifyCode))
	sign, err := rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		log.Fatal("sign VerifyBillEmail error: " + err.Error())
	}
	req.Sign = sign
}

func VerifyBillEmail(client pb.ProviderRegisterServiceClient, verifyCode string) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.VerifyBillEmailReq{NodeId: node.NodeId,
		Timestamp:  uint64(time.Now().Unix()),
		VerifyCode: verifyCode}
	signVerifyBillEmailReq(req, node.PriKey)
	resp, err := client.VerifyBillEmail(ctx, req)
	if err != nil {
		return 0, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func signResendVerifyCodeReq(req *pb.ResendVerifyCodeReq, priKey *rsa.PrivateKey) {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	sign, err := rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, hasher.Sum(nil))
	if err != nil {
		log.Fatal("sign VerifyBillEmail error: " + err.Error())
	}
	req.Sign = sign
}

func ResendVerifyCode(client pb.ProviderRegisterServiceClient) (success bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.ResendVerifyCodeReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	signResendVerifyCodeReq(req, node.PriKey)
	resp, err := client.ResendVerifyCode(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func GetTrackerServer(client pb.ProviderRegisterServiceClient, nodeId []byte, timestamp uint64,
	sign []byte) (server map[string]uint32, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetTrackerServer(ctx, &pb.GetTrackerServerReq{NodeId: nodeId,
		Timestamp: timestamp,
		Sign:      sign})
	if err != nil {
		return nil, err
	}
	res := make(map[string]uint32, len(resp.Server))
	for _, s := range resp.Server {
		res[s.Server] = s.Port
	}
	return res, nil
}
