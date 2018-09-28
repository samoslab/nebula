package register_provider_pb

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"math"

	util_bytes "github.com/samoslab/nebula/util/bytes"
)

func (self *RegisterReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.NodeIdEnc)
	hasher.Write(self.PublicKeyEnc)
	hasher.Write(self.EncryptKeyEnc)
	hasher.Write(self.WalletAddressEnc)
	hasher.Write(self.BillEmailEnc)
	hasher.Write(util_bytes.FromUint64(self.MainStorageVolume))
	hasher.Write(util_bytes.FromUint64(self.UpBandwidth))
	hasher.Write(util_bytes.FromUint64(self.DownBandwidth))
	hasher.Write(util_bytes.FromUint64(self.TestUpBandwidth))
	hasher.Write(util_bytes.FromUint64(self.TestDownBandwidth))
	hasher.Write(util_bytes.FromUint64(math.Float64bits(self.Availability)))
	hasher.Write(util_bytes.FromUint32(self.Port))
	hasher.Write(self.HostEnc)
	hasher.Write(self.DynamicDomainEnc)
	for _, val := range self.ExtraStorageVolume {
		hasher.Write(util_bytes.FromUint64(val))
	}
	hasher.Write(self.PublicKeyHash)
	if self.ConfirmInner {
		hasher.Write([]byte{1})
	}
	return hasher.Sum(nil)
}

func (self *RegisterReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *RegisterReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *VerifyBillEmailReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write([]byte(self.VerifyCode))
	return hasher.Sum(nil)
}

func (self *VerifyBillEmailReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *VerifyBillEmailReq) VerifySign(pubKey *rsa.PublicKey) error {
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

func (self *AddExtraStorageReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	return hasher.Sum(nil)
}

func (self *AddExtraStorageReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *AddExtraStorageReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *GetTrackerServerReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	return hasher.Sum(nil)
}

func (self *GetTrackerServerReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *GetTrackerServerReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *GetCollectorServerReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	return hasher.Sum(nil)
}

func (self *GetCollectorServerReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *GetCollectorServerReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *RefreshIpReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(util_bytes.FromUint32(self.Port))
	return hasher.Sum(nil)
}

func (self *RefreshIpReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *RefreshIpReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *SwitchPrivateReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	return hasher.Sum(nil)
}

func (self *SwitchPrivateReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *SwitchPrivateReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *SwitchPublicReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.PublicKeyHash)
	hasher.Write(util_bytes.FromUint32(self.Port))
	hasher.Write(self.HostEnc)
	hasher.Write(self.DynamicDomainEnc)
	return hasher.Sum(nil)
}

func (self *SwitchPublicReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *SwitchPublicReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}
func (self *PrivateAliveReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(util_bytes.FromUint32(self.Version))
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(util_bytes.FromUint64(self.Total))
	hasher.Write(util_bytes.FromUint64(self.MaxFileSize))
	return hasher.Sum(nil)
}

func (self *PrivateAliveReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *PrivateAliveReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}
