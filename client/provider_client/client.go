package provider_client

import (
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"time"

	collectClient "github.com/samoslab/nebula/client/collector_client"
	"github.com/samoslab/nebula/client/common"
	pb "github.com/samoslab/nebula/provider/pb"
	tcppb "github.com/samoslab/nebula/tracker/collector/client/pb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const streamDataSize = 32 * 1024
const smallFileSize = 512 * 1024

func now() uint64 {
	return uint64(time.Now().UnixNano())
}

// SetActionLog set action log for collect
func SetActionLog(err error, al *tcppb.ActionLog) {
	if al != nil {
		al.EndTime, al.Info = now(), err.Error()
	}
}

func newActionLogFromStoreReq(req *pb.StoreReq) *tcppb.ActionLog {
	return &tcppb.ActionLog{Type: 1,
		Ticket:    req.Ticket,
		FileHash:  req.FileKey,
		FileSize:  req.FileSize,
		BlockHash: req.BlockKey,
		BlockSize: req.BlockSize,
		BeginTime: now()}
}

func newActionLogFromRetrieveReq(req *pb.RetrieveReq) *tcppb.ActionLog {
	return &tcppb.ActionLog{Type: 2,
		Ticket:    req.Ticket,
		FileHash:  req.FileKey,
		FileSize:  req.FileSize,
		BlockHash: req.BlockKey,
		BlockSize: req.BlockSize,
		BeginTime: now()}
}

// Ping test connectivity
func Ping(client pb.ProviderServiceClient) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := client.Ping(ctx, &pb.PingReq{})
	return err
}

// StorePiece store blocks to privider
func StorePiece(log logrus.FieldLogger, client pb.ProviderServiceClient, uploadPara *common.UploadParameter, auth []byte, ticket string, tm uint64, pm *common.ProgressManager) error {
	fileInfo := uploadPara.HF
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
	req := &pb.StoreReq{
		Ticket:    ticket,
		Auth:      auth,
		Timestamp: tm,
		FileKey:   uploadPara.OriginFileHash,
		FileSize:  uploadPara.OriginFileSize,
		BlockKey:  fileInfo.FileHash,
		BlockSize: fileSize}
	al := newActionLogFromStoreReq(req)
	defer collectClient.Collect(al)
	if fileSize < smallFileSize {
		req.Data, err = ioutil.ReadAll(file)
		if err != nil {
			SetActionLog(err, al)
			return err
		}
		al.TransportSize = uint64(len(req.Data))
		resp, err := client.StoreSmall(context.Background(), req)
		if err != nil {
			SetActionLog(err, al)
			return err
		}
		if !resp.Success {
			log.Errorf("RPC return false")
			SetActionLog(err, al)
			return errors.New("RPC return false")
		}
		// for progress
		if realfile != "" {
			if err := pm.SetIncrement(realfile, uint64(len(req.Data))); err != nil {
				log.Errorf("file %s not in progress map", realfile)
			}
		}
		al.Success, al.EndTime = true, now()
		return nil
	}

	stream, err := client.Store(context.Background())
	if err != nil {
		log.Errorf("RPC Store failed: %s", err.Error())
		SetActionLog(err, al)
		return err
	}
	defer stream.CloseSend()
	buf := make([]byte, streamDataSize)
	first := true
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Errorf("read file failed: %s", err.Error())
			SetActionLog(err, al)
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
				SetActionLog(err, al)
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
		if bytesRead < streamDataSize {
			break
		}
	}
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		log.Errorf("RPC CloseAndRecv failed: %s", err.Error())

		SetActionLog(err, al)
		return err
	}
	if !storeResp.Success {
		log.Error("RPC return false")
		SetActionLog(err, al)
		return errors.New("RPC return false")
	}
	al.Success, al.EndTime = true, now()
	return nil
}

// Retrieve download file from provider piece by piece
func Retrieve(log logrus.FieldLogger, client pb.ProviderServiceClient, filePath string, auth []byte, ticket string, tm uint64, fileKey, blockKey []byte, fileSize, blockSize uint64, pm *common.ProgressManager) error {
	fileHashString := hex.EncodeToString(blockKey)
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
	req := &pb.RetrieveReq{
		Ticket:    ticket,
		FileKey:   fileKey,
		Auth:      auth,
		FileSize:  fileSize,
		Timestamp: tm,
		BlockKey:  blockKey,
		BlockSize: blockSize}
	al := newActionLogFromRetrieveReq(req)
	defer collectClient.Collect(al)
	if fileSize < smallFileSize {
		resp, err := client.RetrieveSmall(context.Background(), req)
		if err != nil {
			SetActionLog(err, al)
			return err
		}
		if _, err = file.Write(resp.Data); err != nil {
			SetActionLog(err, al)
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
			SetActionLog(err, al)
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
	al.Success, al.EndTime, al.TransportSize = true, now(), uint64(fileSize)
	return nil
}
