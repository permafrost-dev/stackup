package functionsextension

import (
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	devextension "github.com/stackup-app/stackup/lib/scripting/extensions/dev_extension"
	"github.com/stackup-app/stackup/lib/semver"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type JavaScriptFunctions struct {
	Engine types.JavaScriptEngineContract
	types.ScriptExtensionContract
}

func Create(engine types.JavaScriptEngineContract) *JavaScriptFunctions {
	return &JavaScriptFunctions{Engine: engine}
}

func (jsf JavaScriptFunctions) GetName() string {
	return "functions"
}

func (jsf JavaScriptFunctions) OnInstall(engine types.JavaScriptEngineContract) {
	jsf.Register()
}

func (jsf *JavaScriptFunctions) Register() {
	jsf.Engine.GetVm().Set("binaryExists", jsf.createBinaryExists)
	jsf.Engine.GetVm().Set("composerJson", jsf.createComposerJsonFunction)
	jsf.Engine.GetVm().Set("env", jsf.createJavascriptFunctionEnv)
	jsf.Engine.GetVm().Set("exec", jsf.createJavascriptFunctionExec)
	jsf.Engine.GetVm().Set("exists", jsf.createJavascriptFunctionExists)
	jsf.Engine.GetVm().Set("fetch", jsf.createFetchFunction)
	jsf.Engine.GetVm().Set("fetchJson", jsf.createFetchJsonFunction)
	jsf.Engine.GetVm().Set("fileContains", jsf.createFileContainsFunction)
	jsf.Engine.GetVm().Set("getCwd", jsf.createGetCurrentWorkingDirectory)
	jsf.Engine.GetVm().Set("getVar", jsf.createGetVarFunction)
	jsf.Engine.GetVm().Set("hasEnv", jsf.createHasEnvFunction)
	jsf.Engine.GetVm().Set("hasFlag", jsf.createJavascriptFunctionHasFlag)
	jsf.Engine.GetVm().Set("hasVar", jsf.createHasVarFunction)
	jsf.Engine.GetVm().Set("outputOf", jsf.createOutputOfFunction)
	jsf.Engine.GetVm().Set("packageJson", jsf.createPackageJsonFunction)
	jsf.Engine.GetVm().Set("platform", jsf.createPlatformFunction)
	jsf.Engine.GetVm().Set("requirementsTxt", jsf.createRequirementsTxtFunction)
	jsf.Engine.GetVm().Set("script", jsf.createScriptFunction)
	jsf.Engine.GetVm().Set("selectTaskWhen", jsf.createSelectTaskWhen)
	jsf.Engine.GetVm().Set("semver", jsf.createSemverFunction)
	jsf.Engine.GetVm().Set("setVar", jsf.createSetVarFunction)
	jsf.Engine.GetVm().Set("setTimeout", jsf.createSetTimeoutFunction)
	jsf.Engine.GetVm().Set("statusMessage", jsf.createStatusMessageFunction)
	jsf.Engine.GetVm().Set("task", jsf.createTaskFunction)
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
	result, _ := jsf.Engine.GetGateway().GetUrl(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createFetchJsonFunction(call otto.FunctionCall) otto.Value {
	var result interface{}
	gw := jsf.Engine.GetGateway()
	utils.GetUrlJson(call.Argument(0).String(), &result, &gw)

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createRequirementsTxtFunction(call otto.FunctionCall) otto.Value {
	result, _ := devextension.LoadRequirementsTxt(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createPackageJsonFunction(call otto.FunctionCall) otto.Value {
	result, _ := devextension.LoadPackageJson(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createComposerJsonFunction(call otto.FunctionCall) otto.Value {
	result, _ := devextension.LoadComposerJson(call.Argument(0).String())

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
	result := utils.SearchFileForString(call.Argument(0).String(), call.Argument(1).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createStatusMessageFunction(call otto.FunctionCall) otto.Value {
	support.StatusMessage(call.Argument(0).String(), false)

	return getResult(call, true)
}

func (jsf *JavaScriptFunctions) createHasVarFunction(call otto.FunctionCall) otto.Value {
	_, result := jsf.Engine.GetAppVars().Load(call.Argument(0).String())

	return getResult(call, result)
}

func (jsf *JavaScriptFunctions) createGetVarFunction(call otto.FunctionCall) otto.Value {
	v, _ := jsf.Engine.GetAppVars().Load(call.Argument(0).String())

	return v.(otto.Value)
}

func (jsf *JavaScriptFunctions) createSetVarFunction(call otto.FunctionCall) otto.Value {
	jsf.Engine.GetAppVars().Store(call.Argument(0).String(), call.Argument(1))

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
		temp, _ := jsf.Engine.GetAppVars().Load(taskName[1:])
		taskName = temp.(string)
	}

	task, _ := jsf.Engine.GetFindTaskById(taskName)
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
	taskName := trueTaskName

	if !conditional {
		taskName = falseTaskName
	}

	t, _ := jsf.Engine.GetFindTaskById(taskName)

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
