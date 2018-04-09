package provider_pb

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"time"

	util_bytes "github.com/spolabs/nebula/util/bytes"
)

const timestamp_expired = 900
const method_store = "Store"
const method_retrieve = "Retrieve"
const method_get_fragment = "GetFragment"

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

func (self *GetFragmentReq) CheckAuth(publicKeyBytes []byte) error {
	return checkAuth(publicKeyBytes, method_get_fragment, self.Key, uint64(self.Size), self.Timestamp, "", self.Auth)
}
