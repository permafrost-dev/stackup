package app

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dotenv-org/godotenvvault"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/debug"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/integrations"
	"github.com/stackup-app/stackup/lib/messages"
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
	Integrations   map[string]integrations.Integration
	ProcessMap     *sync.Map
	CommandStartCb types.CommandCallback
	ExitAppFunc    func()
	types.AppWorkflowContract
}

func CreateWorkflow(gw *gateway.Gateway, processMap *sync.Map) *StackupWorkflow {
	result := &StackupWorkflow{
		Settings:      &settings.Settings{},
		Preconditions: []*WorkflowPrecondition{},
		Tasks:         []*Task{},
		State:         WorkflowState{},
		Includes:      []WorkflowInclude{},
		Gateway:       gw,
		ProcessMap:    processMap,
		Integrations:  map[string]integrations.Integration{},
	}

	result.Integrations = integrations.List(result.AsContract)

	return result
}

func (workflow *StackupWorkflow) AsContract() types.AppWorkflowContract {
	return workflow
}

func (workflow *StackupWorkflow) FindTaskById(id string) (any, bool) {
	return workflow.GetTaskById(id)
}

func (workflow *StackupWorkflow) GetTaskById(id string) (*Task, bool) {
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

func (workflow *StackupWorkflow) GetEnvSection() []string {
	return workflow.Env
}

func (workflow *StackupWorkflow) TryLoadDotEnvVaultFile() {
	if !utils.ArrayContains(workflow.Env, "dotenv://vault") {
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
	return utils.CombineArrays([]*TaskReference{}, workflow.Startup, workflow.Shutdown, workflow.Servers)
}

func (workflow *StackupWorkflow) Initialize(engine *scripting.JavaScriptEngine, configPath string) {
	workflow.JsEngine = engine

	utils.ImportEnvDefsIntoEnvironment(workflow.Env)

	for _, integration := range workflow.Integrations {
		if integration.IsEnabled() {
			err := integration.Run()

			if err != nil {
				fmt.Printf(" integration error [%s]: %v\n", integration.Name(), err)
			}
		}
	}

	workflow.InitializeSections()
	workflow.processIncludes()
}

func (workflow *StackupWorkflow) InitializeSections() {
	for _, t := range workflow.Tasks {
		t.Initialize(workflow)
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
	utils.SetIfEmpty(&workflow.Settings.Defaults.Tasks.Path, consts.DEFAULT_CWD_SETTING)
	utils.SetIfEmpty(&workflow.Settings.Defaults.Tasks.Platforms, consts.ALL_PLATFORMS)
	utils.SetIfEmpty(&workflow.Settings.Domains.Allowed, consts.DEFAULT_ALLOWED_DOMAINS)
	utils.SetIfEmpty(&workflow.Settings.Domains.Blocked, []string{})
	utils.SetIfEmpty(&workflow.Settings.Cache.TtlMinutes, consts.DEFAULT_CACHE_TTL_MINUTES)
	utils.SetIfEmpty(&workflow.Settings.DotEnvFiles, []string{".env"})
	utils.SetIfEmpty(&workflow.Settings.Gateway.Middleware, consts.DEFAULT_GATEWAY_MIDDLEWARE)

	workflow.expandEnvVars(&workflow.Settings.Notifications.Slack.ChannelIds)
	workflow.expandEnvVars(&workflow.Settings.Notifications.Telegram.ChatIds)

	for _, host := range workflow.Settings.Domains.Hosts {
		if host.Gateway == "allow" || host.Gateway == "" {
			workflow.Settings.Domains.Allowed = append(workflow.Settings.Domains.Allowed, host.Hostname)
		}
		if host.Gateway == "block" {
			workflow.Settings.Domains.Blocked = append(workflow.Settings.Domains.Blocked, host.Hostname)
		}
		if len(host.Headers) > 0 {
			workflow.Gateway.DomainHeaders.Store(host.Hostname, host.Headers)
		}
	}

	utils.UniqueInPlace(&workflow.Settings.Domains.Allowed)
	utils.UniqueInPlace(&workflow.Settings.Domains.Blocked)

	// workflow.setDefaultOptionsForTasks()
}

func (workflow *StackupWorkflow) expandEnvVars(items *[]string) {
	expanded := make([]string, len(*items))

	for i, item := range *items {
		expanded[i] = os.ExpandEnv(item)
	}

	copy(*items, expanded)
}

// copy the default task settings into each task if the settings are not already set
// func (workflow *StackupWorkflow) setDefaultOptionsForTasks() {
// 	for _, task := range workflow.Tasks {
// 		task.SetDefaultSettings(workflow.Settings)
// 	}
// }

// processIncludes loads the includes and processes all included files in the workflow asynchronously,
// so the order in which they are loaded is not guaranteed.
func (workflow *StackupWorkflow) processIncludes() {
	var wgPreload sync.WaitGroup

	// cache requests so async loading doesn't cause the same file to be loaded multiple times
	for _, url := range workflow.getPossibleIncludedChecksumUrls() {
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
			workflow.processInclude(&inc)
		}(include)
	}
	wgLoadIncludes.Wait()

	workflow.InitializeSections()
}

func (workflow *StackupWorkflow) getIncludedUrls() []string {
	result := []string{}

	for _, include := range workflow.Includes {
		result = append(result, include.FullUrl())
	}

	return utils.GetUniqueStrings(result)
}

func (workflow *StackupWorkflow) getPossibleIncludedChecksumUrls() []string {
	result := []string{}

	for _, url := range workflow.getIncludedUrls() {
		result = append(result, checksums.GetChecksumUrls(url)...)
	}

	return result
}

func (workflow *StackupWorkflow) tryLoadingCachedData(include *WorkflowInclude) bool {
	data, loaded := workflow.Cache.Get(include.Identifier())
	include.setLoadedFromCache(loaded, data)

	return loaded
}

func (workflow *StackupWorkflow) loadRemoteFileInclude(include *WorkflowInclude) (error, bool) {
	var err error = nil
	var contents string

	if contents, err = workflow.Gateway.GetUrl(include.FullUrl()); err != nil {
		return err, false
	}

	include.SetContents(contents, true)

	return err, err == nil
}

func (workflow *StackupWorkflow) handleChecksumVerification(include *WorkflowInclude) bool {
	var result bool = include.ValidateChecksum()

	if include.ValidationState.IsMismatch() && workflow.Settings.ExitOnChecksumMismatch {
		support.FailureMessageWithXMark(messages.ExitDueToChecksumMismatch())
		workflow.ExitAppFunc()
	}

	return result
}

func (workflow *StackupWorkflow) loadAndImportInclude(rawYaml string) error {
	var template IncludedTemplate

	if err := yaml.Unmarshal([]byte(rawYaml), &template); err != nil {
		return err
	}

	template.Initialize(workflow)

	workflow.Tasks = append(workflow.Tasks, template.Tasks...)
	workflow.Preconditions = append(workflow.Preconditions, template.Preconditions...)
	workflow.Startup = append(workflow.Startup, template.Startup...)
	workflow.Shutdown = append(workflow.Shutdown, template.Shutdown...)
	workflow.Servers = append(workflow.Servers, template.Servers...)
	workflow.Init = strings.TrimSpace(workflow.Init + "\n" + template.Init)

	return nil
}

func (workflow *StackupWorkflow) processInclude(include *WorkflowInclude) error {
	include.Initialize(workflow)

	var err error = nil
	loaded := workflow.tryLoadingCachedData(include)

	if !loaded {
		debug.Logf("include not loaded from cache: %s", include.DisplayName())

		err, loaded = workflow.loadRemoteFileInclude(include)
		if !loaded {
			support.FailureMessageWithXMark(messages.RemoteIncludeStatus("rejected: "+err.Error(), include.DisplayName()))
			return err
		}
	}

	if !loaded {
		support.FailureMessageWithXMark(messages.RemoteIncludeStatus("failed", include.DisplayName()))
		return errors.New(messages.RemoteIncludeCannotLoad(include.DisplayName()))
	}

	if err := workflow.loadAndImportInclude(include.Contents); err != nil {
		support.FailureMessageWithXMark(messages.RemoteIncludeStatus("cache load failed", include.DisplayName()))
		return err
	}

	if !workflow.handleChecksumVerification(include) {
		// the app terminiates during handleChecksumVerification if the 'exit-on-checksum-mismatch' setting is enabled
		// so we can only show a wanring message here.
		support.WarningMessage(messages.RemoteIncludeChecksumMismatch(include.DisplayName()))
		return nil
	}

	support.SuccessMessageWithCheck(messages.RemoteIncludeStatus(include.loadedStatusText(), include.DisplayName()))

	return nil
}
