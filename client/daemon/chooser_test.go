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
	expectPro := backupPros[0]

	backMap := createBackupProvicer(backupPros)
	assert.Equal(t, 2, len(backMap))
	for _, v := range backMap {
		assert.Equal(t, 1, len(v))
		assert.Equal(t, expectPro, v[0])
	}

	choosed := chooseBackupProvicer([]byte("hash2"), backMap)
	assert.Equal(t, expectPro, choosed)
	choosed = chooseBackupProvicer([]byte("hash1"), backMap)
	assert.Nil(t, choosed)
	choosed = chooseBackupProvicer([]byte("hash2"), backMap)
	assert.Nil(t, choosed)
	choosed = chooseBackupProvicer([]byte("hash3"), backMap)
	assert.Nil(t, choosed)
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
	expectPros := []*mpb.BlockProviderAuth{pros[0], pros[2]}
	avaliablePros, err := generateAvaliablePro(sortPros)
	assert.NoError(t, err)
	assert.Equal(t, expectPros, avaliablePros)
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
			sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: 1001, OriginIndex: i})
		} else {
			sortPros = append(sortPros, &SortablePro{Pro: bpa, Delay: 999, OriginIndex: i})
		}
	}

	expectPros := []*mpb.BlockProviderAuth{pros[0], pros[5], pros[4], pros[3]}
	avaliablePros, err := generateAvaliablePro(sortPros)
	assert.NoError(t, err)
	assert.Equal(t, expectPros, avaliablePros)
}
