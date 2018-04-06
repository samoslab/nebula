package register_client

import (
	"context"
	"time"

	pb "github.com/spolabs/nebula/tracker/register/provider/pb"
)

func GetPublicKey(client pb.ProviderRegisterServiceClient) (pubKey []byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.GetPublicKey(ctx, &pb.GetPublicKeyReq{})
	if err != nil {
		return nil, err
	}
	return resp.PublicKey, nil
}

func Register(client pb.ProviderRegisterServiceClient, nodeIdEnc []byte, publicKeyEnc []byte,
	encryptKeyEnc []byte, walletAddressEnc []byte, billEmailEnc []byte, mainStorageVolume uint64,
	upBandwidth uint64, downBandwidth uint64, testUpBandwidth uint64, testDownBandwidth uint64,
	availability float32, port uint32, ipEnc []byte, dynamicDomainEnc []byte,
	extraStorageVolume []uint64, sign []byte) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.Register(ctx, &pb.RegisterReq{NodeIdEnc: nodeIdEnc,
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
		IpEnc:              ipEnc,
		DynamicDomainEnc:   dynamicDomainEnc,
		ExtraStorageVolume: extraStorageVolume,
		Sign:               sign})
	if err != nil {
		return 0, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func VerifyBillEmail(client pb.ProviderRegisterServiceClient, nodeId []byte, timestamp uint64,
	verifyCode string, sign []byte) (code uint32, errMsg string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.VerifyBillEmail(ctx, &pb.VerifyBillEmailReq{NodeId: nodeId,
		Timestamp:  timestamp,
		VerifyCode: verifyCode,
		Sign:       sign})
	if err != nil {
		return 0, "", err
	}
	return resp.Code, resp.ErrMsg, nil
}

func ResendVerifyCode(client pb.ProviderRegisterServiceClient, nodeId []byte, timestamp uint64,
	sign []byte) (success bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := client.ResendVerifyCode(ctx, &pb.ResendVerifyCodeReq{NodeId: nodeId,
		Timestamp: timestamp,
		Sign:      sign})
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
