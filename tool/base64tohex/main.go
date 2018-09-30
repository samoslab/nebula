package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("% string", os.Args[0])
		return
	}
	bts, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(hex.EncodeToString(bts))
}
