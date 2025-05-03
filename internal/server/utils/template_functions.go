package utils

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Dobefu/go-web-starter/internal/static"
)

func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"dict": func(values ...any) (dict map[string]any) {
			// If the length of the map is not an even number, it is malformed.
			if (len(values) % 2) != 0 {
				return nil
			}

			dict = make(map[string]any, (len(values) / 2))

			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)

				if !ok {
					continue
				}

				dict[key] = values[i+1]
			}

			return dict
		},
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"startswith": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
		"readfile": func(icon string) string {
			subFS, err := static.GetStaticFS()

			if err != nil {
				return ""
			}

			content, err := fs.ReadFile(subFS, filepath.Join("icons", fmt.Sprintf("%s.svg", icon)))

			if err != nil {
				return ""
			}

			return string(content)
		},
		"replace": func(s string, old string, new string) string {
			return strings.ReplaceAll(s, old, new)
		},
		"slice": func(items ...any) []any {
			return items
		},
		"dump": func(v any) template.HTML {
			json, err := json.MarshalIndent(v, "", "  ")

			if err != nil {
				return template.HTML(fmt.Sprintf("<pre>%#v</pre>", v))
			}

			return template.HTML(fmt.Sprintf("<pre>%s</pre>", string(json)))
		},
		"trimTrailingNewline": func(s string) string {
			return strings.TrimRight(s, "\r\n")
		},
	}
}
