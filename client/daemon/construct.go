package daemon

// DownFile list files format, used when download file
type DownFile struct {
	ID       string `json:"id"`
	FileSize uint64 `json:"filesize"`
	FileName string `json:"filename"`
	FileHash string `json:"filehash"`
	Folder   bool   `json:"folder"`
}

// FilePages list file
type FilePages struct {
	Total uint32      `json:"total"`
	Files []*DownFile `json:"files"`
}

// DirPair dir and its parent is a pair
type DirPair struct {
	Name   string
	Parent string
	Folder bool
}
