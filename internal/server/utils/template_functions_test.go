package utils

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDict(t *testing.T) {
	dict := TemplateFuncMap()["dict"].(func(...any) map[string]any)

	tests := []struct {
		name string
		args []any
		want map[string]any
	}{
		{"valid", []any{"key1", "value1", "key2", 42}, map[string]any{"key1": "value1", "key2": 42}},
		{"odd args", []any{"key1", "value1", "key2"}, nil},
		{"non-string key", []any{123, "value1", "key2", "value2"}, map[string]any{"key2": "value2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dict(tt.args...)

			if tt.want == nil {
				assert.Nil(t, got)
				return
			}

			assert.Equal(t, tt.want, got)
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
		{"simple HTML", "<p>Hello</p>", "<p>Hello</p>"},
		{"empty string", "", ""},
		{"escaped HTML", "<div>&lt;Hello&gt;</div>", "<div>&lt;Hello&gt;</div>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(raw(tt.in))
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStartsWith(t *testing.T) {
	startswith := TemplateFuncMap()["startswith"].(func(string, string) bool)

	tests := []struct {
		name   string
		s      string
		prefix string
		want   bool
	}{
		{"matching prefix", "hello world", "hello", true},
		{"non-matching prefix", "hello world", "world", false},
		{"empty string", "", "", true},
		{"empty prefix", "hello", "", true},
		{"case sensitive", "Hello", "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := startswith(tt.s, tt.prefix)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadFile(t *testing.T) {
	readfile := TemplateFuncMap()["readfile"].(func(string) string)

	tests := []struct {
		name string
		icon string
		want bool
	}{
		{"valid icon", "close", true},
		{"nonexistent icon", "nonexistent", false},
		{"empty icon name", "", false},
		{"directory traversal attempt", "../../../etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readfile(tt.icon)

			if tt.want {
				assert.NotEmpty(t, got, "readfile() should return non-empty string for valid icon")
			} else {
				assert.Empty(t, got, "readfile() should return empty string for invalid icon")
			}
		})
	}
}

func TestReplace(t *testing.T) {
	replace := TemplateFuncMap()["replace"].(func(s string, old string, new string) string)

	assert.Equal(t, "toast", replace("test", "e", "oa"))
}
