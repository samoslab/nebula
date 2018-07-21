package daemon

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/samoslab/nebula/client/config"
	"github.com/samoslab/nebula/util/file"
)

// GetConfigFile get config file
func GetConfigFile() (string, string) {
	defaultConfig := filepath.Join(file.UserHome(), config.DefaultConfigDir, config.DefaultConfig)
	defaultAppDir, _ := filepath.Split(defaultConfig)
	return defaultAppDir, defaultConfig
}

// GetFileModTime get file modify time
func GetFileModTime(filename string) (int64, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}

// GetFileSize get file size
func GetFileSize(filename string) (int64, error) {
	fileInfo, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

// GetChunkSizeAndNum get chunk size and num by file size
func GetChunkSizeAndNum(fileSize int64, partitionSize int64) (int64, int) {
	chunkNum := int(math.Ceil(float64(fileSize) / float64(partitionSize)))
	chunkSize := fileSize / int64(chunkNum)
	return chunkSize, chunkNum
}

// ReverseCalcuatePartFileSize calculate part.file size
func ReverseCalcuatePartFileSize(fileSize int64, partitionNum, currentPartition int) int64 {
	chunkSize := fileSize / int64(partitionNum)
	// last part
	if currentPartition == partitionNum-1 {
		return fileSize - chunkSize*int64(currentPartition)
	}
	return chunkSize
}

func sameDisk(src, dst string) bool {
	rootSrc := strings.ToUpper(src[0:2])
	rootDst := strings.ToUpper(dst[0:2])
	return rootSrc == rootDst
}

func RenameCrossOS(src, dest string) error {
	if runtime.GOOS == "windows" {
		//move in same disk, copy in diff disk
		if sameDisk(src, dest) {
			return os.Rename(src, dest)
		}
		_, err := CopyFile(src, dest)
		if err != nil {
			return err
		}

		return os.Remove(src)
	}

	return os.Rename(src, dest)
}

func CopyFile(srcName, dstName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func SaveFile(fileName string, content []byte) error {
	// open output file
	fo, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer fo.Close()
	if _, err := fo.Write(content); err != nil {
		return err
	}
	return nil
}
