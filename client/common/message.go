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
	Type   string `json:"type"`
	Source string `json:"source"`
	Code   uint32 `json:"code"`
	Err    string `json:"error"`
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

func MakeSuccDoneMsg(tp string, fileName string) *DoneMsg {
	return &DoneMsg{
		Type:   tp,
		Source: fileName,
		Code:   0,
		Err:    "",
	}
}

type ProgressMsg struct {
	Type     string  `json:"type"`
	FileName string  `json:"filename"`
	Progress float64 `json:"progress"`
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

func MakeSuccProgressMsg(tp, fileName string, progress float64) *ProgressMsg {
	return &ProgressMsg{
		Type:     tp,
		FileName: fileName,
		Progress: progress,
		Code:     0,
		Err:      "",
	}
}
