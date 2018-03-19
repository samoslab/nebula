package main

import (
	"io"
	"os"
)

func main() {
	var fo *os.File
	first := true
	// open input file
	fi, err := os.Open("/home/lijt/go/bin/godoc")
	if err != nil {
		panic(err)
	}
	// close fi on exit and check for its returned error
	defer func() {
		if err := fi.Close(); err != nil {
			panic(err)
		}
	}()
	if first {
		// open output file
		fo, err := os.OpenFile(
			"/tmp/t/test.blk",
			os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
			0666)
		if err != nil {
			panic(err)
		}
		// close fo on exit and check for its returned error
		defer func() {
			if err := fo.Close(); err != nil {
				panic(err)
			}
		}()
	}

	// make a buffer to keep chunks that are read
	buf := make([]byte, 32*1024)
	for {
		// read a chunk
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}

		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			panic(err)
		}
	}
}
