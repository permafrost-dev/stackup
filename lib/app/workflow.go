package app

import (
	"strings"
)

type StackupWorkflow struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	Version       string            `yaml:"version"`
	Settings      *WorkflowSettings `yaml:"settings"`
	Init          string            `yaml:"init"`
	Preconditions []Precondition    `yaml:"preconditions"`
	Tasks         []*Task           `yaml:"tasks"`
	Startup       []TaskReference   `yaml:"startup"`
	Shutdown      []TaskReference   `yaml:"shutdown"`
	Servers       []TaskReference   `yaml:"servers"`
	Scheduler     []ScheduledTask   `yaml:"scheduler"`
	State         *StackupWorkflowState
}

type WorkflowSettings struct {
	Defaults *WorkflowSettingsDefaults `yaml:"defaults"`
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

func (workflow *StackupWorkflow) FindTaskById(id string) *Task {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Id, id) && len(task.Id) > 0 {
			return task
		}
	}

	return nil
}

func (workflow *StackupWorkflow) Initialize() {
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

	for _, task := range workflow.Tasks {
		task.Initialize()
	}

	if len(workflow.Init) > 0 {
		App.JsEngine.Evaluate(workflow.Init)
	}
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
