package daemon

import (
	"testing"

	"github.com/samoslab/nebula/client/common"
	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	req := &common.UploadReq{
		Filename:  "/root/abc.txt",
		Dest:      "/tmp/abc",
		IsEncrypt: false,
		Sno:       0,
	}

	task := NewTask(common.TaskUploadFileType, req)
	assert.Equal(t, req, task.Payload.(*common.UploadReq))
}
