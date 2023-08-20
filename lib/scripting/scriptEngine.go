package scripting

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type JavaScriptEngine struct {
	Vm                     *otto.Otto
	AppVars                *sync.Map
	AppGateway             *gateway.Gateway
	Functions              *JavaScriptFunctions
	GetWorkflowContract    *func() *types.AppWorkflowContract
	GetApplicationIconPath func() string
	types.JavaScriptEngineContract
}

func CreateNewJavascriptEngine(vars *sync.Map, gateway *gateway.Gateway, getWorkflowContract func() *types.AppWorkflowContract, getAppIconFunc func() string) *JavaScriptEngine {
	result := JavaScriptEngine{
		Vm:                     otto.New(),
		AppVars:                vars,
		AppGateway:             gateway,
		GetApplicationIconPath: getAppIconFunc,
		GetWorkflowContract:    &getWorkflowContract,
	}

	result.Init(getWorkflowContract())

	return &result
}

func (e *JavaScriptEngine) toInterface() interface{} {
	return e
}

func (e *JavaScriptEngine) AsContract() types.JavaScriptEngineContract {
	return e.toInterface().(types.JavaScriptEngineContract)
}

func (e *JavaScriptEngine) AsContractPtr() *types.JavaScriptEngineContract {
	var result types.JavaScriptEngineContract = e.AsContract()
	return &result
}

func (e *JavaScriptEngine) Init(workflow *types.AppWorkflowContract) {
	fmt.Printf("Init()\n")

	e.Vm = otto.New()
	e.Functions = CreateJavascriptFunctions(e)
	e.Functions.Init()
	e.CreateEnvironmentVariables()
	CreateScriptFsObject(e)
	CreateScriptAppObject(e)
	CreateScriptVarsObject(e)
	CreateScriptDevObject(e)
	CreateScriptNetObject(e)
	CreateScripNotificationsObject(workflow, e)
}

func (e *JavaScriptEngine) CreateAppVariables() {
	e.AppVars.Range(func(key, value any) bool {
		e.Vm.Set("$"+(key.(string)), value)
		return true
	})
}

func (e *JavaScriptEngine) CreateEnvironmentVariables() {
	for _, env := range os.Environ() {
		parts := strings.Split(env, "=")
		e.Vm.Set("$"+parts[0], parts[1])
	}
}

func (e *JavaScriptEngine) ToValue(value otto.Value) any {
	if value.IsBoolean() {
		v, _ := value.ToBoolean()
		return v
	}

	if value.IsString() {
		v, _ := value.ToString()
		return v
	}

	if value.IsNumber() {
		v, _ := value.ToInteger()
		return v
	}

	if value.IsObject() {
		v, _ := value.Object().Value().Export()
		return v
	}

	if value.IsNull() {
		return nil
	}

	if value.IsUndefined() {
		return nil
	}

	if value.IsNaN() {
		return nil
	}

	r, _ := value.ToString()

	return r
}

func (e *JavaScriptEngine) Evaluate(script string) any {
	tempScript := strings.TrimSpace(script)

	if len(tempScript) == 0 {
		return nil
	}

	if e.IsEvaluatableScriptString(tempScript) {
		tempScript = e.GetEvaluatableScriptString(tempScript)
	}

	fmt.Printf("Vm: %v\n", e)

	result, err := e.Vm.Run(tempScript)

	if err != nil {
		support.WarningMessage(fmt.Sprintf("script error: %v\n", err))
		return nil
	}

	if result.IsObject() {
		v := result.Object()
		keys := result.Object().Keys()

		if utils.StringArrayContains(keys, "Id") && utils.StringArrayContains(keys, "Name") && utils.StringArrayContains(keys, "Command") {
			v2, _ := v.Value().Object().Get("Id")
			return v2.String()
		}
	}

	if result.IsBoolean() {
		v, _ := result.ToBoolean()
		return v
	}

	if result.IsString() {
		v, _ := result.ToString()

		if e.IsEvaluatableScriptString(v) {
			return e.Evaluate(v)
		}

		return v
	}

	if result.IsNumber() {
		v, _ := result.ToInteger()
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
	temp := strings.TrimSpace(s)

	return strings.HasPrefix(temp, "{{") && strings.HasSuffix(temp, "}}")
}

func (e *JavaScriptEngine) MakeStringEvaluatable(script string) string {
	if e.IsEvaluatableScriptString(script) {
		return script
	}

	if len(script) == 0 {
		return ""
	}

	return "{{ " + script + " }}"
}
