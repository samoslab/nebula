package daemon

import (
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/samoslab/nebula/client/common"
)

type Task struct {
	ID      string           `json:"task_id"`
	Payload common.UploadReq `json:"payload"`
}

func Hash(msg string) string {
	Md5Inst := md5.New()
	Md5Inst.Write([]byte(msg))
	result := Md5Inst.Sum([]byte(""))
	return fmt.Sprintf("%x", result)
}

func NewTask(req common.UploadReq) *Task {
	id := Hash(fmt.Sprintf("%s:%s:%d:%v", req.Filename, req.Dest, req.Sno, req.IsEncrypt))
	return &Task{
		ID:      id,
		Payload: req,
	}
}

type TaskManager struct {
	mutex sync.Mutex
	queue *Queue
}

func NewTaskManager() *TaskManager {
	return &TaskManager{
		queue: New(),
	}
}

func (t *TaskManager) Shutdown() {
}

func (t *TaskManager) Add(req common.UploadReq) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	task := NewTask(req)
	t.queue.Add(task)
}

func (t *TaskManager) Count() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.queue.Length()
}

func (t *TaskManager) First() common.UploadReq {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	r := t.queue.Remove()
	task := r.(*Task)
	return task.Payload
}
