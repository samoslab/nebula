package daemon

import (
	"sync"

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
		t.Payload = req.(common.UploadReq)
	}
	return t
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

func (t *TaskManager) Add(task Task) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.queue.Add(task)
}

func (t *TaskManager) Count() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.queue.Length()
}

func (t *TaskManager) First() Task {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	r := t.queue.Remove()
	return r.(Task)
}
