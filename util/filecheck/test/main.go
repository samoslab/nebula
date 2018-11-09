package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	util_filecheck "github.com/samoslab/nebula/util/filecheck"
)

var runner *util_filecheck.GenMetadataRunner
var paths map[string]struct{} = make(map[string]struct{}, 2048)

func main() {
	runner = util_filecheck.NewRunner()
	go runner.Run()
	sig := make(chan bool)
	go clean(sig)
	filepath.Walk("/Volumes/Dev/maven-repository/", walkfunc)
	time.Sleep(300 * time.Second)
	runner.Quit()
	sig <- true
}

func walkfunc(path string, info os.FileInfo, err error) (er error) {
	if info.IsDir() {
		return
	}
	if !strings.HasSuffix(info.Name(), ".jar") {
		return
	}
	runner.AddPath(path, 32768)
	paths[path] = struct{}{}
	return
}

func clean(sig chan bool) {
	for i := 0; i < 2000; i++ {
		time.Sleep(1 * time.Second)
		select {
		case <-sig:
			return
		default:
			for k := range paths {
				exist, _, generator, _, _, _, er := runner.GetResult(k)
				if exist {
					if er == nil {
						fmt.Println(base64.StdEncoding.EncodeToString(generator))
					} else {
						fmt.Println(er)
					}
					runner.RemoveResult(k)
					delete(paths, k)
				}
			}
		}
	}
}
