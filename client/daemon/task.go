package daemon

import (
	"github.com/samoslab/nebula/client/common"
)

// Task task executed in back-end
type Task struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// NewTask create a task according to type and request
func NewTask(tp string, req interface{}) Task {
	t := Task{Type: tp}
	switch tp {
	case common.TaskUploadFileType:
		t.Payload = req.(*common.UploadReq)
	case common.TaskUploadDirType:
		t.Payload = req.(*common.UploadDirReq)
	case common.TaskDownloadFileType:
		t.Payload = req.(*common.DownloadReq)
	case common.TaskDownloadDirType:
		t.Payload = req.(*common.DownloadDirReq)
	}
	return t
}
