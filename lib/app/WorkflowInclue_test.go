package app_test

import (
	"strings"
	"testing"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowIncludeGetChecksumAlgorithm(t *testing.T) {
	hashUrls := map[string]checksums.ChecksumAlgorithm{
		"https://test/sha256sum":     checksums.ChecksumAlgorithmSha256,
		"https://test/sha256sum.txt": checksums.ChecksumAlgorithmSha256,
		"https://test/sha512sum":     checksums.ChecksumAlgorithmSha512,
		"https://test/sha512sum.txt": checksums.ChecksumAlgorithmSha512,
	}

	hashLengths := map[checksums.ChecksumAlgorithm]int{
		//checksums.ChecksumAlgorithmSha1:   40,
		checksums.ChecksumAlgorithmSha256: 64,
		// checksums.ChecksumAlgorithmSha3: 96,
		checksums.ChecksumAlgorithmSha512: 128,
	}

	for url, expected := range hashUrls {
		wi := app.WorkflowInclude{ChecksumUrl: url}
		wi.UpdateChecksumAlgorithm()
		assert.True(t, wi.HashAlgorithm.IsValid())
		assert.Equal(t, expected.String(), wi.HashAlgorithm.String(), "expected %s for url %s", expected, url)
		if expected.IsSupportedAlgorithm() {
			assert.Equal(t, expected.GetHashLength(), wi.HashAlgorithm.GetHashLength())
		}
	}

	for name, length := range hashLengths {
		wi := app.WorkflowInclude{FoundChecksum: strings.Repeat("a", length)}
		wi.UpdateHash()
		wi.UpdateChecksumAlgorithm()

		assert.True(t, wi.HashAlgorithm.IsValid())
		assert.Equal(t, name.String(), wi.HashAlgorithm.String())
	}
}
