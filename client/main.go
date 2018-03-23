package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	client "github.com/spolabs/nebula/client/provider_client"
	pb "github.com/spolabs/nebula/provider/pb"
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
	hash := sha256Sum(path)
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

func sha256Sum(file string) []byte {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("open file failed: %s", err.Error())
		return nil
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		fmt.Printf("read file failed: %s", err.Error())
		return nil
	}
	return h.Sum(nil)
}
