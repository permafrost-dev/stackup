package workflow_test

import (
	"strings"
	"testing"

	"github.com/stackup-app/stackup/lib/workflow"
	"github.com/stretchr/testify/assert"
)

func TestWorkflowIncdeGetChecksumAlforithm(t *testing.T) {
	hashUrls := map[string]string{
		"https://test/sha256sum":     "sha256",
		"https://test/sha256sum.txt": "sha256",
		"https://test/sha512sum":     "sha512",
		"https://test/sha512sum.txt": "sha512",
	}

	hashLengths := map[string]int{
		"md4":    16,
		"md5":    32,
		"sha1":   40,
		"sha256": 64,
		"sha384": 96,
		"sha512": 128,
	}

	for url, expected := range hashUrls {
		wi := workflow.WorkflowInclude{ChecksumUrl: url}
		assert.Equal(t, expected, wi.GetChecksumAlgorithm(), "expected %s for url %s", expected, url)
	}

	for name, length := range hashLengths {
		wi := workflow.WorkflowInclude{FoundChecksum: strings.Repeat("a", length)}
		assert.Equal(t, name, wi.GetChecksumAlgorithm())
	}
}
