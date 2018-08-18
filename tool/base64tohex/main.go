package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

func main() {
	bts, err := base64.StdEncoding.DecodeString("gtpIRY3ksd7iIWs3TIt4Rabwz9A=")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hex.EncodeToString(bts))
}
