package file

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserHome(t *testing.T) {
	home := UserHome()
	require.NotEqual(t, home, "")
}
