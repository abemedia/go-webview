package embedded

import (
	"fmt"
	"os"
	"path/filepath"
)

const version = "0.8.3"

func init() {
	dir := os.TempDir() // filepath.Join(os.TempDir(), "webview-"+version)
	file := filepath.Join(dir, name)

	if _, err := os.Stat(file); err != nil {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			panic(err)
		}
		if err := os.WriteFile(file, lib, os.ModePerm); err != nil { //nolint:gosec
			panic(err)
		}
	}

	fmt.Println("embedded", file)

	if err := os.Setenv("PATH", dir+";"+os.Getenv("PATH")); err != nil {
		panic(err)
	}
}
