package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Dobefu/go-web-starter/internal/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/html"
)

//go:embed components/**/*.gohtml layouts/*.gohtml pages/*.gohtml email/*.gohtml email/**/*.gohtml
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

	m := minify.New()
	m.Add("text/html", &html.Minifier{
		KeepDocumentTags: true,
		KeepWhitespace:   false,
		KeepEndTags:      false,
		KeepQuotes:       false,
		TemplateDelims:   html.GoTemplateDelims,
	})

	for _, tmplPath := range templateFiles {
		content, err := fs.ReadFile(fsys, tmplPath)
		if err != nil {
			return fmt.Errorf("error reading template %s: %w", tmplPath, err)
		}

		preprocessed := preprocessTemplate(string(content))
		minified, err := m.String("text/html", preprocessed)

		if err != nil {
			return fmt.Errorf("error minifying template %s: %w", tmplPath, err)
		}

		name := filepath.Base(tmplPath)
		t := tmpl.New(name).Funcs(utils.TemplateFuncMap())

		_, err = t.Parse(minified)

		if err != nil {
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

func preprocessTemplate(content string) string {
	// Join split HTML attributes into a single line.
	content = regexp.MustCompile(`\s{0,9}\n\s{0,9}([a-zA-Z-]{1,9}=)`).ReplaceAllString(content, " $1")

	return content
}
