package templates

import (
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockFS struct {
	fstest.MapFS
	readErr bool
	walkErr bool
}

func (m mockFS) Open(name string) (fs.File, error) {
	if m.readErr {
		return nil, fmt.Errorf("error reading template %s: mock error", name)
	}

	return m.MapFS.Open(name)
}

func (m mockFS) ReadFile(name string) ([]byte, error) {
	if m.readErr {
		return nil, fmt.Errorf("error reading template %s: mock error", name)
	}

	return m.MapFS.ReadFile(name)
}

func (m mockFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if m.walkErr {
		return nil, fmt.Errorf("mock walk error")
	}
	return m.MapFS.ReadDir(name)
}

func TestGetTemplateFiles(t *testing.T) {
	t.Run("valid template directory", func(t *testing.T) {
		files, err := GetTemplateFiles()
		assert.NoError(t, err)
		assert.NotEmpty(t, files)

		for _, file := range files {
			assert.Equal(t, ".gohtml", filepath.Ext(file))
		}
	})
}

func TestGetTemplateContent(t *testing.T) {
	tests := []struct {
		name         string
		templatePath string
		wantErr      bool
	}{
		{
			name:         "existing template",
			templatePath: "",
			wantErr:      false,
		},
		{
			name:         "non-existent template",
			templatePath: "non-existent.gohtml",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.templatePath == "" {
				files, err := GetTemplateFiles()
				assert.NoError(t, err)
				assert.NotEmpty(t, files)
				tt.templatePath = files[0]
			}

			content, err := GetTemplateContent(tt.templatePath)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, content)
			}
		})
	}
}

func TestLoadTemplates(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) (*gin.Engine, fs.FS)
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid templates",
			setup: func(t *testing.T) (*gin.Engine, fs.FS) {
				return gin.New(), TemplateFS
			},
			wantErr: false,
		},
		{
			name: "invalid template syntax",
			setup: func(t *testing.T) (*gin.Engine, fs.FS) {
				fs := fstest.MapFS{
					"invalid.gohtml": &fstest.MapFile{
						Data: []byte("{{if .InvalidSyntax}}"),
					},
				}
				return gin.New(), fs
			},
			wantErr: true,
			errMsg:  "error parsing template",
		},
		{
			name: "no template files",
			setup: func(t *testing.T) (*gin.Engine, fs.FS) {
				return gin.New(), fstest.MapFS{}
			},
			wantErr: true,
			errMsg:  "no template files found",
		},
		{
			name: "unreadable template file",
			setup: func(t *testing.T) (*gin.Engine, fs.FS) {
				fs := mockFS{
					MapFS: fstest.MapFS{
						"test.gohtml": &fstest.MapFile{
							Data: []byte("test"),
						},
					},
					readErr: true,
				}
				return gin.New(), fs
			},
			wantErr: true,
			errMsg:  "error reading template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, fs := tt.setup(t)
			err := LoadTemplatesFromFS(router, fs)

			if tt.wantErr {
				assert.Error(t, err)

				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFindTemplateFiles(t *testing.T) {
	t.Run("embedded templates", func(t *testing.T) {
		files, err := findTemplateFiles(TemplateFS)
		assert.NoError(t, err)
		assert.NotEmpty(t, files)

		for _, file := range files {
			assert.Equal(t, ".gohtml", filepath.Ext(file))
		}
	})

	t.Run("filesystem error", func(t *testing.T) {
		_, err := findTemplateFiles(mockFS{walkErr: true})
		assert.Error(t, err)
	})
}

func TestTemplateCache(t *testing.T) {
	cache := GetTemplateCache()
	assert.NotNil(t, cache)

	tmpl := template.New("test")
	cache.Set("test", tmpl)

	retrieved, ok := cache.Get("test")
	assert.True(t, ok)
	assert.Equal(t, tmpl, retrieved)

	_, ok = cache.Get("non-existent")
	assert.False(t, ok)
}

func TestLoadTemplatesWrapper(t *testing.T) {
	router := gin.New()
	err := LoadTemplates(router)
	assert.NoError(t, err)
}
