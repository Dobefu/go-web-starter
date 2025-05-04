package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTextFromHtml(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output string
	}{
		{
			name:   "empty string",
			input:  "",
			output: "",
		},
		{
			name:   "empty paragraph",
			input:  "<p> </p>",
			output: "",
		},
		{
			name:   "script tag",
			input:  "<script>alert('test')</script>",
			output: "",
		},
		{
			name:   "single paragraph",
			input:  "<p>test</p>",
			output: "test\n\n",
		},
		{
			name:   "anchor tag",
			input:  `<a class>link</a>`,
			output: "link\n\n",
		},
		{
			name:   "anchor tag with link",
			input:  `<a href="/href">link</a>`,
			output: "link (/href)\n\n",
		},
		{
			name:   "footer attribute",
			input:  "<div footer>test</div>",
			output: "\ntest\n\n",
		},
		{
			name:   "title",
			input:  `<p class="text-xl">test</p>`,
			output: "test\n----\n\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := getTextFromHtml(tt.input)
			assert.Equal(t, tt.output, result)
		})
	}
}
