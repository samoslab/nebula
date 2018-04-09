package node

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/hex"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/spolabs/nebula/provider/config"
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
	conf := config.GetProviderConfig()
	pubKeyBytes, err := hex.DecodeString(conf.PublicKey)
	if err != nil {
		log.Fatalf("DecodeString Public Key failed: %s", err)
	}
	if conf.NodeId != hex.EncodeToString(sha1Sum(pubKeyBytes)) {
		log.Fatalln("NodeId is not match PublicKey")
	}
	pubK, err := x509.ParsePKCS1PublicKey(pubKeyBytes)
	if err != nil {
		log.Fatalf("ParsePKCS1PublicKey failed: %s", err)
	}
	priKeyBytes, err := hex.DecodeString(conf.PrivateKey)
	if err != nil {
		log.Fatalf("DecodeString Private Key failed: %s", err)
	}
	priK, err := x509.ParsePKCS1PrivateKey(priKeyBytes)
	if err != nil {
		log.Fatalf("ParsePKCS1PrivateKey failed: %s", err)
	}
	m := make(map[string][]byte, len(conf.EncryptKey))
	for k, v := range conf.EncryptKey {
		m[k], err = hex.DecodeString(v)
		if err != nil {
			log.Fatalf("DecodeString EncryptKey %s failed: %s", v, err)
		}
	}
	nodeId, err := hex.DecodeString(conf.NodeId)
	if err != nil {
		log.Fatalf("DecodeString node id hex string failed: %s", err)
	}

	return &Node{NodeId: nodeId, PubKey: pubK, PriKey: priK, PubKeyBytes: pubKeyBytes, EncryptKey: m}
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
		n.NodeId = sha1Sum(n.PubKeyBytes)
		if count_preceding_zero_bits(sha1Sum(n.NodeId)) < difficulty {
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

func sha1Sum(content []byte) []byte {
	h := sha1.New()
	h.Write(content)
	return h.Sum(nil)
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
