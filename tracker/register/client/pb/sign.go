package register_client_pb

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	util_bytes "github.com/samoslab/nebula/util/bytes"
)

func (self *VerifyContactEmailReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write([]byte(self.VerifyCode))
	return hasher.Sum(nil)
}

func (self *VerifyContactEmailReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *VerifyContactEmailReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *ResendVerifyCodeReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	return hasher.Sum(nil)
}

func (self *ResendVerifyCodeReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *ResendVerifyCodeReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *BuyPackageReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(util_bytes.FromUint64(uint64(self.PackageId)))
	hasher.Write(util_bytes.FromUint32(self.Quanlity))
	if self.CancelUnpaid {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (self *BuyPackageReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *BuyPackageReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *MyAllOrderReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	if self.OnlyNotExpired {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (self *MyAllOrderReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *MyAllOrderReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *OrderInfoReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.OrderId)
	return hasher.Sum(nil)
}

func (self *OrderInfoReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *OrderInfoReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *RemoveOrderReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.OrderId)
	return hasher.Sum(nil)
}

func (self *RemoveOrderReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *RemoveOrderReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *RechargeAddressReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	return hasher.Sum(nil)
}

func (self *RechargeAddressReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *RechargeAddressReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *PayOrderReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.OrderId)
	return hasher.Sum(nil)
}

func (self *PayOrderReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *PayOrderReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *UsageAmountReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	return hasher.Sum(nil)
}

func (self *UsageAmountReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *UsageAmountReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}
