package register_client

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/tracker/register/provider/pb"
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

func Register(client pb.ProviderRegisterServiceClient, nodeIdEnc []byte, publicKeyEnc []byte,
	encryptKeyEnc []byte, walletAddressEnc []byte, billEmailEnc []byte, mainStorageVolume uint64,
	upBandwidth uint64, downBandwidth uint64, testUpBandwidth uint64, testDownBandwidth uint64,
	availability float64, port uint32, hostEnc []byte, dynamicDomainEnc []byte,
	extraStorageVolume []uint64, priKey *rsa.PrivateKey) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &pb.RegisterReq{Timestamp: uint64(time.Now().Unix()),
		NodeIdEnc:          nodeIdEnc,
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
	req.SignReq(priKey)
	resp, err := client.Register(ctx, req)
	if err != nil {
		return 1000, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func VerifyBillEmail(client pb.ProviderRegisterServiceClient, verifyCode string) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.VerifyBillEmailReq{NodeId: node.NodeId,
		Timestamp:  uint64(time.Now().Unix()),
		VerifyCode: verifyCode}
	req.SignReq(node.PriKey)
	resp, err := client.VerifyBillEmail(ctx, req)
	if err != nil {
		return 0, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func ResendVerifyCode(client pb.ProviderRegisterServiceClient) (success bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.ResendVerifyCodeReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	req.SignReq(node.PriKey)
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
