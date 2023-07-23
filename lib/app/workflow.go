package app

import "strings"

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

func (workflow *StackupWorkflow) FindTaskById(id string) *Task {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Id, id) && len(task.Id) > 0 {
			return task
		}
	}

	return nil
}
