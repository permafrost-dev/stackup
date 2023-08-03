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
	for _, line := range lines {
		hash, fn, err := matchHashAndFilename(line)
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

func FindChecksumForFileFromUrl(checksums []*Checksum, url string) *Checksum {
	for _, checksum := range checksums {
		if path.Base(checksum.Filename) == path.Base(url) {
			return checksum
		}
	}

	return nil
}

func matchHashAndFilename(input string) (string, string, error) {
	// Define the regular expression pattern
	pattern := `^([a-f0-9]{64})\s+(.+)$`

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
