package node

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"fmt"
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
		if count_preceding_zero_bits(n.NodeId) < difficulty {
			break
		}
	}
	return n
}

func count_preceding_zero_bits(nodeId []byte) int {
	// arg := sha1Sum(nodeId)
	// TODO
	return 100
}

func sha1Sum(content []byte) []byte {
	h := sha1.New()
	h.Write(content)
	return h.Sum(nil)
}
