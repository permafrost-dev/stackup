package lib

import (
	"strings"
	"time"

	"github.com/permafrost-dev/stack-supervisor/utils"
)

type Task struct {
	Name string `yaml:"name"`
	// Id      string        `yaml:"id"`
	Bin       string        `yaml:"bin"`
	Command   string        `yaml:"command"`
	Sync      bool          `yaml:"sync,omitempty"`
	If        string        `yaml:"if,omitempty"`
	Delay     time.Duration `yaml:"delay,omitempty"`
	Platforms []string      `yaml:"platforms,omitempty"`
	Result    bool
}

func (t *Task) Init(config *StackConfig) {
	// t.Name = utils.ReplaceConfigurationKeyVariablesInMap(t.Name, config, "config")
	for _, match := range utils.MatchDotProperties(t.Bin) {
		// println(strings.Trim(match, "${}"))
		t.Bin = strings.ReplaceAll(t.Bin, match, utils.GetDotPropertyStruct(config, strings.Trim(match, "${}")).(string))
	}
	//t.Bin = utils.ReplaceConfigurationKeyVariablesInMap(t.Bin, config, "config")
	t.Command = utils.ReplaceConfigurationKeyVariablesInMap(t.Command, config, "config")
	// t.If = utils.ReplaceConfigurationKeyVariablesInMap(t.If, config, "config")
}

func (t *Task) Evaluate(app *Application) {
	engine := NewEngine(app)
	result, err := engine.Evaluate(t.If)

	t.Result = result.(bool)

	if err != nil {
		t.Result = false
	}
}
