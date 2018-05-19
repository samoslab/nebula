package aes

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
func Test(t *testing.T) {
	test(t, 16, 13)
	test(t, 16, 16)
	test(t, 16, 19)
	test(t, 16, 392)
	test(t, 24, 23)
	test(t, 24, 24)
	test(t, 24, 29)
	test(t, 24, 223)
	test(t, 32, 31)
	test(t, 32, 32)
	test(t, 32, 37)
	test(t, 32, 345)
}

func test(t *testing.T, keySize int, dataSize int) {
	var key, data, en, de []byte
	var err error
	key = randAesKey(keySize)
	data = randAesKey(dataSize)

	en, err = Encrypt(data, key)
	checkErr(err)
	de, err = Decrypt(en, key)
	checkErr(err)
	if !bytes.Equal(data, de) {
		t.Error("failed")
	}
}

func randAesKey(bits int) []byte {
	token := make([]byte, bits)
	_, err := rand.Read(token)
	if err != nil {
		checkErr(err)
	}
	return token
}
