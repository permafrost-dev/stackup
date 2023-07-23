package scripting

import (
	"os"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/utils"
)

func CreateJavascriptFunctions(vm *otto.Otto) {
	vm.Set("binaryExists", createBinaryExists)
	vm.Set("env", createJavascriptFunctionEnv)
	vm.Set("exec", createJavascriptFunctionExec)
	vm.Set("exists", createJavascriptFunctionExists)
	vm.Set("getCwd", createGetCurrentWorkingDirectory)
	vm.Set("hasFlag", createJavascriptFunctionHasFlag)
}

func getResult(call otto.FunctionCall, v any) otto.Value {
	result, _ := call.Otto.ToValue(v)

	return result
}

// func createSelectTaskCommandWhen(call otto.FunctionCall) otto.Value {
//     result := false
//     taskName := call.Argument(0).String()
//     task := call.Argument(1).Object()

//     if taskName == task.Get("name").String() {
//         result = true
//     }

//     return getResult(call, result)
// }

func createGetCurrentWorkingDirectory(call otto.FunctionCall) otto.Value {
	result, _ := os.Getwd()

	return getResult(call, result)
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
	result := utils.RunCommandInPath(call.Argument(0).String(), ".", false)

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
