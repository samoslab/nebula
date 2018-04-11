package metadata_pb

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	util_hash "github.com/samoslab/nebula/util/hash"
)

func TestMkFolderReq(t *testing.T) {
	req := MkFolderReq{NodeId: util_hash.Sha1([]byte("test-node-id")),
		Timestamp: uint64(time.Now().Unix()),
		Parent:    &FilePath{&FilePath_Path{"/folder1/folder2"}},
		Folder:    []string{"f1", "f2"}}
	priKey, err := rsa.GenerateKey(rand.Reader, 256*8)
	if err != nil {
		t.Errorf("failed")
	}
	pubKey := &priKey.PublicKey
	if req.SignReq(priKey) != nil {
		t.Errorf("failed")
	}
	if req.VerifySign(pubKey) != nil {
		t.Errorf("failed")
	}
	req = MkFolderReq{NodeId: util_hash.Sha1([]byte("test-node-id")),
		Timestamp: uint64(time.Now().Unix()),
		Parent:    &FilePath{&FilePath_Id{[]byte("parent-id")}},
		Folder:    []string{"f1", "f2"}}
	priKey, err = rsa.GenerateKey(rand.Reader, 256*8)
	if err != nil {
		t.Errorf("failed")
	}
	pubKey = &priKey.PublicKey
	if req.SignReq(priKey) != nil {
		t.Errorf("failed")
	}
	if req.VerifySign(pubKey) != nil {
		t.Errorf("failed")
	}
}
