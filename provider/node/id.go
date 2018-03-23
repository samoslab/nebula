package node

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spolabs/nebula/provider/config"
)

type AesKey []byte // AesKey[0] is version, AesKey[1:] is real AES key
type Node struct {
	NodeId     []byte
	PubKey     *rsa.PublicKey
	PriKey     *rsa.PrivateKey
	EncryptKey map[string][]byte
}

func LoadFormConfig() *Node {
	conf := config.GetProviderConfig()
	pubKeyBytes := []byte(conf.PublicKey)
	if conf.NodeId != hex.EncodeToString(sha1Sum(pubKeyBytes)) {
		log.Fatalln("NodeId is not match PublicKey")
	}
	pubK, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		log.Fatalf("ParsePKCS1PublicKey failed: %s\n", err)
	}
	priK, err := x509.ParsePKCS1PrivateKey([]byte(conf.PrivateKey))
	if err != nil {
		log.Fatalf("ParsePKCS1PrivateKey failed: %s\n", err)
	}
	m := make(map[string][]byte, len(conf.EncryptKey))
	for k, v := range conf.EncryptKey {
		m[k] = []byte(v)
	}
	nodeId, err := hex.DecodeString(conf.NodeId)
	if err != nil {
		log.Fatalf("DecodeString node id hex string failed: %s\n", err)
	}

	return &Node{NodeId: nodeId, PubKey: pubK, PriKey: priK, EncryptKey: m}
}
func NewNode(difficulty int) *Node {
	n := &Node{}
	for {
		pk, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			fmt.Printf("GenerateKey failed:%s\n", err.Error())
		}
		n.PriKey = pk
		n.PubKey = &pk.PublicKey
		byteSlice := x509.MarshalPKCS1PublicKey(n.PubKey)
		n.NodeId = sha1Sum(byteSlice)
		if count_preceding_zero_bits(sha1Sum(n.NodeId)) < difficulty {
			break
		}
	}
	return n
}

func count_preceding_zero_bits(nodeIdHash []byte) int {
	res := 0
	for b := range nodeIdHash {
		str := strconv.FormatInt(int64(b), 2)
		if len(str) > 1 {
			res += (8 - len(str))
			break
		} else if str == "0" {
			res += 8
		} else {
			res += 7
			break
		}
	}
	return res
}

func sha1Sum(content []byte) []byte {
	h := sha1.New()
	h.Write(content)
	return h.Sum(nil)
}

func (self *Node) PrivateKeyStr() string {
	return string(x509.MarshalPKCS1PrivateKey(self.PriKey))
}
func (self *Node) PublicKeyStr() string {
	return string(self.PublicKeyBytes())
}

func (self *Node) PublicKeyBytes() []byte {
	return x509.MarshalPKCS1PublicKey(self.PubKey)
}
