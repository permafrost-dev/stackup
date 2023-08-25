package app

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/utils"
)

type WorkflowInclude struct {
	Url               string   `yaml:"url"`
	Headers           []string `yaml:"headers"`
	File              string   `yaml:"file"`
	ChecksumUrl       string   `yaml:"checksum-url"`
	VerifyChecksum    bool     `yaml:"verify,omitempty"`
	AccessKey         string   `yaml:"access-key"`
	SecretKey         string   `yaml:"secret-key"`
	Secure            bool     `yaml:"secure"`
	ChecksumIsValid   *bool
	ValidationState   ChecksumVerificationState
	Contents          string
	Hash              string
	FoundChecksum     string
	HashAlgorithm     string
	ChecksumValidated bool
	FromCache         bool
	Workflow          *StackupWorkflow
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

func (wi *WorkflowInclude) getChecksumFromContents(contents string) string {
	if checksum := checksums.FindChecksumForFilename(path.Base(wi.FullUrl()), contents); checksum != nil {
		return checksum.Hash
	}

	fmt.Printf("error parsing checksums\n")

	return "---"
}

func (wi *WorkflowInclude) ValidateChecksum() (bool, string, error) {
	wi.ChecksumValidated = false
	wi.ValidationState = ChecksumVerificationStateNotVerified

	if !wi.Workflow.Settings.ChecksumVerification {
		return true, wi.Hash, nil
	}

	wi.ValidationState = ChecksumVerificationStatePending

	checksumUrls := []string{
		wi.FullUrlPath() + "/checksums.txt",
		wi.FullUrlPath() + "/checksums.sha256.txt",
		wi.FullUrlPath() + "/checksums.sha512.txt",
		wi.FullUrlPath() + "sha256sum",
		wi.FullUrlPath() + "sha512sum",
		wi.FullUrl() + ".sha256",
		wi.FullUrl() + ".sha512",
	}

	var checksumContents string = ""
	for _, url := range checksumUrls {
		var err error
		temp, err := wi.Workflow.Gateway.GetUrl(url)

		if err == nil && temp != "" {
			hasChecksum := strings.Contains(temp, path.Base(wi.FullUrl()))

			if !hasChecksum {
				hasChecksum = strings.HasSuffix(url, ".sha256") || strings.HasSuffix(url, ".sha512") && strings.Contains(url, path.Base(wi.FullUrl()))
			}

			if !hasChecksum {
				continue
			}

			checksumContents = temp
			wi.ChecksumUrl = url

			break
		}
	}

	if checksumContents != "" {
		wi.FoundChecksum = wi.getChecksumFromContents(checksumContents)
		wi.HashAlgorithm = wi.GetChecksumAlgorithm()
	}

	if wi.HashAlgorithm == "unknown" || wi.HashAlgorithm == "" {
		wi.ValidationState = ChecksumVerificationStateError
		return false, "", fmt.Errorf("unable to find valid checksum file for %s", wi.DisplayUrl())
	}

	wi.UpdateHash()
	// fmt.Printf("Found checksum: %s\n", wi.FoundChecksum)
	// fmt.Printf("Calculated checksum: %s\n", wi.Hash)
	// fmt.Printf("Checksum algorithm: %s\n", wi.HashAlgorithm)

	wi.SetChecksumIsValid(strings.EqualFold(wi.Hash, wi.FoundChecksum))
	wi.ChecksumValidated = *wi.ChecksumIsValid
	// fmt.Printf("Checksum is valid: %v\n", *wi.ChecksumIsValid)

	if *wi.ChecksumIsValid {
		wi.ValidationState = ChecksumVerificationStateVerified
	} else {
		wi.ValidationState = ChecksumVerificationStateMismatch
	}

	return *wi.ChecksumIsValid, wi.Hash, nil
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
	temp, _ := utils.ReplaceFilenameInUrl(wi.FullUrl(), "")

	return strings.TrimSuffix(temp, "/")
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

func (wi *WorkflowInclude) GetChecksumAlgorithm() string {
	var patterns map[string]*regexp.Regexp = map[string]*regexp.Regexp{
		"sha256": regexp.MustCompile("sha256(sum|\\.txt)"),
		"sha512": regexp.MustCompile("sha512(sum|\\.txt)"),
	}

	for name, pattern := range patterns {
		// fmt.Printf("Checking %s against %s\n", wi.ChecksumUrl, pattern)
		if pattern.MatchString(wi.ChecksumUrl) {
			return name
		}
	}

	// hash length to name mapping
	hashTypeMap := map[int]string{
		16:  "md4",
		32:  "md5",
		40:  "sha1",
		64:  "sha256",
		96:  "sha384",
		128: "sha512",
	}

	if hashType, ok := hashTypeMap[len(wi.FoundChecksum)]; ok {
		return hashType
	}

	return "unknown"
}

func (wi *WorkflowInclude) SetChecksumIsValid(value bool) {
	wi.ChecksumIsValid = &value
}

func (wi *WorkflowInclude) UpdateHash() {
	wi.Hash = checksums.CalculateSha256Hash(wi.Contents)
	wi.HashAlgorithm = "sha256"
}

func (wi *WorkflowInclude) NewCacheEntry() *cache.CacheEntry {
	expires := carbon.Now().AddMinutes(wi.Workflow.Settings.Cache.TtlMinutes)

	result := wi.Workflow.Cache.CreateEntry(
		wi.Contents,
		&expires,
		wi.Hash,
		wi.HashAlgorithm,
		nil,
	)

	return result
}
