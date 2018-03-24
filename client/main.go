package main

import (
	"fmt"
	"os"

	client "github.com/spolabs/nebula/client/provider_client"
	pb "github.com/spolabs/nebula/provider/pb"
	util_hash "github.com/spolabs/nebula/util/hash"
	"google.golang.org/grpc"
)

const stream_data_size = 32 * 1024

func main() {
	conn, err := grpc.Dial("127.0.0.1:6666", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		return
	}
	defer conn.Close()
	psc := pb.NewProviderServiceClient(conn)
	fmt.Println("==========test Ping RPC==========")
	err = client.Ping(psc)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Store RPC==========")
	path := "/home/lijt/go/bin/godoc"
	hash, err := util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = client.Store(psc, path, []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Retrieve RPC==========")
	err = client.Retrieve(psc, "/tmp/godoc", []byte("mock-auth"), "mock-ticket", hash)
	if err != nil {
		fmt.Println(err)
	}
}
