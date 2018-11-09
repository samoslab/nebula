package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	util_filecheck "github.com/samoslab/nebula/util/filecheck"
)

var runner *util_filecheck.GenMetadataRunner
var paths = make(map[string]struct{}, 2048)
var pathsLock = new(sync.RWMutex)

func main() {
	runner = util_filecheck.NewRunner()
	go runner.Run()
	sig := make(chan bool)
	go clean(sig)
	filepath.Walk("/", walkfunc)
	for {
		time.Sleep(30 * time.Second)
		if getPathsLen() == 0 {
			break
		}
	}
	runner.Quit()
	sig <- true
}

func walkfunc(path string, info os.FileInfo, err error) (er error) {
	if info.IsDir() {
		return
	}
	// if !strings.HasSuffix(info.Name(), ".jar") {
	// 	return
	// }
	runner.AddPath(path, 32768)
	pathsLock.Lock()
	defer pathsLock.Unlock()
	paths[path] = struct{}{}
	return
}

func clean(sig chan bool) {
	for {
		time.Sleep(1 * time.Second)
		select {
		case <-sig:
			return
		default:
			getAndClean()
		}
	}
}
func getAndClean() {
	for k := range copyMap() {
		exist, _, generator, _, _, _, er := runner.GetResult(k)
		if exist {
			if er == nil {
				fmt.Println(base64.StdEncoding.EncodeToString(generator))
			} else {
				fmt.Println(er)
			}
			runner.RemoveResult(k)
			removeFromMap(k)
		}
	}
}

func removeFromMap(path string) {
	pathsLock.Lock()
	defer pathsLock.Unlock()
	delete(paths, path)
}

func copyMap() map[string]struct{} {
	pathsLock.RLock()
	defer pathsLock.RUnlock()
	m := make(map[string]struct{}, len(paths))
	for k, v := range paths {
		m[k] = v
	}
	return m
}

func getPathsLen() int {
	pathsLock.RLock()
	defer pathsLock.RUnlock()
	return len(paths)
}
