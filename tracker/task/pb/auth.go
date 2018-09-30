package task_pb

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

func (self *TaskListResp) genAuth(publicKeyBytes []byte) []byte {
	hash := hmac.New(sha256.New, publicKeyBytes)
	hash.Write(util_bytes.FromUint64(self.Timestamp))
	if len(self.Task) > 0 {
		for _, task := range self.Task {
			hash.Write(task.Id)
			hash.Write(util_bytes.FromUint64(task.Creation))
			hash.Write([]byte(task.Type.String()))
			hash.Write(task.FileId)
			hash.Write(task.FileHash)
			hash.Write(util_bytes.FromUint64(task.FileSize))
			hash.Write(task.BlockHash)
			hash.Write(util_bytes.FromUint64(task.BlockSize))
			if len(task.OppositeId) > 0 {
				for _, oid := range task.OppositeId {
					hash.Write([]byte(oid))
				}
			}
			if len(task.ProofId) > 0 {
				hash.Write(task.ProofId)
			}
		}
	}
	return hash.Sum(nil)
}

func (self *TaskListResp) GenAuth(publicKeyBytes []byte) {
	self.Auth = self.genAuth(publicKeyBytes)
}

func (self *TaskListResp) CheckAuth(publicKeyBytes []byte) error {
	interval := time.Now().Unix() - int64(self.Timestamp)
	if interval > timestamp_expired || interval < timestamp_ahead {
		return errors.New("auth expired")
	}
	if len(self.Auth) > 0 && bytes.Equal(self.Auth, self.genAuth(publicKeyBytes)) {
		return nil
	}
	return errors.New("auth verify failed")
}
