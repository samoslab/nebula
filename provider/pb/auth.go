package provider_pb

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"time"

	util_bytes "github.com/samoslab/nebula/util/bytes"
)

const timestamp_expired = 900
const method_store = "Store"
const method_retrieve = "Retrieve"
const method_get_fragment = "GetFragment"
const method_remove = "Remove"

func genAuth(publicKeyBytes []byte, method string, key []byte, fileSize uint64, timestamp uint64, ticket string) []byte {
	hash := hmac.New(sha256.New, publicKeyBytes)
	hash.Write([]byte(method))
	hash.Write(key)
	hash.Write(util_bytes.FromUint64(fileSize))
	hash.Write(util_bytes.FromUint64(timestamp))
	hash.Write([]byte(ticket))
	return hash.Sum(nil)
}

func checkAuth(publicKeyBytes []byte, method string, key []byte, fileSize uint64, timestamp uint64, ticket string, auth []byte) error {
	if uint64(time.Now().Unix())-timestamp > timestamp_expired {
		return errors.New("auth expired")
	}
	if len(key) == 0 {
		return errors.New("wrong key")
	}
	if len(auth) > 0 && bytes.Equal(auth, genAuth(publicKeyBytes, method, key, fileSize, timestamp, ticket)) {
		return nil
	}
	return errors.New("auth verify failed")
}

func (self *StoreReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_store, self.Key, self.FileSize, self.Timestamp, self.Ticket, self.Auth)
}

func (self *RetrieveReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_retrieve, self.Key, self.FileSize, self.Timestamp, self.Ticket, self.Auth)
}

func (self *RemoveReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_remove, self.Key, self.FileSize, self.Timestamp, "", self.Auth)
}

func (self *GetFragmentReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_get_fragment, self.Key, uint64(self.Size), self.Timestamp, "", self.Auth)
}

func GenRetrieveAuth(publicKeyBytes []byte, hash []byte, size uint64, timestamp uint64, ticket string) []byte {
	return genAuth(publicKeyBytes, method_retrieve, hash, size, timestamp, ticket)
}

func GenStoreAuth(publicKeyBytes []byte, hash []byte, size uint64, timestamp uint64, ticket string) []byte {
	return genAuth(publicKeyBytes, method_store, hash, size, timestamp, ticket)
}

func GenGetFragmentAuth(publicKeyBytes []byte, hash []byte, size uint64, timestamp uint64) []byte {
	return genAuth(publicKeyBytes, method_get_fragment, hash, size, timestamp, "")
}

func GenRemoveAuth(publicKeyBytes []byte, hash []byte, size uint64, timestamp uint64) []byte {
	return genAuth(publicKeyBytes, method_remove, hash, size, timestamp, "")
}
