package varsextension

import (
	"github.com/stackup-app/stackup/lib/types"
)

type ScriptVars struct {
	engine types.JavaScriptEngineContract
}

func Create(e types.JavaScriptEngineContract) *ScriptVars {
	return &ScriptVars{
		engine: e,
	}
}

func (sv *ScriptVars) GetName() string {
	return "vars"
}

func (ex *ScriptVars) OnInstall(engine types.JavaScriptEngineContract) {
	engine.GetVm().Set(ex.GetName(), ex)
}

func (sv *ScriptVars) Get(name string) any {
	v, _ := sv.engine.GetAppVars().Load(name)

	return v
}

func (sv *ScriptVars) Set(name string, value any) {
	sv.engine.GetAppVars().Store("$"+name, value)
	sv.engine.GetVm().Set("$"+name, value)
}

func (sv *ScriptVars) Has(name string) bool {
	_, result := sv.engine.GetAppVars().Load(name)

	return result
}
