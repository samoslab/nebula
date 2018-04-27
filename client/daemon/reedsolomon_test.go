package daemon

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func getFileMd5(filename string) (string, error) {
	f, err := os.OpenFile(filename, os.O_RDONLY, 0644)
	if err != nil {
		return "", err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", md5.Sum(b)), nil
}

func TestEncoder(t *testing.T) {
	outDir := ""
	dataShards := 2
	parShards := 1
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	fmt.Printf("current dir %s\n", currentDir)
	fname := filepath.Join(currentDir, "testdata/test.zip")
	log, err := NewLogger("", true)
	require.NoError(t, err)
	originMd5, err := getFileMd5(fname)
	fmt.Printf("origin md5 %s\n", originMd5)
	require.NoError(t, err)
	hashFiles, err := RsEncoder(log, outDir, fname, dataShards, parShards)
	require.NoError(t, err)
	require.Equal(t, 3, len(hashFiles))
	outDir, file := filepath.Split(fname)
	for i, fileInfo := range hashFiles {
		require.Equal(t, fileInfo.FileName, filepath.Join(outDir, fmt.Sprintf("%s.%d", file, i)))
		require.Equal(t, fileInfo.SliceIndex, i)
		_, err = os.Stat(fileInfo.FileName)
		require.NoError(t, err)
	}

	err = RsDecoder(log, fname, "", dataShards, parShards)
	require.NoError(t, err)
	recoveryMd5, err := getFileMd5(fname)
	fmt.Printf("recovery md5 %s\n", recoveryMd5)
	require.Equal(t, originMd5, recoveryMd5)
	for _, fileInfo := range hashFiles {
		fmt.Printf("%s\n", fileInfo.FileName)
		if err := os.Remove(fileInfo.FileName); err != nil {
			log.Errorf("delete %s failed, error %v", fileInfo.FileName, err)
		}
		_, err = os.Stat(fileInfo.FileName)
		require.True(t, os.IsNotExist(err))
	}

}
