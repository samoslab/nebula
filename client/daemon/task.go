package daemon

import (
	"github.com/samoslab/nebula/client/common"
)

type Task struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

func NewTask(tp string, req interface{}) Task {
	t := Task{Type: tp}
	switch tp {
	case common.TaskUploadFileType:
		t.Payload = req.(*common.UploadReq)
	case common.TaskUploadDirType:
		t.Payload = req.(*common.UploadDirReq)
	}
	return t
}
