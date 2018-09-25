package provider_pb

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"time"

	util_bytes "github.com/samoslab/nebula/util/bytes"
)

const timestamp_expired = 1800
const timestamp_ahead = -300

const method_store = "Store"
const method_retrieve = "Retrieve"
const method_get_fragment = "GetFragment"
const method_remove = "Remove"

func genAuth(publicKeyBytes []byte, method string, fileKey []byte, fileSize uint64, blockKey []byte, blockSize uint64, timestamp uint64, ticket string, fromProvider bool) []byte {
	if len(blockKey) == 0 {
		blockKey = fileKey
	}
	if blockSize == 0 {
		blockSize = fileSize
	}
	hash := hmac.New(sha256.New, publicKeyBytes)
	hash.Write([]byte(method))
	hash.Write(fileKey)
	hash.Write(util_bytes.FromUint64(fileSize))
	hash.Write(blockKey)
	hash.Write(util_bytes.FromUint64(blockSize))
	hash.Write(util_bytes.FromUint64(timestamp))
	hash.Write([]byte(ticket))
	if fromProvider {
		hash.Write([]byte{1})
	}
	return hash.Sum(nil)
}

func checkAuth(publicKeyBytes []byte, method string, fileKey []byte, fileSize uint64, blockKey []byte, blockSize uint64, timestamp uint64, ticket string, auth []byte, fromProvider bool) error {
	interval := time.Now().Unix() - int64(timestamp)
	if interval > timestamp_expired || interval < timestamp_ahead {
		return errors.New("auth expired")
	}
	if len(blockKey) == 0 {
		return errors.New("wrong key")
	}
	if len(auth) > 0 && bytes.Equal(auth, genAuth(publicKeyBytes, method, fileKey, fileSize, blockKey, blockSize, timestamp, ticket, fromProvider)) {
		return nil
	}
	return errors.New("auth verify failed")
}

func (self *StoreReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_store, self.FileKey, self.FileSize, self.BlockKey, self.BlockSize, self.Timestamp, self.Ticket, self.Auth, self.FromProvider)
}

func (self *RetrieveReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_retrieve, self.FileKey, self.FileSize, self.BlockKey, self.BlockSize, self.Timestamp, self.Ticket, self.Auth, self.FromProvider)
}

func (self *RemoveReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_remove, nil, 0, self.Key, self.Size, self.Timestamp, "", self.Auth, false)
}

func (self *GetFragmentReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_get_fragment, nil, 0, self.Key, uint64(self.Size), self.Timestamp, "", self.Auth, false)
}

func GenRetrieveAuth(publicKeyBytes []byte, fileKey []byte, fileSize uint64, blockKey []byte, blockSize uint64, timestamp uint64, ticket string, fromProvider bool) []byte {
	return genAuth(publicKeyBytes, method_retrieve, fileKey, fileSize, blockKey, blockSize, timestamp, ticket, fromProvider)
}

func GenStoreAuth(publicKeyBytes []byte, fileKey []byte, fileSize uint64, blockKey []byte, blockSize uint64, timestamp uint64, ticket string, fromProvider bool) []byte {
	return genAuth(publicKeyBytes, method_store, fileKey, fileSize, blockKey, blockSize, timestamp, ticket, fromProvider)
}

func GenGetFragmentAuth(publicKeyBytes []byte, hash []byte, size uint32, timestamp uint64) []byte {
	return genAuth(publicKeyBytes, method_get_fragment, nil, 0, hash, uint64(size), timestamp, "", false)
}

func GenRemoveAuth(publicKeyBytes []byte, hash []byte, size uint64, timestamp uint64) []byte {
	return genAuth(publicKeyBytes, method_remove, nil, 0, hash, size, timestamp, "", false)
}

func (self *StoreReq) GenAuth(publicKeyBytes []byte) {
	self.Auth = genAuth(publicKeyBytes, method_store, self.FileKey, self.FileSize, self.BlockKey, self.BlockSize, self.Timestamp, self.Ticket, self.FromProvider)
}
func (self *RetrieveReq) GenAuth(publicKeyBytes []byte) {
	self.Auth = genAuth(publicKeyBytes, method_retrieve, self.FileKey, self.FileSize, self.BlockKey, self.BlockSize, self.Timestamp, self.Ticket, self.FromProvider)
}
func (self *RemoveReq) GenAuth(publicKeyBytes []byte) {
	self.Auth = genAuth(publicKeyBytes, method_remove, nil, 0, self.Key, self.Size, self.Timestamp, "", false)
}
func (self *GetFragmentReq) GenAuth(publicKeyBytes []byte) {
	self.Auth = genAuth(publicKeyBytes, method_get_fragment, nil, 0, self.Key, uint64(self.Size), self.Timestamp, "", false)
}

func (self *CheckAvailableReq) genAuth(publicKeyBytes []byte) []byte {
	hash := hmac.New(sha256.New, publicKeyBytes)
	hash.Write(util_bytes.FromUint64(self.Timestamp))
	return hash.Sum(nil)
}

func (self *CheckAvailableReq) GenAuth(publicKeyBytes []byte) {
	self.Auth = self.genAuth(publicKeyBytes)
}

func (self *CheckAvailableReq) CheckAuth(publicKeyBytes []byte) error {
	interval := time.Now().Unix() - int64(self.Timestamp)
	if interval > timestamp_expired || interval < timestamp_ahead {
		return errors.New("auth expired")
	}
	if len(self.Auth) > 0 && bytes.Equal(self.Auth, self.genAuth(publicKeyBytes)) {
		return nil
	}
	return errors.New("auth verify failed")
}

func (self *CheckFileReq) genAuth(publicKeyBytes []byte) []byte {
	hash := hmac.New(sha256.New, publicKeyBytes)
	hash.Write(util_bytes.FromUint64(self.Timestamp))
	hash.Write(self.Key)
	hash.Write(util_bytes.FromUint64(self.Size))
	hash.Write(util_bytes.FromUint32(self.ChunkSize))
	for k, v := range self.ChunkSeq {
		hash.Write(util_bytes.FromUint32(k))
		hash.Write(v)
	}
	return hash.Sum(nil)
}

func (self *CheckFileReq) CheckAuth(publicKeyBytes []byte) error {
	interval := time.Now().Unix() - int64(self.Timestamp)
	if interval > timestamp_expired || interval < timestamp_ahead {
		return errors.New("auth expired")
	}
	if len(self.Auth) > 0 && bytes.Equal(self.Auth, self.genAuth(publicKeyBytes)) {
		return nil
	}
	return errors.New("auth verify failed")
}

func (self *CheckFileReq) GenAuth(publicKeyBytes []byte) {
	self.Auth = self.genAuth(publicKeyBytes)
}
