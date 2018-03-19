package server

import (
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	"github.com/pkg/errors"
	pb "github.com/spolabs/nebula/provider/pb"
)

const stream_data_size = 32 * 1024

type ProviderServer struct {
}

func (self *ProviderServer) wrapErr(err error, info string) error {
	return errors.Wrapf(err, info)
}

func (self *ProviderServer) checkAuth(auth string, key string) error {
	// TODO check auth
	return nil
}

func (self *ProviderServer) Ping(ctx context.Context, req *pb.PingReq) (*pb.PingResp, error) {
	fmt.Println(req.NodeId)
	return &pb.PingResp{NodeId: ""}, nil //TODO return real nodeId
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
			if err = self.checkAuth(req.Auth, req.Key); err != nil {
				fmt.Printf("check auth failed: %s", err.Error())
				return err
			}
			file, err = os.OpenFile(
				"/tmp/t/"+req.Key+".blk",
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
	if err := stream.SendAndClose(&pb.StoreResp{Result: true}); err != nil {
		fmt.Printf("RPC SendAndClose failed: %s", err.Error())
		return self.wrapErr(err, "SendAndClose failed")
	}
	return nil
}

func (self *ProviderServer) Retrieve(req *pb.RetrieveReq, stream pb.ProviderService_RetrieveServer) error {
	if err := self.checkAuth(req.Auth, req.Key); err != nil {
		return err
	}
	file, err := os.Open("/tmp/t/" + req.Key + ".blk")
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
