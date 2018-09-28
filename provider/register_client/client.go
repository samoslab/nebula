package register_client

import (
	"context"
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/tracker/register/provider/pb"
	"google.golang.org/grpc"
)

func GetPublicKey(client pb.ProviderRegisterServiceClient) (pubKey []byte, publicKeyHash []byte, ip string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetPublicKey(ctx, &pb.GetPublicKeyReq{})
	if err != nil {
		return nil, nil, "", err
	}
	return resp.PublicKey, resp.PublicKeyHash, resp.Ip, nil
}

func RegisterPublic(client pb.ProviderRegisterServiceClient, publicKeyHash []byte, nodeIdEnc []byte, publicKeyEnc []byte,
	encryptKeyEnc []byte, walletAddressEnc []byte, billEmailEnc []byte, mainStorageVolume uint64,
	upBandwidth uint64, downBandwidth uint64, testUpBandwidth uint64, testDownBandwidth uint64,
	availability float64, port uint32, hostEnc []byte, dynamicDomainEnc []byte,
	extraStorageVolume []uint64, priKey *rsa.PrivateKey) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req := &pb.RegisterReq{Timestamp: uint64(time.Now().Unix()),
		PublicKeyHash:      publicKeyHash,
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
		return 9999, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func RegisterPrivate(client pb.ProviderRegisterServiceClient, publicKeyHash []byte, nodeIdEnc []byte, publicKeyEnc []byte,
	encryptKeyEnc []byte, walletAddressEnc []byte, billEmailEnc []byte, mainStorageVolume uint64,
	upBandwidth uint64, downBandwidth uint64, testUpBandwidth uint64, testDownBandwidth uint64,
	availability float64, extraStorageVolume []uint64, priKey *rsa.PrivateKey) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req := &pb.RegisterReq{Timestamp: uint64(time.Now().Unix()),
		PublicKeyHash:      publicKeyHash,
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
		ExtraStorageVolume: extraStorageVolume,
		ConfirmInner:       true}
	req.SignReq(priKey)
	resp, err := client.Register(ctx, req)
	if err != nil {
		return 9999, "", err
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

func AddExtraStorage(client pb.ProviderRegisterServiceClient, volume uint64) (success bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.AddExtraStorageReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix()),
		Volume:    volume}
	req.SignReq(node.PriKey)
	resp, err := client.AddExtraStorage(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func SwitchPrivate(client pb.ProviderRegisterServiceClient) (success bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.SwitchPrivateReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	req.SignReq(node.PriKey)
	resp, err := client.SwitchPrivate(ctx, req)
	if err != nil {
		return false, err
	}
	return resp.Success, nil
}

func SwitchPublic(client pb.ProviderRegisterServiceClient, publicKeyHash []byte, port uint32, hostEnc []byte,
	dynamicDomainEnc []byte) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.SwitchPublicReq{Timestamp: uint64(time.Now().Unix()),
		NodeId:           node.NodeId,
		PublicKeyHash:    publicKeyHash,
		Port:             port,
		HostEnc:          hostEnc,
		DynamicDomainEnc: dynamicDomainEnc}
	req.SignReq(node.PriKey)
	resp, err := client.SwitchPublic(ctx, req)
	if err != nil {
		return 9999, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func PrivateAlive(trackerServer string) error {
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial tracker %s failed: %s\n", trackerServer, err.Error())
		return err
	}
	defer conn.Close()
	prsc := pb.NewProviderRegisterServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.PrivateAliveReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	req.SignReq(node.PriKey)
	_, err = prsc.PrivateAlive(ctx, req)
	if err != nil {
		fmt.Printf("PrivateAlive failed: %s\n", err.Error())
	}
	return err
}

func GetTrackerServer(client pb.ProviderRegisterServiceClient) (server map[string]uint32, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.GetTrackerServerReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	req.SignReq(node.PriKey)
	resp, err := client.GetTrackerServer(ctx, req)
	if err != nil {
		return nil, err
	}
	res := make(map[string]uint32, len(resp.Server))
	for _, s := range resp.Server {
		res[s.Server] = s.Port
	}
	return res, nil
}

func GetCollectorServer(client pb.ProviderRegisterServiceClient) (server map[string]uint32, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.GetCollectorServerReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	req.SignReq(node.PriKey)
	resp, err := client.GetCollectorServer(ctx, req)
	if err != nil {
		return nil, err
	}
	res := make(map[string]uint32, len(resp.Server))
	for _, s := range resp.Server {
		res[s.Server] = s.Port
	}
	return res, nil
}

func RefreshIp(client pb.ProviderRegisterServiceClient, port uint32) (ip string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.RefreshIpReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix()),
		Port:      port}
	req.SignReq(node.PriKey)
	resp, err := client.RefreshIp(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.Ip, nil

}
