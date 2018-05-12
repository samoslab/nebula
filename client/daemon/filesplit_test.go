package daemon

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChunSizeAndNum(t *testing.T) {
	testCase := []struct {
		FileSize      int64
		PartitionSize int64
		ChunkSize     int64
		ChunkNum      int
	}{
		{
			FileSize:      int64(10000),
			PartitionSize: int64(999),
			ChunkSize:     int64(909),
			ChunkNum:      11,
		},
		{
			FileSize:      int64(259 * 1024 * 1024),
			PartitionSize: PartitionMaxSize,
			ChunkSize:     int64(259 * 1024 * 512),
			ChunkNum:      2,
		},
		{
			FileSize:      int64(999 * 1024 * 1024),
			PartitionSize: PartitionMaxSize,
			ChunkSize:     int64(999 * 1024 * 256),
			ChunkNum:      4,
		},
	}
	for _, cs := range testCase {
		chunkSize, chunkNum := GetChunkSizeAndNum(cs.FileSize, cs.PartitionSize)
		assert.Equal(t, cs.ChunkSize, chunkSize)
		assert.Equal(t, cs.ChunkNum, chunkNum)
	}
}

func TestReverseCalcuPartFileSize(t *testing.T) {
	fileSize := int64(3479)
	partitionNum := 3
	assert.Equal(t, int64(1159), ReverseCalcuatePartFileSize(fileSize, partitionNum, 0))
	assert.Equal(t, int64(1159), ReverseCalcuatePartFileSize(fileSize, partitionNum, 1))
	assert.Equal(t, int64(1161), ReverseCalcuatePartFileSize(fileSize, partitionNum, 2))
}

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
