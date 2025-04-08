package config

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

var BuildHash string
var executablePath = os.Executable

func init() {
	BuildHash = generateBuildHash()
}

func generateBuildHash() string {
	executable, err := executablePath()

	if err != nil {
		return "-"
	}

	file, err := os.Open(executable)
	defer func() { _ = file.Close() }()

	if err != nil {
		return "-"
	}

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		return "-"
	}

	return fmt.Sprintf("%x\n", hash.Sum(nil))
}
