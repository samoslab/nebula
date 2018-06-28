package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func main() {
	bts, err := hex.DecodeString("53eb1208b1d1559b56120e227d7db5f7e9a52b34")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(base64.StdEncoding.EncodeToString(bts))
}
