package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// UnifiedResponse for all reponse format
type UnifiedResponse struct {
	Errmsg string `json:"errmsg"`
	Code   int    `json:"code"`
	Data   json.RawMessage
}

// MakeUnifiedHTTPResponse make uniformed format http response
func MakeUnifiedHTTPResponse(code int, data interface{}, errmsg string) (UnifiedResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return UnifiedResponse{}, err
	}
	return UnifiedResponse{
		Code:   code,
		Data:   jsonData,
		Errmsg: errmsg,
	}, nil
}

// DecodeUnifiedHTTPResponse decode to uniform response
func DecodeUnifiedHTTPResponse(rsp *http.Response) (*UnifiedResponse, error) {
	byteBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return DecodeResponse(byteBody)
}

// DecodeResponse decode http body
func DecodeResponse(response []byte) (*UnifiedResponse, error) {
	result := &UnifiedResponse{}
	if err := json.Unmarshal(response, result); err != nil {
		err = fmt.Errorf("Invalid json response : %v", err)
		return nil, err
	}
	return result, nil
}

// SendRequest send http request with signature header
func SendRequest(method, url, token string, reqBody io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	timestamp := fmt.Sprintf("%d", Now())
	hash := hmac.New(sha256.New, []byte(token))
	hash.Write([]byte(timestamp))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("auth", hex.EncodeToString(hash.Sum(nil)))

	rsp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return rsp, err
}

// PartitionFile partition for upload file prepare
type PartitionFile struct {
	OriginFileName string
	OriginFileHash []byte
	OriginFileSize uint64
	FileName       string
	Pieces         []HashFile
}

// UploadParameter parameter for stream store
type UploadParameter struct {
	OriginFileHash []byte
	OriginFileSize uint64
	HF             HashFile
	Checksum       bool
}

// HashFile file info for reedsolomon
type HashFile struct {
	FileSize   int64
	FileName   string
	FileHash   []byte
	SliceIndex int
}

// Now return current unix timestamp
func Now() uint64 {
	return uint64(time.Now().UTC().Unix())
}
