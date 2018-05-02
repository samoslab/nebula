package node

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"strconv"

	"github.com/samoslab/nebula/provider/config"
	util_hash "github.com/samoslab/nebula/util/hash"
	log "github.com/sirupsen/logrus"
)

type AesKey []byte // AesKey[0] is version, AesKey[1:] is real AES key
type Node struct {
	NodeId      []byte
	PubKey      *rsa.PublicKey
	PriKey      *rsa.PrivateKey
	PubKeyBytes []byte
	EncryptKey  map[string][]byte
}

func LoadFormConfig() *Node {
	node := &Node{}
	var err error
	node.NodeId, node.PubKey, node.PriKey, node.PubKeyBytes, node.EncryptKey, err = config.ParseNode()
	if err != nil {
		log.Fatalln(err)
	}
	return node
}

const RSA_KEY_BYTES = 256

func NewNode(difficulty int) *Node {
	n := &Node{}
	for {
		pk, err := rsa.GenerateKey(rand.Reader, 256*8)
		if err != nil {
			log.Fatalf("GenerateKey failed:%s", err.Error())
		}
		n.PriKey = pk
		n.PubKey = &pk.PublicKey
		n.PubKeyBytes = x509.MarshalPKCS1PublicKey(n.PubKey)
		n.NodeId = util_hash.Sha1(n.PubKeyBytes)
		if count_preceding_zero_bits(util_hash.Sha1(n.NodeId)) < difficulty {
			break
		}
	}
	m := make(map[string][]byte, 1)
	m["0"] = randAesKey(256)
	n.EncryptKey = m
	return n
}

func randAesKey(bits int) []byte {
	token := make([]byte, bits)
	_, err := rand.Read(token)
	if err != nil {
		log.Errorf("generate AES key err: %s", err)
	}
	return token
}
func count_preceding_zero_bits(nodeIdHash []byte) int {
	res := 0
	for _, b := range nodeIdHash {
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

func (self *Node) NodeIdStr() string {
	return hex.EncodeToString(self.NodeId)
}

func (self *Node) PrivateKeyStr() string {
	return hex.EncodeToString(x509.MarshalPKCS1PrivateKey(self.PriKey))
}
func (self *Node) PublicKeyStr() string {
	return hex.EncodeToString(self.PubKeyBytes)
}
