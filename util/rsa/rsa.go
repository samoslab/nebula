package rsa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
)

func EncryptLong(pubKey *rsa.PublicKey, data []byte, bts int) ([]byte, error) {
	if len(data) <= bts-11 {
		return rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	}
	var err error
	var encrypt []byte
	var buffer bytes.Buffer
	for i := 0; i < len(data); {
		end := i + bts - 11
		if end > len(data) {
			end = len(data)
		}
		encrypt, err = rsa.EncryptPKCS1v15(rand.Reader, pubKey, data[i:end])
		if err != nil {
			return nil, err
		}
		buffer.Write(encrypt)
		i = end
	}
	return buffer.Bytes(), nil
}

func DecryptLong(priKey *rsa.PrivateKey, data []byte, bts int) ([]byte, error) {
	if len(data) <= bts {
		return rsa.DecryptPKCS1v15(rand.Reader, priKey, data)
	}
	var err error
	var decrypt []byte
	var buffer bytes.Buffer
	for i := 0; i < len(data); {
		end := i + bts
		if end > len(data) {
			end = len(data)
		}
		decrypt, err = rsa.DecryptPKCS1v15(rand.Reader, priKey, data[i:end])
		if err != nil {
			return nil, err
		}
		buffer.Write(decrypt)
		i = end
	}
	return buffer.Bytes(), nil
}
