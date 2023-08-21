package utils

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gobwas/glob"
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

func RunCommandInPath(input string, dir string, silent bool) (*exec.Cmd, error) {
	c := StartCommand(input, dir, silent)
	if err := c.Run(); err != nil {
		return c, err
	}

	return c, nil
}

func StartCommand(input string, cwd string, silent bool) *exec.Cmd {
	// Split the input into command and arguments
	parts := strings.Split(input, " ")
	cmd := parts[0]
	args := parts[1:]

	c := exec.Command(cmd, args...)
	c.Dir = cwd
	if !silent {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}

	return c
}

func WorkingDir(filenames ...string) string {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}

	parts := []string{dir}
	parts = append(parts, filenames...)

	return path.Join(parts...)
}

func FindFirstExistingFile(filenames []string) (string, error) {
	for _, filename := range filenames {
		if IsFile(filename) {
			return filename, nil
		}
	}
	return "", os.ErrNotExist
}

func GetUrlContents(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	// Read the response body into a byte slice
	body, err := io.ReadAll(resp.Body)
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
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

func IsDir(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}

	return info.IsDir()
}

func FileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

func MatchesPattern(s string, pattern string) bool {
	regex := regexp.MustCompile(pattern)
	return regex.Match([]byte(s))
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

	// Generate a random number for each character in the short ID
	var result string
	for i := 0; i < charCount; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			continue
		}

		result += string(charset[n.Int64()])
	}

	return result
}

func ReplaceFilenameInUrl(u string, newFilename string) (string, error) {
	parsedUrl, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	// Replace the filename in the URL path
	parsedUrl.Path = path.Join(path.Dir(parsedUrl.Path), newFilename)

	return parsedUrl.String(), nil
}

func GetUniqueStrings(items []string) []string {
	uniqueItems := make([]string, 0)

	for _, item := range items {
		if !StringArrayContains(uniqueItems, item) {
			uniqueItems = append(uniqueItems, item)
		}
	}

	return uniqueItems
}

func EnsureConfigDirExists(appName string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	// Append the directory name to the home directory
	configDir := filepath.Join(homeDir, "."+appName)

	// Ensure the directory exists
	err = os.MkdirAll(configDir, 0744)
	if err != nil {
		return "", err
	}

	return configDir, nil
}

func DomainGlobMatch(pattern string, s string) bool {
	return GlobMatch(pattern, s, true)
}

func GlobMatch(pattern string, s string, optional bool) bool {
	if pattern == "*" {
		return len(s) > 0
	}

	if !optional {
		return glob.
			MustCompile(pattern, '.').
			Match(s)
	}

	match, err := glob.Compile(pattern)
	if err != nil {
		return false
	}

	return match.Match(s)
}

func FsSafeName(name string) string {
	result := strings.TrimSpace(name)
	result = regexp.MustCompile(`[^\w\\-\\._]+`).ReplaceAllString(result, "-")
	result = regexp.MustCompile(`-{2,}`).ReplaceAllString(result, "-")

	return strings.Trim(result, "-")
}

func ReverseArray[T any](items []T) []T {
	length := len(items)
	for i := 0; i < length/2; i++ {
		items[i], items[length-i-1] = items[length-i-1], items[i]
	}
	return items
}

// casts the items in `toAppend` to the same type as the items in `items`, then
// returns a new array containing the two items from both arrays combined
func CastAndCombineArrays[T interface{}, R any](items []*T, toAppend []R) []*T {
	result := []*T{}
	casted := []*T{}

	for _, item := range toAppend {
		var temp interface{} = item
		var castedItem T = (temp).(T)
		casted = append(casted, &castedItem)
	}

	result = append(result, items...)
	result = append(result, casted...)

	return result
}

func CastArrayItems[T interface{}, R any](items []*T, toAppend []R) []*T {
	result := []*T{}
	casted := []*T{}

	for _, item := range toAppend {
		var temp interface{} = item
		var castedItem T = (temp).(T)
		casted = append(casted, &castedItem)
	}

	result = append(result, items...)
	result = append(result, casted...)

	return result
}
