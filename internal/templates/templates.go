package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/gin-gonic/gin"
)

//go:embed components/**/*.gohtml layouts/*.gohtml pages/*.gohtml
var TemplateFS embed.FS

func LoadTemplates(router *gin.Engine) error {
	return LoadTemplatesFromFS(router, TemplateFS)
}

func LoadTemplatesFromFS(router *gin.Engine, fsys fs.FS) error {
	templateFiles, err := findTemplateFiles(fsys)
	if err != nil {
		return err
	}

	if len(templateFiles) == 0 {
		return fmt.Errorf("no template files found")
	}

	cache := GetTemplateCache()
	tmpl := template.New("").Funcs(utils.TemplateFuncMap())

	for _, tmplPath := range templateFiles {
		content, err := fs.ReadFile(fsys, tmplPath)
		if err != nil {
			return fmt.Errorf("error reading template %s: %w", tmplPath, err)
		}

		name := filepath.Base(tmplPath)
		t := tmpl.New(name).Funcs(utils.TemplateFuncMap())

		if _, err := t.Parse(string(content)); err != nil {
			return fmt.Errorf("error parsing template %s: %w", tmplPath, err)
		}

		cache.Set(name, t)
	}

	router.SetFuncMap(utils.TemplateFuncMap())
	router.SetHTMLTemplate(tmpl)

	return nil
}

func GetTemplateFiles() ([]string, error) {
	return findTemplateFiles(TemplateFS)
}

func findTemplateFiles(fsys fs.FS) (files []string, err error) {
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
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
