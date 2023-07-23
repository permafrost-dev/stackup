package app

import (
	"strings"

	"github.com/robertkrimen/otto"
)

type EvaluateFunction func(script string) any

type JavaScriptEngine struct {
	Vm *otto.Otto
}

func CreateNewJavascriptEngine() JavaScriptEngine {
	result := JavaScriptEngine{}
	result.Init()

	return result
}

func (e *JavaScriptEngine) Init() {
	if e.Vm != nil {
		return
	}

	e.Vm = otto.New()
	CreateJavascriptFunctions(e.Vm)
}

func (e *JavaScriptEngine) Evaluate(script string) any {
	if e.IsEvaluatableScriptString(script) {
		script = e.GetEvaluatableScriptString(script)
	}

	result, _ := e.Vm.Run(script)

	if result.IsBoolean() {
		v, _ := result.ToBoolean()
		return v
	}

	if result.IsString() {
		v, _ := result.ToString()
		return v
	}

	if result.IsNumber() {
		v, _ := result.ToInteger()
		return v
	}

	if result.IsObject() {
		v, _ := result.Object().Value().Export()
		return v
	}

	if result.IsNull() {
		return nil
	}

	if result.IsUndefined() {
		return nil
	}

	if result.IsNaN() {
		return nil
	}

	r, _ := result.ToString()

	return r
}

func (e *JavaScriptEngine) GetEvaluatableScriptString(s string) string {
	if e.IsEvaluatableScriptString(s) {
		return s[2 : len(s)-2]
	}
	return s
}

func (e *JavaScriptEngine) IsEvaluatableScriptString(s string) bool {
	temp := strings.TrimSpace(s)

	return strings.HasPrefix(temp, "{{") && strings.HasSuffix(temp, "}}")
}
