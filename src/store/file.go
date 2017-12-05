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

/**
* Group[网盘(文件中转站)|CDN(http)镜像|云盘]

* Dir
* File
* Block
 */

/**
* user
*
 */

type UserMeta struct {
	Pubkey     string  `json:"pubkey"`                //公钥
	Balance    float32 `json:"balance"`               //用户余额，多少SPO
	CreateTime int     `json:"create_time"`           //创建时间
	UpdateTime int     `json:"update_time,omitempty"` //更新时间
}

/*

 网盘额外参数
* 用户同步本地的目录到网盘上，本地目录存在大量的小文件（大文件)

/**
*
*
* 本地用户上传文件，文件分成了block,传完一个block，针对该block的位图设置为1
* 计算bitmap为1的个数，就知道进度
*/
type SyncMeta struct {
	FileHash   string `json:"file_hash"`             //文件的hash
	SyncBitmap string `'json:"sync_bitmap`           //上传块的进度
	CreateTime int    `json:"create_time"`           //创建时间
	UpdateTime int    `json:"update_time,omitempty"` //更新时间
}

/**
* 目录
* 	用户可以创建目录（目录存储是K:V)
*
 */
type DirMeta struct {
	Owner      string     `json:"owner"`                 //文件上传者的公钥
	name       string     `json:"name"`                  //当前目录名
	Files      []FileMeta `json:"files"`                 //多个文件
	Parent     DirMeta    `json:"parent_dir"`            //
	Childs     []DirMeta  `json:"child_dirs"`            //
	CreateTime int        `json:"create_time"`           //创建时间
	UpdateTime int        `json:"update_time,omitempty"` //更新时间
	Cipher     string     `json:"cipher,omitempty"`      //目录token密文，[方法|密文],比如[aes256|xxxxxxxx]，如果不为空，则blocks数据保密
	Comment    string     `json:"comment,omitempty"`     //目录注释
	Tag        string     `json:"tag,omitempty"`         //目录tag
}

type Extends struct {
	IsShare int8   `json:"is_share"`
	Cipher  string `json:"share_cipher,omitempty"`
	//FileMeta密文，[方法|密文],比如[aes256|xxxxxxxx]，如果不为空，则blocks数据保密
	//如果IsShare == 0,
	//BodyHash公开，

}

//
//高安全性文件
//	默认加密
//高可用性文件
//

type FileMeta struct {
	MetaHash []string `json:"meta_hash"` //MetaHash = sha256(BodyHash+Owner)
	//如果是同一个人上传，提示文件已经存在
	//不同的人，copy一份metainfo
	Extends  Extends `json:"extends,omitempty"`
	Size     float64 `json:"size,omitempty"` //文件总大小
	Name     string  `json:"file_name"`      //文件名字
	BodyHash string  `json:"body_hash"`      //文件的hash

	BlockMap map[int]Block `json:"blocks"` //每一块的hash，一个文件多块

	Owner       string `json:"owner"`                  //文件上传者的公钥
	Expire      int    `json:"expire"`                 //过期时间
	Priority    int    `json:"priority"`               //优先级
	Sig         string `json:"sig"`                    //上传用户的签名
	CreateTime  int    `json:"create_time"`            //创建时间
	UpdateTime  int    `json:"update_time,omitempty"`  //更新时间
	Ver         int    `json:"version,omitempty"`      //版本，只有s
	BlockLength int    `json:"block_length,omitempty"` //每一块的大小，默认512K
	Comment     string `json:"comment,omitempty"`      //文件注释
	Tag         string `json:"tag,omitempty"`          //文件tag
	Announce    string `json:"announce,omitempty"`     //最开始接收服务的地址，公钥地址（相当于Tracker)；可以多个，逗号分隔

}
type Block struct {
	Hash string `json:"hash"`
	Data []byte `json:"Data"`
}

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

	Blocks map[int]Block `json:"blocks"` //每一块的hash，一个文件多块

	Owner       string `json:"owner"`                  //文件上传者的公钥
	Expire      int    `json:"expire"`                 //过期时间
	Priority    int    `json:"priority"`               //优先级
	Sig         string `json:"sig"`                    //上传用户的签名
	CreateTime  int    `json:"create_time"`            //创建时间
	UpdateTime  int    `json:"update_time,omitempty"`  //更新时间
	Ver         int    `json:"version,omitempty"`      //版本，只有s
	BlockLength int    `json:"block_length,omitempty"` //每一块的大小，默认512K
	Cipher      string `json:"cipher,omitempty"`       //文件token密文，[方法|密文],比如[aes256|xxxxxxxx]，如果不为空，则blocks数据保密
	Comment     string `json:"comment,omitempty"`      //文件注释
	Tag         string `json:"tag,omitempty"`          //文件tag
	Announce    string `json:"announce,omitempty"`     //最开始接收服务的地址，公钥地址（相当于Tracker)；可以多个，逗号分隔

	//add more

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
