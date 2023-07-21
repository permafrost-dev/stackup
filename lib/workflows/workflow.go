package workflows

import (
	"io/ioutil"
	"os/exec"
	"time"

	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name          string          `yaml:"name"`
	Description   string          `yaml:"description"`
	Version       string          `yaml:"version"`
	Preconditions []Precondition  `yaml:"preconditions"`
	Tasks         []Task          `yaml:"tasks"`
	Servers       []Server        `yaml:"servers"`
	Scheduler     []ScheduledTask `yaml:"scheduler"`
}

type Precondition struct {
	Name  string `yaml:"name"`
	Check string `yaml:"check"`
}
type Task struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	If      string `yaml:"if,omitempty"`
	Silent  bool   `yaml:"silent"`
	Cwd     string `yaml:"cwd"`
	On      string `yaml:"on"`
	Result  *exec.Cmd
}
type Server struct {
	Name      string   `yaml:"name"`
	Command   string   `yaml:"command"`
	Message   string   `yaml:"message,omitempty"`
	Cwd       string   `yaml:"cwd"`
	Platforms []string `yaml:"platforms,omitempty"`
}
type ScheduledTask struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Cron    string `yaml:"cron"`
	Cwd     string `yaml:"cwd"`
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

// func (p *ServerProcess) IsRunning() bool {
// 	return p.Status == "running"
// }

// func (p *ServerProcess) GetCpuUsage() float64 {
// 	return utils.CheckProcessCpuLoad(p.Pid)
// }

// func (p *ServerProcess) GetMemoryUsage() float64 {
// 	return utils.CheckProcessMemoryUsage(p.Pid)
// }

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