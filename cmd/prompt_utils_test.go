package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func withInputOutput(
	input string,
	testFunc func() (string, error),
) (output string, result string, err error) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	r, w, _ := os.Pipe()
	os.Stdin = r

	if input != "" {
		go func() {
			_, _ = io.WriteString(w, input)
			_ = w.Close()
		}()
	} else {
		_ = w.Close()
	}

	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	res, err := testFunc()

	_ = outW.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, outR)

	return buf.String(), res, err
}

func TestPromptForString(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		promptText  string
		expected    string
		expectError bool
	}{
		{"successful input", "test input\n", "Enter text: ", "test input", false},
		{"empty input", "\n", "Enter text: ", "", false},
		{"input with whitespace", "  test with spaces  \n", "Enter text: ", "test with spaces", false},
		{"read error", "", "Enter text: ", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, result, err := withInputOutput(tt.input, func() (string, error) {
				return promptForString(tt.promptText)
			})

			assert.Equal(t, tt.promptText, output)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestPromptForPassword_Table(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		promptText  string
		expectError bool
		mockError   error
	}{
		{"successful password input", "secretpassword", "Enter password: ", false, nil},
		{"empty password", "", "Enter password: ", false, nil},
		{"read error", "", "Enter password: ", true, errors.New("mock error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalReadPassword := readPassword
			defer func() { readPassword = originalReadPassword }()

			readPassword = func(fd int) ([]byte, error) {
				if tt.mockError != nil {
					return nil, tt.mockError
				}

				return []byte(tt.password), nil
			}

			output, result, err := withInputOutput("", func() (string, error) {
				return promptForPassword(tt.promptText)
			})

			expectedOutput := tt.promptText

			if !tt.expectError {
				expectedOutput += "\n"
			}

			assert.Equal(t, expectedOutput, output)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.password, result)
			}
		})
	}
}

func TestPromptForString_Error(t *testing.T) {
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	r, w, _ := os.Pipe()
	os.Stdin = r
	err := w.Close()
	assert.NoError(t, err)

	_, err = promptForString("Enter text: ")

	_ = outW.Close()
	_, _ = io.Copy(io.Discard, outR)

	assert.Error(t, err)
}

func TestPromptForPassword_Error(t *testing.T) {
	originalPrompt := promptForPassword
	defer func() { promptForPassword = originalPrompt }()

	promptForPassword = func(prompt string) (string, error) {
		fmt.Print(prompt)
		return "", errors.New("mock password error")
	}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	outR, outW, err := os.Pipe()
	assert.NoError(t, err)
	os.Stdout = outW

	result, err := promptForPassword("Enter password: ")

	_ = outW.Close()
	_, _ = io.Copy(io.Discard, outR)

	assert.Equal(t, "", result)
	assert.Error(t, err)
}

func TestPromptForPassword_RealImplementation(t *testing.T) {
	originalReadPassword := readPassword
	defer func() { readPassword = originalReadPassword }()

	readPassword = func(fd int) ([]byte, error) {
		return []byte("mockedpassword"), nil
	}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	result, err := promptForPassword("Enter password: ")
	assert.NoError(t, err)

	_ = outW.Close()
	_, _ = io.Copy(io.Discard, outR)

	assert.Equal(t, "mockedpassword", result)
}

func TestPromptForPassword_RealImplementation_Error(t *testing.T) {
	originalReadPassword := readPassword
	defer func() { readPassword = originalReadPassword }()

	readPassword = func(fd int) ([]byte, error) {
		return nil, errors.New("mock error")
	}

	oldStdout := os.Stdout
	defer func() { os.Stdout = oldStdout }()

	outR, outW, _ := os.Pipe()
	os.Stdout = outW

	result, err := promptForPassword("Enter password: ")

	_ = outW.Close()
	_, _ = io.Copy(io.Discard, outR)

	assert.Equal(t, "", result)
	assert.Error(t, err)
}
