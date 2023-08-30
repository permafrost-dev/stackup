package netextension

import (
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type ScriptNet struct {
	gateway types.GatewayContract
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

func (net *ScriptNet) gatewayPtr() *types.GatewayContract {
	return &net.gateway
}

func (net *ScriptNet) Fetch(url string) any {
	// if !net.gateway.Allowed(url) {
	// 	support.FailureMessageWithXMark(" [script] fetch failed: access to " + url + " is not allowed.")
	// 	return ""
	// }

	// if not allowed by the gateway, an error message will be printed, see Gateway class
	result, err := utils.GetUrlContents(url, net.gatewayPtr())

	if err != nil {
		return ""
	}

	return result
}

func (net *ScriptNet) FetchJson(url string) any {
	var result interface{} = nil

	if !net.gateway.Allowed(url) {
		support.FailureMessageWithXMark(" [script] fetchJson failed: access to " + url + " is not allowed.")
		return result
	}

	utils.GetUrlJson(url, result, net.gatewayPtr())

	return result
}

func (net *ScriptNet) DownloadTo(url string, filename string) {
	if !net.gateway.Allowed(url) {
		support.FailureMessageWithXMark(" [script] DownloadTo() failed: access to '" + url + "' is not allowed.")
		return
	}

	net.gateway.SaveUrlToFile(url, filename)
}
