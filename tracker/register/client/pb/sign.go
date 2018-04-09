package register_client_pb

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	util_bytes "github.com/spolabs/nebula/util/bytes"
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
