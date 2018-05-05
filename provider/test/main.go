package main

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	pb "github.com/samoslab/nebula/provider/pb"
	util_hash "github.com/samoslab/nebula/util/hash"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("127.0.0.1:6666", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return
	}
	defer conn.Close()
	psc := pb.NewProviderServiceClient(conn)
	fmt.Println("==========test Ping RPC==========")
	err = Ping(psc)
	if err != nil {
		fmt.Println(err)
	}
	authPublicKeyBytes, err = hex.DecodeString("3082010a0282010100d1fc30370bf979ee401e797b52d5ee86a419d4713e2585e6768e093993a5684400c67aafafdfa7d68587c7ebd05e13d706799d3e519f1dd17294512a5e982a8d96954f3bbea766bd896acc66b5814f9e9bcbbf3da18a5ce97249da59e9ddf119cc3607d13f613b6d9a840b05b2afcfe18235c1d279f031621fc3bde9bb483e8908230378c19866a73b55500e13e1494f7d6b0d75cac2b5048519a9549afada800626a0f1f376b71e4c32a6a38a465161d9e95971ba2f57f64737e18d1dd54608c16d61ea3f2b1fce55c8b5d34acd00bf68d5ec6dfbc4e3c787e6e00120c0247bcd73b5a2fa7eaff8e74a3366afde8d40670d78db0b3309865969f356c2bacf630203010001")
	//authPublicKeyBytes, err = hex.DecodeString("3082010a0282010100d65e7ed317d188900aea78b95ea2bcc440acb77676ade53d607b98855e3f6a9f04431ce46d8985876e5db92bc60a3141e84932f3520adc3c6eca0c2d40a132f185c1c8c52f9ac03335b2527945a8721d3489dc99cc170f16eb3f00c2866413e1d728b0676583e12423b2fcda0b1c1ad5209ee0e48e85c09bb1f2a3fd67432d060473015c56f47e48a710e0ca3f6838f7a439f9c5eee4c723d5ca1f6f67fa78f754605e656b492d2f4138c85953e81e68bf193e54f13b7294aa824982a4b33a55980be10a928606d89e8f35eb7b02393f9a0caae9dfa53d89aa49e53a1c02be3e4cc6d9afbe4971f764d65ff87cacb5dd82fc4c5ae22390e62140641123dddb450203010001")
	if err != nil {
		fmt.Println(err)
	}
	var path, getPath string
	var hash, getHash []byte
	var fileInfo os.FileInfo
	fmt.Println("==========test Big File 1 Store RPC==========")
	path = "/bin/bash"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Big File 1 Retrieve RPC==========")
	getPath = "/tmp/bash"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	err = Remove(psc, hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Big File 2 Store RPC==========")
	path = "/bin/ip"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Big File 2 Retrieve RPC==========")
	getPath = "/tmp/ip"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	err = Remove(psc, hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Big File 3 Store RPC==========")
	path = "/bin/awk"
	hash, err = util_hash.Sha1File(path)
	if err != nil {
		panic(err)
	}
	fileInfo, err = os.Stat(path)
	if err != nil {
		panic(err)
	}
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Big File 3 Retrieve RPC==========")
	getPath = "/tmp/awk"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	GetFragment(psc, hash)
	err = Remove(psc, hash, uint64(fileInfo.Size()))
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
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 1 Retrieve RPC==========")
	getPath = "/tmp/ls"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	err = Remove(psc, hash, uint64(fileInfo.Size()))
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
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 2 Retrieve RPC==========")
	getPath = "/tmp/touch"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	err = Remove(psc, hash, uint64(fileInfo.Size()))
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
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 3 Retrieve RPC==========")
	getPath = "/tmp/tar"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	GetFragment(psc, hash)
	err = Remove(psc, hash, uint64(fileInfo.Size()))
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
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 4 Retrieve RPC==========")
	getPath = "/tmp/sources.list"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	err = Remove(psc, hash, uint64(fileInfo.Size()))
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
	err = Store(psc, path, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("==========test Small File 5 Retrieve RPC==========")
	getPath = "/tmp/sysctl.conf"
	err = Retrieve(psc, getPath, "mock-ticket", hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
	getHash, err = util_hash.Sha1File(getPath)
	if err != nil {
		fmt.Println(err)
	}
	if !bytes.Equal(hash, getHash) {
		fmt.Printf("error: hash: %x getHash: %x\n", hash, getHash)
	}
	err = Remove(psc, hash, uint64(fileInfo.Size()))
	if err != nil {
		fmt.Println(err)
	}
}

var authPublicKeyBytes []byte

const stream_data_size = 32 * 1024

func Ping(client pb.ProviderServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.Ping(ctx, &pb.PingReq{})
	return err
}

func Store(client pb.ProviderServiceClient, filePath string, ticket string, key []byte, size uint64) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed: %s\n", err.Error())
		return err
	}
	defer file.Close()
	req := &pb.StoreReq{Ticket: ticket, FileKey: key, FileSize: size, BlockKey: key, BlockSize: size, Timestamp: uint64(time.Now().Unix())}
	req.GenAuth(authPublicKeyBytes)
	if size < 512*1024 {
		req.Data, err = ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		resp, err := client.StoreSmall(context.Background(), req)
		if err != nil {
			return err
		}
		if !resp.Success {
			fmt.Println("RPC return false")
			return errors.New("RPC return false")
		}
		return nil
	}
	stream, err := client.Store(context.Background())
	if err != nil {
		fmt.Printf("RPC Store failed: %s\n", err.Error())
		return err
	}
	defer stream.CloseSend()
	first := true
	buf := make([]byte, stream_data_size)
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("read file failed: %s\n", err.Error())
			return err
		}
		if first {
			first = false
			req.Data = buf[:bytesRead]
			if err := stream.Send(req); err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("RPC Send StoreReq failed: %s\n", err.Error())
				return err
			}
		} else {
			if err := stream.Send(&pb.StoreReq{Data: buf[:bytesRead]}); err != nil {
				if err == io.EOF {
					break
				}
				fmt.Printf("RPC Send StoreReq failed: %s\n", err.Error())
				return nil
			}
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("RPC CloseAndRecv failed: %s\n", err.Error())
		return err
	}
	if !storeResp.Success {
		fmt.Println("RPC return false")
		return errors.New("RPC return false")
	}
	return nil
}

func Retrieve(client pb.ProviderServiceClient, filePath string, ticket string, key []byte, size uint64) error {
	file, err := os.OpenFile(filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666)
	if err != nil {
		fmt.Printf("open file failed: %s\n", err.Error())
		return err
	}
	defer file.Close()
	req := &pb.RetrieveReq{Ticket: ticket, FileKey: key, FileSize: size, BlockKey: key, BlockSize: size, Timestamp: uint64(time.Now().Unix())}
	req.GenAuth(authPublicKeyBytes)
	if size < 512*1024 {
		resp, err := client.RetrieveSmall(context.Background(), req)
		if err != nil {
			return err
		}
		if _, err = file.Write(resp.Data); err != nil {
			fmt.Printf("write file %d bytes failed : %s\n", len(resp.Data), err.Error())
			return err
		}
		return nil
	}
	stream, err := client.Retrieve(context.Background(), req)
	if err != nil {
		return err
	}
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("RPC Recv failed: %s\n", err.Error())
			return err
		}
		if len(resp.Data) == 0 {
			break
		}
		if _, err = file.Write(resp.Data); err != nil {
			fmt.Printf("write file %d bytes failed : %s\n", len(resp.Data), err.Error())
			return err
		}
	}
	return nil
}

func Remove(client pb.ProviderServiceClient, key []byte, size uint64) error {
	req := &pb.RemoveReq{Key: key, Size: size, Timestamp: uint64(time.Now().Unix())}
	req.GenAuth(authPublicKeyBytes)
	resp, err := client.Remove(context.Background(), req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return errors.New("failed")
	}
	return nil
}

func GetFragment(client pb.ProviderServiceClient, key []byte) error {
	req := &pb.GetFragmentReq{Key: key, Positions: []byte{25, 50, 75}, Size: 64, Timestamp: uint64(time.Now().Unix())}
	req.GenAuth(authPublicKeyBytes)
	resp, err := client.GetFragment(context.Background(), req)
	if err != nil {
		return err
	}
	for _, v := range resp.Data {
		fmt.Print(v)
		fmt.Print(" ")
	}
	fmt.Println()
	return nil
}
