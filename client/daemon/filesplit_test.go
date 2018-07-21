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

func TestFolderUpload(t *testing.T) {
	parent := "/tmp/upload/"
	dest := "/tmp/music"
	dirs := []DirPair{
		DirPair{Parent: "/tmp/", Name: "upload", Folder: true},
		DirPair{Parent: "/tmp/upload/", Name: "/tmp/upload/abc.txt", Folder: false},
		DirPair{Parent: "/tmp/upload", Name: "localmusic", Folder: true},
		DirPair{Parent: "/tmp/upload/localmusic", Name: "/tmp/upload/localmusic/sea.music", Folder: false},
	}

	newDirs := dirAdjust(dirs, parent, dest, "linux")

	expectDirs := []DirPair{
		DirPair{Parent: "/tmp/music", Name: "upload", Folder: true},
		DirPair{Parent: "/tmp/music/upload", Name: "/tmp/upload/abc.txt", Folder: false},
		DirPair{Parent: "/tmp/music/upload", Name: "localmusic", Folder: true},
		DirPair{Parent: "/tmp/music/upload/localmusic", Name: "/tmp/upload/localmusic/sea.music", Folder: false},
	}
	assert.Equal(t, newDirs, expectDirs)

}

func TestWinFolderUpload(t *testing.T) {
	parent := "C:\\windows\\system32"
	dest := "/tmp/music"
	dirs := []DirPair{
		DirPair{Parent: "C:\\windows", Name: "system32", Folder: true},
		DirPair{Parent: "C:\\windows\\system32", Name: "C:\\windows\\system32\\abc.txt", Folder: false},
		DirPair{Parent: "C:\\windows\\system32", Name: "localmusic", Folder: true},
		DirPair{Parent: "C:\\windows\\system32\\localmusic", Name: "C:\\windows\\system32\\localmusic\\sea.music", Folder: false},
	}

	newDirs := dirAdjust(dirs, parent, dest, "windows")

	expectDirs := []DirPair{
		DirPair{Parent: "/tmp/music", Name: "system32", Folder: true},
		DirPair{Parent: "/tmp/music/system32", Name: "C:\\windows\\system32\\abc.txt", Folder: false},
		DirPair{Parent: "/tmp/music/system32", Name: "localmusic", Folder: true},
		DirPair{Parent: "/tmp/music/system32/localmusic", Name: "C:\\windows\\system32\\localmusic\\sea.music", Folder: false},
	}
	assert.Equal(t, newDirs, expectDirs)

}
