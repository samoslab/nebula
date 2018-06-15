package order

import (
	"context"
	"crypto/rsa"
	"encoding/hex"
	"time"

	"github.com/samoslab/nebula/client/common"
	pb "github.com/samoslab/nebula/tracker/register/client/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type OrderManager struct {
	orderClient pb.OrderServiceClient
	Log         logrus.FieldLogger
	privateKey  *rsa.PrivateKey
	NodeId      []byte
}

func NewOrderManager(trackerServer string, log logrus.FieldLogger, privateKey *rsa.PrivateKey, nodeId []byte) *OrderManager {
	conn, err := grpc.Dial(trackerServer, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("RPC Dial failed: %s", err.Error())
		return nil
	}
	oc := pb.NewOrderServiceClient(conn)
	return &OrderManager{
		orderClient: oc,
		Log:         log,
		privateKey:  privateKey,
		NodeId:      nodeId,
	}
}

func (om *OrderManager) GetAllPackages() ([]*pb.Package, error) {
	log := om.Log
	req := &pb.AllPackageReq{
		Version: common.Version,
	}
	rsp, err := om.orderClient.AllPackage(context.Background(), req)
	if err != nil {
		return nil, err
	}

	log.Infof("%+v", rsp)
	return rsp.GetAllPackage(), nil
}

func (om *OrderManager) GetPackageInfo(id uint64) (*pb.Package, error) {
	log := om.Log
	req := &pb.PackageInfoReq{
		Version:   common.Version,
		PackageId: int64(id),
	}
	rsp, err := om.orderClient.PackageInfo(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	return rsp.GetPackage(), nil
}

func (om *OrderManager) BuyPackage(id uint64, canceled bool, quanlity uint32) (*pb.Order, error) {
	log := om.Log
	req := &pb.BuyPackageReq{
		Version:      common.Version,
		NodeId:       om.NodeId,
		Timestamp:    uint64(time.Now().UTC().Unix()),
		PackageId:    int64(id),
		Quanlity:     quanlity,
		CancelUnpaid: canceled,
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.BuyPackage(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	return rsp.GetOrder(), nil
}

func (om *OrderManager) MyAllOrders(expired bool) ([]*pb.Order, error) {
	log := om.Log
	req := &pb.MyAllOrderReq{
		Version:        common.Version,
		NodeId:         om.NodeId,
		Timestamp:      uint64(time.Now().UTC().Unix()),
		OnlyNotExpired: expired,
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.MyAllOrder(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	return rsp.GetMyAllOrder(), nil
}

func (om *OrderManager) GetOrderInfo(orderId string) (*pb.Order, error) {
	log := om.Log
	orderid, err := hex.DecodeString(orderId)
	if err != nil {
		return nil, err
	}
	req := &pb.OrderInfoReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: uint64(time.Now().UTC().Unix()),
		OrderId:   orderid,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.OrderInfo(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)

	return rsp.GetOrder(), nil
}

func (om *OrderManager) RechargeAddress() (*pb.RechargeAddressResp, error) {
	log := om.Log
	req := &pb.RechargeAddressReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: uint64(time.Now().UTC().Unix()),
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.RechargeAddress(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)

	return rsp, nil
}

func (om *OrderManager) PayOrdor(orderId string) (*pb.PayOrderResp, error) {
	log := om.Log
	orderid, err := hex.DecodeString(orderId)
	if err != nil {
		return nil, err
	}
	req := &pb.PayOrderReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: uint64(time.Now().UTC().Unix()),
		OrderId:   orderid,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.PayOrder(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)

	return rsp, nil
}

func (om *OrderManager) UsageAmount() (*pb.UsageAmountResp, error) {
	log := om.Log
	req := &pb.UsageAmountReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: uint64(time.Now().UTC().Unix()),
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.UsageAmount(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	return rsp, nil
}
