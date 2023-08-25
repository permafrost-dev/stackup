package checksums

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/stackup-app/stackup/lib/utils"
)

type Checksum struct {
	Hash     string
	Filename string
}

func ParseChecksumFileContents(contents string) ([]*Checksum, error) {
	lines := strings.Split(strings.TrimSpace(contents), "\n")

	// Parse each line into a Checksum struct
	var checksums []*Checksum
	lines = strings.Split(lines[0], "\n")

	for _, line := range lines {
		line = strings.ReplaceAll(line, "\t", " ")
		re, _ := regexp.Compile(`\s{2,}`)
		line = re.ReplaceAllString(line, " ")
		hash, fn, err := MatchHashAndFilename(line)
		if err != nil {
			continue
		}

		// Create a new Checksum struct and append it to the array
		checksums = append(checksums, &Checksum{
			Hash:     hash,
			Filename: fn,
		})
	}

	return checksums, nil
}

func FindChecksumForFilename(filename string, contents string) *Checksum {
	// re := regexp.MustCompile(`([a-fA-F0-9]{48,})[\t\s]+(` + filename + `)`)
	// matches := re.FindAllStringSubmatch(contents, -1)
	mapped := ParseChecksumsIntoMap(contents)
	result, ok := mapped[filename]

	if !ok {
		return nil
	}

	return result
}

func ParseChecksumsIntoMap(contents string) map[string]*Checksum {
	contents = strings.ReplaceAll(contents, "\\n", "  # \n")

	contents = strings.ReplaceAll(contents, "\t", " ")
	lines := strings.Split(contents, "\\n")
	lines = strings.Split(lines[0], "\n")

	result := map[string]*Checksum{}

	for _, line := range lines {
		hash, fn, _ := MatchHashAndFilename(line)
		if hash == "" || fn == "" {
			continue
		}
		result[fn] = &Checksum{Hash: hash, Filename: fn}
	}

	return result
}

func FindChecksumForFileFromUrl(checksums []*Checksum, url string) *Checksum {
	for _, checksum := range checksums {
		if path.Base(checksum.Filename) == path.Base(url) {
			return checksum
		}
	}

	return nil
}

func MatchHashAndFilename(input string) (string, string, error) {
	// Define the regular expression pattern
	pattern := `([a-fA-F0-9]{48,})[\s\t]+([\w\d_\-\.\/\\]+)`

	// Compile the regular expression
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return "", "", err
	}

	// Match the input string against the regular expression
	matches := regex.FindStringSubmatch(input)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("input string does not match pattern")
	}

	// Extract the hash and filename from the matches
	hash := matches[1]
	filename := matches[2]

	return hash, filename, nil
}

func CalculateSha256Hash(input string) string {
	inputBytes := []byte(input)
	hashBytes := sha256.Sum256(inputBytes)

	// Convert the hash bytes to a hex string
	hashString := hex.EncodeToString(hashBytes[:])

	return hashString
}

func CalculateSha512Hash(input string) string {
	inputBytes := []byte(input)
	hashBytes := sha512.Sum512(inputBytes)
	hashString := hex.EncodeToString(hashBytes[:])

	return hashString
}

func (c *Checksum) FilenameAsUrl(baseUrl string) string {
	result, _ := utils.ReplaceFilenameInUrl(baseUrl, c.Filename)

	return result
}
