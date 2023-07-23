package app

import (
	"os"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
)

func CreateJavascriptFunctions(vm *otto.Otto) {
	vm.Set("binaryExists", createBinaryExists)
	vm.Set("env", createJavascriptFunctionEnv)
	vm.Set("exec", createJavascriptFunctionExec)
	vm.Set("exists", createJavascriptFunctionExists)
	vm.Set("getCwd", createGetCurrentWorkingDirectory)
	vm.Set("hasFlag", createJavascriptFunctionHasFlag)
	vm.Set("script", createScriptFunction)
	vm.Set("selectTaskWhen", createSelectTaskWhen)
}

func getResult(call otto.FunctionCall, v any) otto.Value {
	result, _ := call.Otto.ToValue(v)

	return result
}

func createScriptFunction(call otto.FunctionCall) otto.Value {
	filename := call.Argument(0).String()

	content, err := os.ReadFile(filename)

    if err != nil {
        support.WarningMessage("Could not read script file: " + filename)
        return getResult(call, false)
    }

	result := App.JsEngine.Evaluate(string(content))

	return getResult(call, result)
}

func createSelectTaskWhen(call otto.FunctionCall) otto.Value {
	conditional, _ := call.Argument(0).ToBoolean()
	trueTaskName := call.Argument(1).String()
	falseTaskName := call.Argument(2).String()
	var task *Task

	if conditional {
		task = App.workflow.FindTaskById(trueTaskName)
	} else {
		task = App.workflow.FindTaskById(falseTaskName)
	}

	return getResult(call, task)
}

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
