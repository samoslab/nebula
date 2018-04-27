package provider_client

import (
	"encoding/hex"
	"errors"
	"io"
	"os"
	"time"

	"github.com/samoslab/nebula/client/common"
	pb "github.com/samoslab/nebula/provider/pb"
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

func StorePiece(log logrus.FieldLogger, client pb.ProviderServiceClient, fileInfo common.HashFile, auth []byte, ticket string, tm uint64, progress map[string]common.ProgressCell, fileMap map[string]string) error {
	filePath := fileInfo.FileName
	fileSize := uint64(fileInfo.FileSize)
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
	realfile, ok := fileMap[filePath]
	if !ok {
		log.Errorf("file %s not in reverse partition map", filePath)
	}
	first := true
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("read file failed: %s\n", err.Error())
			return err
		}
		if first {
			first = false
			if err := stream.Send(UpdateStoreReqAuth(&pb.StoreReq{Data: buf[:bytesRead], Ticket: ticket, Auth: auth, Timestamp: tm, Key: fileInfo.FileHash, FileSize: fileSize})); err != nil {
				log.Errorf("RPC Send StoreReq failed: %s\n", err.Error())
				//if err.Error() == "EOF" {
				//continue
				//}
				return err
			}
		} else {
			if err := stream.Send(UpdateStoreReqAuth(&pb.StoreReq{Data: buf[:bytesRead]})); err != nil {
				log.Errorf("RPC Send StoreReq failed: %s\n", err.Error())
				//if err.Error() == "EOF" {
				//continue
				//}
				return err
			}
		}
		// for progress
		if realfile != "" {
			if cell, ok := progress[realfile]; ok {
				cell.Current = cell.Current + uint64(bytesRead)
				progress[realfile] = cell
			} else {
				log.Errorf("file %s not in progress map", realfile)
			}
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	log.Infof("file %s %+v", filePath, progress)
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorf("RPC CloseAndRecv failed: %s\n", err.Error())
		return err
	}
	if !storeResp.Success {
		log.Error("RPC return false")
		return errors.New("RPC return false")
	}
	time.Sleep(time.Second)
	return nil
}

func updateRetrieveReqAuth(obj *pb.RetrieveReq) *pb.RetrieveReq {
	// TODO
	//obj.Auth = []byte("mock-auth")
	return obj

}

// Retrieve download file from provider piece by piece
func Retrieve(log logrus.FieldLogger, client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, key []byte, fileSize, tm uint64, progress map[string]common.ProgressCell, fileMap map[string]string) error {
	fileHashString := hex.EncodeToString(key)
	file, err := os.OpenFile(filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666)
	if err != nil {
		log.Errorf("open file failed: %s\n", err.Error())
		return err
	}
	defer file.Close()
	realfile, ok := fileMap[fileHashString]
	if !ok {
		log.Errorf("file %s not in reverse partition map", fileHashString)
	}
	stream, err := client.Retrieve(context.Background(), updateRetrieveReqAuth(&pb.RetrieveReq{Ticket: ticket, Key: key, Auth: auth, FileSize: fileSize, Timestamp: tm}))
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
		if realfile != "" {
			if cell, ok := progress[realfile]; ok {
				cell.Current = cell.Current + uint64(len(resp.Data))
				progress[realfile] = cell
			} else {
				log.Errorf("file %s not in progress map", realfile)
			}
		}
		if _, err = file.Write(resp.Data); err != nil {
			log.Errorf("write file %d bytes failed : %s\n", len(resp.Data), err.Error())
			return err
		}
	}
	return nil
}
