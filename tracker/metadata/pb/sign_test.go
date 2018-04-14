package metadata_pb

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
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

func TestCheckFileExistReq(t *testing.T) {
	priKey, err := rsa.GenerateKey(rand.Reader, 256*8)
	if err != nil {
		t.Errorf("failed")
	}
	pubKey := &priKey.PublicKey
	pubKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)
	nodeId := util_hash.Sha1(pubKeyBytes)
	ts := uint64(time.Now().Unix())
	pathStr := "/folder1/folder2"
	path := &FilePath{&FilePath_Path{pathStr}}
	hash := util_hash.Sha1([]byte("test-file"))
	size := uint64(98234)
	name := "file.txt"
	req := CheckFileExistReq{NodeId: nodeId,
		Timestamp:   ts,
		Parent:      path,
		FileHash:    hash,
		FileSize:    size,
		FileName:    name,
		FileModTime: ts - 1000,
		FileData:    hash,
		Interactive: true,
		NewVersion:  false}
	req.SignReq(priKey)
	if req.SignReq(priKey) != nil {
		t.Errorf("failed")
	}
	if req.VerifySign(pubKey) != nil {
		t.Errorf("failed")
	}
	req = CheckFileExistReq{NodeId: nodeId,
		Timestamp:   ts,
		Parent:      path,
		FileHash:    hash,
		FileSize:    size,
		FileName:    name,
		FileModTime: ts - 1000,
		FileData:    nil,
		Interactive: true,
		NewVersion:  false}
	req.SignReq(priKey)
	if req.SignReq(priKey) != nil {
		t.Errorf("failed")
	}
	if req.VerifySign(pubKey) != nil {
		t.Errorf("failed")
	}
}
