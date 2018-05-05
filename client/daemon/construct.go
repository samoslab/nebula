package daemon

// DownFile list files format, used when download file
type DownFile struct {
	ID       string `json:"id"`
	FileSize uint64 `json:"filesize"`
	FileName string `json:"filename"`
	FileHash string `json:"filehash"`
	Folder   bool   `json:"folder"`
}

// DirPair dir and its parent is a pair
type DirPair struct {
	Name   string
	Parent string
}
