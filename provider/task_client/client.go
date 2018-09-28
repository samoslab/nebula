package task

import (
	"context"
	"time"

	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/tracker/task/pb"
)

func TaskList(client pb.ProviderTaskServiceClient) (list []*pb.Task, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.TaskListReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix())}
	req.SignReq(node.PriKey)
	resp, er := client.TaskList(ctx, req)
	if err != nil {
		return nil, er
	}
	return resp.Task, nil
}

func GetOppositeInfo(client pb.ProviderTaskServiceClient, taskId []byte) (resp *pb.GetOppositeInfoResp, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.GetOppositeInfoReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix()),
		TaskId:    taskId}
	req.SignReq(node.PriKey)
	resp, er := client.GetOppositeInfo(ctx, req)
	if err != nil {
		return nil, er
	}
	return resp, nil
}

func GetProveInfo(client pb.ProviderTaskServiceClient, taskId []byte, proofId []byte) (chunkSize uint32, chunkSeq map[uint32][]byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.GetProveInfoReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix()),
		TaskId:    taskId}
	req.SignReq(node.PriKey)
	resp, er := client.GetProveInfo(ctx, req)
	if err != nil {
		return 0, nil, er
	}
	return resp.ChunkSize, resp.ChunkSeq, nil
}

func FinishProve(client pb.ProviderTaskServiceClient, taskId []byte, proofId []byte, finishedTime uint64, result []byte, remark string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.FinishProveReq{NodeId: node.NodeId,
		Timestamp:    uint64(time.Now().Unix()),
		TaskId:       taskId,
		FinishedTime: finishedTime,
		Result:       result,
		Remark:       remark}
	req.SignReq(node.PriKey)
	_, err = client.FinishProve(ctx, req)
	return
}

func FinishTask(client pb.ProviderTaskServiceClient, taskId []byte, finishedTime uint64, success bool, remark string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.FinishTaskReq{NodeId: node.NodeId,
		Timestamp:    uint64(time.Now().Unix()),
		TaskId:       taskId,
		FinishedTime: finishedTime,
		Success:      success,
		Remark:       remark}
	req.SignReq(node.PriKey)
	_, err = client.FinishTask(ctx, req)
	return
}
