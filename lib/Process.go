package lib

import "time"

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
}
