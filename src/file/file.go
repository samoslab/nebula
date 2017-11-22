package file

import (
	"errors"
	"fmt"
	"github.com/skycoin/skycoin/src/cipher"
	"os"
	"time"
)

type MetaInfo struct {
	Size     float32 `json:"size"`
	FileName string  `json:"file_name"`
	PubKey   string  `json:"public_key"`

	Hash       []byte `json:"hash"`
	Sig        string `json:"sig"`
	CreateTime int    `json:"create_time"`

	Blocks map[string]string `json:"blocks"`
}

type Block struct {
	PubKey string `json:"public_key"`
	Hash   string `json:"hash"`
	Data   []byte `json:"Data"`
}

type FileName os.file

func (m *MetaInfo) SetMetaInfo(filePath string) error {
	finfo, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}
	m.Size = float32(finfo.Size())
	m.FileName = path.Base(filePath)
	m.PubKey = cipher.pubKey

	m.Hash = cipher.SHA256
	m.Sig = cipher.Sig
	m.CreateTime = time.Now().Unix()
	m.Blocks = ""

	return nil
}

func (m *MetaInfo) Print() {
	data, _ := json.MarshalIndent(*m, "", "    ")
	fmt.Println(string(data))
}

func (b *Block) SetBlock(data string) error {
	b.hash = cipher.SHA256
	b.data = data
	b.PubKey = cipher.pubKey 
	return nil
}

func Split(file *os.File, size int) (map[int]string, error) {
	finfo, err := file.Stat()
	if err != nil {
		//fmt.Println("get file info failed:", file, size)
		return _, errors.New("get file info failed")
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
			fmt.Println(err, "failed to read from:", file)
			break
		}
		if n <= 0 {
			break
		}
		blocks[i] = buf[:n]
	}
	return blocks, nil
}
