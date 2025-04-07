package config

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

var BuildHash string

func init() {
	BuildHash = generateBuildHash()
}

func generateBuildHash() string {
	file, err := os.Open(os.Args[0])
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
