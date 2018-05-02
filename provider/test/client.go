package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	pb "github.com/samoslab/nebula/provider/pb"
	"golang.org/x/net/context"
)

var authPublicKeyBytes []byte

const stream_data_size = 32 * 1024

func Ping(client pb.ProviderServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.Ping(ctx, &pb.PingReq{})
	return err
}

func Store(client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, key []byte, size uint64) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed: %s", err.Error())
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
		fmt.Printf("RPC Store failed: %s", err.Error())
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
			fmt.Printf("read file failed: %s", err.Error())
			return err
		}
		if first {
			first = false
			req.Data = buf[:bytesRead]
			if err := stream.Send(req); err != nil {
				fmt.Printf("RPC Send StoreReq failed: %s", err.Error())
				return err
			}
		} else {
			if err := stream.Send(&pb.StoreReq{Data: buf[:bytesRead]}); err != nil {
				fmt.Printf("RPC Send StoreReq failed: %s", err.Error())
				return nil
			}
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		fmt.Printf("RPC CloseAndRecv failed: %s", err.Error())
		return err
	}
	if !storeResp.Success {
		fmt.Println("RPC return false")
		return errors.New("RPC return false")
	}
	return nil
}

func Retrieve(client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, key []byte, size uint64) error {
	file, err := os.OpenFile(filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666)
	if err != nil {
		fmt.Printf("open file failed: %s", err.Error())
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
			fmt.Printf("write file %d bytes failed : %s", len(resp.Data), err.Error())
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
			fmt.Printf("RPC Recv failed: %s", err.Error())
			return err
		}
		if len(resp.Data) == 0 {
			break
		}
		if _, err = file.Write(resp.Data); err != nil {
			fmt.Printf("write file %d bytes failed : %s", len(resp.Data), err.Error())
			return err
		}
	}
	return nil
}
