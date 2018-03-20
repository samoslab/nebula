package main

import (
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	// open input file
	fi, err := os.Open("/media/win_d/iso/ubuntu/ubuntu-16.04.2-desktop-amd64.iso")
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	start := time.Now().UnixNano()
	o3, err := fi.Seek(1073741824, 0)
	if err != nil {
		panic(err)
	}
	fmt.Printf("seek cost: %d\n", time.Now().UnixNano()-start)
	fmt.Println(o3)
	buf := make([]byte, 32*1024)
	_, err = fi.Seek(0, 0)
	if err != nil {
		panic(err)
	}
	start = time.Now().UnixNano()
	for i := 0; i < 32768; i++ {
		_, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
	}
	fmt.Printf("read cost: %d\n", time.Now().UnixNano()-start)

}
