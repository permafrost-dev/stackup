package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
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

func RunCommandInPath(input string, dir string, silent bool) (*exec.Cmd, error) {
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
		return c, err
	}

	return c, nil
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

func StringArrayContains(arr []string, s string) bool {
	for _, item := range arr {
		if item == s {
			return true
		}
	}

	return false
}

func SaveUrlToFile(url string, filename string) error {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func GenerateTaskUuid() string {
	return GenerateShortID(8)
}

func GenerateShortID(length ...int) string {
	var charCount = 8
	if len(length) > 0 {
		charCount = length[0]
	}

	// Define the character set to use for the short ID
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Define the length of the short ID
	// charCount := 8

	// Generate a random number for each character in the short ID
	var result string
	for i := 0; i < charCount; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return ""
		}

		result += string(charset[n.Int64()])
	}

	return result
}

func CalculateSHA256Hash(input string) string {
	inputBytes := []byte(input)
	hash := sha256.Sum256(inputBytes)
	hashString := hex.EncodeToString(hash[:])

	return strings.ToLower(hashString)
}
