package provider_client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	pb "github.com/samoslab/nebula/provider/pb"
	util_hash "github.com/samoslab/nebula/util/hash"
	"golang.org/x/net/context"
)

const stream_data_size = 32 * 1024

func Ping(client pb.ProviderServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.Ping(ctx, &pb.PingReq{})
	return err
}
func UpdateStoreReqAuth(obj *pb.StoreReq) *pb.StoreReq {
	// TODO
	//obj.Auth = []byte("mock-auth")
	return obj
}

func StorePiece(client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, tm uint64, key []byte, fileSize uint64, first bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed: %s\n", err.Error())
		return err
	}
	defer file.Close()
	stream, err := client.Store(context.Background())
	if err != nil {
		fmt.Printf("RPC Store failed: %s\n", err.Error())
		return err
	}
	defer stream.CloseSend()
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
			if err := stream.Send(UpdateStoreReqAuth(&pb.StoreReq{Data: buf[:bytesRead], Ticket: ticket, Auth: auth, Timestamp: tm, Key: key, FileSize: fileSize})); err != nil {
				fmt.Printf("RPC Send StoreReq failed: %s\n", err.Error())
				return err
			}
		} else {
			if err := stream.Send(&pb.StoreReq{Data: buf[:bytesRead]}); err != nil {
				fmt.Printf("RPC Send StoreReq failed: %s\n", err.Error())
				return err
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
		fmt.Println("RPC return false\n")
		return errors.New("RPC return false")
	}
	return nil
}

func Store(client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, tm uint64, key []byte, fileSize uint64, first bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("open file failed: %s", err.Error())
		return err
	}
	defer file.Close()
	stream, err := client.Store(context.Background())
	if err != nil {
		fmt.Printf("RPC Store failed: %s", err.Error())
		return err
	}
	defer stream.CloseSend()
	buf, err := util_hash.GetFileData(filePath)
	if err != nil {
		fmt.Printf("get file data error %v", err)
		return err
	}

	if first {
		if err := stream.Send(UpdateStoreReqAuth(&pb.StoreReq{Data: buf, Ticket: ticket, Auth: auth, Timestamp: tm, Key: key, FileSize: fileSize})); err != nil {
			fmt.Printf("RPC Send StoreReq failed: %s", err.Error())
			return err
		}
	} else {
		if err := stream.Send(&pb.StoreReq{Data: buf}); err != nil {
			fmt.Printf("RPC Send StoreReq failed: %s", err.Error())
			return nil
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

func updateRetrieveReqAuth(obj *pb.RetrieveReq) *pb.RetrieveReq {
	// TODO
	//obj.Auth = []byte("mock-auth")
	return obj

}
func Retrieve(client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, key []byte, tm, filesize uint64) error {
	file, err := os.OpenFile(filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666)
	if err != nil {
		fmt.Printf("open file failed: %s\n", err.Error())
		return err
	}
	defer file.Close()
	stream, err := client.Retrieve(context.Background(), updateRetrieveReqAuth(&pb.RetrieveReq{Ticket: ticket, Key: key, Auth: auth, FileSize: filesize, Timestamp: tm}))
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
