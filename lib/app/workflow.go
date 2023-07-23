package app

import (
	"time"
)

type StackupWorkflow struct {
	Name          string          `yaml:"name"`
	Description   string          `yaml:"description"`
	Version       string          `yaml:"version"`
	Preconditions []Precondition  `yaml:"preconditions"`
	Tasks         []*Task         `yaml:"tasks"`
	Startup       []StartupItem   `yaml:"startup"`
	Shutdown      []ShutdownItem  `yaml:"shutdown"`
	Servers       []Server        `yaml:"servers"`
	Scheduler     []ScheduledTask `yaml:"scheduler"`
}

type Precondition struct {
	Name  string `yaml:"name"`
	Check string `yaml:"check"`
}

type StartupItem struct {
	Task string `yaml:"task"`
}

type ShutdownItem struct {
	Task string `yaml:"task"`
}

type Server struct {
	Task string `yaml:"task"`
}

type ScheduledTask struct {
	Task string `yaml:"task"`
	Cron string `yaml:"cron"`
}

type WorkflowState struct {
	CurrentStage    string
	CurrentId       string
	CurrentStep     string
	IsComplete      bool
	ServerProcesses []ServerProcess
	TaskStatuses    []TaskStatus
}

type TaskStatus struct {
	Name       string
	Status     string
	StartedAt  time.Time
	FinishedAt time.Time
	RuntimeMs  int64
	Pid        int32
}

type ServerProcess struct {
	Name      string   `yaml:"name"`
	Id        string   `yaml:"id,omitempty"`
	Command   string   `yaml:"command,omitempty"`
	Cwd       string   `yaml:"cwd"`
	Platforms []string `yaml:"platforms,omitempty"`

	StartedAt time.Time
	StoppedAt time.Time
	RuntimeMs int64
	Pid       int32
	Status    string
}

func (workflow *StackupWorkflow) FindTaskById(id string) *Task {
	for _, task := range workflow.Tasks {
		if task.Id == id && len(task.Id) > 0 {
			return task
		}
	}

	return nil
}
