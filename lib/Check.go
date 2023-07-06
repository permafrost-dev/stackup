package lib

import (
	"fmt"

	"github.com/robertkrimen/otto"
	jsengine "github.com/stackup-app/stackup/lib/javascriptEngine"
	"github.com/stackup-app/stackup/utils"
)

type Check struct {
	Name  string `yaml:"name"`
	Check string `yaml:"check"`

	Result string
}

func (c *Check) Init(config *StackConfig) {
	c.Name = utils.ReplaceConfigurationKeyVariablesInMap(c.Name, config, "config")
	c.Check = utils.ReplaceConfigurationKeyVariablesInMap(c.Check, config, "config")
}

func (c *Check) Evaluate(vm *otto.Otto) string {
	result, err := jsengine.EvaluateScript(c.Check)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}

	c.Result = result.String()

	return c.Result
}
