package collector_client

import (
	"fmt"
	"sync"

	"github.com/robfig/cron"
	pb "github.com/samoslab/nebula/tracker/collector/provider/pb"
	// log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func Collect(al *pb.ActionLog) {
	queue <- al
	if len(queue) > send_immediate_min {
		send()
	}
}

const batch_max = 500
const send_immediate_min = 20

var queue = make(chan *pb.ActionLog, 2000)
var cronRunner *cron.Cron
var sendLock *sync.Mutex = &sync.Mutex{}
var conn *grpc.ClientConn

func Start() {
	var err error
	conn, err = grpc.Dial("127.0.0.1:6688", grpc.WithInsecure())
	if err != nil {
		fmt.Printf("RPC Dial failed: %s\n", err.Error())
		panic(err)
	}

	cronRunner = cron.New()
	cronRunner.AddFunc("4,19,34,49 * * * * *", send)
	cronRunner.Start()
}

func Stop() {
	cronRunner.Stop()
	conn.Close()
}

func send() {
	sendLock.Lock()
	defer sendLock.Unlock()
	for {
		if len(queue) == 0 {
			break
		}
		size := len(queue)
		if size > batch_max {
			size = batch_max
		}
	}
}

func sendBatch(size int) {

}

func sendToCollector(req *pb.CollectReq) {
	// pcsc := pb.NewProviderCollectorServiceClient(conn)
	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()
	// node := node.LoadFormConfig()
	// req.SignReq(node.PriKey)
	// _, err := pcsc.Collect(ctx, req)
	// if err != nil {
	// 	log.Warnf("send to collector failed, error: %s", err)
	// }
}
