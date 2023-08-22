package workflow

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dotenv-org/godotenvvault"
	lla "github.com/emirpasic/gods/lists/arraylist"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/golang-module/carbon/v2"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/settings"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name              string                  `yaml:"name"`
	Description       string                  `yaml:"description"`
	Version           string                  `yaml:"version"`
	Settings          *settings.Settings      `yaml:"settings"`
	Env               []string                `yaml:"env"`
	Init              string                  `yaml:"init"`
	Preconditions     []*WorkflowPrecondition `yaml:"preconditions"`
	Tasks             []*Task                 `yaml:"tasks"`
	TaskList          *lla.List
	Startup           []*TaskReference  `yaml:"startup"`
	Shutdown          []*TaskReference  `yaml:"shutdown"`
	Servers           []*TaskReference  `yaml:"servers"`
	Scheduler         []*ScheduledTask  `yaml:"scheduler"`
	Includes          []WorkflowInclude `yaml:"includes"`
	State             StackupWorkflowState
	Cache             *cache.Cache
	JsEngine          *types.JavaScriptEngineContract
	Gateway           *gateway.Gateway
	ProcessMap        *sync.Map
	CommandStartCb    types.CommandCallback
	ExitAppFunc       func()
	IsPrimaryWorkflow bool
	types.AppWorkflowContract
}

func CreateWorkflow(gw *gateway.Gateway) *StackupWorkflow {
	return &StackupWorkflow{
		IsPrimaryWorkflow: true,
		Settings:          &settings.Settings{},
		Preconditions:     []*WorkflowPrecondition{},
		Tasks:             []*Task{},
		TaskList:          lla.New(),
		State:             StackupWorkflowState{},
		Includes:          []WorkflowInclude{},
		Cache:             cache.New("project", ""),
		Gateway:           gw,
		ProcessMap:        &sync.Map{},
	}
}

type StackupWorkflowState struct {
	CurrentTask *Task
	Stack       *lls.Stack
	History     *lls.Stack
}

type CleanupCallback = func()

// sets the current task, and pushes the previous task onto the stack if it was still running.
// returns a cleanup function callback that restores the state to its previous value.
func (ws *StackupWorkflowState) SetCurrent(task *Task) CleanupCallback {
	if ws.CurrentTask != nil {
		ws.Stack.Push(ws.CurrentTask)
	}

	ws.History.Push(task.Uuid)
	ws.CurrentTask = task

	return func() {
		ws.CurrentTask = nil

		value, ok := ws.Stack.Pop()
		if ok {
			ws.CurrentTask = value.(*Task)
		}
	}
}

func (workflow StackupWorkflow) FindTaskById(id string) (any, bool) {
	if len(id) == 0 {
		return nil, false
	}

	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Id, id) {
			return task, true
		}
	}

	return nil, false
}

func (workflow *StackupWorkflow) FindTaskByUuid(uuid string) *Task {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Uuid, uuid) && len(uuid) > 0 {
			return task
		}
	}

	return nil
}

func (workflow *StackupWorkflow) TryLoadDotEnvVaultFile(value string) {
	if !strings.EqualFold(value, "dotenv://vault") {
		return
	}

	if !utils.IsFile(utils.WorkingDir(".env.vault")) {
		return
	}

	vars, err := godotenvvault.Read()
	if err != nil {
		return
	}

	for k, v := range vars {
		os.Setenv(k, v)
	}
}

func (workflow *StackupWorkflow) GetAllTaskReferences() []*TaskReferenceContract {
	refs := []*TaskReferenceContract{}
	refs = utils.CastAndCombineArrays(refs, workflow.Startup)
	refs = utils.CastAndCombineArrays(refs, workflow.Shutdown)
	refs = utils.CastAndCombineArrays(refs, workflow.Servers)
	refs = utils.CastAndCombineArrays(refs, workflow.Scheduler)

	return refs
}

func (workflow *StackupWorkflow) Initialize(configPath string) {
	if workflow.IsPrimaryWorkflow {
		workflow.Cache = cache.New("", configPath)
		workflow.Settings = &settings.Settings{}
	}

	// generate uuids for each task as the initial step, as other code below relies on a uuid existing
	for _, task := range workflow.Tasks {
		task.Uuid = utils.GenerateTaskUuid()
	}

	// init the server, scheduler, and startup/shutdown items
	for _, tr := range workflow.GetAllTaskReferences() {
		(*tr).Initialize(workflow)
	}

	for _, pc := range workflow.Preconditions {
		pc.Workflow = workflow
	}

	if workflow.IsPrimaryWorkflow {
		workflow.processEnvSection()

		for _, task := range workflow.Tasks {
			task.Workflow = workflow
			task.Initialize()
		}

		workflow.Cache.DefaultTtl = workflow.Settings.Cache.TtlMinutes
	}
}

func (workflow *StackupWorkflow) ConfigureDefaultSettings() {
	if workflow.Settings == nil {
		workflow.Settings = &settings.Settings{
			Defaults: settings.WorkflowSettingsDefaults{
				Tasks: settings.WorkflowSettingsDefaultsTasks{},
			},
			Domains: settings.WorkflowSettingsDomains{
				Allowed: []string{},
				Blocked: []string{},
				Hosts:   []settings.WorkflowSettingsDomainsHost{},
			},
			Notifications: settings.WorkflowSettingsNotifications{
				Telegram: settings.WorkflowSettingsNotificationsTelegram{},
				Slack:    settings.WorkflowSettingsNotificationsSlack{},
			},
		}
	}

	workflow.Settings.Defaults.Tasks.Path = consts.DEFAULT_CWD_SETTING
	workflow.Settings.Defaults.Tasks.Platforms = consts.ALL_PLATFORMS

	workflow.Settings.ChecksumVerification = true
	workflow.Settings.Domains.Allowed = consts.DEFAULT_ALLOWED_DOMAINS
	workflow.Settings.Gateway.Middleware = []string{"validateUrl", "verifyFileType", "validateContentType"}

	workflow.Settings.Cache.TtlMinutes = 15

	workflow.Settings.DotEnvFiles = []string{".env"}

	// expand env var references
	tempChannelIds := []string{}
	for _, channelId := range workflow.Settings.Notifications.Slack.ChannelIds {
		tempChannelIds = append(tempChannelIds, os.ExpandEnv(channelId))
	}
	workflow.Settings.Notifications.Slack.ChannelIds = tempChannelIds

	tempChatIds := []string{}
	for _, chatID := range workflow.Settings.Notifications.Telegram.ChatIds {
		tempChatIds = append(tempChatIds, os.ExpandEnv(chatID))
	}
	workflow.Settings.Notifications.Telegram.ChatIds = tempChatIds

	newHosts := []string{}
	for _, host := range workflow.Settings.Domains.Hosts {
		if host.Gateway == "allow" || host.Gateway == "" {
			newHosts = append(newHosts, host.Hostname)
		}
		if len(host.Headers) > 0 {
			workflow.Gateway.DomainHeaders.Store(host.Hostname, host.Headers)
		}
	}

	workflow.Settings.Domains.Allowed = append(workflow.Settings.Domains.Allowed, newHosts...)
	workflow.Settings.Domains.Allowed = utils.GetUniqueStrings(workflow.Settings.Domains.Allowed)

	// copy the default settings into each task if appropriate
	for _, task := range workflow.Tasks {
		if task.Path == "" && len(workflow.Settings.Defaults.Tasks.Path) > 0 {
			task.Path = workflow.Settings.Defaults.Tasks.Path
		}

		task.Silent = workflow.Settings.Defaults.Tasks.Silent
		copy(task.Platforms, workflow.Settings.Defaults.Tasks.Platforms)
	}
}

func (workflow *StackupWorkflow) processEnvSection() {
	if len(workflow.Env) > 0 {
		for _, def := range workflow.Env {
			if strings.EqualFold(def, "dotenv://vault") {
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

// ProcessIncludes loads the includes and processes all included files in the workflow asynchronously,
// so the order in which they loading is not guaranteed.
func (workflow *StackupWorkflow) ProcessIncludes() {
	if !workflow.IsPrimaryWorkflow {
		return
	}

	for _, inc := range workflow.Includes {
		inc.Initialize(workflow)
	}

	// load the includes asynchronously
	var wg sync.WaitGroup
	for _, include := range workflow.Includes {
		wg.Add(1)
		go func(inc WorkflowInclude) {
			defer wg.Done()
			inc.Process(workflow)
		}(include)
	}
	wg.Wait()

	fmt.Printf("workflow now== %v\n", workflow)
}

func (workflow *StackupWorkflow) hasRemoteDomainAccess(include *WorkflowInclude) bool {
	if !include.IsS3Url() && !include.IsRemoteUrl() {
		return true
	}
	if workflow.Gateway.Allowed(include.FullUrl()) {
		return true
	}

	support.FailureMessageWithXMark("remote include (rejected): domain " + include.Domain() + " access denied.")
	return false
}

func (workflow *StackupWorkflow) tryLoadingCachedData(include *WorkflowInclude) *cache.CacheEntry {
	if !workflow.Cache.Has(include.Identifier()) {
		return nil
	}

	var data *cache.CacheEntry
	data, include.FromCache = workflow.Cache.Get(include.Identifier())

	if include.FromCache {
		include.Hash = data.Hash
		include.HashAlgorithm = data.Algorithm
		include.Contents = data.Value
	}

	return data
}

func (workflow *StackupWorkflow) loadRemoteFileInclude(include *WorkflowInclude) error {
	var err error
	var template *StackupWorkflow

	if include.Contents, err = workflow.Gateway.GetUrl(include.FullUrl()); err != nil {
		return err
	}

	if err = yaml.Unmarshal([]byte(include.Contents), &template); err != nil {
		return err
	}

	template.IsPrimaryWorkflow = false

	// template.JsEngine.CreateEnvironmentVariables(os.Environ())
	// template.JsEngine.CreateAppVariables(.Vars)

	template.Settings = workflow.Settings
	template.JsEngine = workflow.JsEngine
	template.Gateway = workflow.Gateway
	template.Cache = workflow.Cache

	for _, task := range template.Tasks {
		task.Workflow = workflow
		// task.Initialize()
	}

	template.Initialize(".")

	// template.Initialize("/tmp")

	if len(template.Init) > 0 {
		workflow.Init += "\n" + template.Init
	}

	workflow.cacheFetchedRemoteInclude(include)

	workflow.initializeAllTemplateTaskReferences(template)
	workflow.importPreconditionsFromIncludedTemplate(template)
	workflow.importTasksFromIncludedTemplate(template)
	//workflow.copySettingsFromIncludedTemplate(&template)

	return nil
}

func (workflow *StackupWorkflow) cacheFetchedRemoteInclude(include *WorkflowInclude) *cache.CacheEntry {
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

	workflow.Cache.Set(include.Identifier(), item, workflow.Settings.Cache.TtlMinutes)

	return item
}

func (workflow *StackupWorkflow) handleChecksumVerification(include *WorkflowInclude) bool {
	if !include.IsRemoteUrl() {
		return true
	}

	if !include.VerifyChecksum {
		return true
	}

	include.ValidationState = ""
	_, include.FoundChecksum, _ = include.ValidateChecksum(include.Contents)

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

	return true
}

func (workflow *StackupWorkflow) handleDataWasCached(data *cache.CacheEntry, include *WorkflowInclude) error {
	var template StackupWorkflow
	var err error

	if err = yaml.Unmarshal([]byte(include.Contents), &template); err != nil {
		support.FailureMessageWithXMark("cache include (failed: " + err.Error() + "): " + include.DisplayName())
		return err
	}

	template.IsPrimaryWorkflow = false

	template.Settings = workflow.Settings
	template.JsEngine = workflow.JsEngine
	template.Gateway = workflow.Gateway
	template.Cache = workflow.Cache
	workflow.Init += "\n" + template.Init

	//template.ConfigureDefaultSettings()

	for _, task := range template.Tasks {
		task.Workflow = workflow
		//task.Initialize()
	}

	template.Initialize(".")

	// template.Initialize("/tmp")

	if len(template.Init) > 0 {
		workflow.Init += "\n" + template.Init
	}

	workflow.initializeAllTemplateTaskReferences(&template)
	workflow.importPreconditionsFromIncludedTemplate(&template)
	workflow.importTasksFromIncludedTemplate(&template)

	return nil
}

// func (workflow StackupWorkflow) handleDataNotCached(data *cache.CacheEntry, include *WorkflowInclude) (*cache.CacheEntry, error) {
// 	var err error = nil

// 	// if data != nil && !data.IsExpired() {
// 	// 	return nil
// 	// }

// 	if include.IsLocalFile() {
// 		b, _ := os.ReadFile(include.Filename())
// 		include.Contents = string(b)
// 	} else if include.IsRemoteUrl() {
// 		include.Contents, err = workflow.Gateway.GetUrl(include.FullUrl(), include.Headers...)
// 	} else if include.IsS3Url() {
// 		include.AccessKey = os.ExpandEnv(include.AccessKey)
// 		include.SecretKey = os.ExpandEnv(include.SecretKey)
// 		include.Contents = downloader.ReadS3FileContents(include.FullUrl(), include.AccessKey, include.SecretKey, include.Secure)
// 	} else {
// 		fmt.Printf("unknown include type: %s\n", include.DisplayName())
// 		return nil, nil
// 	}

// 	if err != nil {
// 		return nil, err
// 	}

// 	include.Hash = checksums.CalculateSha256Hash(include.Contents)
// 	expires := carbon.Now().AddMinutes(workflow.Settings.Cache.TtlMinutes)

// 	item := workflow.Cache.CreateEntry(
// 		include.DisplayName(),
// 		include.Contents,
// 		&expires,
// 		include.Hash,
// 		include.HashAlgorithm,
// 		nil,
// 	)

// 	workflow.Cache.Set(include.Identifier(), item, workflow.Settings.Cache.TtlMinutes)

// 	return item, err
// }

func (workflow *StackupWorkflow) importTasksFromIncludedTemplate(template *StackupWorkflow) {
	for _, t := range template.Tasks {
		t.Workflow = workflow
		t.FromRemote = true
		t.Uuid = utils.GenerateTaskUuid()
		workflow.Tasks = append(workflow.Tasks, t)
	}
}

func (workflow *StackupWorkflow) initializeAllTemplateTaskReferences(template *StackupWorkflow) {
	for _, tr := range template.GetAllTaskReferences() {
		(*tr).Initialize(workflow)
	}
}

// prepend the included preconditions; we reverse the order of the preconditions in the included file,
// then reverse the existing preconditions, append them, then reverse the workflow preconditions again
// to achieve the correct order.
func (workflow *StackupWorkflow) importPreconditionsFromIncludedTemplate(template *StackupWorkflow) {
	workflow.Preconditions = utils.ReverseArray(workflow.Preconditions)
	template.Preconditions = utils.ReverseArray(template.Preconditions)

	for _, p := range template.Preconditions {
		p.FromRemote = true
		workflow.Preconditions = append(workflow.Preconditions, p)
	}

	workflow.Preconditions = utils.ReverseArray(workflow.Preconditions)
}

func (workflow *StackupWorkflow) copySettingsFromIncludedTemplate(template *StackupWorkflow) {
	workflow.Settings.AnonymousStatistics = template.Settings.AnonymousStatistics

	workflow.Settings.Domains.Allowed = append(workflow.Settings.Domains.Allowed, template.Settings.Domains.Allowed...)
	workflow.Settings.Domains.Blocked = append(workflow.Settings.Domains.Blocked, template.Settings.Domains.Blocked...)

	workflow.Settings.Domains.Allowed = utils.GetUniqueStrings(workflow.Settings.Domains.Allowed)
	workflow.Settings.Domains.Blocked = utils.GetUniqueStrings(workflow.Settings.Domains.Blocked)

	workflow.Settings.Cache.TtlMinutes = template.Settings.Cache.TtlMinutes

	//     if template.Settings.Defaults != nil {
	// 	workflow.Settings.Defaults = &settings.WorkflowSettingsDefaults{
	// 		Tasks: &settings.WorkflowSettingsDefaultsTasks{
	// 			Silent:    false,
	// 			Path:      "",
	// 			Platforms: template.Settings.Defaults.Tasks.Platforms,
	// 		},
	// 	}
	// }
	for _, contentType := range workflow.Settings.Gateway.ContentTypes.Blocked {
		workflow.Settings.Gateway.ContentTypes.Blocked = append(workflow.Settings.Gateway.ContentTypes.Blocked, contentType)
	}
	for _, contentType := range workflow.Settings.Gateway.ContentTypes.Allowed {
		workflow.Settings.Gateway.ContentTypes.Allowed = append(workflow.Settings.Gateway.ContentTypes.Allowed, contentType)
	}

	// workflow.Gateway.SetAllowedDomains(workflow.Settings.Domains.Allowed)
	// workflow.Gateway.SetDomainContentTypes("*", workflow.Settings.Gateway.ContentTypes.Allowed)
	// workflow.Gateway.SetBlockedContentTypes("*", workflow.Settings.Gateway.ContentTypes.Blocked)
}

func (workflow StackupWorkflow) GetSettings() *settings.Settings {
	return workflow.Settings
}

func (workflow *StackupWorkflow) GetJsEngine() *types.JavaScriptEngineContract {
	if workflow.JsEngine == nil {
		return nil
	}

	var ref interface{} = workflow.JsEngine
	result := ref.(types.JavaScriptEngineContract)
	return &result
}
