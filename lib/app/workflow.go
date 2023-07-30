package app

import (
	"fmt"
	"strings"

	lla "github.com/emirpasic/gods/lists/arraylist"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
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
	Startup             []TaskReference `yaml:"startup"`
	Shutdown            []TaskReference `yaml:"shutdown"`
	Servers             []TaskReference `yaml:"servers"`
	Scheduler           []ScheduledTask `yaml:"scheduler"`
	State               *StackupWorkflowState
	RemoteTemplateIndex *RemoteTemplateIndex
}

type WorkflowSettings struct {
	Defaults       *WorkflowSettingsDefaults `yaml:"defaults"`
	RemoteIndexUrl string                    `yaml:"remote-index-url"`
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

	fmt.Println(workflow.Settings.RemoteIndexUrl)

	if workflow.Settings.RemoteIndexUrl != "" {
		remoteIndex, err := LoadRemoteTemplateIndex(workflow.Settings.RemoteIndexUrl)

		if err != nil {
			support.WarningMessage("Unable to load remote template index")
		}

		remoteIndex.Loaded = true
		workflow.RemoteTemplateIndex = remoteIndex

	}

	uuidsToRemove := []string{}

	for _, task := range workflow.Tasks {
		task.Initialize()
		task.ProcessInclude()
		if task.Include != "" {
			uuidsToRemove = append(uuidsToRemove, task.Uuid)
		}
	}

	workflow.RemoveTasks(uuidsToRemove)
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
