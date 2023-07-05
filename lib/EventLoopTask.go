package lib

type EventLoopTask struct {
	Name    string `yaml:"name"`
	Bin     string `yaml:"bin"`
	Command string `yaml:"command"`
	Cwd     string `yaml:"cwd"`
}
