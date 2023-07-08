package workflows

import (
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name          string          `yaml:"name"`
	Description   string          `yaml:"description"`
	Version       string          `yaml:"version"`
	Preset        string          `yaml:"preset,omitempty"`
	Binaries      Binaries        `yaml:"binaries"`
	Filenames     Filenames       `yaml:"filenames"`
	Preconditions []Precondition  `yaml:"preconditions"`
	Commands      []Command       `yaml:"commands"`
	Tasks         []Task          `yaml:"tasks"`
	Servers       []Server        `yaml:"servers"`
	Scheduler     []ScheduledTask `yaml:"scheduler"`
	EventLoop     EventLoop       `yaml:"event-loop"`
}

type Command struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
	Silent      bool   `yaml:"silent,omitempty"`
	On          string `yaml:"on"`
}

type Containers struct {
	Compose string `yaml:"compose"`
	Manager string `yaml:"manager"`
}
type Binaries struct {
	Php        string     `yaml:"php"`
	Containers Containers `yaml:"containers"`
}
type Filenames struct {
	Dotenv        []string `yaml:"dotenv"`
	Dockercompose string   `yaml:"dockercompose"`
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
}
type EventLoopJob struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Cwd     string `yaml:"cwd"`
}
type EventLoop struct {
	Interval string         `yaml:"interval"`
	Jobs     []EventLoopJob `yaml:"jobs"`
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

func (w *StackupWorkflow) FindCommand(name string) *Command {
	for _, cmd := range w.Commands {
		if strings.EqualFold(cmd.Name, name) {
			return &cmd
		}
	}
	return nil
}
