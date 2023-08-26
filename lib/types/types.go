package types

import (
	"os/exec"
	"sync"

	"github.com/stackup-app/stackup/lib/settings"
)

type CommandCallback func(cmd *exec.Cmd)

type JavaScriptEngineContract interface {
	IsEvaluatableScriptString(s string) bool
	GetEvaluatableScriptString(s string) string
	MakeStringEvaluatable(script string) string
	Evaluate(script string) any
	CreateAppVariables(vars *sync.Map)
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
	GetUrl(url string, headers ...string) (string, error)
}

type ScriptExtensionContract interface {
	Install()
	GetName() string
}

type ScriptExtension struct {
	Name string
}

func CreateNewExtension(name string) *ScriptExtension {
	return &ScriptExtension{
		Name: name,
	}
}

func (se *ScriptExtension) IsExtensionInstalled(name string) bool {
	return false
}
