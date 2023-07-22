package workflows

import (
	"io/ioutil"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name          string          `yaml:"name"`
	Description   string          `yaml:"description"`
	Version       string          `yaml:"version"`
	Preconditions []Precondition  `yaml:"preconditions"`
	Tasks         []Task          `yaml:"tasks"`
	Startup       []StartupItem   `yaml:"startup"`
	Shutdown      []ShutdownItem  `yaml:"shutdown"`
	Servers       []Server        `yaml:"servers"`
	Scheduler     []ScheduledTask `yaml:"scheduler"`
}

type Precondition struct {
	Name  string `yaml:"name"`
	Check string `yaml:"check"`
}

type Task struct {
	Name      string   `yaml:"name"`
	Command   string   `yaml:"command"`
	If        string   `yaml:"if,omitempty"`
	Id        string   `yaml:"id,omitempty"`
	Silent    bool     `yaml:"silent"`
	Path      string   `yaml:"path"`
	Type      string   `yaml:"type"`
	Platforms []string `yaml:"platforms,omitempty"`
	MaxRuns   int      `yaml:"maxRuns,omitempty"`
	Result    *exec.Cmd
	RunCount  int
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

func LoadWorkflowFile(filename string) StackupWorkflow {
	var result StackupWorkflow

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return StackupWorkflow{}
	}

	err = yaml.Unmarshal(contents, &result)
	if err != nil {
		return StackupWorkflow{}
	}

	return result
}

func (workflow *StackupWorkflow) FindTaskById(id string) *Task {
	for _, task := range workflow.Tasks {
		if task.Id == id && len(task.Id) > 0 {
			return &task
		}
	}

	return nil
}

func (task *Task) CanRunOnCurrentPlatform() bool {
	if task.Platforms == nil || len(task.Platforms) == 0 {
		return true
	}

	foundPlatform := false

	for _, name := range task.Platforms {
		if strings.EqualFold(runtime.GOOS, name) {
			foundPlatform = true
			break
		}
	}

	return foundPlatform
}

func (task *Task) Initialize() {
	task.RunCount = 0

	if task.MaxRuns <= 0 {
		task.MaxRuns = 999999999
	}
}
