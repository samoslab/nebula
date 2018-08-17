package order

import (
	"context"
	"crypto/rsa"
	"encoding/hex"
	"strconv"

	"github.com/samoslab/nebula/client/common"
	pb "github.com/samoslab/nebula/tracker/register/client/pb"
	rsalong "github.com/samoslab/nebula/util/rsa"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

// OrderManager order manager
type OrderManager struct {
	NodeId      []byte
	privateKey  *rsa.PrivateKey
	Log         logrus.FieldLogger
	orderClient pb.OrderServiceClient
}

// Package descript package, cannot using int64 because js int max is 2^53-1
type Package struct {
	Id          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Price       uint64 `json:"price,omitempty"`
	Remark      string `json:"remark,omitempty"`
	Volume      uint32 `json:"volume,omitempty"`
	Netflow     uint32 `json:"netflow,omitempty"`
	UpNetflow   uint32 `json:"upNetflow,omitempty"`
	ValidDays   uint32 `json:"validDays,omitempty"`
	DownNetflow uint32 `json:"downNetflow,omitempty"`
}

// Order order infos that return to front-end
type Order struct {
	Id          string   `json:"id,omitempty"`
	Creation    uint64   `json:"creation,omitempty"`
	PackageId   string   `json:"packageId,omitempty"`
	Package     *Package `json:"package,omitempty"`
	Quanlity    uint32   `json:"quanlity,omitempty"`
	TotalAmount uint64   `json:"totalAmount,omitempty"`
	Upgraded    bool     `json:"upgraded,omitempty"`
	Discount    string   `json:"discount,omitempty"`
	Volume      uint32   `json:"volume,omitempty"`
	Netflow     uint32   `json:"netflow,omitempty"`
	UpNetflow   uint32   `json:"upNetflow,omitempty"`
	DownNetflow uint32   `json:"downNetflow,omitempty"`
	ValidDays   uint32   `json:"validDays,omitempty"`
	StartTime   uint64   `json:"startTime,omitempty"`
	EndTime     uint64   `json:"endTime,omitempty"`
	Paid        bool     `json:"paid,omitempty"`
	PayTime     uint64   `json:"payTime,omitempty"`
	Remark      string   `json:"remark,omitempty"`
}

func newPackageFromPbPackage(p *pb.Package) *Package {
	return &Package{
		Name:        p.Name,
		Price:       p.Price,
		Volume:      p.Volume,
		Netflow:     p.Netflow,
		ValidDays:   p.ValidDays,
		UpNetflow:   p.UpNetflow,
		DownNetflow: p.DownNetflow,
		Id:          strconv.FormatUint(uint64(p.Id), 10),
	}
}

// NewOrderFromPbOrder create Order from protobuf Order, diff is Id
func NewOrderFromPbOrder(o *pb.Order) *Order {
	return &Order{
		Paid:        o.Paid,
		Volume:      o.Volume,
		Remark:      o.Remark,
		EndTime:     o.EndTime,
		Netflow:     o.Netflow,
		PayTime:     o.PayTime,
		Creation:    o.Creation,
		Quanlity:    o.Quanlity,
		Upgraded:    o.Upgraded,
		Discount:    o.Discount,
		UpNetflow:   o.UpNetflow,
		ValidDays:   o.ValidDays,
		StartTime:   o.StartTime,
		TotalAmount: o.TotalAmount,
		DownNetflow: o.DownNetflow,
		Id:          hex.EncodeToString(o.Id),
		Package:     newPackageFromPbPackage(o.Package),
		PackageId:   strconv.FormatUint(uint64(o.PackageId), 10),
	}
}

// NewOrderManager create order manager ,only communicate with tracker server
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
		NodeId:      nodeId,
		privateKey:  privateKey,
	}
}

// GetAllPackages return all packges
func (om *OrderManager) GetAllPackages() ([]*Package, error) {
	log := om.Log
	req := &pb.AllPackageReq{
		Version: common.Version,
	}
	rsp, err := om.orderClient.AllPackage(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	packages := []*Package{}
	for _, p := range rsp.GetAllPackage() {
		packages = append(packages, newPackageFromPbPackage(p))
	}
	return packages, nil
}

// GetPackageInfo returns package by package id
func (om *OrderManager) GetPackageInfo(pid string) (*Package, error) {
	log := om.Log
	id, err := strconv.ParseUint(pid, 10, 0)
	if err != nil {
		return nil, err
	}
	req := &pb.PackageInfoReq{
		PackageId: int64(id),
		Version:   common.Version,
	}
	rsp, err := om.orderClient.PackageInfo(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	return newPackageFromPbPackage(rsp.GetPackage()), nil
}

// BuyPackage buy package
func (om *OrderManager) BuyPackage(pid string, canceled bool, quanlity uint32) (*Order, error) {
	log := om.Log
	id, err := strconv.ParseUint(pid, 10, 0)
	if err != nil {
		return nil, err
	}
	req := &pb.BuyPackageReq{
		Quanlity:     quanlity,
		CancelUnpaid: canceled,
		NodeId:       om.NodeId,
		PackageId:    int64(id),
		Timestamp:    common.Now(),
		Version:      common.Version,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.BuyPackage(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}
	return NewOrderFromPbOrder(rsp.GetOrder()), nil
}

// DiscountPackage package discount
func (om *OrderManager) DiscountPackage(pid string) (map[uint32]string, error) {
	log := om.Log
	id, err := strconv.ParseUint(pid, 10, 0)
	if err != nil {
		return nil, err
	}
	req := &pb.PackageDiscountReq{
		PackageId: int64(id),
		Version:   common.Version,
	}
	rsp, err := om.orderClient.PackageDiscount(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("%+v", rsp)
	return rsp.GetDiscount(), nil
}

// MyAllOrders returns all orders of current client
func (om *OrderManager) MyAllOrders(expired bool) ([]*Order, error) {
	log := om.Log
	req := &pb.MyAllOrderReq{
		OnlyNotExpired: expired,
		NodeId:         om.NodeId,
		Timestamp:      common.Now(),
		Version:        common.Version,
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.MyAllOrder(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("All orders %+v", rsp)
	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}
	allOrder := []*Order{}
	for _, o := range rsp.GetMyAllOrder() {
		allOrder = append(allOrder, NewOrderFromPbOrder(o))
	}
	return allOrder, nil
}

// GetOrderInfo returns order info by order id
func (om *OrderManager) GetOrderInfo(orderId string) (*Order, error) {
	log := om.Log
	orderid, err := hex.DecodeString(orderId)
	if err != nil {
		return nil, err
	}
	req := &pb.OrderInfoReq{
		OrderId:   orderid,
		NodeId:    om.NodeId,
		Timestamp: common.Now(),
		Version:   common.Version,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}

	rsp, err := om.orderClient.OrderInfo(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("Get order info %+v", rsp)
	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}

	return NewOrderFromPbOrder(rsp.GetOrder()), nil
}

// AddressBalance balance of current client
type AddressBalance struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"`
}

// RechargeAddress get balance of current client
func (om *OrderManager) RechargeAddress() (*AddressBalance, error) {
	log := om.Log
	req := &pb.RechargeAddressReq{
		NodeId:    om.NodeId,
		Timestamp: common.Now(),
		Version:   common.Version,
	}
	err := req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.RechargeAddress(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("Recharge address %+v", rsp)

	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}
	address, err := rsalong.DecryptLong(om.privateKey, rsp.GetRechargeAddressEnc(), 256)
	if err != nil {
		return nil, err
	}
	ab := &AddressBalance{
		Address: string(address),
		Balance: rsp.GetBalance(),
	}

	return ab, nil
}

// PayOrder pay order
func (om *OrderManager) PayOrdor(orderId string) (*pb.PayOrderResp, error) {
	log := om.Log
	orderid, err := hex.DecodeString(orderId)
	if err != nil {
		return nil, err
	}
	req := &pb.PayOrderReq{
		OrderId:   orderid,
		NodeId:    om.NodeId,
		Timestamp: common.Now(),
		Version:   common.Version,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.PayOrder(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("Pay order %+v", rsp)
	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}

	return rsp, nil
}

// UsageAmount usage amount data store
type UsageAmount struct {
	UsageVolume      uint32 `json:"usageVolume"`
	UsageNetflow     uint32 `json:"usageNetflow"`
	UsageUpNetflow   uint32 `json:"usageUpNetflow"`
	UsageDownNetflow uint32 `json:"usageDownNetflow"`
	Volume           uint32 `json:"volume,omitempty"`
	Netflow          uint32 `json:"netflow,omitempty"`
	EndTime          uint64 `json:"endTime,omitempty"`
	UpNetflow        uint32 `json:"upNetflow,omitempty"`
	PackageId        int64  `json:"packageId,omitempty"`
	DownNetflow      uint32 `json:"downNetflow,omitempty"`
}

// NewUsageAmount new usage amount
func NewUsageAmount(rsp *pb.UsageAmountResp) *UsageAmount {
	return &UsageAmount{
		Volume:           rsp.Volume,
		EndTime:          rsp.EndTime,
		Netflow:          rsp.Netflow,
		UpNetflow:        rsp.UpNetflow,
		PackageId:        rsp.PackageId,
		DownNetflow:      rsp.DownNetflow,
		UsageVolume:      rsp.UsageVolume,
		UsageNetflow:     rsp.UsageNetflow,
		UsageUpNetflow:   rsp.UsageUpNetflow,
		UsageDownNetflow: rsp.UsageDownNetflow,
	}
}

// UsageAmount usage amount of package
func (om *OrderManager) UsageAmount() (*UsageAmount, error) {
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
		return nil, err
	}
	log.Infof("Usage amount %+v", rsp)
	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}
	return NewUsageAmount(rsp), nil
}

// RemoveOrder remove order
func (om *OrderManager) RemoveOrdor(orderId string) (*pb.RemoveOrderResp, error) {
	log := om.Log
	orderid, err := hex.DecodeString(orderId)
	if err != nil {
		return nil, err
	}
	req := &pb.RemoveOrderReq{
		Version:   common.Version,
		NodeId:    om.NodeId,
		Timestamp: common.Now(),
		OrderId:   orderid,
	}
	err = req.SignReq(om.privateKey)
	if err != nil {
		return nil, err
	}
	rsp, err := om.orderClient.RemoveOrder(context.Background(), req)
	if err != nil {
		return nil, err
	}
	log.Infof("Remove order %+v", rsp)
	if rsp.GetCode() != 0 {
		return nil, common.NewStatusErr(rsp.Code, rsp.ErrMsg)
	}

	return rsp, nil
}
