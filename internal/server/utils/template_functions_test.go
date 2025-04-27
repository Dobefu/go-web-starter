package utils

import (
	"html/template"
	"testing"

	"fmt"
	"io/fs"

	"github.com/Dobefu/go-web-starter/internal/static"
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

func TestReadFile_GetStaticFSError(t *testing.T) {
	orig := static.GetStaticFS

	static.GetStaticFS = func() (fs.FS, error) {
		return nil, fmt.Errorf("forced error")
	}

	defer func() { static.GetStaticFS = orig }()

	readfile := TemplateFuncMap()["readfile"].(func(string) string)
	assert.Equal(t, "", readfile("anyicon"))
}

func TestReplace(t *testing.T) {
	replace := TemplateFuncMap()["replace"].(func(s string, old string, new string) string)

	assert.Equal(t, "toast", replace("test", "e", "oa"))
}

func TestSlice(t *testing.T) {
	slice := TemplateFuncMap()["slice"].(func(...any) []any)

	tests := []struct {
		name string
		args []any
		want []any
	}{
		{"empty slice", []any{}, []any{}},
		{"single item", []any{42}, []any{42}},
		{"multiple items", []any{"a", 2, true}, []any{"a", 2, true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, slice(tt.args...))
		})
	}
}

func TestDump(t *testing.T) {
	dump := TemplateFuncMap()["dump"].(func(any) template.HTML)

	tests := []struct {
		name     string
		input    any
		wantSubs []string
		wantPre  bool
	}{
		{
			name:     "pretty print JSON for map",
			input:    map[string]any{"foo": "bar", "num": 42},
			wantSubs: []string{`"foo": "bar"`, `"num": 42`},
			wantPre:  true,
		},
		{
			name:     "pretty print JSON for slice",
			input:    []int{1, 2, 3},
			wantSubs: []string{"1", "2", "3"},
			wantPre:  true,
		},
		{
			name:     "fallback for non-JSON-marshalable type",
			input:    make(chan int),
			wantSubs: []string{"chan int"},
			wantPre:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := string(dump(tt.input))

			for _, sub := range tt.wantSubs {
				assert.Contains(t, output, sub)
			}

			if tt.wantPre {
				assert.True(t, output[0:5] == "<pre>" && output[len(output)-6:] == "</pre>")
			}
		})
	}
}

func TestTrimTrailingNewline(t *testing.T) {
	trimTrailingNewline := TemplateFuncMap()["trimTrailingNewline"].(func(string) string)

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no newline", "foo", "foo"},
		{"single newline", "foo\n", "foo"},
		{"single carriage return", "foo\r", "foo"},
		{"windows newline", "foo\r\n", "foo"},
		{"multiple newlines", "foo\n\n", "foo"},
		{"multiple carriage returns", "foo\r\r", "foo"},
		{"mixed newlines", "foo\r\n\n", "foo"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := trimTrailingNewline(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}
