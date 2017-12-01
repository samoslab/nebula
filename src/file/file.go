package main

import (
	//"errors"
	"fmt"
	//"github.com/skycoin/skycoin/src/cipher"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"reflect"
	"time"
)

/***
* 场景：
* 1,查询文件存在不 get_metainfo_by_hash
* 2,查询块存在不 
*
 */

type MetaInfo struct {
	//一个文件存多个块
	Size float64 `json:"size,omitempty"`      //文件总大小
	Name string  `json:"file_name,omitempty"` //文件名字
	Hash string  `json:"file_hash,omitempty"` //文件的hash
	
	//一个block存放多个文件
	Sizes   []float32 `json:"sizes,omitempty"`      //多文件的大小
	Names   []string  `json:"file_names,omitempty"` //多个文件
	Hashs   []string  `json:"file_hashs,omitempty"` //多个文件的hash
	Offsets []float32 `json:"offsets,omitempty"`    //偏移

	Blocks      []string `json:"blocks"`       //每一块的hash，多文件存一个块是

	
	Publisher   string   `json:"publisher"`    //文件上传者的公钥
	Expire      int      `json:"expire"`       //过期时间
	Priority    int      `json:"priority"`     //优先级
	Sig         string   `json:"sig"`          //上传用户的签名
	CreateTime  int      `json:"create_time"`  //创建时间
	UpdateTime  int      `json:"update_time,omitempty"`  //更新时间
	Ver         int      `json:"version,omitempty"`      //版本，只有s
	BlockLength int      `json:"block_length,omitempty"` //每一块的大小，默认512K
	Cipher      string   `json:"cipher,omitempty"`       //文件token密文，[方法|密文],比如[aes256|xxxxxxxx]，如果不为空，则blocks数据保密
	Comment     string   `json:"comment,omitempty"`      //文件注释
	Tag         string   `json:"tag,omitempty"`          //文件tag
	Announce    string   `json:"announce,omitempty"`     //最开始接收服务的地址，公钥地址（相当于Tracker)；可以多个，逗号分隔

	//add more

}

type Block struct {
	Hash string `json:"hash"`
	Data []byte `json:"Data"`
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
