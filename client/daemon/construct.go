package daemon

type HashFile struct {
	FileSize   int64
	FileName   string
	FileHash   []byte
	SliceIndex int
}

type MyPart struct {
	Filename string
	Pieces   []HashFile
}

type DownFile struct {
	ID       string `json:"id"`
	FileSize uint64 `json:"filesize"`
	FileName string `json:"filename"`
	FileHash string `json:"filehash"`
	Folder   bool   `json:"folder"`
}

type DirPair struct {
	Name   string
	Parent string
}
