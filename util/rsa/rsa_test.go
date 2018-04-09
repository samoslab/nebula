package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"testing"
)

func Test(t *testing.T) {
	bts := 256
	priKey, _ := rsa.GenerateKey(rand.Reader, bts*8)
	pubKey := &priKey.PublicKey
	str := "Hello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundHello, playgroundplaygrplaygroundHello"
	encrypt, _ := EncryptLong(pubKey, []byte(str), bts)
	decrypt, _ := DecryptLong(priKey, encrypt, bts)
	if str != string(decrypt) {
		t.Errorf(string(decrypt))
	}
	pubKeyBytes := x509.MarshalPKCS1PublicKey(pubKey)
	encrypt, _ = EncryptLong(pubKey, pubKeyBytes, bts)
	decrypt, _ = DecryptLong(priKey, encrypt, bts)
	if !bytes.Equal(pubKeyBytes, decrypt) {
		t.Error(len(decrypt))
	}
}
