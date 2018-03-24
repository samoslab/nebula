package file

import (
	"io"
	"os"
)

func ExistsWithInfo(path string) (exists bool, fileInfo os.FileInfo) {
	fileInfo, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}
	return true, fileInfo
}

func Exists(path string) bool {
	res, _ := ExistsWithInfo(path)
	return res
}

func ConcatFile(writeFile string, readFile string, removeReadFileIfSuccess bool) error {
	fi, err := os.Open(readFile)
	if err != nil {
		return err
	}
	defer fi.Close()
	fo, err := os.OpenFile(writeFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer fo.Close()
	buf := make([]byte, 8192)
	for {
		// read a chunk
		n, err := fi.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		// write a chunk
		if _, err := fo.Write(buf[:n]); err != nil {
			return err
		}
	}
	if removeReadFileIfSuccess {
		err = os.Remove(readFile)
		if err != nil {
			return nil
		}
	}
	return nil
}
