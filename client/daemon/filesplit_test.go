package daemon

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileSplit(t *testing.T) {
	outDir := "testdata"
	fName := "testdata/odd_filesize.txt"
	fileSize, err := GetFileSize(fName)
	assert.NoError(t, err)
	assert.Equal(t, int64(3479), fileSize)
	// size: 3479 , partition size 1237
	partitionSize := int64(1237)
	chunkSize, chunkNum := GetChunkSizeAndNum(fileSize, partitionSize)
	assert.Equal(t, 3, chunkNum)
	assert.Equal(t, int64(1159), chunkSize)
	partFiles, err := FileSplit(outDir, fName, fileSize, chunkSize, int64(chunkNum))
	assert.NoError(t, err)
	assert.Equal(t, chunkNum, len(partFiles))
	expectPartFiles := []struct {
		name string
		size int64
	}{
		{
			name: "testdata/odd_filesize.txt.part.0",
			size: 1159,
		},
		{
			name: "testdata/odd_filesize.txt.part.1",
			size: 1159,
		},
		{
			name: "testdata/odd_filesize.txt.part.2",
			size: 3479 - 1159 - 1159,
		},
	}
	for i, file := range partFiles {
		assert.Equal(t, file, expectPartFiles[i].name)
		fileSize, err := GetFileSize(file)
		assert.NoError(t, err)
		assert.Equal(t, fileSize, expectPartFiles[i].size)
	}
	for _, file := range partFiles {
		err := os.Remove(file)
		assert.NoError(t, err)
	}
}
