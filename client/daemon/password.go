package daemon

import (
	"crypto/sha256"
	"fmt"

	"github.com/samoslab/nebula/util/aes"
)

func genEncryptKey(sno uint32, password string) ([]byte, error) {
	digestinfo := fmt.Sprintf("msg:%s:%d", InfoForEncrypt, sno)
	encryptData, err := aes.Encrypt([]byte(digestinfo), []byte(password))
	if err != nil {
		return nil, err
	}

	h := sha256.New()
	h.Write(encryptData)
	shaData := h.Sum(nil)
	return shaData, nil
}

func padding(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "0"
	}
	return s
}

func passwordPadding(originPasswd string, sno uint32) (string, error) {
	realPasswd := ""
	length := len(originPasswd)
	switch sno {
	case 0:
		if length > 16 {
			fmt.Errorf("password length must less than 16")
		}
		realPasswd = originPasswd + padding(16-length)
	case 1:
		if length > 32 {
			fmt.Errorf("password length must less than 32")
		}
		realPasswd = originPasswd + padding(32-length)
	default:
		return "", fmt.Errorf("space %d not exist", sno)
	}

	return realPasswd, nil
}

func verifyPassword(sno uint32, password string, encryData []byte) bool {
	encryInfo, err := genEncryptKey(sno, password)
	if err != nil {
		return false
	}
	return string(encryInfo) == string(encryData)
}
