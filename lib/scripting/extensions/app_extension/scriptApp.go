package appextension

import (
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	appvers "github.com/stackup-app/stackup/lib/version"
)

const name = "app"

// const version = "0.0.1"
// const description = "Provides access to application methods."

type ScriptApp struct {
}

func Create() *ScriptApp {
	return &ScriptApp{}
}

func (app *ScriptApp) GetName() string {
	return name
}

func (ex *ScriptApp) OnInstall(engine types.JavaScriptEngineContract) {
	engine.GetVm().Set(ex.GetName(), ex)
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
	return appvers.APP_VERSION
}
