package scripting

type ScriptVars struct {
	engine *JavaScriptEngine
}

func CreateScriptVarsObject(e *JavaScriptEngine) {
	obj := &ScriptVars{
		engine: e,
	}
	e.Vm.Set("vars", obj)
}

func (sv *ScriptVars) Get(name string) any {
	v, _ := sv.engine.AppVars.Load(name)

	return v
}

func (sv *ScriptVars) Set(name string, value any) {
	sv.engine.AppVars.Store("$"+name, value)
	sv.engine.Vm.Set("$"+name, value)
}

func (sv *ScriptVars) Has(name string) bool {
	_, result := sv.engine.AppVars.Load(name)

	return result
}
