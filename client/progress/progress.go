package progress

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/samoslab/nebula/client/common"
)

// ProgressCell for progress bar
type ProgressCell struct {
	Total   uint64
	Current uint64
	Rate    float64
	Time    uint64
}

// ProgressManager progress stats
type ProgressManager struct {
	Progress             map[string]ProgressCell
	PartitionToOriginMap map[string]string // a.txt.1 -> a.txt ; a.txt.2 -> a.txt for progress
	Mutex                sync.Mutex
}

// NewProgressManager create progress status manager
func NewProgressManager() *ProgressManager {
	pm := &ProgressManager{}
	pm.Progress = map[string]ProgressCell{}
	pm.PartitionToOriginMap = map[string]string{}
	return pm
}

// SetProgress set current progress file size
func (pm *ProgressManager) SetProgress(fileName string, currentSize, totalSize uint64) {
	pm.Progress[fileName] = ProgressCell{Total: totalSize, Current: currentSize, Rate: 0.0, Time: common.Now()}
}

// SetPartitionMap set progress file map
func (pm *ProgressManager) SetPartitionMap(fileName, originFile string) {
	pm.PartitionToOriginMap[fileName] = originFile
}

// SetIncrement set increment
func (pm *ProgressManager) SetIncrement(fileName string, increment uint64) error {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	if cell, ok := pm.Progress[fileName]; ok {
		cell.Current = cell.Current + increment
		cell.Time = common.Now()
		pm.Progress[fileName] = cell
		return nil
	}
	return errors.New("not in progress map")
}

func match(fileMap map[string]struct{}, file string) bool {
	if len(fileMap) == 0 {
		return true
	}
	_, ok := fileMap[file]
	return ok
}

// GetProgress return progress data
func (pm *ProgressManager) GetProgress(files []string) (map[string]float64, error) {
	mp := map[string]struct{}{}
	for _, file := range files {
		mp[file] = struct{}{}
	}
	a := map[string]float64{}
	for k, v := range pm.Progress {
		if !match(mp, k) {
			continue
		}
		if v.Total != 0 {
			rate := fmt.Sprintf("%0.2f", float64(v.Current)/float64(v.Total))
			x, err := strconv.ParseFloat(rate, 10)
			if err != nil {
				return a, err
			}
			a[k] = x
		} else {
			a[k] = 0.0
		}
	}
	return a, nil
}

// GetProgress return progress data
func (pm *ProgressManager) GetProgressingMsg(files []string) ([]string, error) {
	result := []string{}
	mp := map[string]struct{}{}
	for _, file := range files {
		mp[file] = struct{}{}
	}
	a := map[string]float64{}
	for k, v := range pm.Progress {
		fmt.Printf("get %s, %+v\n", k, v)
		if !match(mp, k) {
			continue
		}
		if v.Total != 0 {
			rate := fmt.Sprintf("%0.2f", float64(v.Current)/float64(v.Total))
			x, err := strconv.ParseFloat(rate, 10)
			if err != nil {
				return result, err
			}
			// skip finished time bigger than 60s
			if int(x) == 1 && (common.Now()-v.Time) > 60 {
				continue
			}
			a[k] = x
		} else {
			a[k] = 0.0
		}
	}
	for k, v := range a {
		msg := common.MakeSuccProgressMsg(common.TaskUploadProgressType, k, v)
		result = append(result, msg.Serialize())
	}
	return result, nil
}
