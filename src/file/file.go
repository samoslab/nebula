package main

import (
	//"errors"
	"fmt"
	//"github.com/skycoin/skycoin/src/cipher"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"time"
	"reflect"
)

type MetaInfo struct {
	Size       float32       `json:"size"`
	FileName   string        `json:"file_name"`
	PubKey     []byte        `json:"public_key"`
	Hash       string        `json:"hash"`
	Sig        string        `json:"sig"`
	CreateTime int64         `json:"create_time"`
	Blocks     map[int]Block `json:"blocks"`
}

type Block struct {
	PubKey []byte `json:"public_key"`
	Hash   string `json:"hash"`
	Data   []byte `json:"Data"`
}

func main() {
	m := MetaInfo{}
	m.SetMetaInfo("/tmp/a", []byte("abcd"))
	fmt.Println(m)
}

func (m *MetaInfo) SetMetaInfo(filePath string, pubKey []byte) error {
	finfo, err := os.Stat(filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}

	fp, err := os.Open(filePath)
	defer fp.Close()
	if err != nil {
		log.Fatalln(err)
	}

	m.Size = float32(finfo.Size())
	m.FileName = finfo.Name()
	m.PubKey = pubKey

	m.Hash = "" //cipher.SHA256
	m.Sig = ""  //cipher.Sig
	m.CreateTime = time.Now().Unix()

	blocks := split(fp)
	for key, content := range bytes.Fields(blocks) {
		content, err := block(content, []byte("abc"))
		if nil != err {
			log.Fatalln(err)
		}
		fmt.Println("type:", reflect.TypeOf(content))
		fmt.Println(key)
		//block := Block{}
		//block = content
		m.Blocks[key] = Block{}
		//block
	}

	return nil
}

func (m *MetaInfo) Print() {
	data, _ := json.MarshalIndent(*m, "", "    ")
	fmt.Println(string(data))
}

func block(data []byte, pubKey []byte) (Block, error) {
	block := Block{
		Hash:   "",
		Data:   data,
		PubKey: pubKey,
	}
	return block, nil
}

func split(f *os.File) []byte {
	var blocks []byte
	for {
		buf := make([]byte, 1024)
		switch nr, err := f.Read(buf[:]); true {
		case nr < 0:
			fmt.Fprintf(os.Stderr, "cat: error reading: %s\n", err.Error())
			os.Exit(1)
		case nr == 0: // EOF
			return blocks
		case nr > 0:
			blocks = append(blocks, buf...)
		}
	}
}
