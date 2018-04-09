package daemon

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncoder(t *testing.T) {
	outDir := "/tmp"
	dataShards := 4
	parShards := 2
	fname := "testdata/test.zip"
	hashFiles, err := RsEncoder(outDir, fname, dataShards, parShards)
	require.NoError(t, err)
	require.Equal(t, 6, len(hashFiles))
	_, file := filepath.Split(fname)
	for i, fileInfo := range hashFiles {
		require.Equal(t, fileInfo.FileName, filepath.Join(outDir, fmt.Sprintf("%s.%d", file, i)))
		require.Equal(t, fileInfo.SliceIndex, i)
	}

	fname = "/tmp/test.zip"
	outfname := "testdata/test.zip.recovery"
	err = RsDecoder(fname, outfname, dataShards, parShards)
	require.NoError(t, err)
}
