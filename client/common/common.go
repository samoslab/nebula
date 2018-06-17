package common

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ProgressCell for progress bar
type ProgressCell struct {
	Total   uint64
	Current uint64
	Rate    float64
}

// UnifiedResponse for all reponse format
type UnifiedResponse struct {
	Errmsg string `json:"errmsg"`
	Code   int    `json:"code"`
	Data   json.RawMessage
}

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

func DecodeUnifiedHTTPResponse(rsp *http.Response) (*UnifiedResponse, error) {
	byteBody, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	return DecodeResponse(byteBody)
}

func DecodeResponse(response []byte) (*UnifiedResponse, error) {
	result := &UnifiedResponse{}
	if err := json.Unmarshal(response, result); err != nil {
		err = fmt.Errorf("Invalid json response : %v", err)
		return nil, err
	}
	return result, nil
}

func SendRequest(method, url, token string, reqBody io.Reader) (*http.Response, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	timestamp := fmt.Sprintf("%d", time.Now().UTC().Unix())
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

// ProgressManager progress stats

type ProgressManager struct {
	Progress             map[string]ProgressCell
	PartitionToOriginMap map[string]string // a.txt.1 -> a.txt ; a.txt.2 -> a.txt for progress
	Mutex                sync.Mutex
}

// NewProgressManager create progress status manager
func NewProgressManager() *ProgressManager {
	pm := &ProgressManager{}
	pm.Progress = map[string]ProgressCell{}
	pm.PartitionToOriginMap = map[string]string{}
	return pm
}

// SetProgress set current progress file size
func (pm *ProgressManager) SetProgress(fileName string, currentSize, totalSize uint64) {
	pm.Progress[fileName] = ProgressCell{Total: totalSize, Current: currentSize, Rate: 0.0}
}

// SetPartitionMap set progress file map
func (pm *ProgressManager) SetPartitionMap(fileName, originFile string) {
	pm.PartitionToOriginMap[fileName] = originFile
}

// SetIncrement set increment
func (pm *ProgressManager) SetIncrement(fileName string, increment uint64) error {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	if cell, ok := pm.Progress[fileName]; ok {
		cell.Current = cell.Current + increment
		pm.Progress[fileName] = cell
		return nil
	}
	return errors.New("not in progress map")
}

func match(fileMap map[string]struct{}, file string) bool {
	if len(fileMap) == 0 {
		return true
	}
	_, ok := fileMap[file]
	return ok
}

// GetProgress return progress data
func (pm *ProgressManager) GetProgress(files []string) (map[string]float64, error) {
	mp := map[string]struct{}{}
	for _, file := range files {
		mp[file] = struct{}{}
	}
	a := map[string]float64{}
	for k, v := range pm.Progress {
		if !match(mp, k) {
			continue
		}
		if v.Total != 0 {
			rate := fmt.Sprintf("%0.2f", float64(v.Current)/float64(v.Total))
			x, err := strconv.ParseFloat(rate, 10)
			if err != nil {
				return a, err
			}
			a[k] = x
		} else {
			a[k] = 0.0
		}
	}
	return a, nil
}

func Now() uint64 {
	return uint64(time.Now().UTC().Unix())
}
