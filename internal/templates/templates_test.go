package templates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetTemplateFiles(t *testing.T) {
	files, err := GetTemplateFiles()
	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	for _, file := range files {
		assert.Equal(t, ".gohtml", filepath.Ext(file))
	}
}

func TestGetTemplateContent(t *testing.T) {
	files, err := GetTemplateFiles()
	assert.NoError(t, err)
	assert.NotEmpty(t, files)

	content, err := GetTemplateContent(files[0])
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
}

func TestLoadTemplates(t *testing.T) {
	router := gin.New()

	err := LoadTemplates(router)
	assert.NoError(t, err)
}

func TestLoadTemplatesInvalidTemplate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "invalid-templates")
	assert.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	templateDir := filepath.Join(tmpDir, "templates")
	err = os.MkdirAll(templateDir, 0755)
	assert.NoError(t, err)

	invalidTemplate := filepath.Join(templateDir, "invalid.gohtml")
	err = os.WriteFile(invalidTemplate, []byte("{{if .InvalidSyntax}}"), 0644)
	assert.NoError(t, err)

	router := gin.New()

	err = LoadTemplatesFromFS(router, os.DirFS(tmpDir))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error parsing template")
}

func TestGetTemplateContentNonExistent(t *testing.T) {
	_, err := GetTemplateContent("non-existent.gohtml")
	assert.Error(t, err)
}
