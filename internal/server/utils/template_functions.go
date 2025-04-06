package utils

import (
	"html/template"
)

func TemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		"dict": func(values ...interface{}) (dict map[string]interface{}) {
			// If the length of the map is not an even number, it is malformed.
			if (len(values) % 2) != 0 {
				return nil
			}

			dict = make(map[string]interface{}, (len(values) / 2))

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
	}
}
