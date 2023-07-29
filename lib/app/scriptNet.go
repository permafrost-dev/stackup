package app

import (
	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/utils"
)

type ScriptNet struct {
}

func CreateScriptNetObject(vm *otto.Otto) {
	obj := &ScriptNet{}
	vm.Set("net", obj)
}

func (net *ScriptNet) Fetch(url string) any {
	result, _ := utils.GetUrlContents(url)

	return result
}

func (net *ScriptNet) FetchJson(url string) any {
	result, _ := utils.GetUrlJson(url)

	return result
}

func (net *ScriptNet) DownloadTo(url string, filename string) {
	utils.SaveUrlToFile(url, filename)
}
