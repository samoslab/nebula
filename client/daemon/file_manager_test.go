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

	choosed := chooseBackupProvicer(0, backMap)
	assert.Equal(t, 0, choosed)
	choosed = chooseBackupProvicer(4, backMap)
	assert.Equal(t, 5, choosed)
	choosed = chooseBackupProvicer(5, backMap)
	assert.Equal(t, -1, choosed)
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

func TestPasswordPadding(t *testing.T) {
	passwd := "abcde"
	realPasswd, err := passwordPadding(passwd, 0)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(realPasswd))
	assert.Equal(t, passwd+"00000000000", realPasswd)
	realPasswd, err = passwordPadding(passwd, 1)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(realPasswd))
	assert.Equal(t, passwd+"000000000000000000000000000", realPasswd)
	realPasswd, err = passwordPadding(passwd, 2)
	assert.Error(t, err)

}
