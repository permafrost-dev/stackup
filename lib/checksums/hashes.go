package checksums

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
)

func CalculateDefaultHash(input string) (string, ChecksumAlgorithm) {
	return CalculateSha256Hash(input)
}

func CalculateSha256Hash(input string) (string, ChecksumAlgorithm) {
	hashBytes := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hashBytes[:]), ChecksumAlgorithmSha256
}

func CalculateSha512Hash(input string) (string, ChecksumAlgorithm) {
	hashBytes := sha512.Sum512([]byte(input))
	return hex.EncodeToString(hashBytes[:]), ChecksumAlgorithmSha512
}
