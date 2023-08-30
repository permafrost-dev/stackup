package netextension

import (
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type ScriptNet struct {
	gateway types.GatewayContract
	// engine types.JavaScriptEngineContract
	// types.ScriptExtensionContract
}

func Create(gw types.GatewayContract) *ScriptNet {
	return &ScriptNet{
		gateway: gw,
	}
}

func (net *ScriptNet) GetName() string {
	return "net"
}

func (ex *ScriptNet) OnInstall(engine types.JavaScriptEngineContract) {
	engine.GetVm().Set(ex.GetName(), ex)
}

func (net *ScriptNet) Fetch(url string) any {
	if !net.gateway.Allowed(url) {
		support.FailureMessageWithXMark("fetch failed: access to " + url + " is not allowed.")
		return ""
	}

	gw := net.gateway
	result, _ := utils.GetUrlContents(url, &gw)

	return result
}

func (net *ScriptNet) FetchJson(url string) any {
	if !net.gateway.Allowed(url) {
		support.FailureMessageWithXMark("fetchJson failed: access to " + url + " is not allowed.")
		return interface{}(nil)
	}

	var result interface{} = nil
	gw := net.gateway
	utils.GetUrlJson(url, result, &gw)

	return result
}

func (net *ScriptNet) DownloadTo(url string, filename string) {
	if !net.gateway.Allowed(url) {
		support.FailureMessageWithXMark(" [script] DownloadTo() failed: access to '" + url + "' is not allowed.")
		return
	}

	net.gateway.SaveUrlToFile(url, filename)
}
