package scripting

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/semver"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
)

type JavaScriptFunctions struct {
	Engine *JavaScriptEngine
}

func (e *JavaScriptEngine) CreateJavascriptFunctions() {
	jsf := JavaScriptFunctions{Engine: e}
	jsf.Register()
}

func (jsf *JavaScriptFunctions) Register() {
	jsf.Engine.Vm.Set("binaryExists", jsf.createBinaryExists)
	jsf.Engine.Vm.Set("composerJson", jsf.createComposerJsonFunction)
	jsf.Engine.Vm.Set("env", jsf.createJavascriptFunctionEnv)
	jsf.Engine.Vm.Set("exec", jsf.createJavascriptFunctionExec)
	jsf.Engine.Vm.Set("exists", jsf.createJavascriptFunctionExists)
	jsf.Engine.Vm.Set("fetch", jsf.createFetchFunction)
	jsf.Engine.Vm.Set("fetchJson", jsf.createFetchJsonFunction)
	jsf.Engine.Vm.Set("fileContains", jsf.createFileContainsFunction)
	jsf.Engine.Vm.Set("getCwd", jsf.createGetCurrentWorkingDirectory)
	jsf.Engine.Vm.Set("getVar", jsf.createGetVarFunction)
	jsf.Engine.Vm.Set("hasEnv", jsf.createHasEnvFunction)
	jsf.Engine.Vm.Set("hasFlag", jsf.createJavascriptFunctionHasFlag)
	jsf.Engine.Vm.Set("hasVar", jsf.createHasVarFunction)
	jsf.Engine.Vm.Set("outputOf", jsf.createOutputOfFunction)
	jsf.Engine.Vm.Set("packageJson", jsf.createPackageJsonFunction)
	jsf.Engine.Vm.Set("platform", jsf.createPlatformFunction)
	jsf.Engine.Vm.Set("requirementsTxt", jsf.createRequirementsTxtFunction)
	jsf.Engine.Vm.Set("script", jsf.createScriptFunction)
	jsf.Engine.Vm.Set("selectTaskWhen", jsf.createSelectTaskWhen)
	jsf.Engine.Vm.Set("semver", jsf.createSemverFunction)
	jsf.Engine.Vm.Set("setVar", jsf.createSetVarFunction)
	jsf.Engine.Vm.Set("setTimeout", jsf.createSetTimeoutFunction)
	jsf.Engine.Vm.Set("statusMessage", jsf.createStatusMessageFunction)
	jsf.Engine.Vm.Set("task", jsf.createTaskFunction)
}

func getResult(call otto.FunctionCall, v any) otto.Value {
	result, _ := call.Otto.ToValue(v)

	return result
}

func (jsf *JavaScriptFunctions) createSetTimeoutFunction(call otto.FunctionCall) otto.Value {
	// Get the callback function and delay time from the arguments
	callback := call.Argument(0)
	delay, _ := call.Argument(1).ToInteger()

	// Create a channel to wait for the delay time
	done := make(chan bool, 1)
	go func() {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		done <- true
	}()

	// Call the callback function after the delay time
	go func() {
		<-done
		callback.Call(callback)
	}()

	value, _ := otto.ToValue(nil)
	return value
}

func (jsf *JavaScriptFunctions) createFetchFunction(call otto.FunctionCall) otto.Value {
	result, _ := utils.GetUrlContents(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createFetchJsonFunction(call otto.FunctionCall) otto.Value {
	result, _ := utils.GetUrlJson(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createRequirementsTxtFunction(call otto.FunctionCall) otto.Value {
	result, _ := LoadRequirementsTxt(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createPackageJsonFunction(call otto.FunctionCall) otto.Value {
	result, _ := LoadPackageJson(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createComposerJsonFunction(call otto.FunctionCall) otto.Value {
	result, _ := LoadComposerJson(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createSemverFunction(call otto.FunctionCall) otto.Value {
	result := semver.ParseSemverString(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createOutputOfFunction(call otto.FunctionCall) otto.Value {
	result := support.GetCommandOutput(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createFileContainsFunction(call otto.FunctionCall) otto.Value {
	result := support.SearchFileForString(call.Argument(0).String(), call.Argument(1).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createStatusMessageFunction(call otto.FunctionCall) otto.Value {
	support.StatusMessage(call.Argument(0).String(), false)

	return getResult(call, true)
}

func (jsf *JavaScriptFunctions) createHasVarFunction(call otto.FunctionCall) otto.Value {
	_, result := jsf.Engine.AppVars.Load(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createGetVarFunction(call otto.FunctionCall) otto.Value {
	v, _ := jsf.Engine.AppVars.Load(call.Argument(0).String())

	return v.(otto.Value)
}

func (jsf *JavaScriptFunctions) createSetVarFunction(call otto.FunctionCall) otto.Value {
	jsf.Engine.AppVars.Store(call.Argument(0).String(), call.Argument(1))

	return getResult(call, true)
}

func (jsf *JavaScriptFunctions) createHasEnvFunction(call otto.FunctionCall) otto.Value {
	_, result := os.LookupEnv(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createPlatformFunction(call otto.FunctionCall) otto.Value {
	result := runtime.GOOS

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createTaskFunction(call otto.FunctionCall) otto.Value {
	taskName := call.Argument(0).String()

	if strings.HasPrefix(taskName, "$") && len(taskName) > 1 {
		temp, _ := jsf.Engine.AppVars.Load(taskName[1:])
		taskName = temp.(string)
	}
	task, _ := (*jsf.Engine.GetWorkflowContract).FindTaskById(taskName)

	return getResult(call, task)
}

func (jsf *JavaScriptFunctions) createScriptFunction(call otto.FunctionCall) otto.Value {
	filename := call.Argument(0).String()

	content, err := os.ReadFile(filename)

	if err != nil {
		support.WarningMessage("Could not read script file: " + filename)
		return getResult(call, false)
	}

	result := jsf.Engine.Evaluate(string(content))

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createSelectTaskWhen(call otto.FunctionCall) otto.Value {
	conditional, _ := call.Argument(0).ToBoolean()
	trueTaskName := call.Argument(1).String()
	falseTaskName := call.Argument(2).String()
	//var task *types.AppWorkflowTaskContract
	//var wf *types.AppWorkflowContract = jsf.Engine.GetWorkflowContract
	temp := (*jsf.Engine.GetWorkflowContract)

	var t any

	if conditional {
		t, _ = temp.FindTaskById(trueTaskName)
	} else {
		t, _ = temp.FindTaskById(falseTaskName)
	}

	return getResult(call, t)
}

func (jsf *JavaScriptFunctions) createGetCurrentWorkingDirectory(call otto.FunctionCall) otto.Value {
	result, _ := os.Getwd()

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createBinaryExists(call otto.FunctionCall) otto.Value {
	result := utils.BinaryExistsInPath(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createJavascriptFunctionExists(call otto.FunctionCall) otto.Value {
	_, err := os.Stat(call.Argument(0).String())
	result := !os.IsNotExist(err)

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createJavascriptFunctionEnv(call otto.FunctionCall) otto.Value {
	result := os.Getenv(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createJavascriptFunctionExec(call otto.FunctionCall) otto.Value {
	result, err := utils.RunCommandInPath(call.Argument(0).String(), ".", false)

	if err != nil {
		support.WarningMessage(err.Error())
	}

	finalResult, err := call.Otto.ToValue(result)

	if err != nil {
		support.WarningMessage(err.Error())
	}

	return finalResult
}

func (jsf *JavaScriptFunctions) createJavascriptFunctionHasFlag(call otto.FunctionCall) otto.Value {
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
