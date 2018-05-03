package collector_client

import (
	"fmt"

	pb "github.com/samoslab/nebula/tracker/collector/provider/pb"
)

func Collect(al *pb.ActionLog) {
	fmt.Println(al)
	// TODO
}
