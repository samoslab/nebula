package progress

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/samoslab/nebula/client/common"
)

// ProgressCell for progress bar
type ProgressCell struct {
	Sended     bool    `json:"-"`
	Type       string  `json:"type"`
	Total      uint64  `json:"-"`
	Current    uint64  `json:"-"`
	Time       uint64  `json:"-"`
	Rate       float64 `json:"rate"`
	Local      string  `json:"local"`
	LastReaded bool    `json:"-"`
	SpaceNo    int     `json:"spaco_no"`
}

func calRate(current, total uint64) float64 {
	if total == 0 {
		return 1.0
	}
	rateS := fmt.Sprintf("%0.2f", float64(current)/float64(total))
	rate, err := strconv.ParseFloat(rateS, 10)
	if err != nil {
		rate = 0.0
	}
	return rate
}

type ProgressReadable struct {
	Progress map[string]ProgressCell `json:"progress"`
}

// ProgressManager progress stats
type ProgressManager struct {
	Mutex                sync.Mutex
	PartitionToOriginMap map[string]string // a.txt.1 -> a.txt ; a.txt.2 -> a.txt for progress
	Progress             map[string]ProgressCell
}

// NewProgressManager create progress status manager
func NewProgressManager() *ProgressManager {
	pm := &ProgressManager{}
	pm.Progress = map[string]ProgressCell{}
	pm.PartitionToOriginMap = map[string]string{}
	return pm
}

// SetProgress set current progress file size
func (pm *ProgressManager) SetProgress(tp, fileName string, currentSize, totalSize uint64, sno uint32, local string) {
	pm.Progress[fileName] = ProgressCell{Type: tp, Total: totalSize, Current: currentSize, Rate: calRate(currentSize, totalSize), Time: common.Now(), SpaceNo: int(sno), Local: local}
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
			cell.Rate = calRate(cell.Current, cell.Total)
			cell.LastReaded = false
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
func (pm *ProgressManager) GetProgress(files []string) (ProgressReadable, error) {
	pm.Mutex.Lock()
	defer pm.Mutex.Unlock()
	if len(files) == 0 {
		return ProgressReadable{Progress: pm.Progress}, nil
	}
	mp := map[string]struct{}{}
	for _, file := range files {
		mp[file] = struct{}{}
	}
	a := map[string]ProgressCell{}
	for k, v := range pm.Progress {
		if !match(mp, k) {
			continue
		}
		a[k] = v
	}
	return ProgressReadable{Progress: a}, nil
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
		if !v.Sended && !v.LastReaded {
			msg := common.MakeSuccProgressMsg(v.Type, k, v.Rate, v.SpaceNo, v.Local)
			result = append(result, msg.Serialize())
			v.LastReaded = true
			if int(v.Rate) == 1 {
				v.Sended = true
			}
			pm.Mutex.Lock()
			pm.Progress[k] = v
			pm.Mutex.Unlock()
		}
	}
	return result, nil
}
