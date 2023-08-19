package types

import "os/exec"

type CommandCallback func(cmd *exec.Cmd)

type JavaScriptEngineContract interface {
	IsEvaluatableScriptString(s string) bool
	GetEvaluatableScriptString(s string) string
	MakeStringEvaluatable(script string) string
}
