package task_pb

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	util_bytes "github.com/samoslab/nebula/util/bytes"
)

func (self *TaskListReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(util_bytes.FromUint32(self.Category))
	return hasher.Sum(nil)
}

func (self *TaskListReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *TaskListReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *GetOppositeInfoReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.TaskId)
	return hasher.Sum(nil)
}

func (self *GetOppositeInfoReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *GetOppositeInfoReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *GetProveInfoReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.TaskId)
	return hasher.Sum(nil)
}

func (self *GetProveInfoReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *GetProveInfoReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *FinishProveReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.TaskId)
	hasher.Write(self.ProofId)
	hasher.Write(util_bytes.FromUint64(self.FinishedTime))
	hasher.Write(self.Result)
	if len(self.Remark) > 0 {
		hasher.Write([]byte(self.Remark))
	}
	return hasher.Sum(nil)
}

func (self *FinishProveReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *FinishProveReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *FinishTaskReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.TaskId)
	hasher.Write(util_bytes.FromUint64(self.FinishedTime))
	if self.Success {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	if len(self.Remark) > 0 {
		hasher.Write([]byte(self.Remark))
	}
	return hasher.Sum(nil)
}

func (self *FinishTaskReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *FinishTaskReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}
