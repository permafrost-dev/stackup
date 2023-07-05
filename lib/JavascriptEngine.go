package lib

import (
	"fmt"

	"github.com/robertkrimen/otto"
)

type JavaScriptEngine struct {
	Vm *otto.Otto
}

func (e *JavaScriptEngine) Init() {
	if e.Vm != nil {
		return
	}

	e.Vm = otto.New()

	// Define a JavaScript function in Go
	e.Vm.Set("sayHello", func(call otto.FunctionCall) otto.Value {
		fmt.Printf("Hello, %s.\n", call.Argument(0).String())
		return otto.Value{}
	})

	// // Call the JavaScript function from Go
	// vm.Run(`
	// 	sayHello("World");
	// `)
}

func (e *JavaScriptEngine) Evaluate(script string) otto.Value {
	result, _ := e.Vm.Run(script)
	return result
}
