package scripting

import (
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
)

type ScriptNet struct {
	engine *JavaScriptEngine
}

func CreateScriptNetObject(e *JavaScriptEngine) {
	obj := &ScriptNet{
		engine: e,
	}
	e.Vm.Set("net", obj)
}

func (net *ScriptNet) Fetch(url string) any {
	if !net.engine.AppGateway.Allowed(url) {
		support.FailureMessageWithXMark("fetch failed: access to " + url + " is not allowed.")
		return ""
	}

	result, _ := utils.GetUrlContents(url)

	return result
}

func (net *ScriptNet) FetchJson(url string) any {
	if !net.engine.AppGateway.Allowed(url) {
		support.FailureMessageWithXMark("fetchJson failed: access to " + url + " is not allowed.")
		return interface{}(nil)
	}

	var result interface{} = nil
	utils.GetUrlJson(url, result)

	return result
}

func (net *ScriptNet) DownloadTo(url string, filename string) {
	if !net.engine.AppGateway.Allowed(url) {
		support.FailureMessageWithXMark(" [script] DownloadTo() failed: access to '" + url + "' is not allowed.")
		return
	}

	net.engine.AppGateway.SaveUrlToFile(url, filename)
}
