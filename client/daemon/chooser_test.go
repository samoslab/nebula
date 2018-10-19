package daemon

import (
	"testing"

	"github.com/samoslab/nebula/client/common"
	mpb "github.com/samoslab/nebula/tracker/metadata/pb"
	"github.com/stretchr/testify/assert"
)

func createProviders() []*mpb.BlockProviderAuth {
	pros := []*mpb.BlockProviderAuth{
		&mpb.BlockProviderAuth{
			NodeId: []byte("node1"),
			Spare:  false,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash1")},
			},
		},
		&mpb.BlockProviderAuth{
			NodeId: []byte("node2"),
			Spare:  false,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash2")},
			},
		},
		&mpb.BlockProviderAuth{
			NodeId: []byte("node3"),
			Spare:  true,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash1")},
				&mpb.PieceHashAuth{Hash: []byte("hash2")},
			},
		},
	}
	return pros
}

func TestProviderBackupMap(t *testing.T) {
	pros := createProviders()
	sortPros := []SortablePro{}
	for i, bpa := range pros {
		if i == 1 {
			sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: common.NetworkUnreachable, OriginIndex: i})
		} else {
			sortPros = append(sortPros, SortablePro{Pro: bpa, Delay: i, OriginIndex: i})
		}
	}
	workPros := []SortablePro{}
	backupPros := []*SortablePro{}
	for _, proInfo := range sortPros {
		if !proInfo.Pro.GetSpare() {
			workPros = append(workPros, proInfo)
		} else {
			backupPros = append(backupPros, &proInfo)
		}
	}
	assert.Equal(t, 2, len(workPros))
	assert.Equal(t, 1, len(backupPros))
}

func TestGenerateAvaliablePro(t *testing.T) {
	pros := createProviders()
	sortPros := []*SortablePro{}
	for i, bpa := range pros {
		if i == 1 {
			sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: common.NetworkUnreachable, OriginIndex: i})
		} else {
			sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: i, OriginIndex: i})
		}
	}
}

func createSecondProviders() []*mpb.BlockProviderAuth {
	pros := []*mpb.BlockProviderAuth{
		&mpb.BlockProviderAuth{
			NodeId: []byte("node0"),
			Spare:  false,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash0")},
			},
		},
		&mpb.BlockProviderAuth{
			NodeId: []byte("node1"),
			Spare:  false,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash1")},
			},
		},
		&mpb.BlockProviderAuth{
			NodeId: []byte("node2"),
			Spare:  false,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash2")},
			},
		},
		&mpb.BlockProviderAuth{
			NodeId: []byte("node3"),
			Spare:  false,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash3")},
			},
		},
		&mpb.BlockProviderAuth{
			NodeId: []byte("node4"),
			Spare:  true,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash0")},
				&mpb.PieceHashAuth{Hash: []byte("hash2")},
			},
		},
		&mpb.BlockProviderAuth{
			NodeId: []byte("node5"),
			Spare:  true,
			HashAuth: []*mpb.PieceHashAuth{
				&mpb.PieceHashAuth{Hash: []byte("hash1")},
				&mpb.PieceHashAuth{Hash: []byte("hash3")},
			},
		},
	}
	return pros
}

func TestGenerateSecondAvaliablePro(t *testing.T) {
	pros := createSecondProviders()
	sortPros := []*SortablePro{}
	for i, bpa := range pros {
		if i == 1 {
			sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: common.NetworkUnreachable, OriginIndex: i})
		} else if i == 2 {
			sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: common.MaxInvalidDelay + 1, OriginIndex: i})
		} else {
			sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: common.MaxInvalidDelay - 1, OriginIndex: i})
		}
	}

}
