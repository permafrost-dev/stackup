package app

import (
	"fmt"
	"strings"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/support"
)

type EvaluateFunction func(script string) any

type JavaScriptEngine struct {
	Vm *otto.Otto
}

func CreateNewJavascriptEngine() *JavaScriptEngine {
	result := JavaScriptEngine{}
	result.Init()

	return &result
}

func (e *JavaScriptEngine) Init() {
	if e.Vm != nil {
		return
	}

	e.Vm = otto.New()

	CreateJavascriptFunctions(e.Vm)
	CreateScriptFsObject(e.Vm)
	CreateScriptAppObject(e.Vm)
	CreateScriptVarsObject(e.Vm)
}

func (e *JavaScriptEngine) ToValue(value otto.Value) any {

	if value.IsBoolean() {
		v, _ := value.ToBoolean()
		return v
	}

	if value.IsString() {
		v, _ := value.ToString()
		return v
	}

	if value.IsNumber() {
		v, _ := value.ToInteger()
		return v
	}

	if value.IsObject() {
		v, _ := value.Object().Value().Export()
		return v
	}

	if value.IsNull() {
		return nil
	}

	if value.IsUndefined() {
		return nil
	}

	if value.IsNaN() {
		return nil
	}

	r, _ := value.ToString()

	return r
}

func (e *JavaScriptEngine) Evaluate(script string) any {
	tempScript := strings.TrimSpace(script)

	if e.IsEvaluatableScriptString(tempScript) {
		tempScript = e.GetEvaluatableScriptString(tempScript)
	}

	result, err := e.Vm.Run(tempScript)

	if err != nil {
		support.WarningMessage(fmt.Sprintf("script error: %v\n", err))
		return nil
	}

	if result.IsBoolean() {
		v, _ := result.ToBoolean()
		return v
	}

	if result.IsString() {
		v, _ := result.ToString()

		if e.IsEvaluatableScriptString(v) {
			return e.Evaluate(v)
		}

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

func (e *JavaScriptEngine) MakeStringEvaluatable(script string) string {
	if e.IsEvaluatableScriptString(script) {
		return script
	}

	if len(script) == 0 {
		return ""
	}

	return "{{ " + script + " }}"
}
