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
	return cmd.Process.Kill()
}

func WaitForStartOfNextMinute() {
	WaitForStartOfNextInterval(time.Minute)
}

func WaitForStartOfNextInterval(interval time.Duration) {
	time.Sleep(time.Until(time.Now().Truncate(interval).Add(interval)))
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
	} else {
		c.Stdout = nil
		c.Stderr = nil
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

func GetUrlJson(url string, result any) error {
	body, err := GetUrlContents(url)
	if err != nil {
		return err
	}

	// var data interface{}
	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return err
	}

	return nil
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

func FileSize(filename string) int64 {
	var result int64 = 0

	if info, err := os.Stat(filename); err == nil {
		result = info.Size()
	}

	return result
}

func FileExists(filename string) bool {
	if st, err := os.Stat(filename); err == nil {
		return !st.IsDir()
	}
	return false
}

func PathExists(pathname string) bool {
	if st, err := os.Stat(pathname); err == nil {
		return st.IsDir()
	}
	return false
}

func RemoveFile(filename string) error {
	if !FileExists(filename) && !PathExists(filename) {
		return nil
	}

	return os.Remove(filename)
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

func SaveStringToFile(contents string, filename string) error {
	return os.WriteFile(filename, []byte(contents), 0644)
}

func GenerateTaskUuid() string {
	return GenerateShortID(8)
}

func GenerateShortID(length ...int) string {
	var charCount = 8
	var result string = ""

	if len(length) > 0 {
		charCount = length[0]
	}

	// Define the character set to use for the short ID
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Generate a random number for each character in the short ID
	for i := 0; i < charCount; i++ {
		if n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset)))); err == nil {
			result += string(charset[n.Int64()])
		}
	}

	return result
}

func UrlBasePath(u string) string {
	var parsedUrl *url.URL
	var err error

	if parsedUrl, err = url.Parse(u); err != nil {
		// failed to parse the url, so try to parse it manually
		parts := strings.Split(u, "/")
		length := len(parts)
		if !strings.HasSuffix(u, "/") {
			length--
		}
		return strings.Join(parts[1:length], "/")
	}

	baseFn := path.Base(parsedUrl.Path)
	result := strings.Replace(parsedUrl.String(), "/"+baseFn, "", 1)

	if parsedUrl, err = url.Parse(result); err == nil {
		result = parsedUrl.String()
	}

	result = strings.Replace(result, "?"+parsedUrl.Query().Encode(), "", 1)

	return strings.TrimSuffix(result, "/")
}

func ReplaceFilenameInUrl(u string, newFilename string) string {
	var result string = u + newFilename
	if !strings.HasSuffix(u, "/") {
		result = UrlBasePath(u) + "/" + newFilename
	}

	parsedUrl, err := url.Parse(result)
	if err == nil {
		result = parsedUrl.String()
	}

	return strings.TrimSuffix(result, "/")
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

func GetDefaultConfigurationBasePath(path string, defaultPath string) string {
	var err error

	if path == "~" || path == "" {
		if path, _ = os.UserHomeDir(); err != nil {
			return defaultPath
		}
	}

	return path
}

func MakeConfigurationPath(path string, appName string) string {
	appName = "." + strings.TrimPrefix(appName, ".")
	return filepath.Join(GetDefaultConfigurationBasePath(path, "."), appName)
}

func EnsureConfigDirExists(homeDir string, appName string) (string, error) {
	result := MakeConfigurationPath(GetDefaultConfigurationBasePath(homeDir, "."), appName)

	if err := os.MkdirAll(result, 0744); err != nil {
		return "", err
	}

	return result, nil
}

func DomainGlobMatch(pattern string, s string) bool {
	return GlobMatch(pattern, s, true)
}

func GlobMatch(pattern string, s string, optional bool) bool {
	var match glob.Glob
	var err error

	if pattern == "*" {
		return len(s) > 0
	}

	if match, err = glob.Compile(pattern); err != nil {
		return false
	}

	if !optional {
		match = glob.MustCompile(pattern, '.')
	}

	return match.Match(s)
}

func FsSafeName(name string) string {
	result := strings.TrimSpace(name)
	result = regexp.MustCompile(`[^\w\\-\\._]+`).ReplaceAllString(result, "-")
	result = regexp.MustCompile(`-{2,}`).ReplaceAllString(result, "-")

	return strings.Trim(result, "-")
}

func EnforceSuffix(s string, suffix string) string {
	if suffix == "" {
		return s
	}

	return strings.TrimSuffix(s, suffix) + suffix
}

func ReverseArray[T any](items []T) []T {
	length := len(items)
	for i := 0; i < length/2; i++ {
		items[i], items[length-i-1] = items[length-i-1], items[i]
	}
	return items
}

func CombineArrays[T any](arrays ...[]T) []T {
	result := []T{}

	for _, arr := range arrays {
		result = append(result, arr...)
	}

	return result
}

// casts the items in `toAppend` to the same type as the items in `items`, then
// returns a new array containing the two items from both arrays combined
func CastAndCombineArrays[T any, R any](items []T, toAppend []R) []T {
	casted := []T{}

	for _, item := range toAppend {
		var temp interface{} = item
		var castedItem T = (temp).(T)
		casted = append(casted, castedItem)
	}

	return CombineArrays(items, casted)
}

func Max(args ...int) int {
	if len(args) == 0 {
		return 0
	}

	result := args[0]
	for _, value := range args {
		if value > result {
			result = value
		}
	}

	return result
}

func Min(args ...int) int {
	if len(args) == 0 {
		return 0
	}

	result := args[0]
	for _, value := range args {
		if value < result {
			result = value
		}
	}

	return result
}
