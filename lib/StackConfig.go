package lib

type StackConfig struct {
	Version      string       `yaml:"version"`
	Definitions  []Definition `yaml:"definitions"`
	Applications Applications `yaml:"applications"`
	Options      Options      `yaml:"options"`
	Stack        Stack        `yaml:"stack"`

	Props map[string]interface{}
}

func (c *StackConfig) FindDefinition(name string) *Definition {
	for _, def := range c.Definitions {
		if def.Name == name {
			return &def
		}
	}

	return nil
}
