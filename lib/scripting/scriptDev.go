package scripting

type ScriptDev struct {
}

func CreateScriptDevObject(e *JavaScriptEngine) {
	obj := &ScriptDev{}
	e.Vm.Set("dev", obj)
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
