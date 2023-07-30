package app

import (
	"github.com/robertkrimen/otto"
)

type ScriptVars struct {
}

func CreateScriptVarsObject(vm *otto.Otto) {
	obj := &ScriptVars{}
	vm.Set("vars", obj)
}

func (vars *ScriptVars) Get(name string) any {
	v, _ := App.Vars.Load(name)

	return v
}

func (vars *ScriptVars) Set(name string, value any) {
	App.Vars.Store("$"+name, value)
	App.JsEngine.Vm.Set("$"+name, value)
}

func (vars *ScriptVars) Has(name string) bool {
	_, result := App.Vars.Load(name)

	return result
}
