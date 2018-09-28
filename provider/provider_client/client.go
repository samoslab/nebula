package provider_client

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	client "github.com/samoslab/nebula/provider/collector_client"
	pb "github.com/samoslab/nebula/provider/pb"
	tcppb "github.com/samoslab/nebula/tracker/collector/provider/pb"
	"google.golang.org/grpc"
)

const stream_data_size = 32 * 1024
const small_file_limit = 512 * 1024

func Ping(host string, port uint32, timeout int) (nodeIdHash []byte, latency int64, err error) {
	providerAddr := fmt.Sprintf("%s:%d", host, port)
	conn, err := grpc.Dial(providerAddr, grpc.WithInsecure())
	if err != nil {
		return nil, 0, fmt.Errorf("RPC Dial provider %s failed: %s", providerAddr, err.Error())
	}
	defer conn.Close()
	psc := pb.NewProviderServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	start := time.Now().UnixNano()
	resp, err := psc.Ping(ctx, &pb.PingReq{})
	if err != nil {
		return nil, 0, err
	}
	return resp.NodeIdHash, time.Now().UnixNano() - start, nil
}

func now() uint64 {
	return uint64(time.Now().UnixNano())
}

func setActionLog(err error, al *tcppb.ActionLog) {
	if al != nil {
		al.EndTime, al.Info = now(), err.Error()
	}
}

func newActionLogFromStoreReq(ticket string,
	fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64) *tcppb.ActionLog {
	return &tcppb.ActionLog{Type: 1,
		Ticket:    ticket,
		FileHash:  fileHash,
		FileSize:  fileSize,
		BlockHash: blockHash,
		BlockSize: blockSize,
		BeginTime: now(),
		AsClient:  true}
}

func newActionLogFromRetrieveReq(ticket string,
	fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64) *tcppb.ActionLog {
	return &tcppb.ActionLog{Type: 2,
		Ticket:    ticket,
		FileHash:  fileHash,
		FileSize:  fileSize,
		BlockHash: blockHash,
		BlockSize: blockSize,
		BeginTime: now(),
		AsClient:  true}
}

func StoreSmall(psc pb.ProviderServiceClient, data []byte, auth []byte, timestamp uint64, ticket string,
	fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64) error {
	if blockSize >= small_file_limit || int(blockSize) != len(data) {
		return fmt.Errorf("check data size failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req := &pb.StoreReq{Data: data,
		Auth:      auth,
		Timestamp: timestamp,
		Ticket:    ticket,
		FileKey:   fileHash,
		FileSize:  fileSize,
		BlockKey:  blockHash,
		BlockSize: blockSize}
	al := newActionLogFromStoreReq(ticket, fileHash, fileSize, blockHash, blockSize)
	al.TransportSize = uint64(len(data))
	defer client.Collect(al)
	resp, err := psc.StoreSmall(ctx, req)
	if err != nil {
		setActionLog(err, al)
		return err
	}
	if !resp.Success {
		err := fmt.Errorf("RPC return false")
		setActionLog(err, al)
		return err
	}
	al.Success, al.EndTime = true, now()
	return nil
}

func Store(psc pb.ProviderServiceClient, filePath string, auth []byte, timestamp uint64, ticket string,
	fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64) error {
	fileInfo, er := os.Stat(filePath)
	if er != nil {
		return fmt.Errorf("stat file failed, error: %s", er)
	}
	if blockSize < small_file_limit || int64(blockSize) != fileInfo.Size() {
		return fmt.Errorf("check data size failed")
	}
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file failed: %s", err.Error())
	}
	defer file.Close()

	req := &pb.StoreReq{Auth: auth,
		Timestamp: timestamp,
		Ticket:    ticket,
		FileKey:   fileHash,
		FileSize:  fileSize,
		BlockKey:  blockHash,
		BlockSize: blockSize}
	stream, err := psc.Store(context.Background())
	if err != nil {
		return fmt.Errorf("RPC Store failed: %s", err.Error())
	}
	defer stream.CloseSend()
	first := true
	buf := make([]byte, stream_data_size)
	al := newActionLogFromStoreReq(ticket, fileHash, fileSize, blockHash, blockSize)
	defer client.Collect(al)
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			err = fmt.Errorf("read file failed: %s", err.Error())
			setActionLog(err, al)
			return err
		}
		if first {
			first = false
			req.Data = buf[:bytesRead]
			if err := stream.Send(req); err != nil {
				if err == io.EOF {
					break
				}
				err = fmt.Errorf("RPC Send StoreReq failed: %s", err.Error())
				setActionLog(err, al)
				al.TransportSize += uint64(bytesRead)
				return err
			}
		} else {
			if err := stream.Send(&pb.StoreReq{Data: buf[:bytesRead]}); err != nil {
				if err == io.EOF {
					break
				}
				err = fmt.Errorf("RPC Send StoreReq failed: %s", err.Error())
				setActionLog(err, al)
				al.TransportSize += uint64(bytesRead)
				return err
			}
		}
		al.TransportSize += uint64(bytesRead)
		if bytesRead < stream_data_size {
			break
		}
	}
	storeResp, err := stream.CloseAndRecv()
	if err != nil {
		err = fmt.Errorf("RPC CloseAndRecv failed: %s", err.Error())
		setActionLog(err, al)
		return err
	}
	if !storeResp.Success {
		err = fmt.Errorf("RPC return false")
		setActionLog(err, al)
		return err
	}
	al.Success, al.EndTime = true, now()
	return nil
}

func RetrieveSmall(psc pb.ProviderServiceClient, auth []byte, timestamp uint64, ticket string,
	fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64) (data []byte, err error) {
	if blockSize >= small_file_limit {
		return nil, fmt.Errorf("check data size failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req := &pb.RetrieveReq{Auth: auth,
		Timestamp: timestamp,
		Ticket:    ticket,
		FileKey:   fileHash,
		FileSize:  fileSize,
		BlockKey:  blockHash,
		BlockSize: blockSize}
	al := newActionLogFromRetrieveReq(ticket, fileHash, fileSize, blockHash, blockSize)
	defer client.Collect(al)
	resp, err := psc.RetrieveSmall(ctx, req)
	if err != nil {
		setActionLog(err, al)
		return nil, err
	}
	al.Success, al.EndTime, al.TransportSize = true, now(), uint64(len(data))
	return resp.Data, nil
}

func Retrieve(psc pb.ProviderServiceClient, filePath string, auth []byte, timestamp uint64, ticket string,
	fileHash []byte, fileSize uint64, blockHash []byte, blockSize uint64) error {
	if blockSize < small_file_limit {
		return fmt.Errorf("check data size failed")
	}
	file, err := os.OpenFile(filePath,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0666)
	if err != nil {
		return fmt.Errorf("open file failed: %s", err.Error())
	}
	defer file.Close()
	req := &pb.RetrieveReq{Auth: auth,
		Timestamp: timestamp,
		Ticket:    ticket,
		FileKey:   fileHash,
		FileSize:  fileSize,
		BlockKey:  blockHash,
		BlockSize: blockSize}
	stream, err := psc.Retrieve(context.Background(), req)
	if err != nil {
		return fmt.Errorf("RPC Retrieve failed: %s", err.Error())
	}
	al := newActionLogFromStoreReq(ticket, fileHash, fileSize, blockHash, blockSize)
	defer client.Collect(al)
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			err = fmt.Errorf("RPC Recv failed: %s", err.Error())
			setActionLog(err, al)
			return err
		}
		if len(resp.Data) == 0 {
			break
		}
		al.TransportSize += uint64(len(resp.Data))
		if _, err = file.Write(resp.Data); err != nil {
			err = fmt.Errorf("write file %d bytes failed : %s", len(resp.Data), err.Error())
			setActionLog(err, al)
			return err
		}
	}
	al.Success, al.EndTime = true, now()
	return nil
}
