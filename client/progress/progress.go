package progress

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/samoslab/nebula/client/common"
)

// ProgressCell for progress bar
type ProgressCell struct {
	Type    string
	Total   uint64
	Current uint64
	Rate    float64
	Time    uint64
	Sended  bool
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
func (pm *ProgressManager) SetProgress(tp, fileName string, currentSize, totalSize uint64) {
	pm.Progress[fileName] = ProgressCell{Type: tp, Total: totalSize, Current: currentSize, Rate: 0.0, Time: common.Now()}
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
		if cell.Total > 0 {
			rate := fmt.Sprintf("%0.2f", float64(cell.Current)/float64(cell.Total))
			var err error
			cell.Rate, err = strconv.ParseFloat(rate, 10)
			if err != nil {
				cell.Rate = 0.0
			}
		}
		pm.Progress[fileName] = cell
		return nil
	}
	return fmt.Errorf("%s not in progress map", fileName)
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
		a[k] = v.Rate
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
	for k, v := range pm.Progress {
		if !match(mp, k) {
			continue
		}
		// skip already sended
		if !v.Sended {
			fmt.Printf("send %s, %+v\n", k, v)
			msg := common.MakeSuccProgressMsg(v.Type, k, v.Rate)
			result = append(result, msg.Serialize())
			if int(v.Rate) == 1 {
				v.Sended = true
				pm.Progress[k] = v
			}
		}
	}
	return result, nil
}
