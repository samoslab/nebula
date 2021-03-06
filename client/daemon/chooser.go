package daemon

import (
	"fmt"
	"sort"
	"sync"

	"github.com/samoslab/nebula/client/common"
	client "github.com/samoslab/nebula/client/provider_client"
	mpb "github.com/samoslab/nebula/tracker/metadata/pb"
)

func ChooseBackupProvicer(hash []byte, backProMap map[string][]*SortablePro) *SortablePro {
	hashKey := string(hash)
	if sortPros, ok := backProMap[hashKey]; ok {
		for i, _ := range sortPros {
			if !sortPros[i].Used && sortPros[i].Delay <= common.MaxInvalidDelay {
				sortPros[i].Used = true
				backProMap[hashKey] = sortPros
				return sortPros[i]
			}
		}
	}
	return nil
}

func CreateBackupProvicer(backupPros []*SortablePro) map[string][]*SortablePro {
	backProMap := map[string][]*SortablePro{}
	for _, sortPro := range backupPros {
		for _, pha := range sortPro.Pro.HashAuth {
			backProMap[string(pha.Hash)] = append(backProMap[string(pha.Hash)], sortPro)
		}
	}

	return backProMap
}

func GetBestReplicaProvider(pros []*mpb.ReplicaProvider, needNum int) ([]*mpb.ReplicaProvider, []*mpb.ReplicaProvider, error) {
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

	if len(wellPros) < MinReplicaNum {
		return nil, nil, fmt.Errorf("Replica provider only %d", len(wellPros))
	}

	return wellPros[0:common.Min(needNum, len(wellPros))], wellPros[common.Min(needNum, len(wellPros)):len(wellPros)], nil
}

// SortablePro provider sorted by delay
type SortablePro struct {
	Pro         *mpb.BlockProviderAuth
	Delay       int
	OriginIndex int
	Used        bool
}

// UsingBestProvider ping provider
func UsingBestProvider(pros []*mpb.BlockProviderAuth) ([]*SortablePro, []*SortablePro) {
	sortPros := []*SortablePro{}
	for i, bpa := range pros {
		sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: 0, OriginIndex: i})
	}
	workPros := []*SortablePro{}
	backupPros := []*SortablePro{}
	for _, proInfo := range sortPros {
		if !proInfo.Pro.GetSpare() {
			workPros = append(workPros, proInfo)
		} else {
			backupPros = append(backupPros, proInfo)
		}
	}

	return workPros, backupPros
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
