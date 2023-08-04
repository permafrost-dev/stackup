package app

import (
	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
)

type ScriptNet struct {
}

func CreateScriptNetObject(vm *otto.Otto) {
	obj := &ScriptNet{}
	vm.Set("net", obj)
}

func (net *ScriptNet) Fetch(url string) any {
	if !App.Gatekeeper.CanAccessUrl(url) {
		support.FailureMessageWithXMark("fetch failed: access to " + url + " is not allowed.")
		return ""
	}

	result, _ := utils.GetUrlContents(url)

	return result
}

func (net *ScriptNet) FetchJson(url string) any {
	if !App.Gatekeeper.CanAccessUrl(url) {
		support.FailureMessageWithXMark("fetchJson failed: access to " + url + " is not allowed.")
		return interface{}(nil)
	}

	result, _ := utils.GetUrlJson(url)

	return result
}

func (net *ScriptNet) DownloadTo(url string, filename string) {
	if !App.Gatekeeper.CanAccessUrl(url) {
		support.FailureMessageWithXMark("download failed: access to " + url + " is not allowed.")
		return
	}

	utils.SaveUrlToFile(url, filename)
}
