package app

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dotenv-org/godotenvvault"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/debug"
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
	Startup        []*TaskReference        `yaml:"startup"`
	Shutdown       []*TaskReference        `yaml:"shutdown"`
	Servers        []*TaskReference        `yaml:"servers"`
	Scheduler      []*ScheduledTask        `yaml:"scheduler"`
	Includes       []WorkflowInclude       `yaml:"includes"`
	Debug          bool                    `yaml:"debug"`
	State          WorkflowState
	Cache          *cache.Cache
	JsEngine       *scripting.JavaScriptEngine
	Gateway        *gateway.Gateway
	ProcessMap     *sync.Map
	CommandStartCb types.CommandCallback
	ExitAppFunc    func()
	types.AppWorkflowContract
}

func CreateWorkflow(gw *gateway.Gateway, processMap *sync.Map) *StackupWorkflow {
	return &StackupWorkflow{
		Settings:      &settings.Settings{},
		Preconditions: []*WorkflowPrecondition{},
		Tasks:         []*Task{},
		State:         WorkflowState{},
		Includes:      []WorkflowInclude{},
		Gateway:       gw,
		ProcessMap:    processMap,
	}
}

func (workflow *StackupWorkflow) FindTaskById(id string) (*Task, bool) {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Id, id) && len(id) > 0 {
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

func (workflow *StackupWorkflow) GetAllTaskReferences() []*TaskReference {
	refs := []*TaskReference{}
	refs = utils.CombineArrays(refs, workflow.Startup)
	refs = utils.CombineArrays(refs, workflow.Shutdown)
	refs = utils.CombineArrays(refs, workflow.Servers)

	return refs
}

func (workflow *StackupWorkflow) Initialize(engine *scripting.JavaScriptEngine, configPath string) {
	workflow.JsEngine = engine
	workflow.processEnvSection()
	workflow.InitializeSections()
	workflow.ProcessIncludes()
}

func (workflow *StackupWorkflow) InitializeSections() {
	for _, t := range workflow.Tasks {
		t.Initialize(workflow) //.JsEngine, workflow.CommandStartCb, workflow.State.SetCurrent, workflow.ProcessMap.Store)
		t.SetDefaultSettings(workflow.Settings)
	}

	// init startup, shutdown, servers sections
	for _, t := range workflow.GetAllTaskReferences() {
		t.Initialize(workflow)
	}

	for _, st := range workflow.Scheduler {
		st.Initialize(workflow)
	}

	for _, pc := range workflow.Preconditions {
		pc.Initialize(workflow)
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

	if len(workflow.Settings.Defaults.Tasks.Path) == 0 {
		workflow.Settings.Defaults.Tasks.Path = consts.DEFAULT_CWD_SETTING
	}

	if len(workflow.Settings.Defaults.Tasks.Platforms) == 0 {
		workflow.Settings.Defaults.Tasks.Platforms = consts.ALL_PLATFORMS
	}

	if len(workflow.Settings.Domains.Allowed) == 0 {
		workflow.Settings.Domains.Allowed = consts.DEFAULT_ALLOWED_DOMAINS
	}

	if workflow.Settings.Cache.TtlMinutes <= 0 {
		workflow.Settings.Cache.TtlMinutes = consts.DEFAULT_CACHE_TTL_MINUTES
	}

	if len(workflow.Settings.DotEnvFiles) == 0 {
		workflow.Settings.DotEnvFiles = []string{".env"}
	}

	workflow.Settings.Gateway.Middleware = []string{"validateUrl", "verifyFileType", "validateContentType"}

	workflow.expandEnvVars(workflow.Settings.Notifications.Slack.ChannelIds)
	workflow.expandEnvVars(workflow.Settings.Notifications.Telegram.ChatIds)

	newHosts := []string{}
	for _, host := range workflow.Settings.Domains.Hosts {
		if host.Gateway == "allow" || host.Gateway == "" {
			newHosts = append(newHosts, host.Hostname)
		}
		if len(host.Headers) > 0 {
			workflow.Gateway.DomainHeaders.Store(host.Hostname, host.Headers)
		}
	}

	workflow.Settings.Domains.Allowed = utils.Unique(workflow.Settings.Domains.Allowed, newHosts)

	workflow.setDefaultOptionsForTasks()
}

func (workflow *StackupWorkflow) expandEnvVars(items []string) {
	for i, item := range items {
		items[i] = os.ExpandEnv(item)
	}
}

// copy the default task settings into each task if the settings are not already set
func (workflow *StackupWorkflow) setDefaultOptionsForTasks() {
	for _, task := range workflow.Tasks {
		task.SetDefaultSettings(workflow.Settings)
	}
}

func (workflow *StackupWorkflow) processEnvSection() {
	for _, str := range workflow.Env {
		if strings.EqualFold(str, "dotenv://vault") {
			workflow.TryLoadDotEnvVaultFile(str)
			continue
		}

		if strings.Contains(str, "=") {
			parts := strings.SplitN(str, "=", 2)
			os.Setenv(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
}

// ProcessIncludes loads the includes and processes all included files in the workflow asynchronously,
// so the order in which they are loaded is not guaranteed.
func (workflow *StackupWorkflow) ProcessIncludes() {
	var wgPreload sync.WaitGroup

	// cache requests so async loading doesn't cause the same file to be loaded multiple times
	for _, url := range workflow.GetPossibleIncludedChecksumUrls() {
		wgPreload.Add(1)
		go func(s string) {
			defer wgPreload.Done()
			workflow.Gateway.GetUrl(s)
		}(url)
	}
	wgPreload.Wait()

	var wgLoadIncludes sync.WaitGroup
	for _, include := range workflow.Includes {
		wgLoadIncludes.Add(1)
		go func(inc WorkflowInclude) {
			defer wgLoadIncludes.Done()
			workflow.ProcessInclude(&inc)
		}(include)
	}
	wgLoadIncludes.Wait()

	workflow.InitializeSections()
}

func (workflow *StackupWorkflow) GetIncludedUrls() []string {
	result := []string{}

	for _, include := range workflow.Includes {
		result = append(result, include.FullUrl())
	}

	return utils.GetUniqueStrings(result)
}

func (workflow *StackupWorkflow) GetPossibleIncludedChecksumUrls() []string {
	result := []string{}

	for _, wi := range workflow.Includes {
		checksumUrls := []string{
			wi.FullUrlPath() + "/checksums.txt",
			wi.FullUrlPath() + "/checksums.sha256.txt",
			wi.FullUrlPath() + "/checksums.sha512.txt",
			wi.FullUrlPath() + "/sha256sum",
			wi.FullUrlPath() + "/sha512sum",
			wi.FullUrl() + ".sha256",
			wi.FullUrl() + ".sha512",
		}
		result = utils.CombineArrays(result, checksumUrls)
	}

	return utils.GetUniqueStrings(result)
}

func (workflow *StackupWorkflow) tryLoadingCachedData(include *WorkflowInclude) bool {
	if !workflow.Cache.Has(include.Identifier()) {
		return false
	}

	data, loaded := workflow.Cache.Get(include.Identifier())
	include.SetLoadedFromCache(loaded, data)

	return loaded
}

func (workflow *StackupWorkflow) loadRemoteFileInclude(include *WorkflowInclude) error {
	var err error = nil

	if include.Contents, err = workflow.Gateway.GetUrl(include.FullUrl()); err != nil {
		return err
	}

	include.UpdateHash()
	workflow.Cache.Set(include.Identifier(), include.NewCacheEntry(), workflow.Settings.Cache.TtlMinutes)

	return err
}

func (workflow *StackupWorkflow) handleChecksumVerification(include *WorkflowInclude) bool {
	var result bool
	result = include.ValidateChecksum()

	if include.ValidationState.IsMismatch() && workflow.Settings.ExitOnChecksumMismatch {
		support.FailureMessageWithXMark("Exiting due to checksum mismatch.")
		workflow.ExitAppFunc()
	}

	return result
}

func (workflow *StackupWorkflow) loadAndImportInclude(rawYaml string) error {
	var template IncludedTemplate

	if err := yaml.Unmarshal([]byte(rawYaml), &template); err != nil {
		return err
	}

	for _, task := range template.Tasks {
		task.Initialize(workflow) // workflow.JsEngine, workflow.CommandStartCb, workflow.State.SetCurrent, workflow.ProcessMap.Store)
		workflow.Tasks = append(workflow.Tasks, task)
	}

	for _, precondition := range template.Preconditions {
		precondition.Initialize(workflow)
		workflow.Preconditions = append(workflow.Preconditions, precondition)
	}

	for _, startup := range template.Startup {
		startup.Initialize(workflow)
		workflow.Startup = append(workflow.Startup, startup)
	}

	for _, shutdown := range template.Shutdown {
		shutdown.Initialize(workflow)
		workflow.Shutdown = append(workflow.Shutdown, shutdown)
	}

	for _, server := range template.Servers {
		server.Initialize(workflow)
		workflow.Servers = append(workflow.Servers, server)
	}

	if len(template.Init) > 0 {
		workflow.Init += "\n" + template.Init
	}

	return nil
}

func (workflow *StackupWorkflow) ProcessInclude(include *WorkflowInclude) error {
	include.Initialize(workflow)

	loaded := workflow.tryLoadingCachedData(include)

	if !loaded {
		debug.Logf("include not loaded from cache: %s", include.DisplayUrl())

		if err := workflow.loadRemoteFileInclude(include); err != nil {
			support.FailureMessageWithXMark("remote include (rejected: " + err.Error() + "): " + include.DisplayName())
			return err
		}

		loaded = true
	}

	if loaded {
		if err := workflow.loadAndImportInclude(include.Contents); err != nil {
			support.FailureMessageWithXMark("include from cache failed: (" + err.Error() + "): " + include.DisplayName())
			return err
		}
	}

	if !loaded {
		support.FailureMessageWithXMark("remote include failed: " + include.DisplayName())
		return fmt.Errorf("unable to load remote include: %s", include.DisplayName())
	}

	if !workflow.handleChecksumVerification(include) {
		// support.FailureMessageWithXMark("checksum verification failed: " + include.DisplayName())
		// return fmt.Errorf("remote include checksum verification failed: %s", include.DisplayName())
	}

	support.SuccessMessageWithCheck("remote include (" + include.LoadedStatusText() + "): " + include.DisplayName())

	return nil
}
