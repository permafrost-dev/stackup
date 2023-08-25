package app

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/utils"
)

type WorkflowInclude struct {
	Url             string   `yaml:"url"`
	Headers         []string `yaml:"headers"`
	File            string   `yaml:"file"`
	ChecksumUrl     string   `yaml:"checksum-url"`
	VerifyChecksum  bool     `yaml:"verify,omitempty"`
	AccessKey       string   `yaml:"access-key"`
	SecretKey       string   `yaml:"secret-key"`
	Secure          bool     `yaml:"secure"`
	ChecksumIsValid *bool
	ValidationState ChecksumVerificationState
	Contents        string
	Hash            string
	FoundChecksum   string
	HashAlgorithm   checksums.ChecksumAlgorithm
	FromCache       bool
	Workflow        *StackupWorkflow
}

func (wi *WorkflowInclude) Initialize(workflow *StackupWorkflow) {
	wi.Workflow = workflow

	// expand environment variables in the include headers
	for i, v := range wi.Headers {
		if wi.Workflow.JsEngine.IsEvaluatableScriptString(v) {
			wi.Headers[i] = wi.Workflow.JsEngine.Evaluate(v).(string)
		}
		wi.Headers[i] = os.ExpandEnv(v)
	}

	wi.VerifyChecksum = wi.Workflow.Settings.ChecksumVerification
	wi.ValidationState = ChecksumVerificationStateNotVerified
	wi.ChecksumIsValid = nil
}

func (wi *WorkflowInclude) LoadedStatusText() string {
	result := "fetched"

	if wi.FromCache {
		result = "cached"
	}

	return fmt.Sprintf("%s, %s", result, wi.ValidationState.String())
}

func (wi *WorkflowInclude) SetLoadedFromCache(loaded bool, data *cache.CacheEntry) {
	wi.FromCache = loaded

	if !loaded {
		return
	}

	wi.Hash = data.Hash
	wi.HashAlgorithm = checksums.ParseChecksumAlgorithm(data.Algorithm)
	wi.Contents = data.Value
	wi.UpdateHash()
}

func (wi *WorkflowInclude) UpdateChecksumFromChecksumsFile(url, contents string) {
	cs := checksums.FindFilenameChecksum(path.Base(wi.FullUrl()), contents)
	if cs != nil {
		wi.FoundChecksum = cs.Hash
	}

	wi.UpdateChecksumAlgorithm()
}

func (wi *WorkflowInclude) GetChecksumUrls() []string {
	return []string{
		wi.FullUrlPath() + "/checksums.txt",
		wi.FullUrlPath() + "/checksums.sha256.txt",
		wi.FullUrlPath() + "/checksums.sha512.txt",
		wi.FullUrlPath() + "sha256sum",
		wi.FullUrlPath() + "sha512sum",
		wi.FullUrl() + ".sha256",
		wi.FullUrl() + ".sha512",
	}
}

func (wi *WorkflowInclude) ValidateChecksum() bool {
	wi.UpdateHash()
	wi.ValidationState = ChecksumVerificationStateNotVerified

	if !wi.IsRemoteUrl() || !wi.VerifyChecksum || !wi.Workflow.Settings.ChecksumVerification {
		return true
	}

	wi.ValidationState = ChecksumVerificationStatePending

	for _, url := range wi.GetChecksumUrls() {
		temp, err := wi.Workflow.Gateway.GetUrl(url)

		if err == nil && temp != "" {
			// check if the contents of the checksum file contains the filename of the included file,
			// OR if the url ends with .sha256 or .sha512 AND the url contains the filename of the included file
			hasChecksum := strings.Contains(temp, path.Base(wi.FullUrl()))
			if !hasChecksum {
				hasChecksum = strings.HasSuffix(url, ".sha256") || strings.HasSuffix(url, ".sha512") && strings.Contains(url, path.Base(wi.FullUrl()))
			}

			if !hasChecksum {
				continue
			}

			wi.ChecksumUrl = url
			wi.UpdateChecksumFromChecksumsFile(url, temp)
			break
		}
	}

	if !wi.HashAlgorithm.IsSupportedAlgorithm() {
		wi.ValidationState = ChecksumVerificationStateError
	}

	wi.SetChecksumIsValid(checksums.HashesMatch(wi.Hash, wi.FoundChecksum))

	return wi.ValidationState.IsVerified()
}

func (wi *WorkflowInclude) IsLocalFile() bool {
	return wi.Filename() != "" && !wi.IsRemoteUrl() && !wi.IsS3Url()
}

func (wi *WorkflowInclude) IsRemoteUrl() bool {
	return strings.HasPrefix(wi.FullUrl(), "http")
}

func (wi *WorkflowInclude) IsS3Url() bool {
	return strings.HasPrefix(wi.FullUrl(), "s3:")
}

func (wi *WorkflowInclude) Filename() string {
	return utils.AbsoluteFilePath(wi.File)
}

func (wi *WorkflowInclude) FullUrl() string {
	if strings.HasPrefix(strings.TrimSpace(wi.Url), "gh:") {
		return "https://raw.githubusercontent.com/" + strings.TrimPrefix(wi.Url, "gh:")
	}

	return wi.Url
}

func (wi *WorkflowInclude) Identifier() string {
	if wi.IsLocalFile() {
		return wi.Filename()
	}

	return wi.FullUrl()
}

func (wi *WorkflowInclude) FullUrlPath() string {
	return utils.ReplaceFilenameInUrl(wi.FullUrl(), "")
}

func (wi *WorkflowInclude) DisplayUrl() string {
	displayUrl := strings.Replace(wi.FullUrl(), "https://", "", -1)
	displayUrl = strings.Replace(displayUrl, "github.com/", "", -1)
	displayUrl = strings.Replace(displayUrl, "raw.githubusercontent.com/", "", -1)

	return displayUrl
}

func (wi WorkflowInclude) DisplayName() string {
	if wi.IsRemoteUrl() || wi.IsS3Url() {
		return wi.DisplayUrl()
	}

	if wi.IsLocalFile() {
		return wi.Filename()
	}

	return "<unknown>"
}

func (wi *WorkflowInclude) UpdateChecksumAlgorithm() {
	wi.HashAlgorithm = checksums.DetermineChecksumAlgorithm([]string{wi.FoundChecksum, wi.Hash}, wi.ChecksumUrl)
}

func (wi *WorkflowInclude) SetChecksumIsValid(value bool) bool {
	wi.ChecksumIsValid = &value
	wi.ValidationState.SetVerified(value)

	return value
}

func (wi *WorkflowInclude) UpdateHash() {
	originalHash := wi.Hash

	wi.Hash = checksums.CalculateSha256Hash(wi.Contents)
	wi.UpdateChecksumAlgorithm()

	if wi.Hash != originalHash {
		wi.SetChecksumIsValid(false)
		wi.ValidationState.Reset()
	}
}

func (wi *WorkflowInclude) NewCacheEntry() *cache.CacheEntry {
	wi.UpdateHash()

	return wi.Workflow.Cache.CreateEntry(
		wi.Contents,
		cache.CreateExpiresAtPtr(wi.Workflow.Settings.Cache.TtlMinutes),
		wi.Hash,
		wi.HashAlgorithm.String(),
		cache.CreateCarbonNowPtr(),
	)
}
