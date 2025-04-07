package templates

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed **/*.gohtml
var TemplateFS embed.FS

func LoadTemplates(router *gin.Engine) error {
	templateFiles, err := GetTemplateFiles()

	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "templates")

	if err != nil {
		return err
	}

	defer os.RemoveAll(tmpDir)

	for _, tmpl := range templateFiles {
		content, err := TemplateFS.ReadFile(tmpl)

		if err != nil {
			return err
		}

		fullPath := filepath.Join(tmpDir, tmpl)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return err
		}
	}

	router.LoadHTMLGlob(filepath.Join(tmpDir, "**/*.gohtml"))
	return nil
}

func GetTemplateFiles() ([]string, error) {
	var files []string

	err := fs.WalkDir(TemplateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".gohtml" {
			path = strings.TrimPrefix(path, "./")
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func GetTemplateContent(path string) ([]byte, error) {
	return TemplateFS.ReadFile(path)
}
