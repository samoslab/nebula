package hash

import (
	"crypto/sha1"
	"errors"
	"io"
	"os"
)

//Sha1File calculate file sha1 hash, filePath must be exist
func Sha1File(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}

	return hash.Sum(nil)[:20], nil
}

const file_read_buf_size = 8192

func Sha1FilePiece(filePath string, start uint32, size uint32) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	newPosition, err := file.Seek(int64(start), 0)
	if err != nil {
		return nil, err
	}
	//TODO comment blow if block
	if newPosition != int64(start) {
		return nil, errors.New("Seek file failed")
	}
	hash := sha1.New()
	buf := make([]byte, file_read_buf_size)
	for {
		if size < file_read_buf_size {
			buf = make([]byte, size)
		}
		bytesRead, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if bytesRead < len(buf) {
			return nil, io.EOF
		}
		hash.Write(buf)
		size -= uint32(len(buf))
		if size == 0 {
			break
		}
	}
	return hash.Sum(nil)[:20], nil
}
