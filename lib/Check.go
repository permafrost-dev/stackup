package lib

import "github.com/permafrost-dev/stack-supervisor/utils"

type Check struct {
	Name      string   `yaml:"name"`
	Check     string   `yaml:"check"`
	Platforms []string `yaml:"platforms,omitempty"`
	Result    bool
}

func (c *Check) Init(config *StackConfig) {
	c.Name = utils.ReplaceConfigurationKeyVariablesInMap(c.Name, config, "config")
	c.Check = utils.ReplaceConfigurationKeyVariablesInMap(c.Check, config, "config")
}

func (c *Check) Evaluate(app *Application) {
	engine := NewEngine(app)
	result, err := engine.Evaluate(c.Check)

	c.Result = result.(bool)

	if err != nil {
		c.Result = false
	}
}
