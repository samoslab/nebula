package order

import (
	"context"
	"crypto/rsa"
	"encoding/hex"
	"fmt"

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

type Order struct {
	Id          string      `json:"id,omitempty"`
	Creation    uint64      `json:"creation,omitempty"`
	PackageId   int64       `json:"packageId,omitempty"`
	Package     *pb.Package `json:"package,omitempty"`
	Quanlity    uint32      `json:"quanlity,omitempty"`
	TotalAmount uint64      `json:"totalAmount,omitempty"`
	Upgraded    bool        `json:"upgraded,omitempty"`
	Discount    string      `json:"discount,omitempty"`
	Volume      uint32      `json:"volume,omitempty"`
	Netflow     uint32      `json:"netflow,omitempty"`
	UpNetflow   uint32      `json:"upNetflow,omitempty"`
	DownNetflow uint32      `json:"downNetflow,omitempty"`
	ValidDays   uint32      `json:"validDays,omitempty"`
	StartTime   uint64      `json:"startTime,omitempty"`
	EndTime     uint64      `json:"endTime,omitempty"`
	Paid        bool        `json:"paid,omitempty"`
	PayTime     uint64      `json:"payTime,omitempty"`
	Remark      string      `json:"remark,omitempty"`
}

func NewOrderFromPbOrder(o *pb.Order) *Order {
	return &Order{
		Id:          hex.EncodeToString(o.Id),
		Creation:    o.Creation,
		PackageId:   o.PackageId,
		Package:     o.Package,
		Quanlity:    o.Quanlity,
		TotalAmount: o.TotalAmount,
		Upgraded:    o.Upgraded,
		Discount:    o.Discount,
		Volume:      o.Volume,
		Netflow:     o.Netflow,
		UpNetflow:   o.UpNetflow,
		DownNetflow: o.DownNetflow,
		ValidDays:   o.ValidDays,
		StartTime:   o.StartTime,
		EndTime:     o.EndTime,
		Paid:        o.Paid,
		PayTime:     o.PayTime,
		Remark:      o.Remark,
	}
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
		return nil, common.StatusErrFromError(err)
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
		return nil, common.StatusErrFromError(err)
	}
	log.Infof("%+v", rsp)
	return rsp.GetPackage(), nil
}

func (om *OrderManager) BuyPackage(id uint64, canceled bool, quanlity uint32) (*Order, error) {
	log := om.Log
	req := &pb.BuyPackageReq{
		Version:      common.Version,
		NodeId:       om.NodeId,
		Timestamp:    common.Now(),
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
		return nil, common.StatusErrFromError(err)
	}
	if rsp.GetCode() != 0 {
		return nil, fmt.Errorf("buy packge error %s", rsp.GetErrMsg())
	}
	log.Infof("%+v", rsp)
	return NewOrderFromPbOrder(rsp.GetOrder()), nil
}

func (om *OrderManager) DiscountPackage(id uint64) (map[uint32]string, error) {
	log := om.Log
	req := &pb.PackageDiscountReq{
		Version:   common.Version,
		PackageId: int64(id),
	}
	rsp, err := om.orderClient.PackageDiscount(context.Background(), req)
	if err != nil {
		return nil, common.StatusErrFromError(err)
	}
	log.Infof("%+v", rsp)
	return rsp.GetDiscount(), nil
}

func (om *OrderManager) MyAllOrders(expired bool) ([]*Order, error) {
	log := om.Log
	req := &pb.MyAllOrderReq{
		Version:        common.Version,
		NodeId:         om.NodeId,
		Timestamp:      common.Now(),
		OnlyNotExpired: expired,
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.MyAllOrder(context.Background(), req)
	if err != nil {
		return nil, common.StatusErrFromError(err)
	}
	log.Infof("%+v", rsp)
	allOrder := []*Order{}
	for _, o := range rsp.GetMyAllOrder() {
		allOrder = append(allOrder, NewOrderFromPbOrder(o))
	}
	return allOrder, nil
}

func (om *OrderManager) GetOrderInfo(orderId string) (*Order, error) {
	log := om.Log
	orderid, err := hex.DecodeString(orderId)
	if err != nil {
		return nil, err
	}
	req := &pb.OrderInfoReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: common.Now(),
		OrderId:   orderid,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}

	rsp, err := om.orderClient.OrderInfo(context.Background(), req)
	if err != nil {
		return nil, common.StatusErrFromError(err)
	}
	log.Infof("%+v", rsp)

	return NewOrderFromPbOrder(rsp.GetOrder()), nil
}

type AddressBalance struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

func (om *OrderManager) RechargeAddress() (*AddressBalance, error) {
	log := om.Log
	req := &pb.RechargeAddressReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: common.Now(),
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.RechargeAddress(context.Background(), req)
	if err != nil {
		return nil, common.StatusErrFromError(err)
	}
	log.Infof("%+v", rsp)

	if rsp.GetCode() != 0 {
		return nil, fmt.Errorf("recharge error %v", rsp.GetErrMsg())
	}
	ab := &AddressBalance{
		Address: rsp.GetRechargeAddress(),
		Balance: rsp.GetBalance(),
	}

	return ab, nil
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
		Timestamp: common.Now(),
		OrderId:   orderid,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.PayOrder(context.Background(), req)
	if err != nil {
		return nil, common.StatusErrFromError(err)
	}
	log.Infof("%+v", rsp)

	return rsp, nil
}

func (om *OrderManager) UsageAmount() (*pb.UsageAmountResp, error) {
	log := om.Log
	req := &pb.UsageAmountReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: common.Now(),
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.UsageAmount(context.Background(), req)
	if err != nil {
		return nil, common.StatusErrFromError(err)
	}
	log.Infof("%+v", rsp)
	return rsp, nil
}
