package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func BinaryExistsInPath(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

func KillProcessOnWindows(cmd *exec.Cmd) error {
	cmd.Process.Kill()
	return nil
}

func WaitForStartOfNextMinute() {
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
}

func AbsoluteFilePath(path string) string {
	path, _ = filepath.Abs(path)

	return path
}

func StringToInt(s string, defaultResult int) int {
	i, err := strconv.Atoi(s)

	if err != nil {
		return defaultResult
	}

	return i
}

func RunCommandInPath(input string, dir string, silent bool) *exec.Cmd {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	args := parts[1:]

	c := exec.Command(cmd, args...)
	if !silent {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}
	c.Dir = dir

	if err := c.Run(); err != nil {
		log.Println(err)
	}

	return c
}

func StartCommand(input string, cwd string) (*exec.Cmd, error) {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	args := parts[1:]

	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = cwd

	return c, nil
}

func WorkingDir(filenames ...string) string {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	parts := make([]string, 0)
	parts = append(parts, dir)
	parts = append(parts, filenames...)

	dir = path.Join(parts...)

	return dir
}

func FindFirstExistingFile(filenames []string) (string, error) {
	for _, filename := range filenames {
		if _, err := os.Stat(filename); err == nil {
			return filename, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("not found")
}
