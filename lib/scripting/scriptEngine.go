package scripting

import (
	"fmt"
	"strings"
	"sync"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type FindTaskByIdFunc func(string) (any, error)

type JavaScriptEngine struct {
	Vm                     *otto.Otto
	AppVars                *sync.Map
	AppGateway             *gateway.Gateway
	Functions              *JavaScriptFunctions
	GetApplicationIconPath func() string
	FindTaskById           FindTaskByIdFunc
	Registry               *sync.Map
	InstalledExtensions    *sync.Map
	initialized            bool

	types.JavaScriptEngineContract
}

func CreateNewJavascriptEngine(vars *sync.Map, gateway *gateway.Gateway, findTaskFunc FindTaskByIdFunc, getAppIconFunc func() string) *JavaScriptEngine {
	result := &JavaScriptEngine{
		initialized:            false,
		Vm:                     otto.New(),
		AppVars:                vars,
		AppGateway:             gateway,
		GetApplicationIconPath: getAppIconFunc,
		Registry:               &sync.Map{},
		InstalledExtensions:    &sync.Map{},
		FindTaskById:           findTaskFunc,
	}

	return result
}

func (e *JavaScriptEngine) GetGateway() *types.GatewayContract {
	var result types.GatewayContract = e.AppGateway
	return &result
}

func (e *JavaScriptEngine) RegisterExtension(name string, value types.ScriptExtensionContract) {
	e.Registry.Store(name, value)
}

func (e *JavaScriptEngine) IsExtensionInstalled(name string) bool {
	_, ok := e.InstalledExtensions.Load(name)
	return ok
}

func (e *JavaScriptEngine) InstallRegistry() {
	e.Registry.Range(func(key, value any) bool {
		k := key.(string)
		v := value.(types.ScriptExtensionContract)
		if !e.IsExtensionInstalled(k) {
			e.InstalledExtensions.Store(k, v)
			v.Install()
		}
		return true
	})
}

func (e *JavaScriptEngine) toInterface() interface{} {
	return e
}

func (e *JavaScriptEngine) AsContract() types.JavaScriptEngineContract {
	return e.toInterface().(types.JavaScriptEngineContract)
}

func (e *JavaScriptEngine) Initialize(appVars *sync.Map, environ []string) {
	if !e.initialized {
		e.Vm = otto.New()
		e.CreateJavascriptFunctions()
		e.CreateScriptFsObject()
		CreateScriptAppObject(e)
		CreateScriptVarsObject(e)
		CreateScriptDevObject(e)
		CreateScriptNetObject(e)
		// CreateScripNotificationsObject(workflow, e)
		e.initialized = true
	}

	e.CreateAppVariables(appVars)
	e.CreateEnvironmentVariables(environ)
}

func (e *JavaScriptEngine) CreateAppVariables(vars *sync.Map) {
	vars.Range(func(key, value any) bool {
		e.Vm.Set("$"+(key.(string)), value)
		return true
	})
}

func (e *JavaScriptEngine) CreateEnvironmentVariables(vars []string) {
	for _, env := range vars {
		parts := strings.SplitN(env, "=", 2)
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
