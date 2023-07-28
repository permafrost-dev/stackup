package app

import (
	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/version"
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

func (app *ScriptApp) StatusLine(message string) {
	support.StatusMessageLine(message, false)
}

func (app *ScriptApp) SuccessMessage(message string) {
	support.SuccessMessageWithCheck(message)
}

func (app *ScriptApp) FailureMessage(message string) {
	support.FailureMessageWithXMark(message)
}

func (app *ScriptApp) WarningMessage(message string) {
	support.WarningMessage(message)
}

func (app *ScriptApp) Version() string {
	return version.APP_VERSION
}
