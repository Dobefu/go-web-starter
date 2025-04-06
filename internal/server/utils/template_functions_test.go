package utils

import (
	"html/template"
	"testing"
)

func TestDict(t *testing.T) {
	dict := TemplateFuncMap()["dict"].(func(...interface{}) map[string]interface{})

	tests := []struct {
		name string
		args []interface{}
		want map[string]interface{}
	}{
		{"valid", []interface{}{"key1", "value1", "key2", 42}, map[string]interface{}{"key1": "value1", "key2": 42}},
		{"odd args", []interface{}{"key1", "value1", "key2"}, nil},
		{"non-string key", []interface{}{123, "value1", "key2", "value2"}, map[string]interface{}{"key2": "value2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dict(tt.args...)

			if tt.want == nil {
				if got != nil {
					t.Errorf("dict() = %v, want nil", got)
				}

				return
			}

			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("dict()[%q] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestRaw(t *testing.T) {
	raw := TemplateFuncMap()["raw"].(func(string) template.HTML)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple", "<p>Hello</p>", "<p>Hello</p>"},
		{"empty", "", ""},
		{"special", "<div>&lt;Hello&gt;</div>", "<div>&lt;Hello&gt;</div>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(raw(tt.in)); got != tt.want {
				t.Errorf("raw() = %q, want %q", got, tt.want)
			}
		})
	}
}
