package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderBackupMap(t *testing.T) {
	workedNum := 40
	backupNum := 10
	backMap := createBackupProvicer(workedNum, backupNum)
	assert.Equal(t, 40, len(backMap))
	for _, v := range backMap {
		assert.Equal(t, 2, len(v))
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

	newDirs := dirAdjust(dirs, parent, dest)

	expectDirs := []DirPair{
		DirPair{Parent: "/tmp/music", Name: "upload", Folder: true},
		DirPair{Parent: "/tmp/music/upload", Name: "/tmp/upload/abc.txt", Folder: false},
		DirPair{Parent: "/tmp/music/upload", Name: "localmusic", Folder: true},
		DirPair{Parent: "/tmp/music/upload/localmusic", Name: "/tmp/upload/localmusic/sea.music", Folder: false},
	}
	assert.Equal(t, newDirs, expectDirs)

}
