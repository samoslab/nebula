package metadata_pb

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"

	util_bytes "github.com/spolabs/nebula/util/bytes"
)

func (req *MkFolderReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write([]byte(req.Path))
	for _, f := range req.Folder {
		hasher.Write([]byte(f))
	}
	return hasher.Sum(nil)
}

func (req *MkFolderReq) SignReq(priKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, req.hash())
}

func (req *MkFolderReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, req.hash(), req.Sign)
}

func (req *CheckFileExistReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write([]byte(req.FilePath))
	hasher.Write(req.FileHash)
	hasher.Write(util_bytes.FromUint64(req.FileSize))
	hasher.Write([]byte(req.FileName))
	hasher.Write(util_bytes.FromUint64(req.FileModTime))
	hasher.Write(req.FileData)
	if req.Interactive {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	if req.NewVersion {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (req *CheckFileExistReq) SignReq(priKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, req.hash())
}

func (req *CheckFileExistReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, req.hash(), req.Sign)
}

func (req *UploadFilePrepareReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write(req.FileHash)
	hasher.Write(util_bytes.FromUint64(req.FileSize))
	for _, p := range req.Partition {
		for _, pi := range p.Piece {
			hasher.Write(pi.Hash)
			hasher.Write(util_bytes.FromUint32(pi.Size))
		}
	}
	return hasher.Sum(nil)
}

func (req *UploadFilePrepareReq) SignReq(priKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, req.hash())
}

func (req *UploadFilePrepareReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, req.hash(), req.Sign)
}

func (req *UploadFileDoneReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write([]byte(req.FilePath))
	hasher.Write(req.FileHash)
	hasher.Write(util_bytes.FromUint64(req.FileSize))
	hasher.Write([]byte(req.FileName))
	hasher.Write(util_bytes.FromUint64(req.FileModTime))
	for _, p := range req.Partition {
		for _, b := range p.Block {
			hasher.Write(b.Hash)
			hasher.Write(util_bytes.FromUint32(b.Size))
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
	if req.Interactive {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	if req.NewVersion {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (req *UploadFileDoneReq) SignReq(priKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, req.hash())
}

func (req *UploadFileDoneReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, req.hash(), req.Sign)
}

func (req *ListFilesReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write([]byte(req.Path))
	hasher.Write(util_bytes.FromUint32(req.PageSize))
	hasher.Write(util_bytes.FromUint32(req.PageNum))
	hasher.Write([]byte(req.SortType.String()))
	if req.AscOrder {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (req *ListFilesReq) SignReq(priKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, req.hash())
}

func (req *ListFilesReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, req.hash(), req.Sign)
}

func (req *RetrieveFileReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write(req.FileHash)
	hasher.Write(util_bytes.FromUint64(req.FileSize))
	return hasher.Sum(nil)
}

func (req *RetrieveFileReq) SignReq(priKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, req.hash())
}

func (req *RetrieveFileReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, req.hash(), req.Sign)
}

func (req *RemoveReq) hash() []byte {
	hasher := sha256.New()
	hasher.Write(req.NodeId)
	hasher.Write(util_bytes.FromUint64(req.Timestamp))
	hasher.Write([]byte(req.Path))
	if req.Recursive {
		hasher.Write([]byte{1})
	} else {
		hasher.Write([]byte{0})
	}
	return hasher.Sum(nil)
}

func (req *RemoveReq) SignReq(priKey *rsa.PrivateKey) ([]byte, error) {
	return rsa.SignPKCS1v15(rand.Reader, priKey, crypto.SHA256, req.hash())
}

func (req *RemoveReq) VerifySign(pubKey *rsa.PublicKey) error {
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, req.hash(), req.Sign)
}
