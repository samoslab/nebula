package filetype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileType(t *testing.T) {
	ft := FileType("not-exists.file")
	assert.Equal(t, ft.Type, "unknown")
	assert.Equal(t, ft.Extension, "unknown")
}
