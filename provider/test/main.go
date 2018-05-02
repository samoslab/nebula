package main

import (
	"fmt"
	"os"

	pb "github.com/samoslab/nebula/provider/pb"
	util_hash "github.com/samoslab/nebula/util/hash"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:6666", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s", err.Error())
		return
	}
	defer conn.Close()
	psc := pb.NewProviderServiceClient(conn)
	fmt.Println("==========test Ping RPC==========")
	err = Ping(psc)
	if err != nil {
		fmt.Println(err)
	}
	var path string
	var hash []byte
	var fileInfo os.FileInfo
	fmt.Println("==========test Big File Store RPC==========")
	path = "/home/lijt/go/bin/godoc"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Big File Retrieve RPC==========")
	err = Retrieve(psc, "/tmp/godoc", []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 1 Store RPC==========")
	path = "/bin/ls"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 1 Retrieve RPC==========")
	err = Retrieve(psc, "/tmp/ls", []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 2 Store RPC==========")
	path = "/bin/touch"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 2 Retrieve RPC==========")
	err = Retrieve(psc, "/tmp/touch", []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 3 Store RPC==========")
	path = "/bin/tar"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 3 Retrieve RPC==========")
	err = Retrieve(psc, "/tmp/tar", []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 4 Store RPC==========")
	path = "/etc/apt/sources.list"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 4 Retrieve RPC==========")
	err = Retrieve(psc, "/tmp/sources.list", []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 5 Store RPC==========")
	path = "/etc/sysctl.conf"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 5 Retrieve RPC==========")
	err = Retrieve(psc, "/tmp/sysctl.conf", []byte("mock-auth"), "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
}
