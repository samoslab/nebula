package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	util_filecheck "github.com/samoslab/nebula/util/filecheck"
)

var runner *util_filecheck.GenMetadataRunner
var paths = make(map[string]struct{}, 2048)
var pathsLock = new(sync.RWMutex)

func main() {
	var waitClose sync.WaitGroup
	for i := 0; i < 4; i++ {
		waitClose.Add(1)
		go run(waitClose)
	}
	waitClose.Wait()

}

func run(waitClose sync.WaitGroup) {

	for i := 0; i < 300; i++ {
		filepath.Walk("/Users/lijt/Downloads/uploadfiles", walkfunc)
	}
	waitClose.Done()

}

func walkfunc(path string, info os.FileInfo, err error) (er error) {
	if info.IsDir() {
		return
	}
	// if !strings.HasSuffix(info.Name(), ".jar") {
	// 	return
	// }
	_, generator, _, _, _, err := util_filecheck.GenMetadata(path, getChunkSize(uint64(info.Size())))
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(base64.StdEncoding.EncodeToString(generator))
	}
	return
}

func getChunkSize(fileSize uint64) uint32 {
	if fileSize < 32768 {
		return 2048
	} else if fileSize < 131072 {
		return 4096
	} else if fileSize < 524288 {
		return 8192
	} else if fileSize < 4194304 {
		return 16384
	} else if fileSize < 16777216 {
		return 32768
	} else if fileSize < 33554432 {
		return 65536
	} else if fileSize < 67108864 {
		return 131072
	} else {
		return 262144
	}
}
