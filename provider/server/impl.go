package server

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/net/context"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	pb "github.com/spolabs/nebula/provider/pb"
)

const stream_data_size = 32 * 1024

type ProviderServer struct {
	PathDb *bolt.DB
}

func NewProviderServer(configDir string) *ProviderServer {
	ps := &ProviderServer{}
	var err error
	ps.PathDb, err = bolt.Open("/tmp/my.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
	}
	return ps
}

func (self *ProviderServer) Close() {
	self.PathDb.Close()
}

func (self *ProviderServer) wrapErr(err error, info string) error {
	return errors.Wrapf(err, info)
}

func (self *ProviderServer) checkAuth(method string, auth []byte, ticket string, key string, fileSize uint64, timestamp int64) error {
	// TODO check auth
	return nil
}

func (self *ProviderServer) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{}, nil
}

func (self *ProviderServer) Store(stream pb.ProviderService_StoreServer) error {
	first := true
	var ticket string
	var file *os.File
	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("RPC Recv failed: %s", err.Error())
			return self.wrapErr(err, "failed unexpectadely while reading chunks from stream")
		}
		if first {
			first = false
			ticket = req.Ticket
			if err = self.checkAuth("Store", req.Auth, ticket, req.Key, req.FileSize, req.Timestamp); err != nil {
				fmt.Printf("check auth failed: %s", err.Error())
				return err
			}
			path := self.getPath(req.Key, req.FileSize)
			fileInfo, err := os.Stat(path)
			if err == nil {
				return os.ErrExist
			}
			if fileInfo != nil && fileInfo.IsDir() {
				return errors.New("exist directory")
			}
			file, err = os.OpenFile(
				path,
				os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
				0666)
			if err != nil {
				fmt.Printf("open file failed: %s", err.Error())
				return self.wrapErr(err, "open file failed")
			}
			defer file.Close()
		}
		if len(req.Data) == 0 {
			break
		}
		if _, err = file.Write(req.Data); err != nil {
			fmt.Printf("write file %d bytes failed : %s", len(req.Data), err.Error())
			return self.wrapErr(err, "write file failed")
		}

	}
	// TODO send log with ticket
	fmt.Println(ticket)
	if err := stream.SendAndClose(&pb.StoreResp{Success: true}); err != nil {
		fmt.Printf("RPC SendAndClose failed: %s", err.Error())
		return self.wrapErr(err, "SendAndClose failed")
	}
	return nil
}

func (self *ProviderServer) Retrieve(req *pb.RetrieveReq, stream pb.ProviderService_RetrieveServer) error {
	if err := self.checkAuth("Retrieve", req.Auth, req.Ticket, req.Key, req.FileSize, req.Timestamp); err != nil {
		return err
	}
	file, err := os.Open(self.queryPath(req.Key))
	if err != nil {
		fmt.Printf("open file failed: %s", err.Error())
		return self.wrapErr(err, "open file failed")
	}
	defer file.Close()
	buf := make([]byte, stream_data_size)
	for {
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("read file failed: %s", err.Error())
			return self.wrapErr(err, "read file failed")
		}
		if bytesRead > 0 {
			stream.Send(&pb.RetrieveResp{Data: buf[:bytesRead]})
		}
		if bytesRead < stream_data_size {
			break
		}
	}
	// TODO send log with ticket
	fmt.Println(req.Ticket)
	return nil
}

func (self *ProviderServer) GetFragment(ctx context.Context, req *pb.GetFragmentReq) (*pb.GetFragmentResp, error) {

	return nil, nil
}

const max_combine_file_size = 1048576 //1M

func (self *ProviderServer) queryPath(key string) string {
	// query PathDb
	return "/tmp/" + key + ".blk"
}

func (self *ProviderServer) savePath(key string, path string) {

}
func (self *ProviderServer) getPath(key string, fileSize uint64) string {

	return "/tmp/" + key + ".blk"
}