package daemon

import (
	"log"
	"math"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

// GetConfigFile get config file
func GetConfigFile() (string, string) {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Get OS current user failed: %s", err)
	}
	defaultAppDir := filepath.Join(usr.HomeDir, ".spo-nebula-client")
	defaultConfig := filepath.Join(defaultAppDir, "config.json")
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

func ReverseCalcuatePartFileSize(fileSize int64, partitionNum, currentPartition int) int64 {
	chunkSize := fileSize / int64(partitionNum)
	// last part
	if currentPartition == partitionNum-1 {
		return fileSize - chunkSize*int64(currentPartition)
	}
	return chunkSize
}

func Fping(ips []string) ([]string, error) {
	commands := "fping " + strings.Join(ips, " ")
	cmd := exec.Command("/bin/sh", "-c", commands)
	ip, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	aliveIps := []string{}
	for _, ip := range strings.Split(string(ip), "\n") {
		if strings.HasSuffix(ip, "is alive") {
			aliveIps = append(aliveIps, strings.Trim(ip, " is alive"))
		}
	}
	return aliveIps, nil
}
