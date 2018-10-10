package task

import (
	"context"
	"time"

	"github.com/samoslab/nebula/provider/node"
	pb "github.com/samoslab/nebula/tracker/task/pb"
)

func TaskList(client pb.ProviderTaskServiceClient, remove bool, prove bool, send bool, replicate bool) (list []*pb.Task, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	var cate uint32 = 0
	if remove {
		cate |= 0x1
	}
	if prove {
		cate |= 0x2
	}
	if send {
		cate |= 0x4
	}
	if replicate {
		cate |= 0x8
	}
	if cate == 0 {
		return
	}
	req := &pb.TaskListReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix()),
		Category:  cate}
	req.SignReq(node.PriKey)
	resp, er := client.TaskList(ctx, req)
	if er != nil {
		return nil, er
	}
	if err = resp.CheckAuth(node.PubKeyBytes); err != nil {
		return
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
	if er != nil {
		return nil, er
	}
	return resp, nil
}

func GetProveInfo(client pb.ProviderTaskServiceClient, taskId []byte) (proofId []byte, chunkSize uint32, chunkSeq map[uint32][]byte, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.GetProveInfoReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix()),
		TaskId:    taskId}
	req.SignReq(node.PriKey)
	resp, er := client.GetProveInfo(ctx, req)
	if er != nil {
		return nil, 0, nil, er
	}
	return resp.ProofId, resp.ChunkSize, resp.ChunkSeq, nil
}

func FinishProve(client pb.ProviderTaskServiceClient, taskId []byte, proofId []byte, finishedTime uint64, result []byte, remark string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.FinishProveReq{NodeId: node.NodeId,
		Timestamp:    uint64(time.Now().Unix()),
		TaskId:       taskId,
		ProofId:      proofId,
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

func VerifyBlocks(client pb.ProviderTaskServiceClient, query bool, previous uint64, miss []*pb.HashAndSize) (last uint64, blocks []*pb.HashAndSize, hasNext bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	node := node.LoadFormConfig()
	req := &pb.VerifyBlocksReq{NodeId: node.NodeId,
		Timestamp: uint64(time.Now().Unix()),
		Query:     query,
		Previous:  previous,
		Miss:      miss}
	req.SignReq(node.PriKey)
	resp, er := client.VerifyBlocks(ctx, req)
	if er != nil {
		err = er
	} else {
		last, blocks, hasNext = resp.Last, resp.Blocks, resp.HasNext
	}
	return
}
