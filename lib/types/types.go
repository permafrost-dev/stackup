package types

import (
	"os/exec"

	"github.com/stackup-app/stackup/lib/settings"
)

type CommandCallback func(cmd *exec.Cmd)

type JavaScriptEngineContract interface {
	IsEvaluatableScriptString(s string) bool
	GetEvaluatableScriptString(s string) string
	MakeStringEvaluatable(script string) string
	Evaluate(script string) any
}

type AppWorkflowTaskContract interface {
	// CanRunOnCurrentPlatform() bool
	// CanRunConditionally() bool
	Initialize()
	Run(synchronous bool)
}

type AppWorkflowContract interface {
	FindTaskById(id string) (*AppWorkflowTaskContract, bool)
	GetSettings() *settings.Settings
	GetJsEngine() *JavaScriptEngineContract
}
