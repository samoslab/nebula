package daemon

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordPadding(t *testing.T) {
	passwd := "abcde"
	realPasswd, err := passwordPadding(passwd, 0)
	assert.NoError(t, err)
	assert.Equal(t, 16, len(realPasswd))
	assert.Equal(t, passwd+"00000000000", realPasswd)
	realPasswd, err = passwordPadding(passwd, 1)
	assert.NoError(t, err)
	assert.Equal(t, 32, len(realPasswd))
	assert.Equal(t, passwd+"000000000000000000000000000", realPasswd)
	realPasswd, err = passwordPadding(passwd, 2)
	assert.Error(t, err)

}
