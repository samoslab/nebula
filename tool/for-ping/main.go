package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/samoslab/nebula/provider/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 1 && len(os.Args) != 2 {
		fmt.Printf("%s [ip:port]", os.Args[0])
		return
	}
	listen := ":6666"
	if len(os.Args) == 2 {
		listen = os.Args[1]
	}
	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(520 * 1024))
	go startPingServer(listen, grpcServer)
	defer grpcServer.GracefulStop()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}

func startPingServer(listen string, grpcServer *grpc.Server) {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Printf("failed to listen: %s, error: %s\n", listen, err.Error())
		os.Exit(57)
	}
	pb.RegisterProviderServiceServer(grpcServer, &pingProviderService{})
	grpcServer.Serve(lis)
}

type pingProviderService struct {
}

func (self *pingProviderService) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{}, nil
}
func (self *pingProviderService) Store(stream pb.ProviderService_StoreServer) error {
	return nil
}
func (self *pingProviderService) StoreSmall(ctx context.Context, req *pb.StoreReq) (*pb.StoreResp, error) {
	return nil, nil
}
func (self *pingProviderService) Retrieve(req *pb.RetrieveReq, stream pb.ProviderService_RetrieveServer) error {
	return nil
}
func (self *pingProviderService) RetrieveSmall(ctx context.Context, req *pb.RetrieveReq) (*pb.RetrieveResp, error) {
	return nil, nil
}
func (self *pingProviderService) Remove(ctx context.Context, req *pb.RemoveReq) (*pb.RemoveResp, error) {
	return nil, nil
}
func (self *pingProviderService) GetFragment(ctx context.Context, req *pb.GetFragmentReq) (*pb.GetFragmentResp, error) {
	return nil, nil
}
func (self *pingProviderService) CheckAvailable(ctx context.Context, req *pb.CheckAvailableReq) (resp *pb.CheckAvailableResp, err error) {
	return nil, nil
}
