package app

import (
	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/support"
)

type ScriptApp struct {
}

func CreateScriptAppObject(vm *otto.Otto) {
	obj := &ScriptApp{}
	vm.Set("app", obj)
}

func (app *ScriptApp) StatusMessage(message string) {
	support.StatusMessage(message, false)
}

func (app *ScriptApp) WarningMessage(message string) {
    support.WarningMessage(message)
}
