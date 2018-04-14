package daemon

import "os"

func GetFileModTime(filename string) (int64, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}
