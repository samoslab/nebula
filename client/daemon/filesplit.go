package daemon

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

const chunkSize int64 = 1024 * 1024

// Slice record file slice for split
type Slice struct {
	Begin int64
	End   int64
}

func FileSlice(fileName string, chunkSize int64) ([]Slice, error) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}
	num := int64(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))
	sliceArray := make([]Slice, num)
	var i int64 = 0
	for ; i < int64(num); i++ {
		if i+1 == num {
			sliceArray[i].Begin = i * chunkSize
			sliceArray[i].End = fileInfo.Size()
		} else {
			sliceArray[i].Begin = i * chunkSize
			sliceArray[i].End = (i+1)*chunkSize - 1
		}
	}
	fmt.Printf("slice %+v\n", sliceArray)

	return sliceArray, nil
}

func FileShardNum(fileName string, chunkSize int64) (int, error) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return 0, err
	}

	num := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))
	return num, nil
}

// FileSplit split file by size
func FileSplit(outDir, fileName string, chunkSize int64) ([]string, error) {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}

	num := int(math.Ceil(float64(fileInfo.Size()) / float64(chunkSize)))

	fi, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer fi.Close()
	b := make([]byte, chunkSize)
	partFiles := make([]string, num)
	var i int64 = 1
	for ; i <= int64(num); i++ {

		fi.Seek((i-1)*(chunkSize), 0)
		if len(b) > int((fileInfo.Size() - (i-1)*chunkSize)) {
			b = make([]byte, fileInfo.Size()-(i-1)*chunkSize)
		}

		fi.Read(b)

		filename := filepath.Join(outDir, strconv.Itoa(int(i))+".part")
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		f.Write(b)
		f.Close()
		partFiles = append(partFiles, filename)
	}
	return partFiles, nil
}

func FileJoin(num int) {
	fii, err := os.OpenFile("test.zip.1", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 1; i <= num; i++ {
		f, err := os.OpenFile("./"+strconv.Itoa(int(i))+".db", os.O_RDONLY, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println(err)
			return
		}
		fii.Write(b)
		f.Close()
	}
}
