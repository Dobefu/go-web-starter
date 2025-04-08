package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed **/*.gohtml
var TemplateFS embed.FS

var templateFuncs = template.FuncMap{
	"dict": func(values ...interface{}) (map[string]interface{}, error) {
		if (len(values) % 2) != 0 {
			return nil, nil
		}

		dict := make(map[string]interface{}, len(values)/2)

		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)

			if !ok {
				return nil, nil
			}

			dict[key] = values[i+1]
		}

		return dict, nil
	},
}

func LoadTemplates(router *gin.Engine) error {
	return LoadTemplatesFromFS(router, TemplateFS)
}

func LoadTemplatesFromFS(router *gin.Engine, fsys fs.FS) error {
	templateFiles := make([]string, 0)

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && filepath.Ext(path) == ".gohtml" {
			path = strings.TrimPrefix(path, "./")
			templateFiles = append(templateFiles, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if len(templateFiles) == 0 {
		return fmt.Errorf("no template files found")
	}

	tmpl := template.New("").Funcs(templateFuncs)

	for _, tmplPath := range templateFiles {
		content, err := fs.ReadFile(fsys, tmplPath)

		if err != nil {
			return fmt.Errorf("error reading template %s: %w", tmplPath, err)
		}

		_, err = tmpl.New(tmplPath).Parse(string(content))

		if err != nil {
			return fmt.Errorf("error parsing template %s: %w", tmplPath, err)
		}
	}

	router.SetFuncMap(templateFuncs)
	router.SetHTMLTemplate(tmpl)

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
