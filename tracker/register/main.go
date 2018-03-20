package main

import (
	"fmt"
	"log"
	"net"

	pbc "github.com/spolabs/nebula/tracker/register/client/pb"
	server_client "github.com/spolabs/nebula/tracker/register/client/server"
	pbp "github.com/spolabs/nebula/tracker/register/provider/pb"
	server_provider "github.com/spolabs/nebula/tracker/register/provider/server"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 6666))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pbp.RegisterProviderRegisterServiceServer(grpcServer, &server_provider.ProviderRegisterService{})
	pbc.RegisterClientRegisterServiceServer(grpcServer, &server_client.ClientRegisterService{})
	grpcServer.Serve(lis)
}
