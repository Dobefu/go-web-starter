package static

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticFileSystem(t *testing.T) {
	fs, err := StaticFileSystem()
	assert.NoError(t, err)
	assert.NotNil(t, fs, "expected non-nil filesystem")

	file, err := fs.Open("/favicon.ico")
	assert.NoError(t, err)
	defer file.Close()

	stat, err := file.Stat()
	assert.NoError(t, err)
	assert.False(t, stat.IsDir(), "expected file, got directory")
}
