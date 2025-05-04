package email

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTemplateHtml(t *testing.T) {
	tests := []struct {
		name        string
		body        EmailBody
		expectError bool
	}{
		{
			name:        "empty",
			body:        EmailBody{},
			expectError: true,
		},
		{
			name: "heading",
			body: EmailBody{
				Template: "components/atoms/skip-to-main",
				Data:     nil,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := getTemplateHtml(tt.body)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Greater(t, len(result), 0)
			}
		})
	}
}
