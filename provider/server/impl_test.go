package server

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/spolabs/nebula/provider/node"
	util_bytes "github.com/spolabs/nebula/util/bytes"
)

func TestCheckAuth(t *testing.T) {
	bytes, err := hex.DecodeString("3082010a02820101009fd7cc531bb184ae3c082ffc3dad1f23d7013e3ed17bc8bc4c5abcd4642b9085846356c928fde4fd6289cee4c90f0b4dad42cf1ba391724564d0d410c95ee4430ade835fc2128f2698d8fbaadcc02d639b65aebb7947e25aecf38e0db51a689a62d29a078d49257d006719ab4556c70a9f2140a16a8a54c47e0bc28265558309dbe1fed5e8a47a9b43d51977b1d4a4bfe7e9ef890ce66f66f75dcf9f8a9ce19cdf1a780115d3127fae63e3696faf1049c443b8cac59a0e00e504320bcd72334a3bd5cb95d4fec1a1e0798225281c5b943683d95ea8f8d965be0369b75f6b521c18fbc80129995f947557fd02442dd8732e433be64f6894fd8d4ae316a6dddb510203010001")
	if err != nil {
		t.Error(err)
	}
	no := &node.Node{PubKeyBytes: bytes}
	ps := &ProviderServer{Node: no}
	method := "Store"
	key, _ := hex.DecodeString("c925852911756e6d4b14b425188f5cf67d1d3cfc")
	fileSize := uint64(1293847)
	timestamp := uint64(time.Now().Unix())
	hash := hmac.New(sha256.New, bytes)
	hash.Write([]byte(method))
	hash.Write(key)
	hash.Write(util_bytes.FromUint64(fileSize))
	hash.Write(util_bytes.FromUint64(timestamp))
	auth := hash.Sum(nil)
	if ps.checkAuth(method, auth, key, fileSize, timestamp) != nil {
		t.Error("fail")
	}

}
