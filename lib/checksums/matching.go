package checksums

import (
	"fmt"
	"regexp"
)

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
