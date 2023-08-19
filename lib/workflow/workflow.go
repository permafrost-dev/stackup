package workflow

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
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/settings"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name           string                  `yaml:"name"`
	Description    string                  `yaml:"description"`
	Version        string                  `yaml:"version"`
	Settings       *settings.Settings      `yaml:"settings"`
	Env            []string                `yaml:"env"`
	Init           string                  `yaml:"init"`
	Preconditions  []*WorkflowPrecondition `yaml:"preconditions"`
	Tasks          []*Task                 `yaml:"tasks"`
	TaskList       *lla.List
	Startup        []TaskReference    `yaml:"startup"`
	Shutdown       []TaskReference    `yaml:"shutdown"`
	Servers        []TaskReference    `yaml:"servers"`
	Scheduler      []ScheduledTask    `yaml:"scheduler"`
	Includes       []*WorkflowInclude `yaml:"includes"`
	State          *StackupWorkflowState
	Cache          *cache.Cache
	JsEngine       *scripting.JavaScriptEngine
	Gateway        *gateway.Gateway
	ProcessMap     *sync.Map
	CommandStartCb types.CommandCallback
	ExitAppFunc    func()
}

type StackupWorkflowState struct {
	CurrentTask *Task
	Stack       *lls.Stack
	History     *lls.Stack
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

func (workflow *StackupWorkflow) Initialize(configPath string) {
	workflow.Cache = cache.New("", configPath)

	// generate uuids for each task as the initial step, as other code below relies on a uuid existing
	for _, task := range workflow.Tasks {
		task.Uuid = utils.GenerateTaskUuid()
	}

	for _, tr := range workflow.Startup {
		tr.Initialize(workflow)
	}

	for _, tr := range workflow.Shutdown {
		tr.Initialize(workflow)
	}

	for _, tr := range workflow.Servers {
		tr.Initialize(workflow)
	}

	for _, st := range workflow.Scheduler {
		st.Initialize(workflow)
	}

	workflow.processEnvSection()
	workflow.createMissingSettingsSection()
	workflow.configureDefaultSettings()
	workflow.ProcessIncludes()

	if len(workflow.Init) > 0 {
		workflow.JsEngine.Evaluate(workflow.Init)
	}

	for _, pc := range workflow.Preconditions {
		pc.Initialize(workflow)
	}

	for _, task := range workflow.Tasks {
		task.Initialize(workflow)
	}

	workflow.Cache.DefaultTtl = workflow.Settings.Cache.TtlMinutes
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
		workflow.Settings.Domains = &settings.WorkflowSettingsDomains{Allowed: []string{}}
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
				workflow.Gateway.SetDomainHeaders(host.Hostname, host.Headers)
			}
		}
	}

	workflow.Gateway.SetAllowedDomains(workflow.Settings.Domains.Allowed)

	if workflow.Settings.Gateway == nil {
		workflow.Settings.Gateway = &settings.WorkflowSettingsGateway{
			ContentTypes: &settings.GatewayContentTypes{
				Blocked: []string{},
				Allowed: []string{},
			},
			FileExtensions: &settings.WorkflowSettingsGatewayFileExtensions{
				Allow: []string{},
				Block: []string{},
			},
			Middleware: []string{},
		}
	}

	if workflow.Settings.Gateway != nil && workflow.Settings.Gateway.ContentTypes != nil {
		workflow.Gateway.SetDomainContentTypes("*", workflow.Settings.Gateway.ContentTypes.Allowed)
		workflow.Gateway.SetBlockedContentTypes("*", workflow.Settings.Gateway.ContentTypes.Blocked)
	}

	if workflow.Settings.Cache.TtlMinutes <= 0 {
		workflow.Settings.Cache.TtlMinutes = 15
	}

	if len(workflow.Settings.DotEnvFiles) == 0 {
		workflow.Settings.DotEnvFiles = []string{".env"}
	}

	if workflow.Settings.Notifications == nil {
		workflow.Settings.Notifications = &settings.WorkflowSettingsNotifications{
			Telegram: &settings.WorkflowSettingsNotificationsTelegram{
				APIKey:  "",
				ChatIds: []string{},
			},
			Slack: &settings.WorkflowSettingsNotificationsSlack{
				WebhookUrl: "",
				ChannelIds: []string{},
			},
		}
	}

	if workflow.Settings.Notifications.Telegram == nil {
		workflow.Settings.Notifications.Telegram = &settings.WorkflowSettingsNotificationsTelegram{
			APIKey:  "",
			ChatIds: []string{},
		}
	}

	if workflow.Settings.Notifications.Slack == nil {
		workflow.Settings.Notifications.Slack = &settings.WorkflowSettingsNotificationsSlack{
			WebhookUrl: "",
			ChannelIds: []string{},
		}
	}

	tempChannelIds := []string{}
	for _, channelId := range workflow.Settings.Notifications.Slack.ChannelIds {
		if strings.HasPrefix(channelId, "$") {
			tempChannelIds = append(tempChannelIds, os.ExpandEnv(channelId))
		} else {
			tempChannelIds = append(tempChannelIds, channelId)
		}
	}
	workflow.Settings.Notifications.Slack.ChannelIds = tempChannelIds

	tempChatIds := []string{}
	for _, chatID := range workflow.Settings.Notifications.Telegram.ChatIds {
		if strings.HasPrefix(chatID, "$") {
			tempChatIds = append(tempChatIds, os.ExpandEnv(chatID))
		} else {
			tempChatIds = append(tempChatIds, chatID)
		}
	}
	workflow.Settings.Notifications.Telegram.ChatIds = tempChatIds

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
		gatewayAllow := "allow"
		workflow.Settings = &settings.Settings{
			AnonymousStatistics:  &enableStats,
			DotEnvFiles:          []string{".env"},
			Cache:                &settings.WorkflowSettingsCache{TtlMinutes: 15},
			ChecksumVerification: &verifyChecksums,
			Domains: &settings.WorkflowSettingsDomains{
				Allowed: []string{"raw.githubusercontent.com", "api.github.com"},
				Hosts: []settings.WorkflowSettingsDomainsHosts{
					{Hostname: "raw.githubusercontent.com", Gateway: &gatewayAllow, Headers: nil},
					{Hostname: "api.github.com", Gateway: &gatewayAllow, Headers: nil},
				},
			},
			Defaults: &settings.WorkflowSettingsDefaults{
				Tasks: &settings.WorkflowSettingsDefaultsTasks{
					Silent:    false,
					Path:      workflow.JsEngine.MakeStringEvaluatable("getCwd()"),
					Platforms: []string{"windows", "linux", "darwin"},
				},
			},
			Gateway: &settings.WorkflowSettingsGateway{
				ContentTypes: &settings.GatewayContentTypes{
					Blocked: []string{},
					Allowed: []string{"*"},
				},
			},
			Notifications: &settings.WorkflowSettingsNotifications{
				Telegram: &settings.WorkflowSettingsNotificationsTelegram{
					APIKey:  "",
					ChatIds: []string{},
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
	for _, inc := range workflow.Includes {
		inc.Initialize(workflow)
	}

	// load the includes asynchronously
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

func (workflow *StackupWorkflow) hasRemoteDomainAccess(include *WorkflowInclude) bool {
	if include.IsS3Url() || include.IsRemoteUrl() {
		if !workflow.Gateway.Allowed(include.FullUrl()) {
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
		include.HashAlgorithm = data.Algorithm
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
		workflow.Init += "\n" + template.Init
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

			if !include.ChecksumValidated && workflow.Settings.ExitOnChecksumMismatch {
				support.FailureMessageWithXMark("Exiting due to checksum mismatch.")
				workflow.ExitAppFunc()
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
			include.Contents, err = workflow.Gateway.GetUrl(include.Filename())
		} else if include.IsRemoteUrl() {
			include.Contents, err = workflow.Gateway.GetUrl(include.FullUrl(), include.Headers...)
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
		expires := carbon.Now().AddMinutes(workflow.Settings.Cache.TtlMinutes)

		item := workflow.Cache.CreateEntry(
			include.DisplayName(),
			include.Contents,
			&expires,
			include.Hash,
			include.HashAlgorithm,
			nil,
		)

		workflow.Cache.Set(include.DisplayName(), item, workflow.Settings.Cache.TtlMinutes)
	}

	return err
}

func (workflow *StackupWorkflow) importTasksFromIncludedTemplate(template *IncludedTemplate) {
	for _, t := range template.Tasks {
		t.FromRemote = true
		t.Uuid = utils.GenerateTaskUuid()
		workflow.Tasks = append(workflow.Tasks, t)
	}
}

// prepend the included preconditions; we reverse the order of the preconditions in the included file,
// then reverse the existing preconditions, append them, then reverse the workflow preconditions again
// to achieve the correct order.
func (workflow *StackupWorkflow) importPreconditionsFromIncludedTemplate(template *IncludedTemplate) {
	workflow.Preconditions = workflow.reversePreconditions(workflow.Preconditions)
	template.Preconditions = workflow.reversePreconditions(template.Preconditions)

	for _, p := range template.Preconditions {
		p.FromRemote = true
		workflow.Preconditions = append(workflow.Preconditions, p)
	}

	workflow.Preconditions = workflow.reversePreconditions(workflow.Preconditions)
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
	workflow.Gateway.SetAllowedDomains(workflow.Settings.Domains.Allowed)

	workflow.Settings.Gateway.ContentTypes.Allowed = utils.GetUniqueStrings(workflow.Settings.Gateway.ContentTypes.Allowed)
	workflow.Settings.Gateway.ContentTypes.Blocked = utils.GetUniqueStrings(workflow.Settings.Gateway.ContentTypes.Blocked)
	workflow.Gateway.SetDomainContentTypes("*", workflow.Settings.Gateway.ContentTypes.Allowed)
	workflow.Gateway.SetBlockedContentTypes("*", workflow.Settings.Gateway.ContentTypes.Blocked)
}

func (workflow *StackupWorkflow) GetSettings() *settings.Settings {
	return workflow.Settings
}
