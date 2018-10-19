package main

import (
	"encoding/hex"
	"fmt"
	"os"

	util_bytes "github.com/samoslab/nebula/util/bytes"
	util_num "github.com/samoslab/nebula/util/num"
)

const filename_suffix = ".blk"
const ModFactorExp = 13
const ModFactor = 1 << ModFactorExp
const slash = "/"

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("% string", os.Args[0])
		return
	}
	key, err := hex.DecodeString(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	val := util_bytes.ToUint32(key, len(key)-4)
	sub1 := util_num.FixLength(val&(ModFactor-1), 4)
	sub2 := util_num.FixLength((val>>ModFactorExp)&(ModFactor-1), 4)
	filename := hex.EncodeToString(key)
	fmt.Println(sub1 + slash + sub2 + slash + filename + filename_suffix)
}
