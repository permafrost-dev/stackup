package app

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
)

type WorkflowInclude struct {
	Url               string   `yaml:"url"`
	Headers           []string `yaml:"headers"`
	File              string   `yaml:"file"`
	ChecksumUrl       string   `yaml:"checksum-url"`
	VerifyChecksum    *bool    `yaml:"verify,omitempty"`
	AccessKey         string   `yaml:"access-key"`
	SecretKey         string   `yaml:"secret-key"`
	Secure            bool     `yaml:"secure"`
	ChecksumIsValid   *bool
	ValidationState   string
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
		if App.JsEngine.IsEvaluatableScriptString(v) {
			wi.Headers[i] = App.JsEngine.Evaluate(v).(string)
		}
		wi.Headers[i] = os.ExpandEnv(v)
	}

	// set some default values
	if wi.VerifyChecksum == nil {
		boolValue := true //wi.ChecksumUrl != ""
		wi.VerifyChecksum = &boolValue
	}
	wi.ValidationState = "not validated"
	wi.ChecksumIsValid = nil
}

func (include *WorkflowInclude) Process() {
	var err error = nil
	data := include.Workflow.tryLoadingCachedData(include)

	if !include.Workflow.hasRemoteDomainAccess(include) {
		return
	}

	if err = include.Workflow.handleDataNotCached(include.FromCache, data, include); err != nil {
		fmt.Println(err)
		return
	}

	if !include.Workflow.handleChecksumVerification(include) {
		return
	}

	if err = include.Workflow.loadRemoteFileInclude(include); err != nil {
		fmt.Println(err)
		return
	}

	support.SuccessMessageWithCheck("remote include (" + include.LoadedStatusText() + "): " + include.DisplayName())
}

func (wi *WorkflowInclude) LoadedStatusText() string {
	cachedText := "fetched"
	if wi.FromCache {
		cachedText = "cached"
	}
	if wi.ValidationState != "" {
		cachedText += wi.ValidationState + ", " + cachedText
	}

	return cachedText
}

func (wi *WorkflowInclude) getChecksumFromContents(contents string) string {
	checksumsArr, _ := checksums.ParseChecksumFileContents(contents)
	checksum := checksums.FindChecksumForFileFromUrl(checksumsArr, wi.FullUrl())

	if checksum != nil {
		// fmt.Printf("found checksum: %s\n", checksum.Hash)
		return checksum.Hash
	}

	//try to match a hash using regex
	regex := regexp.MustCompile("^([a-fA-F0-9]{48,})$")
	matches := regex.FindAllString(contents, -1)

	if len(matches) > 0 {
		return matches[0]
	}

	return strings.TrimSpace(contents)
}

func (wi *WorkflowInclude) ValidateChecksum(contents string) (bool, string, error) {
	if wi.VerifyChecksum != nil && *wi.VerifyChecksum == false {
		return true, "", nil
	}

	checksumUrls := []string{
		wi.FullUrlPath() + "/checksums.txt",
		wi.FullUrlPath() + "/checksums.sha256.txt",
		wi.FullUrlPath() + "/checksums.sha512.txt",
		wi.FullUrlPath() + "sha256sum",
		wi.FullUrlPath() + "sha512sum",
		wi.FullUrl() + ".sha256",
		wi.FullUrl() + ".sha512",
	}

	for _, url := range checksumUrls {
		if !App.Gateway.Allowed(url) {
			support.FailureMessageWithXMark("Access to " + url + " is not allowed.")
			continue
		}

		checksumContents, err := utils.GetUrlContents(url)
		if err != nil {
			continue
		}

		wi.ChecksumUrl = url
		wi.FoundChecksum = wi.getChecksumFromContents(checksumContents)
		wi.HashAlgorithm = wi.GetChecksumAlgorithm()

		break
	}

	if wi.HashAlgorithm == "unknown" || wi.HashAlgorithm == "" {
		return false, "", fmt.Errorf("unable to find valid checksum file for %s", wi.DisplayUrl())
	}

	var hash string

	switch wi.HashAlgorithm {
	case "sha256":
		hash = checksums.CalculateSha256Hash(contents)
		break
	case "sha512":
		hash = checksums.CalculateSha512Hash(contents)
		break
	default:
		wi.SetChecksumIsValid(false)
		return false, "", fmt.Errorf("unsupported algorithm: %s", wi.HashAlgorithm)
	}

	if !strings.EqualFold(hash, wi.FoundChecksum) {
		wi.SetChecksumIsValid(false)
		return false, "", nil
	}

	wi.SetChecksumIsValid(true)
	return true, hash, nil
}

func (wi *WorkflowInclude) IsLocalFile() bool {
	return wi.File != "" && utils.IsFile(wi.File)
}

func (wi *WorkflowInclude) IsRemoteUrl() bool {
	return wi.FullUrl() != "" && strings.HasPrefix(wi.FullUrl(), "http")
}

func (wi *WorkflowInclude) IsS3Url() bool {
	return wi.FullUrl() != "" && strings.HasPrefix(wi.FullUrl(), "s3:")
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

func (wi *WorkflowInclude) Domain() string {
	urlStr := wi.FullUrl()
	parsedUrl, err := url.Parse(urlStr)

	if err != nil {
		return urlStr
	}

	return parsedUrl.Hostname()
}

func (wi *WorkflowInclude) FullUrlPath() string {
	temp, _ := utils.ReplaceFilenameInUrl(wi.FullUrl(), "")

	return strings.TrimSuffix(temp, "/")
}

func (wi *WorkflowInclude) DisplayUrl() string {
	displayUrl := strings.Replace(wi.FullUrl(), "https://", "", -1)
	displayUrl = strings.Replace(displayUrl, "github.com/", "", -1)
	displayUrl = strings.Replace(displayUrl, "raw.githubusercontent.com/", "", -1)
	// displayUrl = strings.Replace(displayUrl, "s3://", "", -1)

	return displayUrl
}

func (wi *WorkflowInclude) DisplayName() string {
	if wi.IsRemoteUrl() {
		return wi.DisplayUrl()
	}

	if wi.IsLocalFile() {
		return wi.Filename()
	}

	if wi.IsS3Url() {
		return wi.DisplayUrl()
	}

	return "<unknown>"
}

func (wi *WorkflowInclude) GetChecksumAlgorithm() string {
	if strings.HasSuffix(wi.ChecksumUrl, ".sha256") || strings.HasSuffix(wi.ChecksumUrl, ".sha256.txt") {
		return "sha256"
	}
	if strings.HasSuffix(wi.ChecksumUrl, ".sha512") || strings.HasSuffix(wi.ChecksumUrl, ".sha512.txt") {
		return "sha512"
	}

	if strings.EqualFold(path.Base(wi.ChecksumUrl), "sha256sum") {
		return "sha256"
	}

	if strings.EqualFold(path.Base(wi.ChecksumUrl), "sha512sum") {
		return "sha512"
	}

	// check for md4 hash length:
	if len(wi.FoundChecksum) == 16 {
		return "md4"
	}
	if len(wi.FoundChecksum) == 32 {
		return "md5"
	}
	if len(wi.FoundChecksum) == 40 {
		return "sha1"
	}
	if len(wi.FoundChecksum) == 48 {
		return "sha224"
	}
	if len(wi.FoundChecksum) == 64 {
		return "sha256"
	}
	if len(wi.FoundChecksum) == 96 {
		return "sha384"
	}
	if len(wi.FoundChecksum) == 128 {
		return "sha512"
	}

	return "unknown"
}

func (wi *WorkflowInclude) SetChecksumIsValid(value bool) {
	wi.ChecksumIsValid = &value
}