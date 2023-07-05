package lib

import (
	"strings"

	"github.com/permafrost-dev/stack-supervisor/utils"
)

type Definition struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
	conf  *StackConfig
}

func (d *Definition) Init(config *StackConfig) {
	d.Name = utils.ReplaceConfigurationKeyVariablesInStruct(d.Name, config, "config")
	d.Value = utils.ReplaceConfigurationKeyVariablesInStruct(d.Value, config, "config")
}

func (d *Definition) Evaluate() {
	if !strings.Contains(d.Value, "{{") || !strings.Contains(d.Value, "}}") {
		return
	}

	engine := Engine{}
	value, err := engine.Evaluate(d.Value)

	if err != nil {
		return
	}

	d.Value = value.(string)
}
