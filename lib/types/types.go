package types

import (
	"os/exec"
	"sync"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/settings"
)

type CommandCallback func(cmd *exec.Cmd)

type SetProcessCallback func(key, value any)

type JavaScriptEngineContract interface {
	IsEvaluatableScriptString(s string) bool
	GetEvaluatableScriptString(s string) string
	MakeStringEvaluatable(script string) string
	Evaluate(script string) any
	CreateAppVariables(vars *sync.Map)
	GetVm() *otto.Otto
	GetGateway() GatewayContract
	GetAppVars() *sync.Map
	GetFindTaskById(id string) (any, bool)
}

type AppWorkflowTaskContract interface {
	// CanRunOnCurrentPlatform() bool
	// CanRunConditionally() bool
	Initialize()
	Run(synchronous bool)
}

type AppWorkflowContract interface {
	FindTaskById(id string) (any, bool)
	GetSettings() *settings.Settings
	GetJsEngine() *JavaScriptEngineContract
}

type AppWorkflowContractPtr *AppWorkflowContract

type GatewayContract interface {
	Allowed(url string) bool
	SaveUrlToFile(url string, filename string) error
	GetUrl(url string, headers ...string) (string, error)
}

type ScriptExtensionContract interface {
	OnInstall(engine JavaScriptEngineContract)
	GetName() string
}

type ExtensionInfo struct {
	Name        string
	Version     string
	Description string
	createFn    interface{} // CreateExtensionFunc
}

type ExtensionInfoMap map[string]*ExtensionInfo

type CreateExtensionFunc func(info *ExtensionInfo) any //ScriptExtensionContract

type ScriptExtension struct {
	Name string
}

func CreateNewExtension(name string) *ScriptExtension {
	return &ScriptExtension{
		Name: name,
	}
}

func NewExtensionInfo(name string, version string, description string, createFn interface{}) *ExtensionInfo {
	return &ExtensionInfo{
		Name:        name,
		Version:     version,
		Description: description,
		createFn:    createFn,
	}
}

func (se *ScriptExtension) IsExtensionInstalled(name string) bool {
	return false
}

func (se *ScriptExtension) Install() {
	// do nothing
}

func (se *ScriptExtension) GetName() string {
	return se.Name
}

func (ei *ExtensionInfo) CreateExtension() ScriptExtensionContract {
	result := ei.createFn.(func(info *ExtensionInfo) any)(ei)
	return result.(ScriptExtensionContract)

}

func ExtensionInfoArrayToMap(array ...ExtensionInfo) map[string]*ExtensionInfo {
	result := map[string]*ExtensionInfo{}

	for _, item := range array {
		result[item.Name] = &item
	}

	return result
}
