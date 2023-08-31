package checksums

import (
	"strings"
)

type CalculateHashFunction = func(input string) (string, ChecksumAlgorithm)

type Checksum struct {
	Hash      string
	Algorithm ChecksumAlgorithm
	Filename  string
}

func FindFilenameChecksum(filename string, contents string) *Checksum {
	for _, line := range stringToLines(contents) {
		hash, fn, err := matchHashAndFilename(line)
		if err == nil && strings.EqualFold(fn, filename) {
			return &Checksum{Hash: hash, Filename: fn, Algorithm: ChecksumAlgorithmSha256}
		}
	}

	return nil
}

func DetermineChecksumAlgorithm(hashes []string, checksumUrl string) ChecksumAlgorithm {
	// find the algorithm using length of the hash
	if algorithm := matchAlgorithmByLength(hashes); algorithm.IsSupportedAlgorithm() {
		return algorithm
	}

	// find the algorithm using the checksum url
	if algorithm := matchAlgorithmByUrl(checksumUrl); algorithm.IsSupportedAlgorithm() {
		return algorithm
	}

	return ChecksumAlgorithmUnsupported
}

func HashesMatch(hash1 string, hash2 string) bool {
	return strings.EqualFold(hash1, hash2) && hash1 != ""
}
