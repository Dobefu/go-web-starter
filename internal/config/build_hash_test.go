package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateBuildHash(t *testing.T) {
	origExec := executablePath
	defer func() { executablePath = origExec }()

	t.Run("executable error", func(t *testing.T) {
		executablePath = func() (string, error) { return "", os.ErrNotExist }

		hash := generateBuildHash()
		assert.Equal(t, "-", hash)
	})

	t.Run("file open error", func(t *testing.T) {
		executablePath = func() (string, error) { return "/bogus", nil }

		hash := generateBuildHash()
		assert.Equal(t, "-", hash)
	})

	t.Run("io.Copy error", func(t *testing.T) {
		tmpDir := t.TempDir()
		dirPath := filepath.Join(tmpDir, "dir")

		err := os.Mkdir(dirPath, 0755)
		assert.NoError(t, err)

		executablePath = func() (string, error) { return dirPath, nil }

		hash := generateBuildHash()
		assert.Equal(t, "-", hash)
	})
}
