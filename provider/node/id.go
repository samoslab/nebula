package node

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"fmt"
	"strconv"
)

type NodeId []byte
type Node struct {
	NodeId NodeId
	PubKey rsa.PublicKey
	PriKey rsa.PrivateKey
}

func NewNode(difficulty int) Node {
	n := Node{}
	for {
		pk, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			fmt.Printf("GenerateKey failed:%s\n", err.Error())
		}
		n.PriKey = *pk
		n.PubKey = pk.PublicKey
		byteSlice, err := x509.MarshalPKIXPublicKey(n.PubKey)
		if err != nil {
			fmt.Printf("MarshalPKIXPublicKey failed:%s\n", err.Error())
		}
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
