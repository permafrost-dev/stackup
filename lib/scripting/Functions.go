package scripting

import (
	"fmt"
	"os"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/utils"
)

func CreateJavascriptFunctions(vm *otto.Otto) {
	vm.Set("binaryExists", createBinaryExists)
	vm.Set("env", createJavascriptFunctionEnv)
	vm.Set("exec", createJavascriptFunctionExec)
	vm.Set("exists", createJavascriptFunctionExists)
	vm.Set("hasFlag", createJavascriptFunctionHasFlag)
}

func getResult(call otto.FunctionCall, v any) otto.Value {
	result, _ := call.Otto.ToValue(v)

	return result
}

func createBinaryExists(call otto.FunctionCall) otto.Value {
	result := utils.BinaryExistsInPath(call.Argument(0).String())

	return getResult(call, result)
}

func createJavascriptFunctionExists(call otto.FunctionCall) otto.Value {
	_, err := os.Stat(call.Argument(0).String())
	result := !os.IsNotExist(err)

	return getResult(call, result)
}

func createJavascriptFunctionEnv(call otto.FunctionCall) otto.Value {
	result := os.Getenv(call.Argument(0).String())

	return getResult(call, result)
}

func createJavascriptFunctionExec(call otto.FunctionCall) otto.Value {
	fmt.Println("executing command " + call.Argument(0).String())
	fmt.Println("exec: " + call.Argument(0).String())
	result := utils.RunCommandEx(call.Argument(0).String(), ".")

	finalResult, _ := call.Otto.ToValue(result.ProcessState.Success())

	return finalResult
}

func createJavascriptFunctionHasFlag(call otto.FunctionCall) otto.Value {
	result := false
	flag := call.Argument(0).String()

	// result, _ := vm.ToValue(temp)

	for _, v := range os.Args[1:] {
		if v == flag || v == "--"+flag {
			result = true
			break
		}
	}

	return getResult(call, result)
}

// func EvaluateScript(script string) (otto.Value, error) {
// 	vm := otto.New()

// 	// Define the "add" function
// 	vm.Set("add", func(call otto.FunctionCall) otto.Value {
// 		num1, _ := call.Argument(0).ToInteger()
// 		num2, _ := call.Argument(1).ToInteger()
// 		result, _ := vm.ToValue(num1 + num2)
// 		return result
// 	})

// 	// Define the "subtract" function
// 	vm.Set("subtract", func(call otto.FunctionCall) otto.Value {
// 		num1, _ := call.Argument(0).ToInteger()
// 		num2, _ := call.Argument(1).ToInteger()
// 		result, _ := vm.ToValue(num1 - num2)
// 		return result
// 	})

// 	vm.Set("env", func(call otto.FunctionCall) otto.Value {
// 		value, _ := call.Argument(0).ToString()
// 		tempResult := os.Getenv(value)
// 		result, _ := vm.ToValue(tempResult)

// 		return result
// 	})

// 	vm.Set("hasFlag", func(call otto.FunctionCall) otto.Value {
// 		value, _ := call.Argument(0).ToString()
// 		r, _ := flags.Parse(os.Args[:1])

// 		temp := false
// 		for _, v := range r {
// 			if v == value {
// 				temp = true
// 				break
// 			}
// 		}

// 		result, _ := vm.ToValue(temp)
// 		return result
// 	})

// 	vm.Set("exists", func(call otto.FunctionCall) otto.Value {
// 		value, _ := call.Argument(0).ToString()
// 		_, err := os.Stat(value)

// 		exists := !os.IsNotExist(err)
// 		result, _ := vm.ToValue(exists)

// 		return result
// 	})

// 	result, err := vm.Run(script)
// 	if err != nil {
// 		return otto.Value{}, fmt.Errorf("Error evaluating script: %w", err)
// 	}

// 	return result, nil
// }
