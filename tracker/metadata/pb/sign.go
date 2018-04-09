package metadata_pb

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	util_bytes "github.com/samoslab/nebula/util/bytes"
)

func (self *MkFolderReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write([]byte(self.Path))
	for _, f := range self.Folder {
		hasher.Write([]byte(f))
	}
	return hasher.Sum(nil)
}

func (self *MkFolderReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *MkFolderReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *CheckFileExistReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write([]byte(self.FilePath))
	hasher.Write(self.FileHash)
	hasher.Write(util_bytes.FromUint64(self.FileSize))
	hasher.Write([]byte(self.FileName))
	hasher.Write(util_bytes.FromUint64(self.FileModTime))
	hasher.Write(self.FileData)
	if self.Interactive {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	if self.NewVersion {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (self *CheckFileExistReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *CheckFileExistReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *UploadFilePrepareReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.FileHash)
	hasher.Write(util_bytes.FromUint64(self.FileSize))
	for _, p := range self.Partition {
		for _, pi := range p.Piece {
			hasher.Write(pi.Hash)
			hasher.Write(util_bytes.FromUint32(pi.Size))
		}
	}
	return hasher.Sum(nil)
}

func (self *UploadFilePrepareReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *UploadFilePrepareReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *UploadFileDoneReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write([]byte(self.FilePath))
	hasher.Write(self.FileHash)
	hasher.Write(util_bytes.FromUint64(self.FileSize))
	hasher.Write([]byte(self.FileName))
	hasher.Write(util_bytes.FromUint64(self.FileModTime))
	for _, p := range self.Partition {
		for _, b := range p.Block {
			hasher.Write(b.Hash)
			hasher.Write(util_bytes.FromUint64(b.Size))
			hasher.Write(util_bytes.FromUint32(b.BlockSeq))
			if b.Checksum {
				hasher.Write([]byte{1})
			} else {
				hasher.Write([]byte{0})
			}
			for _, by := range b.StoreNodeId {
				hasher.Write(by)
			}
		}
	}
	if self.Interactive {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	if self.NewVersion {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (self *UploadFileDoneReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *UploadFileDoneReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *ListFilesReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write([]byte(self.Path))
	hasher.Write(util_bytes.FromUint32(self.PageSize))
	hasher.Write(util_bytes.FromUint32(self.PageNum))
	hasher.Write([]byte(self.SortType.String()))
	if self.AscOrder {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (self *ListFilesReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *ListFilesReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *RetrieveFileReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write(self.FileHash)
	hasher.Write(util_bytes.FromUint64(self.FileSize))
	return hasher.Sum(nil)
}

func (self *RetrieveFileReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *RetrieveFileReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}

func (self *RemoveReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(self.NodeId)
	hasher.Write(util_bytes.FromUint64(self.Timestamp))
	hasher.Write([]byte(self.Path))
	if self.Recursive {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (self *RemoveReq) SignReq(priKey *rsa.PrivateKey) (err error) {
	self.Sign, err = rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, self.hash())
	return
}

func (self *RemoveReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, self.hash(), self.Sign)
}
