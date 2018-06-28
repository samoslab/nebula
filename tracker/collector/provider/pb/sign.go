package collector_provider_pb

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	util_bytes "github.com/samoslab/nebula/util/bytes"
)

var byte_slice_true = []byte{1}
var byte_slice_false = []byte{0}

func (self *Batch) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	for _, al := range self.ActionLog {
		hasher.Write(util_bytes.FromUint32(al.Type))
		hasher.Write([]byte(al.Ticket))
		if al.Success {
			hasher.Write(byte_slice_true)
		} else {
			hasher.Write(byte_slice_false)
		}
		hasher.Write(al.FileHash)
		hasher.Write(util_bytes.FromUint64(al.FileSize))
		hasher.Write(al.BlockHash)
		hasher.Write(util_bytes.FromUint64(al.BlockSize))
		hasher.Write(util_bytes.FromUint64(al.BeginTime))
		hasher.Write(util_bytes.FromUint64(al.EndTime))
		hasher.Write(util_bytes.FromUint64(al.TransportSize))
		hasher.Write([]byte(al.Info))
	}
	return hasher.Sum(nil)
}

func (self *Batch) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *Batch) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}
