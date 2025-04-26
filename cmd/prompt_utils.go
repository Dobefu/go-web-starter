package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

var promptForString = func(promptText string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(promptText)
	input, err := reader.ReadString('\n')

	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}

	return strings.TrimSpace(input), nil
}

var readPassword = term.ReadPassword

var promptForPassword = func(promptText string) (string, error) {
	fmt.Print(promptText)
	bytePassword, err := readPassword(int(os.Stdin.Fd()))

	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}

	fmt.Println()

	return string(bytePassword), nil
}
