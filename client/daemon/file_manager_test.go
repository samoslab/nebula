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

}
