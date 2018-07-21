package daemon

import (
	"testing"

	"github.com/samoslab/nebula/client/common"
	"github.com/stretchr/testify/assert"
)

func TestTask(t *testing.T) {
	tm := NewTaskManager()
	req := common.UploadReq{
		Filename:  "/root/abc.txt",
		Dest:      "/tmp/abc",
		IsEncrypt: false,
		Sno:       0,
	}

	tm.Add(req)
	assert.Equal(t, 1, tm.Count())

	req1 := common.UploadReq{
		Filename:  "/home/xxx/abc.txt",
		Dest:      "/tmp/xxx",
		IsEncrypt: true,
		Sno:       1,
	}

	tm.Add(req1)
	assert.Equal(t, 2, tm.Count())

	origin := tm.First()
	assert.Equal(t, req, origin)

	origin = tm.First()
	assert.Equal(t, req1, origin)
	assert.Equal(t, 0, tm.Count())
}
