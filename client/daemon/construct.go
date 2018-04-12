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
