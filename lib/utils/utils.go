package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
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

func ChangeWorkingDirectory(path string) error {
	err := os.Chdir(path)
	if err != nil {
		return err
	}

	return nil
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

func GetUrlContents(url string) (string, error) {
	// Send an HTTP GET request to the URL
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body into a byte slice
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Convert the byte slice to a string and return it
	return string(body), nil
}

func GetUrlJson(url string) (interface{}, error) {
	body, err := GetUrlContents(url)
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal([]byte(body), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func IsFile(filename string) bool {
	return !IsDir(filename)
}

func IsDir(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func MatchPattern(s string, pattern string) []string {
	regex := regexp.MustCompile(pattern)
	if !regex.MatchString(s) {
		return []string{}
	}

	matches := regex.FindAllString(s, -1)
	return matches
}

func MatchesPattern(s string, pattern string) bool {
	regex := regexp.MustCompile(pattern)
	return regex.MatchString(s)
}
