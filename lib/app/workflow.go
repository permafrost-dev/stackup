package app

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/dotenv-org/godotenvvault"
	lla "github.com/emirpasic/gods/lists/arraylist"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/downloader"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name          string                  `yaml:"name"`
	Description   string                  `yaml:"description"`
	Version       string                  `yaml:"version"`
	Settings      *WorkflowSettings       `yaml:"settings"`
	Env           []string                `yaml:"env"`
	Init          string                  `yaml:"init"`
	Preconditions []*WorkflowPrecondition `yaml:"preconditions"`
	Tasks         []*Task                 `yaml:"tasks"`
	TaskList      *lla.List
	Startup       []TaskReference    `yaml:"startup"`
	Shutdown      []TaskReference    `yaml:"shutdown"`
	Servers       []TaskReference    `yaml:"servers"`
	Scheduler     []ScheduledTask    `yaml:"scheduler"`
	Includes      []*WorkflowInclude `yaml:"includes"`
	State         *StackupWorkflowState
	Cache         *cache.Cache
}
type WorkflowSettings struct {
	Defaults               *WorkflowSettingsDefaults `yaml:"defaults"`
	ExitOnChecksumMismatch bool                      `yaml:"exit-on-checksum-mismatch"`
	ChecksumVerification   *bool                     `yaml:"checksum-verification"`
	DotEnvFiles            []string                  `yaml:"dotenv"`
	Cache                  *WorkflowSettingsCache    `yaml:"cache"`
	Domains                *WorkflowSettingsDomains  `yaml:"domains"`
	AnonymousStatistics    *bool                     `yaml:"anonymous-stats"`
	Gateway                *WorkflowSettingsGateway  `yaml:"gateway"`
}
type GatewayContentTypes struct {
	Blocked []string `yaml:"blocked"`
	Allowed []string `yaml:"allowed"`
}
type WorkflowSettingsGateway struct {
	ContentTypes *GatewayContentTypes `yaml:"content-types"`
}

type WorkflowSettingsDomains struct {
	Allowed []string                       `yaml:"allowed"`
	Hosts   []WorkflowSettingsDomainsHosts `yaml:"hosts"`
}

type WorkflowSettingsDomainsHosts struct {
	Hostname string   `yaml:"hostname"`
	Gateway  *string  `yaml:"gateway"`
	Headers  []string `yaml:"headers"`
}

type WorkflowSettingsCache struct {
	TtlMinutes int `yaml:"ttl-minutes"`
}
type WorkflowSettingsDefaults struct {
	Tasks *WorkflowSettingsDefaultsTasks `yaml:"tasks"`
}

type WorkflowSettingsDefaultsTasks struct {
	Silent    bool     `yaml:"silent"`
	Path      string   `yaml:"path"`
	Platforms []string `yaml:"platforms"`
}

type StackupWorkflowState struct {
	CurrentTask *Task
	Stack       *lls.Stack
	History     *lls.Stack
}

func GetState() *StackupWorkflowState {
	return App.Workflow.State
}

func (workflow *StackupWorkflow) FindTaskById(id string) *Task {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Id, id) && len(task.Id) > 0 {
			return task
		}
	}

	return nil
}

func (workflow *StackupWorkflow) FindTaskByUuid(uuid string) *Task {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Uuid, uuid) && len(uuid) > 0 {
			return task
		}
	}

	return nil
}

func (workflow *StackupWorkflow) TaskIdToUuid(id string) string {
	task := workflow.FindTaskById(id)

	if task == nil {
		return ""
	}

	return task.Uuid
}

func (workflow *StackupWorkflow) reversePreconditions(items []*WorkflowPrecondition) []*WorkflowPrecondition {
	length := len(items)
	for i := 0; i < length/2; i++ {
		items[i], items[length-i-1] = items[length-i-1], items[i]
	}

	return items
}

func (workflow *StackupWorkflow) TryLoadDotEnvVaultFile(value string) bool {
	if !utils.IsFile(utils.WorkingDir(".env.vault")) {
		return false
	}

	parsedUrl, err := url.Parse(value)
	if err != nil || parsedUrl.Scheme != "dotenv" || parsedUrl.Hostname() != "vault" {
		return false
	}

	vars, err := godotenvvault.Read()
	if err != nil {
		return false
	}

	for key, value := range vars {
		os.Setenv(key, value)
	}

	return true
}

func (workflow *StackupWorkflow) Initialize() {
	workflow.Cache = cache.CreateCache("")

	// generate uuids for each task as the initial step, as other code below relies on a uuid existing
	for _, task := range workflow.Tasks {
		task.Uuid = utils.GenerateTaskUuid()
	}

	workflow.processEnvSection()
	workflow.createMissingSettingsSection()
	workflow.configureDefaultSettings()
	workflow.ProcessIncludes()

	if len(workflow.Init) > 0 {
		App.JsEngine.Evaluate(workflow.Init)
	}

	for _, pc := range workflow.Preconditions {
		pc.Initialize()
	}

	for _, task := range workflow.Tasks {
		task.Initialize()
	}
}

func (workflow *StackupWorkflow) configureDefaultSettings() {
	if workflow.Settings.ChecksumVerification == nil {
		verifyChecksums := true
		workflow.Settings.ChecksumVerification = &verifyChecksums
	}

	if workflow.Settings.AnonymousStatistics == nil {
		enableStats := false
		workflow.Settings.AnonymousStatistics = &enableStats
	}

	if workflow.Settings.Domains == nil {
		workflow.Settings.Domains = &WorkflowSettingsDomains{Allowed: []string{}}
	}

	if len(workflow.Settings.Domains.Allowed) == 0 {
		workflow.Settings.Domains.Allowed = []string{"raw.githubusercontent.com", "api.github.com"}
	}

	if len(workflow.Settings.Domains.Hosts) > 0 {
		for _, host := range workflow.Settings.Domains.Hosts {
			if host.Gateway != nil && *host.Gateway == "allow" {
				workflow.Settings.Domains.Allowed = append(workflow.Settings.Domains.Allowed, host.Hostname)
			}
			if len(host.Headers) > 0 {
				App.Gateway.SetDomainHeaders(host.Hostname, host.Headers)
			}
		}
	}

	App.Gateway.SetAllowedDomains(workflow.Settings.Domains.Allowed)

	if workflow.Settings.Gateway == nil {
		workflow.Settings.Gateway = &WorkflowSettingsGateway{
			ContentTypes: &GatewayContentTypes{
				Blocked: []string{},
				Allowed: []string{},
			},
		}
	}

	if workflow.Settings.Gateway != nil && workflow.Settings.Gateway.ContentTypes != nil {
		App.Gateway.SetDomainContentTypes("*", workflow.Settings.Gateway.ContentTypes.Allowed)
		App.Gateway.SetBlockedContentTypes("*", workflow.Settings.Gateway.ContentTypes.Blocked)
	}

	if workflow.Settings.Cache.TtlMinutes <= 0 {
		workflow.Settings.Cache.TtlMinutes = 5
	}

	if len(workflow.Settings.DotEnvFiles) == 0 {
		workflow.Settings.DotEnvFiles = []string{".env"}
	}

	// copy the default settings into each task if appropriate
	for _, task := range workflow.Tasks {
		if task.Path == "" && len(workflow.Settings.Defaults.Tasks.Path) > 0 {
			task.Path = workflow.Settings.Defaults.Tasks.Path
		}

		if !task.Silent && workflow.Settings.Defaults.Tasks.Silent {
			task.Silent = workflow.Settings.Defaults.Tasks.Silent
		}

		if (task.Platforms == nil || len(task.Platforms) == 0) && len(workflow.Settings.Defaults.Tasks.Platforms) > 0 {
			task.Platforms = workflow.Settings.Defaults.Tasks.Platforms
		}
	}
}

func (workflow *StackupWorkflow) createMissingSettingsSection() {
	// no default settings were provided, so create sensible defaults
	if workflow.Settings == nil {
		verifyChecksums := true
		enableStats := false
		gatewayAllow := "allowed"
		workflow.Settings = &WorkflowSettings{
			AnonymousStatistics:  &enableStats,
			DotEnvFiles:          []string{".env"},
			Cache:                &WorkflowSettingsCache{TtlMinutes: 5},
			ChecksumVerification: &verifyChecksums,
			Domains: &WorkflowSettingsDomains{
				Allowed: []string{"raw.githubusercontent.com", "api.github.com"},
				Hosts: []WorkflowSettingsDomainsHosts{
					{Hostname: "raw.githubusercontent.com", Gateway: &gatewayAllow, Headers: nil},
					{Hostname: "api.github.com", Gateway: &gatewayAllow, Headers: nil},
				},
			},
			Defaults: &WorkflowSettingsDefaults{
				Tasks: &WorkflowSettingsDefaultsTasks{
					Silent:    false,
					Path:      App.JsEngine.MakeStringEvaluatable("getCwd()"),
					Platforms: []string{"windows", "linux", "darwin"},
				},
			},
			Gateway: &WorkflowSettingsGateway{
				ContentTypes: &GatewayContentTypes{
					Blocked: []string{},
					Allowed: []string{"*"},
				},
			},
		}
	}
}

func (workflow *StackupWorkflow) processEnvSection() {
	if len(workflow.Env) > 0 {
		for _, def := range workflow.Env {
			if strings.Contains(def, "://") {
				workflow.TryLoadDotEnvVaultFile(def)
				continue
			}

			if strings.Contains(def, "=") {
				key, value, _ := strings.Cut(def, "=")
				os.Setenv(key, value)
			}
		}
	}
}

func (workflow *StackupWorkflow) RemoveTasks(uuidsToRemove []string) {
	// Create a map of UUIDs to remove for faster lookup
	uuidMap := make(map[string]bool)
	for _, uuid := range uuidsToRemove {
		uuidMap[uuid] = true
	}

	// Remove tasks with UUIDs in the uuidMap
	var newTasks []*Task
	for _, task := range workflow.Tasks {
		if !uuidMap[task.Uuid] {
			newTasks = append(newTasks, task)
		}
	}
	workflow.Tasks = newTasks
}

func (workflow *StackupWorkflow) ProcessIncludes() {
	// initialize the includes
	for _, inc := range workflow.Includes {
		inc.Initialize(workflow)
	}

	var wg sync.WaitGroup
	for _, include := range workflow.Includes {
		wg.Add(1)
		go func(inc *WorkflowInclude) {
			defer wg.Done()
			inc.Process()
		}(include)
	}
	wg.Wait()
}

func (*StackupWorkflow) hasRemoteDomainAccess(include *WorkflowInclude) bool {
	if include.IsS3Url() || include.IsRemoteUrl() {
		if !App.Gateway.Allowed(include.FullUrl()) {
			support.FailureMessageWithXMark("remote include (rejected): domain " + include.Domain() + " access denied.")
			return false
		}
	}

	return true
}

func (workflow *StackupWorkflow) tryLoadingCachedData(include *WorkflowInclude) *cache.CacheEntry {
	data, found := workflow.Cache.Get(include.DisplayName())
	include.FromCache = found

	if include.FromCache {
		include.Hash = data.Hash
		include.HashAlgorithm = data.Hash
		include.Contents = data.Value
	}

	return data
}

func (workflow *StackupWorkflow) loadRemoteFileInclude(include *WorkflowInclude) error {
	var template IncludedTemplate
	err := yaml.Unmarshal([]byte(include.Contents), &template)

	if err != nil {
		return err
	}

	if len(template.Init) > 0 {
		App.Workflow.Init += "\n" + template.Init
	}

	workflow.importPreconditionsFromIncludedTemplate(&template)
	workflow.importTasksFromIncludedTemplate(&template)
	workflow.copySettingsFromIncludedTemplate(&template)

	return nil
}

func (workflow *StackupWorkflow) handleChecksumVerification(include *WorkflowInclude) bool {
	if workflow.Settings.ChecksumVerification != nil && *workflow.Settings.ChecksumVerification {
		include.ChecksumValidated, include.FoundChecksum, _ = include.ValidateChecksum(include.Contents)
	}

	include.ValidationState = ""

	if workflow.Settings.ChecksumVerification != nil && *workflow.Settings.ChecksumVerification && include.IsRemoteUrl() {
		if *include.VerifyChecksum == true || include.VerifyChecksum == nil {
			if include.ChecksumValidated {
				include.ValidationState = "verified"
			}

			if !include.ChecksumValidated && include.FoundChecksum != "" {
				include.ValidationState = "verification failed"
			}

			if !include.ChecksumValidated && App.Workflow.Settings.ExitOnChecksumMismatch {
				support.FailureMessageWithXMark("Exiting due to checksum mismatch.")
				App.exitApp()
				return false
			}
		}
	}
	return true
}

func (workflow *StackupWorkflow) handleDataNotCached(found bool, data *cache.CacheEntry, include *WorkflowInclude) error {
	var err error = nil

	if !found || data.IsExpired() {
		if include.IsLocalFile() {
			include.Contents, err = App.Gateway.GetUrl(include.Filename())
		} else if include.IsRemoteUrl() {
			include.Contents, err = App.Gateway.GetUrl(include.FullUrl(), include.Headers...)
		} else if include.IsS3Url() {
			include.AccessKey = os.ExpandEnv(include.AccessKey)
			include.SecretKey = os.ExpandEnv(include.SecretKey)
			include.Contents = downloader.ReadS3FileContents(include.FullUrl(), include.AccessKey, include.SecretKey, include.Secure)
		} else {
			fmt.Printf("unknown include type: %s\n", include.DisplayName())
			return nil
		}

		if err != nil {
			return err
		}

		include.Hash = checksums.CalculateSha256Hash(include.Contents)
		expires := carbon.Now().AddMinutes(App.Workflow.Settings.Cache.TtlMinutes)
		now := carbon.Now()

		item := cache.CreateCacheEntry(
			include.DisplayName(),
			include.Contents,
			&expires,
			include.Hash,
			include.HashAlgorithm,
			&now,
		)

		workflow.Cache.Set(include.DisplayName(), item, App.Workflow.Settings.Cache.TtlMinutes)
	}

	return err
}

func (*StackupWorkflow) importTasksFromIncludedTemplate(template *IncludedTemplate) {
	for _, t := range template.Tasks {
		t.FromRemote = true
		t.Uuid = utils.GenerateTaskUuid()
		App.Workflow.Tasks = append(App.Workflow.Tasks, t)
	}
}

// prepend the included preconditions; we reverse the order of the preconditions in the included file,
// then reverse the existing preconditions, append them, then reverse the workflow preconditions again
// to achieve the correct order.
func (*StackupWorkflow) importPreconditionsFromIncludedTemplate(template *IncludedTemplate) {
	App.Workflow.Preconditions = App.Workflow.reversePreconditions(App.Workflow.Preconditions)
	template.Preconditions = App.Workflow.reversePreconditions(template.Preconditions)

	for _, p := range template.Preconditions {
		p.FromRemote = true
		App.Workflow.Preconditions = append(App.Workflow.Preconditions, p)
	}

	App.Workflow.Preconditions = App.Workflow.reversePreconditions(App.Workflow.Preconditions)
}

func (workflow *StackupWorkflow) copySettingsFromIncludedTemplate(template *IncludedTemplate) {
	if template.Settings != nil {
		if template.Settings.ChecksumVerification != nil {
			workflow.Settings.ChecksumVerification = template.Settings.ChecksumVerification
		}
		if template.Settings.AnonymousStatistics != nil {
			workflow.Settings.AnonymousStatistics = template.Settings.AnonymousStatistics
		}
		if template.Settings.Domains != nil {
			for _, domain := range template.Settings.Domains.Allowed {
				workflow.Settings.Domains.Allowed = append(workflow.Settings.Domains.Allowed, domain)
			}
		}
		if template.Settings.Cache != nil {
			workflow.Settings.Cache = template.Settings.Cache
		}
		if template.Settings.Defaults != nil {
			workflow.Settings.Defaults = template.Settings.Defaults
		}
		if workflow.Settings.Gateway != nil && workflow.Settings.Gateway.ContentTypes != nil {
			for _, contentType := range workflow.Settings.Gateway.ContentTypes.Blocked {
				workflow.Settings.Gateway.ContentTypes.Blocked = append(workflow.Settings.Gateway.ContentTypes.Blocked, contentType)
			}
			for _, contentType := range workflow.Settings.Gateway.ContentTypes.Allowed {
				workflow.Settings.Gateway.ContentTypes.Allowed = append(workflow.Settings.Gateway.ContentTypes.Allowed, contentType)
			}
		}
		if template.Settings.Gateway != nil {
			workflow.Settings.Gateway = template.Settings.Gateway
		}
	}

	workflow.Settings.Domains.Allowed = utils.GetUniqueStrings(workflow.Settings.Domains.Allowed)
	App.Gateway.SetAllowedDomains(workflow.Settings.Domains.Allowed)

	workflow.Settings.Gateway.ContentTypes.Allowed = utils.GetUniqueStrings(workflow.Settings.Gateway.ContentTypes.Allowed)
	workflow.Settings.Gateway.ContentTypes.Blocked = utils.GetUniqueStrings(workflow.Settings.Gateway.ContentTypes.Blocked)
	App.Gateway.SetDomainContentTypes("*", workflow.Settings.Gateway.ContentTypes.Allowed)
	App.Gateway.SetBlockedContentTypes("*", workflow.Settings.Gateway.ContentTypes.Blocked)
}
