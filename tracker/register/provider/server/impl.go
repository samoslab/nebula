package server

import (
	"golang.org/x/net/context"

	pb "github.com/spolabs/nebula/tracker/register/provider/pb"
)

type ProviderRegisterService struct {
}

func (self *ProviderRegisterService) GetPublicKey(ctx context.Context, req *pb.Empty) (*pb.PublicKeyResp, error) {
	return &pb.PublicKeyResp{PublicKey: []byte("test provider public key")}, nil //TODO return public key
}

func (self *ProviderRegisterService) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	return &pb.RegisterResp{Success: true}, nil // TODO add real logic
}
