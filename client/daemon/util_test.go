package daemon

import (
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
