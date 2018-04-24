package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoder(t *testing.T) {
	outDir := "/tmp"
	dataShards := 4
	parShards := 2
	fname := "testdata/test.zip"
	log, err := NewLogger("", true)
	require.NoError(t, err)
	hashFiles, err := RsEncoder(log, outDir, fname, dataShards, parShards)
	require.NoError(t, err)
	require.Equal(t, 6, len(hashFiles))
	_, file := filepath.Split(fname)
	for i, fileInfo := range hashFiles {
		require.Equal(t, fileInfo.FileName, filepath.Join(outDir, fmt.Sprintf("%s.%d", file, i)))
		require.Equal(t, fileInfo.SliceIndex, i)
		_, err = os.Stat(fileInfo.FileName)
		require.NoError(t, err)
	}

	fname = "/tmp/test.zip"
	outfname := "testdata/test.zip.recovery"
	err = RsDecoder(log, fname, outfname, dataShards, parShards)
	require.NoError(t, err)
	for _, fileInfo := range hashFiles {
		fmt.Printf("%s\n", fileInfo.FileName)
		if err := os.Remove(fileInfo.FileName); err != nil {
			log.Errorf("delete %s failed, error %v", fileInfo.FileName, err)
		}
		_, err = os.Stat(fileInfo.FileName)
		require.True(t, os.IsNotExist(err))
	}

}
