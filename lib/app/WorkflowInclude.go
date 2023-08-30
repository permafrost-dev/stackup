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

	for k, v := range mapppd {
		if strings.HasPrefix(url, k) {
			return strings.Replace(url, k, v, 1)
		}
	}

	return url
}

func IsHashUrlForFileUrl(urlstr, filename string) bool {
	return strings.HasSuffix(urlstr, filename+".sha256") || strings.HasSuffix(urlstr, filename+".sha512")
}

func (wi *WorkflowInclude) Initialize(workflow *StackupWorkflow) {
	wi.Workflow = workflow
	wi.expandHeaders()
	wi.setDefaults()
}

func (wi *WorkflowInclude) IncludeType() IncludeType {
	return DetermineIncludeType(wi.FullUrl(), wi.Filename())
}

func (wi *WorkflowInclude) SetContents(contents string, storeInCache bool) {
	wi.Contents = contents
	wi.UpdateHash()

	if storeInCache {
		wi.Workflow.Cache.Set(wi.Identifier(), wi.NewCacheEntry(), wi.Workflow.Settings.Cache.TtlMinutes)
	}
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

func (wi *WorkflowInclude) loadedStatusText() string {
	result := "fetched"

	if wi.FromCache {
		result = "cached"
	}

	return fmt.Sprintf("%s, %s", result, wi.ValidationState.String())
}

func (wi *WorkflowInclude) setLoadedFromCache(loaded bool, data *cache.CacheEntry) {
	wi.FromCache = loaded

	if !loaded {
		return
	}

	wi.Hash = data.Hash
	wi.HashAlgorithm = checksums.ParseChecksumAlgorithm(data.Algorithm)
	wi.SetContents(data.Value, false)
}

func (wi *WorkflowInclude) UpdateChecksumFromChecksumsFile(contents string) {
	cs := checksums.FindFilenameChecksum(path.Base(wi.FullUrl()), contents)
	if cs != nil {
		wi.FoundChecksum = cs.Hash
	}

	wi.UpdateChecksumAlgorithm()
}

func (wi *WorkflowInclude) shouldVerifyChecksum() bool {
	return wi.IncludeType() == IncludeTypeHttp || wi.VerifyChecksum || wi.Workflow.Settings.ChecksumVerification
}

func (wi *WorkflowInclude) ValidateChecksum() bool {
	wi.UpdateHash()
	wi.ValidationState.Reset()

	if !wi.shouldVerifyChecksum() {
		return true
	}

	wi.ValidationState = ChecksumVerificationStatePending
	found := false

	for _, url := range wi.Workflow.getPossibleIncludedChecksumUrls() {
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
		wi.UpdateChecksumFromChecksumsFile(urlText)

		break
	}

	matched := found && !wi.ValidationState.IsError() && checksums.HashesMatch(wi.Hash, wi.FoundChecksum)
	wi.ValidationState.SetVerified(matched)

	if !found {
		wi.ValidationState = ChecksumVerificationStateError
	}

	return wi.ValidationState.IsVerified()
}

func (wi *WorkflowInclude) Filename() string {
	return utils.AbsoluteFilePath(wi.File)
}

func (wi *WorkflowInclude) FullUrl() string {
	return expandUrlPrefixes(wi.Url)
}

func (wi *WorkflowInclude) Identifier() string {
	return utils.ConsistentUniqueId(wi.FullUrl() + wi.Filename())
}

func (wi WorkflowInclude) DisplayName() string {
	return utils.FirstNonEmpty(
		utils.FormatDisplayUrl(wi.FullUrl()),
		wi.Filename(),
		"<unknown>",
	)
}

func (wi *WorkflowInclude) UpdateChecksumAlgorithm() {
	wi.HashAlgorithm = checksums.DetermineChecksumAlgorithm([]string{wi.FoundChecksum, wi.Hash}, wi.ChecksumUrl)
}

func (wi *WorkflowInclude) UpdateHash() {
	originalHash := wi.Hash

	wi.Hash, wi.HashAlgorithm = checksums.CalculateSha256Hash(wi.Contents)
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
