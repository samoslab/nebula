package provider_client

import (
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/samoslab/nebula/client/common"
	pb "github.com/samoslab/nebula/provider/pb"
	util_hash "github.com/samoslab/nebula/util/hash"
	"github.com/sirupsen/logrus"
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

func StorePiece(log logrus.FieldLogger, client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, tm uint64, key []byte, fileSize uint64, progress map[string]common.ProgressCell) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("open file failed: %s\n", err.Error())
		return err
	}
	defer file.Close()
	stream, err := client.Store(context.Background())
	if err != nil {
		log.Errorf("RPC Store failed: %s\n", err.Error())
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
			log.Errorf("read file failed: %s\n", err.Error())
			return err
		}
		log.Infof("send %d bytes ", bytesRead)
		if err := stream.Send(UpdateStoreReqAuth(&pb.StoreReq{Data: buf[:bytesRead], Ticket: ticket, Auth: auth, Timestamp: tm, Key: key, FileSize: fileSize})); err != nil {
			log.Errorf("RPC Send StoreReq failed: %s\n", err.Error())
			if err.Error() == "EOF" {
				continue
			}
			return err
		}
		// for progress
		if cell, ok := progress[filePath]; ok {
			alreadySent := cell.Current + uint64(bytesRead)
			cell.Current = alreadySent
			progress[filePath] = cell
		} else {
			log.Errorf("file %s not in progress map", filePath)
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorf("RPC CloseAndRecv failed: %s\n", err.Error())
		return err
	}
	if !storeResp.Success {
		fmt.Println("RPC return false\n")
		return errors.New("RPC return false")
	}
	return nil
}

func Store(log logrus.FieldLogger, client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, tm uint64, key []byte, fileSize uint64) error {
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("open file failed: %s", err.Error())
		return err
	}
	defer file.Close()
	stream, err := client.Store(context.Background())
	if err != nil {
		log.Errorf("RPC Store failed: %s", err.Error())
		return err
	}
	defer stream.CloseSend()
	buf, err := util_hash.GetFileData(filePath)
	if err != nil {
		log.Errorf("get file data error %v", err)
		return err
	}

	if err := stream.Send(UpdateStoreReqAuth(&pb.StoreReq{Data: buf, Ticket: ticket, Auth: auth, Timestamp: tm, Key: key, FileSize: fileSize})); err != nil {
		log.Errorf("RPC Send StoreReq failed: %s", err.Error())
		return err
	}
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorf("RPC CloseAndRecv failed: %s", err.Error())
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
func Retrieve(log logrus.FieldLogger, client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, key []byte, tm, filesize uint64) error {
	file, err := os.OpenFile(filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666)
	if err != nil {
		log.Errorf("open file failed: %s\n", err.Error())
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
			log.Errorf("RPC Recv failed: %s\n", err.Error())
			return err
		}
		if len(resp.Data) == 0 {
			break
		}
		if _, err = file.Write(resp.Data); err != nil {
			log.Errorf("write file %d bytes failed : %s\n", len(resp.Data), err.Error())
			return err
		}
	}
	return nil
}
