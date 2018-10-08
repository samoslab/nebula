package main

import (
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("% string", os.Args[0])
		return
	}
	bts, err := hex.DecodeString(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%x-%x-%x-%x-%x\n", bts[:4], bts[4:6], bts[6:8], bts[8:10], bts[10:])
}
