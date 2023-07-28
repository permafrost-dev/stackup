package app

import (
	"encoding/json"
	"io/ioutil"

	"github.com/robertkrimen/otto"
)

type Composer struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	License     string            `json:"license"`
	Require     map[string]string `json:"require"`
	RequireDev  map[string]string `json:"require-dev"`
}

func CreateScriptComposerFunction(vm *otto.Otto) {
	vm.Set("composer", LoadComposerJson)
}

func LoadComposerJson(filename string) (*Composer, error) {
	composer := Composer{}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return &composer, err
	}

	err = json.Unmarshal(contents, &composer)
	if err != nil {
		return &composer, err
	}

	return &composer, nil
}

func (composer *Composer) HasDependency(name string) bool {
	for _, dependency := range composer.GetDependencies() {
		if dependency == name {
			return true
		}
	}
	return false
}

func (composer *Composer) GetDependencies() []string {
	dependencies := []string{}
	if composer.hasProperty("require") {
		require := composer.Require
		for name := range require {
			dependencies = append(dependencies, name)
		}
	}

	return dependencies
}

func (composer *Composer) GetDevDependencies() []string {
	dependencies := []string{}
	if composer.hasProperty("require-dev") {
		require := composer.RequireDev
		for name := range require {
			dependencies = append(dependencies, name)
		}
	}

	return dependencies
}

func (composer *Composer) GetDependency(name string) string {
	if composer.hasProperty("require") {
		require := composer.Require
		if version, ok := require[name]; ok {
			return version
		}
	}

	return ""
}

func (composer *Composer) GetDevDependency(name string) string {
	if composer.hasProperty("require-dev") {
		require := composer.RequireDev
		if version, ok := require[name]; ok {
			return version
		}
	}

	return ""
}

func (composer *Composer) hasProperty(name string) bool {
	return composer.getProperty(name) != nil
}

func (composer *Composer) getProperty(name string) interface{} {
	switch name {
	case "name":
		return composer.Name
	case "type":
		return composer.Type
	case "description":
		return composer.Description
	case "keywords":
		return composer.Keywords
	case "license":
		return composer.License
	case "require":
		return composer.Require
	case "require-dev":
		return composer.RequireDev
	default:
		return nil
	}
}
