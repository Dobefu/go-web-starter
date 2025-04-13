package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/css/dist/* static/js/dist/* static/icons/* static/favicon.*
var StaticFS embed.FS

func GetStaticFS() (fs.FS, error) {
	return fs.Sub(StaticFS, "static")
}

func StaticFileSystem() (http.FileSystem, error) {
	subFS, err := GetStaticFS()

	if err != nil {
		return nil, err
	}

	return http.FS(subFS), nil
}
