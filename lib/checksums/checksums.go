package checksums

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

type Checksum struct {
	Hash      string
	Algorithm ChecksumAlgorithm
	Filename  string
}

func FindFilenameChecksum(filename string, contents string) *Checksum {
	result, ok := parseChecksumFileContentsIntoMap(contents)[filename]
	if !ok {
		return nil
	}
	return result
}

func parseChecksumFileContentsIntoMap(contents string) map[string]*Checksum {
	lines := stringToLines(contents)
	result := map[string]*Checksum{}

	for _, line := range lines {
		hash, fn, _ := matchHashAndFilename(line)
		if hash != "" && fn != "" {
			result[fn] = &Checksum{Hash: hash, Filename: fn, Algorithm: ChecksumAlgorithmSha256}
		}
	}

	return result
}

func stringToLines(contents string) []string {
	contents = strings.TrimSpace(contents)
	contents = strings.ReplaceAll(contents, "\\n", "  # \n")
	contents = strings.ReplaceAll(contents, "\t", " ")

	lines := strings.Split(contents, "\\n")
	lines = strings.Split(lines[0], "\n")

	return lines
}

func matchHashAndFilename(input string) (string, string, error) {
	pattern := `([a-fA-F0-9]{48,})[\s\t]+([\w\d_\-\.\/\\]+)`
	regex := regexp.MustCompile(pattern)

	matches := regex.FindStringSubmatch(input)
	if len(matches) != 3 {
		return "", "", fmt.Errorf("input string does not match pattern")
	}

	hash := matches[1]
	filename := matches[2]

	return hash, filename, nil
}

func CalculateSha256Hash(input string) string {
	hashBytes := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hashBytes[:])
}

func CalculateSha512Hash(input string) string {
	hashBytes := sha512.Sum512([]byte(input))
	return hex.EncodeToString(hashBytes[:])
}

func DetermineChecksumAlgorithm(hashes []string, checksumUrl string) ChecksumAlgorithm {
	result := ChecksumAlgorithmUnsupported

	// find the algorithm using length of the hash
	if algorithm := matchAlgorithmByLength(hashes); !result.IsSupportedAlgorithm() && algorithm.IsSupportedAlgorithm() {
		result = algorithm
	}

	// find the algorithm using the checksum url
	if algorithm := matchAlgorithmByUrl(checksumUrl); !result.IsSupportedAlgorithm() && algorithm.IsSupportedAlgorithm() {
		result = algorithm
	}

	return result
}

func matchAlgorithmByUrl(url string) ChecksumAlgorithm {
	// regex to match against checksum filenames
	patterns := map[ChecksumAlgorithm]*regexp.Regexp{
		ChecksumAlgorithmSha256: regexp.MustCompile("sha256(sum|\\.txt)"),
		ChecksumAlgorithmSha512: regexp.MustCompile("sha512(sum|\\.txt)"),
	}

	for name, pattern := range patterns {
		if pattern.MatchString(url) {
			return name
		}
	}

	return ChecksumAlgorithmUnsupported
}

func matchAlgorithmByLength(hashes []string) ChecksumAlgorithm {
	// hash length to name mapping
	hashTypeMap := map[int]ChecksumAlgorithm{
		16:  ParseChecksumAlgorithm("md4"),
		32:  ParseChecksumAlgorithm("md5"),
		40:  ParseChecksumAlgorithm("sha1"),
		64:  ParseChecksumAlgorithm("sha256"),
		96:  ParseChecksumAlgorithm("sha384"),
		128: ParseChecksumAlgorithm("sha512"),
	}

	for _, hash := range hashes {
		if hashType, ok := hashTypeMap[len(hash)]; ok {
			return hashType
		}
	}

	return ChecksumAlgorithmUnsupported
}

func HashesMatch(hash1 string, hash2 string) bool {
	return strings.EqualFold(hash1, hash2) && hash1 != ""
}
