package daemon

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/sirupsen/logrus"

	"github.com/samoslab/nebula/client/common"
	"github.com/samoslab/nebula/util/dbutil"
)

var (
	// task bucket
	taskBkt = []byte("client_task")
)

// Status send task Status
type Status int8

const (
	// StatusGotTask just got task
	StatusGotTask Status = iota
	// StatusDone coins sent and confirmed
	StatusDone
	// StatusUnknown fallback value
	StatusUnknown
)

var statusString = []string{
	StatusGotTask: "got_task",
	StatusDone:    "done",
	StatusUnknown: "unknown",
}

func (s Status) String() string {
	return statusString[s]
}

// store storage for task
type store struct {
	db  *bolt.DB
	log logrus.FieldLogger
}

// TaskFilter filter function
type TaskFilter func(rwd TaskInfo) bool

// newStore creates a store instance
func newStore(db *bolt.DB, log logrus.FieldLogger) (*store, error) {
	if db == nil {
		return nil, errors.New("new task store failed, db is nil")
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(taskBkt); err != nil {
			return dbutil.NewCreateBucketFailedErr(taskBkt, err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return &store{
		db:  db,
		log: log.WithField("prefix", "task.store"),
	}, nil
}

type TaskInfo struct {
	Key       string
	Status    Status
	Task      Task
	UpdatedAt uint64
	Seq       uint64
	Err       string
}

//StoreTask save task into db
func (s *store) StoreTask(task Task) (TaskInfo, error) {
	var taskInfo TaskInfo
	if err := s.db.Update(func(tx *bolt.Tx) error {
		seq, err := dbutil.NextSequence(tx, taskBkt)
		if err != nil {
			return err
		}
		taskkey := fmt.Sprintf("client-task:%d", seq)
		taskInfo = TaskInfo{Seq: seq, Task: task, Status: StatusGotTask, Key: taskkey, UpdatedAt: common.Now()}
		if err := dbutil.PutBucketValue(tx, taskBkt, taskkey, taskInfo); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return TaskInfo{}, err
	}
	return taskInfo, nil
}

//StoreTask finish task
func (s *store) UpdateTaskInfo(taskkey string, update func(TaskInfo) TaskInfo) (TaskInfo, error) {
	return s.UpdateTaskInfoCallback(taskkey, update, func(di TaskInfo) error { return nil })
}

// UpdateTaskInfoCallback updates deposit info. The update func takes a TaskInfo
// and returns a modified copy of it.  After updating the TaskInfo, it calls callback,
// inside of the transaction.  If the callback returns an error, the TaskInfo update
// is rolled back.
func (s *store) UpdateTaskInfoCallback(taskKey string, update func(TaskInfo) TaskInfo, callback func(TaskInfo) error) (TaskInfo, error) {
	log := s.log.WithField("taskKey", taskKey)

	var ri TaskInfo
	if err := s.db.Update(func(tx *bolt.Tx) error {
		if err := dbutil.GetBucketObject(tx, taskBkt, taskKey, &ri); err != nil {
			return err
		}

		log = log.WithField("taskInfo", ri)

		if ri.Key != taskKey {
			log.Error("TaskInfo.Key does not match taskKey")
			err := fmt.Errorf("TaskInfo %+v saved under different key %s", ri, taskKey)
			return err
		}

		ri = update(ri)
		ri.UpdatedAt = common.Now()

		if err := dbutil.PutBucketValue(tx, taskBkt, taskKey, ri); err != nil {
			return err
		}

		return callback(ri)

	}); err != nil {
		return TaskInfo{}, err
	}

	return ri, nil
}

// GetUnProcessTask returns filtered task info
func (s *store) GetTaskArray(flt TaskFilter) ([]TaskInfo, error) {
	var tasks []TaskInfo

	if err := s.db.View(func(tx *bolt.Tx) error {
		return dbutil.ForEach(tx, taskBkt, func(k, v []byte) error {
			var task TaskInfo
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}
			if flt(task) {
				tasks = append(tasks, task)
			}

			return nil
		})
	}); err != nil {
		return nil, err
	}

	return tasks, nil
}

//GetTask returns task
func (s *store) GetTask(taskID string) (TaskInfo, error) {
	var taskInfo TaskInfo
	if err := s.db.View(func(tx *bolt.Tx) error {
		if err := dbutil.GetBucketObject(tx, taskBkt, taskID, &taskInfo); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return TaskInfo{}, err
	}
	return taskInfo, nil
}
