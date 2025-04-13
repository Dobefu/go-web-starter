package utils

import (
	"html/template"
	"strings"
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
		"raw": func(s string) template.HTML {
			return template.HTML(s)
		},
		"startswith": func(s, prefix string) bool {
			return strings.HasPrefix(s, prefix)
		},
	}
}
