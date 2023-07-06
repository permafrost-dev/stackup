package lib

import (
	"strings"

	"github.com/robertkrimen/otto"
	jsengine "github.com/stackup-app/stackup/lib/javascriptEngine"
)

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
	getResult := func(v otto.Value) any {
		switch strings.ToLower(v.Class()) {
		case "string":
			result, _ := v.ToString()
			return result
		case "boolean":
			result, _ := v.ToBoolean()
			return result
		case "number":
			result, _ := v.ToFloat()
			return result
		case "object":
			result, _ := v.Export()
			return result
		case "undefined":
			return nil
		default:
			return nil
		}
	}

	result, _ := e.Vm.Run(script)

	return getResult(result)
}
