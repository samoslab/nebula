package provider_client

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

func StorePiece(log logrus.FieldLogger, client pb.ProviderServiceClient, fileInfo common.HashFile, auth []byte, ticket string, tm uint64, pm *common.ProgressManager) error {
	filePath := fileInfo.FileName
	fileSize := uint64(fileInfo.FileSize)
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("open file failed: %s", err.Error())
		return err
	}
	defer file.Close()
	realfile, ok := pm.PartitionToOriginMap[filePath]
	if !ok {
		log.Errorf("file %s not in reverse partition map", filePath)
	}
	fmt.Printf("auth %x\n", auth)
	fmt.Printf("ticket %x\n", ticket)
	fmt.Printf("filehash %x\n", fileInfo.FileHash)
	fmt.Printf("blockhash %x\n", fileInfo.FileHash)
	fmt.Printf("tm %d\n", tm)
	fmt.Printf("size %d\n", fileInfo.FileSize)
	req := &pb.StoreReq{Ticket: ticket, Auth: auth, Timestamp: tm, FileKey: fileInfo.FileHash, FileSize: fileSize, BlockKey: fileInfo.FileHash, BlockSize: fileSize}
	if fileSize < 512*1024 {
		req.Data, err = ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		resp, err := client.StoreSmall(context.Background(), req)
		if err != nil {
			return err
		}
		if !resp.Success {
			log.Errorf("RPC return false")
			return errors.New("RPC return false")
		}
		// for progress
		if realfile != "" {
			if err := pm.SetIncrement(realfile, uint64(len(req.Data))); err != nil {
				log.Errorf("file %s not in progress map", realfile)
			}
		}
		return nil
	}

	stream, err := client.Store(context.Background())
	if err != nil {
		log.Errorf("RPC Store failed: %s", err.Error())
		return err
	}
	defer stream.CloseSend()
	buf := make([]byte, stream_data_size)
	first := true
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("read file failed: %s", err.Error())
			return err
		}
		if first {
			first = false
			req.Data = buf[:bytesRead]
			if err := stream.Send(req); err != nil {
				log.Errorf("RPC First Send StoreReq failed: %s", err.Error())
				if err == io.EOF {
					break
				}
				return err
			}
			log.Infof("RPC First Send StoreReq SUCCESS")
		} else {
			if err := stream.Send(&pb.StoreReq{Data: buf[:bytesRead]}); err != nil {
				log.Errorf("RPC Send non-first StoreReq failed: %s", err.Error())
				//if err == io.EOF {
				//break
				//}
				return err
			}
		}
		// for progress
		if realfile != "" {
			if err := pm.SetIncrement(realfile, uint64(bytesRead)); err != nil {
				log.Errorf("file %s not in progress map", realfile)
			}
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorf("RPC CloseAndRecv failed: %s", err.Error())
		return err
	}
	if !storeResp.Success {
		log.Error("RPC return false")
		return errors.New("RPC return false")
	}
	return nil
}

// Retrieve download file from provider piece by piece
func Retrieve(log logrus.FieldLogger, client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, tm uint64, key []byte, fileSize uint64, pm *common.ProgressManager) error {
	fileHashString := hex.EncodeToString(key)
	file, err := os.OpenFile(filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666)
	if err != nil {
		log.Errorf("open file failed: %s", err.Error())
		return err
	}
	defer file.Close()
	realfile, ok := pm.PartitionToOriginMap[fileHashString]
	if !ok {
		log.Errorf("file %s not in reverse partition map", fileHashString)
	}
	req := &pb.RetrieveReq{Ticket: ticket, FileKey: key, Auth: auth, FileSize: fileSize, Timestamp: tm, BlockKey: key, BlockSize: fileSize}
	if fileSize < 512*1024 {
		resp, err := client.RetrieveSmall(context.Background(), req)
		if err != nil {
			return err
		}
		if _, err = file.Write(resp.Data); err != nil {
			log.Errorf("write file %d bytes failed : %s", len(resp.Data), err.Error())
			return err
		}
		if realfile != "" {
			if err := pm.SetIncrement(realfile, uint64(len(resp.Data))); err != nil {
				log.Errorf("file %s not in progress map", realfile)
			}
		}
		return nil
	}
	stream, err := client.Retrieve(context.Background(), req)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("RPC Recv failed: %s", err.Error())
			return err
		}
		if len(resp.Data) == 0 {
			break
		}
		if realfile != "" {
			if err := pm.SetIncrement(realfile, uint64(len(resp.Data))); err != nil {
				log.Errorf("file %s not in progress map", realfile)
			}
		}
		if _, err = file.Write(resp.Data); err != nil {
			log.Errorf("write file %d bytes failed : %s", len(resp.Data), err.Error())
			return err
		}
	}
	return nil
}
