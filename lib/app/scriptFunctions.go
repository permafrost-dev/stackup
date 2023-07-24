package app

import (
	"os"
	"runtime"

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
	vm.Set("hasEnv", createHasEnvFunction)
	vm.Set("hasFlag", createJavascriptFunctionHasFlag)
	vm.Set("platform", createPlatformFunction)
	vm.Set("script", createScriptFunction)
	vm.Set("selectTaskWhen", createSelectTaskWhen)
	vm.Set("task", createTaskFunction)
	vm.Set("workflow", createWorkflowFunction)
}

func getResult(call otto.FunctionCall, v any) otto.Value {
	result, _ := call.Otto.ToValue(v)

	return result
}

func createHasEnvFunction(call otto.FunctionCall) otto.Value {
	_, result := os.LookupEnv(call.Argument(0).String())

	return getResult(call, result)
}

func createWorkflowFunction(call otto.FunctionCall) otto.Value {
	result := App.Workflow

	return getResult(call, result)
}

func createPlatformFunction(call otto.FunctionCall) otto.Value {
	result := runtime.GOOS

	return getResult(call, result)
}

func createTaskFunction(call otto.FunctionCall) otto.Value {
	taskName := call.Argument(1).String()
	task := App.Workflow.FindTaskById(taskName)

	return getResult(call, task)
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
		task = App.Workflow.FindTaskById(trueTaskName)
	} else {
		task = App.Workflow.FindTaskById(falseTaskName)
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
