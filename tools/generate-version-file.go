package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	VERSION     = "0.0.0"
	TARGET_FILE = "lib/version/app-version.go"
)

func RunGitCommand(args ...string) string {
	cmd := exec.Command("git", args...)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return ""
	}

	return strings.TrimSpace(out.String())
}

func main() {
	result := RunGitCommand("rev-parse", "HEAD")[0:8]

	file, err := os.Create(TARGET_FILE)
	if err != nil {
		fmt.Printf("Cannot create %s file: %v\n", TARGET_FILE, err)
		os.Exit(1)
	}
	defer file.Close()

	file.WriteString(fmt.Sprintf("package version\n\nconst APP_VERSION = \"%s-%s\"\n", VERSION, result))
}
