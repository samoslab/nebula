package file

import (
	"errors"
	"fmt"
	"os"
)

type MetaInfo struct {
	Size      float32 `json:"size"`
	FileName  string  `json:"file_name"`
	PublicKey string  `json:"public_key"`

	HashContent []byte `json:"hash_content"`
	Sig         string `json:"sig"`
	CreateTime  int    `json:"create_time"`

	Blocks map[string]string `json:"blocks"`
}

type Block struct {
	PublicKey string `json:"public_key"`
	Hash      string `json:"hash"`
	Data      []byte `json:"Data"`
}

func (m *MetaInfo) SetMetaInfo(filePath string, publicKey string) error {
	fp, err := os.Open(filePath)
	defer fp.Close()
	if err != nil {
		fmt.Println(filePath, err)
		return err
	}
	m.Size = 1.11
	m.FileName = "a"
	m.PublicKey = publicKey

	m.HashContent = ""
	m.Sig = ""
	m.CreateTime = ""
	m.Blocks = ""

	return nil
}

func (m *MetaInfo) Print() {
	data, _ := json.MarshalIndent(*m, "", "    ")
	fmt.Println(string(data))
}

func (b *Block) SetBlock(data string, publicKey string) error {
	b.hash = ""
	b.data = data
	b.PublicKey = publicKey
	return nil
}

func Split(file *os.File, size int) (map[int]string, error) {
	finfo, err := file.Stat()
	if err != nil {
		fmt.Println("get file info failed:", file, size)
		return _, err
	}

	bufsize := 1024 * 1024
	if size < bufsize {
		bufsize = size
	}

	buf := make([]byte, bufsize)
	num := (int(finfo.Size()) + size - 1) / size

	blocks = make(map[int]string)
	for i := 0; i < num; i++ {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			fmt.Println(err2, "failed to read from:", file)
			break
		}
		if n <= 0 {
			break
		}
		blocks[i] = buf[:n]
	}
	return blocks, nil
}
