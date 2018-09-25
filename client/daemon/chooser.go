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

func chooseBackupProvicer(hash []byte, backProMap map[string][]*SortablePro) *SortablePro {
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

func createBackupProvicer(backupPros []*SortablePro) map[string][]*SortablePro {
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

type SortablePro struct {
	Pro         *mpb.BlockProviderAuth
	Delay       int
	OriginIndex int
	Used        bool
}

// UsingBestProvider ping provider
func UsingBestProvider(pros []*mpb.BlockProviderAuth) ([]*mpb.BlockProviderAuth, error) {
	sortPros := []*SortablePro{}
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
		sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: pingTime, OriginIndex: i})
	}

	return generateAvaliablePro(sortPros)
}

func generateAvaliablePro(sortPros []*SortablePro) ([]*mpb.BlockProviderAuth, error) {
	workPros := []*SortablePro{}
	backupPros := []*SortablePro{}
	for _, proInfo := range sortPros {
		if !proInfo.Pro.GetSpare() {
			workPros = append(workPros, proInfo)
		} else {
			backupPros = append(backupPros, proInfo)
		}
	}

	backupProMap := createBackupProvicer(backupPros)

	availablePros := []*mpb.BlockProviderAuth{}
	for _, proInfo := range workPros {
		if proInfo.Delay == common.NetworkUnreachable || proInfo.Delay > common.MaxInvalidDelay {
			// provider cannot connect , choose one from backup
			log.Errorf("Provider %+v cannot connected", proInfo.Pro)
			replacePro := chooseBackupProvicer(proInfo.Pro.HashAuth[0].Hash, backupProMap)
			if replacePro == nil {
				log.Errorf("No availbe provider for provider %d", proInfo.OriginIndex)
				return nil, fmt.Errorf("no more backup provider can be choosed")
			}
			availablePros = append(availablePros, replacePro.Pro)
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
