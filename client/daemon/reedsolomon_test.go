package daemon

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/samoslab/nebula/client/util/logger"
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
	fname := filepath.Join(currentDir, "testdata/test.zip")
	log, err := logger.NewLogger("", true)
	require.NoError(t, err)
	originMd5, err := getFileMd5(fname)
	require.NoError(t, err)
	originSize, err := GetFileSize(fname)
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

	err = RsDecoder(log, fname, "", originSize, dataShards, parShards)
	require.NoError(t, err)
	recoveryMd5, err := getFileMd5(fname)
	require.Equal(t, originMd5, recoveryMd5)
	for _, fileInfo := range hashFiles {
		if err := os.Remove(fileInfo.FileName); err != nil {
			log.Errorf("delete %s failed, error %v", fileInfo.FileName, err)
		}
		_, err = os.Stat(fileInfo.FileName)
		require.True(t, os.IsNotExist(err))
	}

}

//
func TestOddFileSize(t *testing.T) {
	outDir := ""
	for _, dataShards := range []int{3, 4, 5, 6, 7, 8, 9, 10} {
		for parShards := 2; parShards < dataShards; parShards++ {
			currentDir, err := os.Getwd()
			require.NoError(t, err)
			fname := filepath.Join(currentDir, "testdata/odd_filesize.txt")
			log, err := logger.NewLogger("", false)
			require.NoError(t, err)
			originMd5, err := getFileMd5(fname)
			require.NoError(t, err)
			originSize, err := GetFileSize(fname)
			require.NoError(t, err)
			hashFiles, err := RsEncoder(log, outDir, fname, dataShards, parShards)
			require.NoError(t, err)
			require.Equal(t, dataShards+parShards, len(hashFiles))
			outDir, file := filepath.Split(fname)
			for i, fileInfo := range hashFiles {
				require.Equal(t, fileInfo.FileName, filepath.Join(outDir, fmt.Sprintf("%s.%d", file, i)))
				require.Equal(t, fileInfo.SliceIndex, i)
				_, err = os.Stat(fileInfo.FileName)
				require.NoError(t, err)
			}

			err = RsDecoder(log, fname, "", originSize, dataShards, parShards)
			require.NoError(t, err)
			recoveryMd5, err := getFileMd5(fname)
			require.Equal(t, originMd5, recoveryMd5)
			for _, fileInfo := range hashFiles {
				if err := os.Remove(fileInfo.FileName); err != nil {
					log.Errorf("delete %s failed, error %v", fileInfo.FileName, err)
				}
				_, err = os.Stat(fileInfo.FileName)
				require.True(t, os.IsNotExist(err))
			}
		}
	}

}
