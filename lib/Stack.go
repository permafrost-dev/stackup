package lib

type Stack struct {
	Name        string          `yaml:"name"`
	Description string          `yaml:"description,omitempty"`
	Version     string          `yaml:"version,omitempty"`
	Definitions []Definition    `yaml:"definitions"`
	Checks      []Check         `yaml:"checks"`
	Tasks       []Task          `yaml:"tasks"`
	Processes   []Process       `yaml:"processes"`
	EventLoop   []EventLoopTask `yaml:"event-loop"`
	Scheduled   []Scheduled     `yaml:"scheduled,omitempty"`
}
