package daemon

import (
	"fmt"
	"sort"
	"sync"

	"github.com/samoslab/nebula/client/common"
	client "github.com/samoslab/nebula/client/provider_client"
	mpb "github.com/samoslab/nebula/tracker/metadata/pb"
	"github.com/yanzay/log"
)

type IndexStatus struct {
	Index int
	Used  bool
}

func chooseBackupProvicer(current int, backupMap map[int][]IndexStatus) int {
	choosed := -1
	if arr, ok := backupMap[current]; ok {
		for i, _ := range arr {
			if !arr[i].Used {
				choosed = arr[i].Index
				arr[i].Used = true
				backupMap[current] = arr
				return choosed
			}
		}
	}
	return choosed
}

func createBackupProvicer(workedNum, backupNum int) map[int][]IndexStatus {
	backupMap := map[int][]IndexStatus{}
	if workedNum != 40 || backupNum != 10 {
		return backupMap
	}
	// workedNum = 40 , backupNum = 10
	// span = 40 /10 * 2 = 8 nextGroup = 10 /2 = 5
	// 0-7 --> [0, 5] ; 8-15 --> [1, 6] ; 16-23 --> [2, 7] ; 24-31 -->[3, 8]; 32-39 --> [4, 9]
	span := (workedNum / backupNum) * 2
	nextGroup := backupNum / 2
	for i := 0; i < workedNum; i++ {
		backupMap[i] = append(backupMap[i], IndexStatus{Index: i / span, Used: false})
		backupMap[i] = append(backupMap[i], IndexStatus{Index: i/span + nextGroup, Used: false})
	}

	return backupMap
}

func GetBestReplicaProvider(pros []*mpb.ReplicaProvider, needNum int) ([]*mpb.ReplicaProvider, error) {
	if len(pros) <= needNum {
		return pros, nil
	}
	type SortablePro struct {
		Pro   *mpb.ReplicaProvider
		Delay int
	}

	sortPros := []SortablePro{}
	// ping concurrent
	pingResultMap := map[int]int{}
	var pingResultMutex sync.Mutex

	var wg sync.WaitGroup

	for i, bpa := range pros {
		wg.Add(1)
		go func(i int, bpa *mpb.ReplicaProvider) {
			defer wg.Done()
			pingTime := client.GetPingTime(bpa.GetServer(), bpa.GetPort())
			pingResultMutex.Lock()
			defer pingResultMutex.Unlock()
			pingResultMap[i] = pingTime
		}(i, bpa)
	}
	wg.Wait()
	for i, bpa := range pros {
		pingTime, _ := pingResultMap[i]
		sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: pingTime})
	}

	sort.Slice(sortPros, func(i, j int) bool { return sortPros[i].Delay < sortPros[j].Delay })
	wellPros := []*mpb.ReplicaProvider{}
	for _, pro := range sortPros {
		if pro.Delay != common.NetworkUnreachable {
			wellPros = append(wellPros, pro.Pro)
		}
	}
	return wellPros[0:needNum], nil
}

// UsingBestProvider ping provider
func UsingBestProvider(pros []*mpb.BlockProviderAuth) ([]*mpb.BlockProviderAuth, error) {
	normalPros := []*mpb.BlockProviderAuth{}
	for _, proInfo := range pros {
		if !proInfo.GetSpare() {
			normalPros = append(normalPros, proInfo)
		}
	}
	return normalPros, nil
	type SortablePro struct {
		Pro         *mpb.BlockProviderAuth
		Delay       int
		OriginIndex int
	}

	sortPros := []SortablePro{}
	// TODO can ping concurrent
	pingResultMap := map[int]int{}
	var pingResultMutex sync.Mutex

	var wg sync.WaitGroup

	for i, bpa := range pros {
		wg.Add(1)
		go func(i int, bpa *mpb.BlockProviderAuth) {
			defer wg.Done()
			pingTime := client.GetPingTime(bpa.GetServer(), bpa.GetPort())
			pingResultMutex.Lock()
			defer pingResultMutex.Unlock()
			pingResultMap[i] = pingTime
		}(i, bpa)
	}
	wg.Wait()
	for i, bpa := range pros {
		pingTime, _ := pingResultMap[i]
		sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: pingTime, OriginIndex: i})
	}

	workPros := []SortablePro{}
	backupPros := []SortablePro{}
	for _, proInfo := range sortPros {
		if !proInfo.Pro.GetSpare() {
			workPros = append(workPros, proInfo)
		} else {
			backupPros = append(backupPros, proInfo)
		}
	}

	workedNum := len(workPros)
	backupNum := len(backupPros)

	backupMap := createBackupProvicer(workedNum, backupNum)

	availablePros := []*mpb.BlockProviderAuth{}
	for _, proInfo := range workPros {
		if proInfo.Delay == common.NetworkUnreachable {
			// provider cannot connect , choose one from backup
			log.Errorf("Provider %+v cannot connected", proInfo.Pro)
			if backupNum == 0 {
				log.Errorf("No backup provider for provider %d", proInfo.OriginIndex)
				return nil, fmt.Errorf("one of provider cannot connected and no backup provider")
			}
			choosed := chooseBackupProvicer(proInfo.OriginIndex, backupMap)
			if choosed == -1 {
				log.Errorf("No availbe provider for provider %d", proInfo.OriginIndex)
				return nil, fmt.Errorf("no more backup provider can be choosed")
			}
			availablePros = append(availablePros, backupPros[choosed].Pro)
		} else {
			availablePros = append(availablePros, proInfo.Pro)
		}
	}

	return availablePros, nil
}

// BestRetrieveNode ping retrieve node
func BestRetrieveNode(pros []*mpb.RetrieveNode) *mpb.RetrieveNode {
	//todo if provider ip is same
	type SortablePro struct {
		Pro   *mpb.RetrieveNode
		Delay int
	}

	sortPros := []SortablePro{}
	for _, bpa := range pros {
		pingTime := client.GetPingTime(bpa.GetServer(), bpa.GetPort())
		sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: pingTime})
	}

	sort.Slice(sortPros, func(i, j int) bool { return sortPros[i].Delay < sortPros[j].Delay })

	availablePros := []*mpb.RetrieveNode{}
	for _, proInfo := range sortPros {
		availablePros = append(availablePros, proInfo.Pro)
	}

	return availablePros[0]
}
