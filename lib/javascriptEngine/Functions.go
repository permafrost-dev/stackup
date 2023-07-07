package jsengine

import (
	"fmt"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/utils"
)

type JavascriptFunctions struct {
	Vm          *otto.Otto
	Initialized bool
}

func NewJavascriptFunctions(vm *otto.Otto) JavascriptFunctions {
	result := JavascriptFunctions{Vm: vm, Initialized: false}
	result.Init()

	return result
}

func (jf *JavascriptFunctions) Init() {
	jf.Vm.Set("exists", CreateJavascriptFunctionExists)
	jf.Vm.Set("env", CreateJavascriptFunctionEnv)
	jf.Vm.Set("exec", CreateJavascriptFunctionExec)
	jf.Vm.Set("hasFlag", CreateJavascriptFunctionHasFlag)

	fmt.Println("created javascript functions")

	jf.Initialized = true
}

func CreateJavascriptFunctionExists(call otto.FunctionCall) otto.Value {
	_, err := os.Stat(call.Argument(0).String())
	result := !os.IsNotExist(err)
	finalResult, _ := call.Otto.ToValue(result)

	return finalResult
}

func CreateJavascriptFunctionEnv(call otto.FunctionCall) otto.Value {
	result := os.Getenv(call.Argument(0).String())
	finalResult, _ := call.Otto.ToValue(result)

	return finalResult
}

func CreateJavascriptFunctionExec(call otto.FunctionCall) otto.Value {
	fmt.Println("executing command " + call.Argument(0).String())
	fmt.Println("exec: " + call.Argument(0).String())
	result := utils.RunCommandEx(call.Argument(0).String(), ".")

	finalResult, _ := call.Otto.ToValue(result.ProcessState.Success())

	return finalResult
}

func CreateJavascriptFunctionHasFlag(call otto.FunctionCall) otto.Value {
	result := false
	flag := call.Argument(0).String()

	for _, v := range os.Args[1:] {
		if v == flag {
			result = true
			break
		}
	}

	finalResult, _ := call.Otto.ToValue(result)

	return finalResult
}

func EvaluateScript(script string) (otto.Value, error) {
	vm := otto.New()

	// Define the "add" function
	vm.Set("add", func(call otto.FunctionCall) otto.Value {
		num1, _ := call.Argument(0).ToInteger()
		num2, _ := call.Argument(1).ToInteger()
		result, _ := vm.ToValue(num1 + num2)
		return result
	})

	// Define the "subtract" function
	vm.Set("subtract", func(call otto.FunctionCall) otto.Value {
		num1, _ := call.Argument(0).ToInteger()
		num2, _ := call.Argument(1).ToInteger()
		result, _ := vm.ToValue(num1 - num2)
		return result
	})

	vm.Set("env", func(call otto.FunctionCall) otto.Value {
		value, _ := call.Argument(0).ToString()
		tempResult := os.Getenv(value)
		result, _ := vm.ToValue(tempResult)

		return result
	})

	vm.Set("hasFlag", func(call otto.FunctionCall) otto.Value {
		value, _ := call.Argument(0).ToString()
		r, _ := flags.Parse(os.Args[:1])

		temp := false
		for _, v := range r {
			if v == value {
				temp = true
				break
			}
		}

		result, _ := vm.ToValue(temp)
		return result
	})

	vm.Set("exists", func(call otto.FunctionCall) otto.Value {
		value, _ := call.Argument(0).ToString()
		_, err := os.Stat(value)

		var temp bool

		if os.IsNotExist(err) {
			temp = false
		} else {
			temp = true
		}

		result, _ := vm.ToValue(temp)

		return result
	})

	result, err := vm.Run(script)
	if err != nil {
		return otto.Value{}, fmt.Errorf("Error evaluating script: %w", err)
	}

	return result, nil
}
