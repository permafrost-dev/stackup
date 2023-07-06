package lib

import (
	"time"

	"github.com/stackup-app/stackup/utils"
)

type Process struct {
	Name         string        `yaml:"name"`
	Id           string        `yaml:"id,omitempty"`
	Dependencies []string      `yaml:"dependencies,omitempty"`
	Bin          string        `yaml:"bin"`
	Command      string        `yaml:"command,omitempty"`
	Args         string        `yaml:"args"`
	Cwd          string        `yaml:"cwd"`
	Commands     string        `yaml:"commands,omitempty"`
	Delay        time.Duration `yaml:"delay,omitempty"`
	Platforms    []string      `yaml:"platforms,omitempty"`

	StartedAt time.Time
	StoppedAt time.Time
	RuntimeMs int64
	Pid       int32
	Status    string
}

func (p *Process) IsRunning() bool {
	return p.Status == "running"
}

func (p *Process) GetCpuUsage() float64 {
	return utils.CheckProcessCpuLoad(p.Pid)
}

func (p *Process) GetMemoryUsage() float64 {
	return utils.CheckProcessMemoryUsage(p.Pid)
}
