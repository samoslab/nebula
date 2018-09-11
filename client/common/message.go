package common

import (
	"encoding/json"
)

type Message struct {
	Type MsgType `json:"type"`
	Data json.RawMessage
}

func MakeMsg(tp MsgType, data interface{}) (Message, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return Message{}, err
	}
	return Message{
		Type: tp,
		Data: jsonData,
	}, nil
}

type MsgType int

const (
	TypeUploadFileDone MsgType = iota
	TypeUploadDirDone
	TypeDownloadFileDone
	TypeDownloadDirDone
	TypeUploadProgress
	TypeDownloadProgress
	TypeUnknown
)

var typeString = []string{
	TypeUploadFileDone:   "UploadFileDone",
	TypeUploadDirDone:    "UploadDirDone",
	TypeDownloadFileDone: "DownloadFileDone",
	TypeDownloadDirDone:  "DownloadDirDone",
	TypeUploadProgress:   "UploadProgress",
	TypeDownloadProgress: "DownloadProgress",
	TypeUnknown:          "Unknown",
}

func (m MsgType) String() string {
	return typeString[m]
}

type DoneMsg struct {
	Type    string `json:"type"`
	Key     string `json:"key"`
	Local   string `json:"local"`
	SpaceNo uint32 `json:"space_no"`
	Code    uint32 `json:"code"`
	Err     string `json:"error"`
}

func (m *DoneMsg) SetError(code uint32, err error) *DoneMsg {
	m.Code = code
	m.Err = err.Error()
	return m
}

func (m *DoneMsg) Serialize() string {
	data, _ := json.Marshal(m)
	return string(data)
}

func MakeSuccDoneMsg(tp string, fileName string, spaceNo uint32) *DoneMsg {
	return &DoneMsg{
		Type:    tp,
		Key:     fileName,
		SpaceNo: spaceNo,
		Code:    0,
		Err:     "",
	}
}

type ProgressMsg struct {
	Type     string  `json:"type"`
	Key      string  `json:"key"`
	Local    string  `json:"local"`
	Progress float64 `json:"progress"`
	SpaceNo  int     `json:"space_no"`
	Code     uint32  `json:"code"`
	Err      string  `json:"error"`
}

func (m *ProgressMsg) SetError(code uint32, err error) *ProgressMsg {
	m.Code = code
	m.Err = err.Error()
	return m
}

func (m *ProgressMsg) Serialize() string {
	data, _ := json.Marshal(m)
	return string(data)
}

func MakeSuccProgressMsg(tp, fileName string, progress float64, sno int, local string) *ProgressMsg {
	return &ProgressMsg{
		Type:     tp,
		Key:      fileName,
		Progress: progress,
		Local:    local,
		SpaceNo:  sno,
		Code:     0,
		Err:      "",
	}
}
