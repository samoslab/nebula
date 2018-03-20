package server

import (
	"golang.org/x/net/context"

	pb "github.com/spolabs/nebula/tracker/register/client/pb"
)

type ClientRegisterService struct {
}

func (self *ClientRegisterService) GetPublicKey(ctx context.Context, req *pb.Empty) (*pb.PublicKeyResp, error) {
	return &pb.PublicKeyResp{PublicKey: []byte("test client public key")}, nil //TODO return public key
}

func (self *ClientRegisterService) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterResp, error) {
	return &pb.RegisterResp{Success: true}, nil // TODO add real logic
}
