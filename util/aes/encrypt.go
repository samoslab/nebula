package aes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io/ioutil"
)

func Encrypt(origData, key []byte) ([]byte, error) {
	//https://github.com/polaris1119/myblog_article_code/blob/master/aes/aes.go
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	origData = pkcs5Padding(origData, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func Decrypt(crypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(crypted))
	blockMode.CryptBlocks(origData, crypted)
	origData = pkcs5UnPadding(origData)
	return origData, nil
}

func zeroPadding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{0}, padding)
	return append(ciphertext, padtext...)
}

func zeroUnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

// RandStr return random string
func RandStr(str_size int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, str_size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// EncryptFile encrypt file using aes alg
func EncryptFile(inputfile string, key []byte, outputfile string) error {
	b, err := ioutil.ReadFile(inputfile) //Read the target file
	if err != nil {
		fmt.Printf("Unable to open the input file!\n")
		return err
	}
	ciphertext, err := Encrypt(b, key)
	if err != nil {
		fmt.Printf("encrypt file failed")
		return err
	}
	err = ioutil.WriteFile(outputfile, ciphertext, 0644)
	if err != nil {
		fmt.Printf("Unable to create encrypted file!\n")
		return err
	}

	return nil
}

// DecryptFile decrypt file
func DecryptFile(inputfile string, key []byte, outputfile string) error {
	z, err := ioutil.ReadFile(inputfile)
	if err != nil {
		return err
	}
	result, err := Decrypt(z, key)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(outputfile, result, 0644)
	if err != nil {
		fmt.Printf("Unable to create decrypted file!\n")
		return err
	}
	return nil
}
