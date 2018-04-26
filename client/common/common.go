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
