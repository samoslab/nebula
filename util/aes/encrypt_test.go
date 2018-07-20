package aes

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestEncryptWithDecrypt(t *testing.T) {
	content := "acnde"
	key := "12345678abcdef00"
	data, err := Encrypt([]byte(content), []byte(key))
	assert.NoError(t, err)
	assert.Equal(t, true, len(data) > 0)

	origin, err := Decrypt(data, []byte(key))
	assert.NoError(t, err)
	assert.Equal(t, string(origin), content)

	content = "123123123bdsafadsfyasdf7123ndsafadsf,adsf,asfjewrewrwacnde"
	key = "12345678abcdef00abcdefgh00000000"
	data, err = Encrypt([]byte(content), []byte(key))
	assert.NoError(t, err)
	assert.Equal(t, true, len(data) > 0)

	origin, err = Decrypt(data, []byte(key))
	assert.NoError(t, err)
	assert.Equal(t, string(origin), content)
	data = data[0 : len(data)-2]
	key = "12345678abcdef00abcdefgh00000000"
	assert.Panics(t, func() { Decrypt(data, []byte(key)) })
}
