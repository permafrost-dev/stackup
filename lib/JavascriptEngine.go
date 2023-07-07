package lib

import (
	"strings"

	"github.com/robertkrimen/otto"
	jsengine "github.com/stackup-app/stackup/lib/javascriptEngine"
)

func CreateNewJavascriptEngine() JavaScriptEngine {
	result := JavaScriptEngine{}
	result.Init()

	return result
}

type JavaScriptEngine struct {
	Functions jsengine.JavascriptFunctions
	Vm        *otto.Otto
}

func (e *JavaScriptEngine) Init() {
	if e.Vm != nil {
		return
	}

	e.Vm = otto.New()
	e.Functions = jsengine.NewJavascriptFunctions(e.Vm)
}

func (e *JavaScriptEngine) Evaluate(script string) any {
	// getResult := func(v otto.Value) any {
	// 	switch strings.ToLower(v.Class()) {
	// 	case "string":
	// 		result, _ := v.ToString()
	// 		return result
	// 	case "boolean":
	// 		result, _ := v.ToBoolean()
	// 		return result
	// 	case "number":
	// 		result, _ := v.ToFloat()
	// 		return result
	// 	case "object":
	// 		result, _ := v.Export()
	// 		return result
	// 	case "undefined":
	// 		return nil
	// 	default:
	// 		return nil
	// 	}
	// }

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
	return strings.HasPrefix(s, "{{") && strings.HasSuffix(s, "}}")
}
