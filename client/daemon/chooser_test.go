package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderBackupMap(t *testing.T) {
	workedNum := 40
	backupNum := 10
	backMap := createBackupProvicer(workedNum, backupNum)
	assert.Equal(t, 40, len(backMap))
	for _, v := range backMap {
		assert.Equal(t, 2, len(v))
	}

	usedMap := map[int]struct{}{}
	choosed := chooseBackupProvicer(0, backMap, usedMap)
	assert.Equal(t, 0, choosed)
	choosed = chooseBackupProvicer(4, backMap, usedMap)
	assert.Equal(t, 5, choosed)
	choosed = chooseBackupProvicer(5, backMap, usedMap)
	assert.Equal(t, -1, choosed)

	choosed = chooseBackupProvicer(33, backMap, usedMap)
	assert.Equal(t, 4, choosed)
	choosed = chooseBackupProvicer(39, backMap, usedMap)
	assert.Equal(t, 9, choosed)
	choosed = chooseBackupProvicer(5, backMap, usedMap)
	assert.Equal(t, -1, choosed)

	backMap = createBackupProvicer(6, 2)
	assert.Equal(t, 0, len(backMap))
	usedMap = map[int]struct{}{}
	choosed = chooseBackupProvicer(0, backMap, usedMap)
	assert.Equal(t, -1, choosed)
}
