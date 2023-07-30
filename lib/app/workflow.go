package app

import (
	"fmt"
	"strings"

	lla "github.com/emirpasic/gods/lists/arraylist"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name                string            `yaml:"name"`
	Description         string            `yaml:"description"`
	Version             string            `yaml:"version"`
	Settings            *WorkflowSettings `yaml:"settings"`
	Init                string            `yaml:"init"`
	Preconditions       []Precondition    `yaml:"preconditions"`
	Tasks               []*Task           `yaml:"tasks"`
	TaskList            *lla.List
	Startup             []TaskReference    `yaml:"startup"`
	Shutdown            []TaskReference    `yaml:"shutdown"`
	Servers             []TaskReference    `yaml:"servers"`
	Scheduler           []ScheduledTask    `yaml:"scheduler"`
	Includes            []*WorkflowInclude `yaml:"includes"`
	State               *StackupWorkflowState
	RemoteTemplateIndex *RemoteTemplateIndex
}

type WorkflowInclude struct {
	Url string `yaml:"url"`
}

type WorkflowSettings struct {
	Defaults               *WorkflowSettingsDefaults `yaml:"defaults"`
	RemoteIndexUrl         string                    `yaml:"remote-index-url"`
	ExitOnChecksumMismatch bool                      `yaml:"exit-on-checksum-mismatch"`
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

type Precondition struct {
	Name  string `yaml:"name"`
	Check string `yaml:"check"`
}

type TaskReference struct {
	Task string `yaml:"task"`
}

type ScheduledTask struct {
	Task string `yaml:"task"`
	Cron string `yaml:"cron"`
}

func GetState() *StackupWorkflowState {
	return App.Workflow.State
}

func (wi *WorkflowInclude) DisplayUrl() string {
	displayUrl := strings.Replace(wi.Url, "https://", "", -1)
	displayUrl = strings.Replace(displayUrl, "github.com/", "", -1)
	displayUrl = strings.Replace(displayUrl, "raw.githubusercontent.com/", "", -1)

	return displayUrl
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

func (workflow *StackupWorkflow) Initialize() {
	// generate uuids for each task as the initial step, as other code below relies on a uuid existing
	for _, task := range workflow.Tasks {
		task.Uuid = utils.GenerateTaskUuid()
	}

	// no default settings were provided, so create sensible defaults
	if workflow.Settings == nil {
		workflow.Settings = &WorkflowSettings{
			Defaults: &WorkflowSettingsDefaults{
				Tasks: &WorkflowSettingsDefaultsTasks{
					Silent:    false,
					Path:      App.JsEngine.MakeStringEvaluatable("getCwd()"),
					Platforms: []string{"windows", "linux", "darwin"},
				},
			},
		}
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

	workflow.ProcessIncludes()

	if len(workflow.Init) > 0 {
		App.JsEngine.Evaluate(workflow.Init)
	}

	for _, task := range workflow.Tasks {
		task.Initialize()
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
	workflow.RemoteTemplateIndex = &RemoteTemplateIndex{Loaded: false}

	if workflow.Settings.RemoteIndexUrl != "" {
		remoteIndex, err := LoadRemoteTemplateIndex(workflow.Settings.RemoteIndexUrl)

		if err != nil {
			support.WarningMessage("Unable to load remote template index")
		}

		remoteIndex.Loaded = true
		workflow.RemoteTemplateIndex = remoteIndex
		support.SuccessMessageWithCheck("Downloaded remote template index file.")
	}

	for _, include := range workflow.Includes {
		workflow.ProcessInclude(include)
	}
}

func (workflow *StackupWorkflow) ProcessInclude(include *WorkflowInclude) bool {
	if !strings.HasPrefix(strings.TrimSpace(include.Url), "https") || include.Url == "" {
		return false
	}

	contents, err := utils.GetUrlContents(include.Url)

	if err != nil {
		fmt.Println(err)
		return false
	}

	if App.Workflow.RemoteTemplateIndex.Loaded {
		support.StatusMessage("Validating checksum for remote template: "+include.DisplayUrl(), false)
		remoteMeta := App.Workflow.RemoteTemplateIndex.GetTemplate(include.Url)
		validated, err := remoteMeta.ValidateChecksum(contents)

		if err != nil {
			support.PrintXMarkLine()
			fmt.Println(err)
			return false
		}

		if !validated {
			support.PrintXMarkLine()
			support.WarningMessage("Checksum mismatch for remote template: " + include.DisplayUrl())

			if App.Workflow.Settings.ExitOnChecksumMismatch {
				support.FailureMessageWithXMark("Exiting due to checksum mismatch.")
				App.exitApp()
				return false
			}
		} else {
			support.PrintCheckMarkLine()
		}
	}

	template := &IncludedTemplate{}
	err = yaml.Unmarshal([]byte(contents), template)

	if err != nil {
		fmt.Println(err)
		return false
	}

	if len(template.Init) > 0 {
		workflow.Init += "\n\n" + template.Init
	}

	for _, t := range template.Tasks {
		t.FromRemote = true
		App.Workflow.Tasks = append(App.Workflow.Tasks, t)
	}

	support.SuccessMessageWithCheck("Loaded remote configuration file: " + include.DisplayUrl())

	return true
}

func (tr *TaskReference) TaskId() string {
	if App.JsEngine.IsEvaluatableScriptString(tr.Task) {
		return App.JsEngine.Evaluate(tr.Task).(string)
	}

	return tr.Task
}

func (st *ScheduledTask) TaskId() string {
	if App.JsEngine.IsEvaluatableScriptString(st.Task) {
		return App.JsEngine.Evaluate(st.Task).(string)
	}

	return st.Task
}
