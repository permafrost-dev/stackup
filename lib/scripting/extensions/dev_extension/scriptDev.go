package devextension

import (
	"github.com/stackup-app/stackup/lib/types"
)

const name = "dev"

// const version = "0.0.1"
// const description = "Provides access to development tools."

type ScriptDev struct {
	types.ScriptExtensionContract
}

func Create() *ScriptDev {
	return &ScriptDev{}
}

func (ex *ScriptDev) OnInstall(engine types.JavaScriptEngineContract) {
	engine.GetVm().Set(ex.GetName(), ex)
}

func (dev *ScriptDev) GetName() string {
	return name
}

func (dev *ScriptDev) ComposerJson(filename string) *Composer {
	result, _ := LoadComposerJson(filename)
	return result
}

func (dev *ScriptDev) PackageJson(filename string) *PackageJSON {
	result, _ := LoadPackageJson(filename)
	return result
}

func (dev *ScriptDev) RequirementsTxt(filename string) *RequirementsTxt {
	result, _ := LoadRequirementsTxt(filename)
	return result
}
