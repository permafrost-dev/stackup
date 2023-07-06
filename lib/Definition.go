package lib

import "github.com/stackup-app/stackup/utils"

type Definition struct {
	Name      string `yaml:"name"`
	Value     string `yaml:"value"`
	conf      *StackConfig
	NewResult string
}

func (d *Definition) Init(config *StackConfig) {
	d.Name = utils.ReplaceConfigurationKeyVariablesInStruct(d.Name, config, "config")
	d.Value = utils.ReplaceConfigurationKeyVariablesInStruct(d.Value, config, "config")
	// temp := d.evaluate()
	// d.Value = temp
	// d.NewResult = temp-
}

func (d *Definition) evaluate() {
	// if !strings.Contains(d.Value, "{{") || !strings.Contains(d.Value, "}}") {
	// 	return
	// }
	// engine := NewEngine(GetApplication())
	// value, _ := engine.Evaluate(d.Value)

	// // if err != nil {
	// // 	return
	// // }

	// return value.(string)
	// return "test"
}
