package main

import (
	"fmt"
	"log"
	"net"

	pb "github.com/spolabs/nebula/provider/pb"
	"github.com/spolabs/nebula/provider/server"
	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 6666))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterProviderServiceServer(grpcServer, &server.NewProviderServer())
	grpcServer.Serve(lis)
}
