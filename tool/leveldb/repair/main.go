package main

import (
	"fmt"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println(os.Args[0] + " path")
		return
	}
	_, err := leveldb.RecoverFile(os.Args[1], nil)
	if err != nil {
		fmt.Println(err)
	}
}
