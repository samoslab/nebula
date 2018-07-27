package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	pb "github.com/samoslab/nebula/provider/pb"
	"google.golang.org/grpc"
)

func main() {
	if len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Println("ping host post [timeout]")
		return
	}
	timeout := 10
	if len(os.Args) == 4 {
		var err error
		timeout, err = strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Println("timeout must be integer")
			return
		}
	}
	start := time.Now().UnixNano()
	conn, err := grpc.Dial(os.Args[1]+":"+os.Args[2], grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		return
	}
	defer conn.Close()
	psc := pb.NewProviderServiceClient(conn)
	err = Ping(psc, timeout)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Ping success. cost: %dms\n", (time.Now().UnixNano()-start)/1000000)
	}
}

func Ping(client pb.ProviderServiceClient, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*int(time.Second)))
	defer cancel()
	_, err := client.Ping(ctx, &pb.PingReq{})
	return err
}
