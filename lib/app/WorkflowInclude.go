package app

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/utils"
)

type WorkflowInclude struct {
	Url              string   `yaml:"url"`
	Headers          []string `yaml:"headers"`
	File             string   `yaml:"file"`
	ChecksumUrl      string   `yaml:"checksum-url"`
	VerifyChecksum   bool     `yaml:"verify,omitempty"`
	AccessKey        string   `yaml:"access-key"`
	SecretKey        string   `yaml:"secret-key"`
	Secure           bool     `yaml:"secure"`
	identifierString string
	ValidationState  ChecksumVerificationState
	Contents         string
	Hash             string
	FoundChecksum    string
	HashAlgorithm    checksums.ChecksumAlgorithm
	FromCache        bool
	Workflow         *StackupWorkflow
}

func expandUrlPrefixes(url string) string {
	var mapppd = map[string]string{
		"gh:": "https://raw.githubusercontent.com/",
		"s3:": "https://s3.amazonaws.com/",
	}

	result := url

	for k, v := range mapppd {
		if strings.HasPrefix(url, k) {
			result = strings.Replace(result, k, v, 1)
		}
	}

	return result
}

func formatDisplayUrl(urlstr string) string {
	parsed, _ := url.Parse(urlstr)

	return parsed.Hostname() + "/" + parsed.Path
}

func GetChecksumUrls(fullUrl string) []string {
	url, _ := url.Parse(fullUrl)
	reqFn := path.Base(url.Path)
	url.Path = path.Dir(url.Path)

	return []string{
		url.JoinPath("checksums.txt").String(),
		url.JoinPath("checksums.sha256.txt").String(),
		url.JoinPath("checksums.sha512.txt").String(),
		url.JoinPath("sha256sum").String(),
		url.JoinPath("sha512sum").String(),
		url.JoinPath("sha512sum").String(),
		url.JoinPath(reqFn + ".sha256").String(),
		url.JoinPath(reqFn + ".sha512").String(),
	}
}

func IsHashUrlForFileUrl(urlstr, filename string) bool {
	return strings.HasSuffix(urlstr, filename+".sha256") || strings.HasSuffix(urlstr, filename+".sha512")
}

func (wi *WorkflowInclude) Initialize(workflow *StackupWorkflow) {
	wi.Workflow = workflow
	wi.expandHeaders()
	wi.setDefaults()
	wi.updateIdentifier()
}

func (wi *WorkflowInclude) updateIdentifier() string {
	wi.identifierString = utils.ConsistentUniqueId(wi.FullUrl() + wi.Filename())

	return wi.identifierString
}

func (wi *WorkflowInclude) expandHeaders() {
	for i, v := range wi.Headers {
		if wi.Workflow.JsEngine.IsEvaluatableScriptString(v) {
			wi.Headers[i] = wi.Workflow.JsEngine.Evaluate(v).(string)
		}
		wi.Headers[i] = os.ExpandEnv(v)
	}
}

func (wi *WorkflowInclude) setDefaults() {
	wi.VerifyChecksum = wi.Workflow.Settings.ChecksumVerification
	wi.ValidationState = ChecksumVerificationStateNotVerified
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

func (wi *WorkflowInclude) ValidateChecksum() bool {
	wi.UpdateHash()
	wi.ValidationState.Reset()
	//wi.ValidationState = ChecksumVerificationStateNotVerified

	if !wi.IsRemoteUrl() || !wi.VerifyChecksum || !wi.Workflow.Settings.ChecksumVerification {
		return true
	}

	// wi.TransitionToNext(nil, false)
	wi.ValidationState = ChecksumVerificationStatePending
	found := false

	for _, url := range GetChecksumUrls(wi.FullUrl()) {
		urlText, err := wi.Workflow.Gateway.GetUrl(url)
		if err != nil || urlText == "" {
			continue
		}

		baseFn := path.Base(wi.FullUrl())
		if !strings.Contains(urlText, baseFn) && !IsHashUrlForFileUrl(url, baseFn) {
			continue
		}

		found = true
		wi.ChecksumUrl = url
		wi.UpdateChecksumFromChecksumsFile(url, urlText)
		// wi.TransitionToNext(wi.HashAlgorithm.UnsupportedError(), false)
		// wi.ValidationState = ChecksumVerificationStateError
		// test
		//wi.ValidationState =

		break
	}

	matched := found && !wi.ValidationState.IsError() && checksums.HashesMatch(wi.Hash, wi.FoundChecksum)

	if !found {
		wi.ValidationState = ChecksumVerificationStateError
	}
	if matched {
		wi.ValidationState = ChecksumVerificationStateVerified
	}

	// wi.TransitionToNext(nil, matched)
	//
	wi.ValidationState.SetVerified(matched)

	fmt.Printf("wi.ValidationState: %v\n", wi.ValidationState)

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
	if wi.File == "" {
		return ""
	}
	return utils.AbsoluteFilePath(wi.File)
}

func (wi *WorkflowInclude) FullUrl() string {
	if wi.Url == "" {
		return ""
	}
	return expandUrlPrefixes(wi.Url)
}

func (wi *WorkflowInclude) Identifier() string {
	if wi.identifierString == "" {
		return wi.updateIdentifier()
	}
	return wi.identifierString
}

func (wi WorkflowInclude) DisplayName() string {
	return utils.RemoveEmptyValues(formatDisplayUrl(wi.FullUrl()), wi.Filename(), "<unknown>")[0]
}

func (wi *WorkflowInclude) UpdateChecksumAlgorithm() {
	wi.HashAlgorithm = checksums.DetermineChecksumAlgorithm([]string{wi.FoundChecksum, wi.Hash}, wi.ChecksumUrl)
}

func (wi *WorkflowInclude) UpdateHash() {
	originalHash := wi.Hash

	wi.Hash = checksums.CalculateSha256Hash(wi.Contents)
	wi.HashAlgorithm = checksums.ChecksumAlgorithmSha256
	wi.ValidationState.ResetIf(wi.Hash != originalHash)
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

func (wi *WorkflowInclude) TransitionToNext(err error, matched bool) {
	possibleStates := _ChecksumVerificationStateTransitionMap[wi.ValidationState]

	if len(possibleStates) == 0 {
		return
	}

	if len(possibleStates) == 1 {
		wi.ValidationState = possibleStates[0]
		return
	}

	for _, state := range possibleStates {
		if state == ChecksumVerificationStateError {
			wi.ValidationState = ChecksumVerificationStateError
			return
		}
	}

	for _, state := range possibleStates {
		for _, finalState := range AllFinalCHecksumVerificationStates {
			if state == finalState {
				wi.ValidationState.SetVerified(true)
				return
			}
		}
	}
}
