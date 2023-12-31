package scripting

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/robertkrimen/otto"
	appextension "github.com/stackup-app/stackup/lib/scripting/extensions/app_extension"
	devextension "github.com/stackup-app/stackup/lib/scripting/extensions/dev_extension"
	fsextension "github.com/stackup-app/stackup/lib/scripting/extensions/fs_extension"
	functionsextension "github.com/stackup-app/stackup/lib/scripting/extensions/functions_extension"
	netextension "github.com/stackup-app/stackup/lib/scripting/extensions/net_extension"
	varsextension "github.com/stackup-app/stackup/lib/scripting/extensions/vars_extension"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type FindTaskByIdFunc func(string) (any, bool)

type JavaScriptEngine struct {
	Vm                  *otto.Otto
	InstalledExtensions *sync.Map
	initialized         bool
	AppIntf             types.AppInterface
	types.JavaScriptEngineContract
}

func CreateNewJavascriptEngine(appIntf types.AppInterface) *JavaScriptEngine {
	result := &JavaScriptEngine{
		initialized:         false,
		Vm:                  otto.New(),
		AppIntf:             appIntf,
		InstalledExtensions: &sync.Map{},
	}

	return result
}

func (e *JavaScriptEngine) GetApplicationIconPath() string {
	return e.App().GetApplicationIconPath()
}

func (e *JavaScriptEngine) GetAppVars() *sync.Map {
	return e.App().GetVars()
}

func (e *JavaScriptEngine) toInterface() interface{} {
	return e
}

func (e *JavaScriptEngine) App() types.AppInterface {
	return e.AppIntf
}

func (e *JavaScriptEngine) AsContract() types.JavaScriptEngineContract {
	return e.toInterface().(types.JavaScriptEngineContract)
}

func (e *JavaScriptEngine) Initialize() {
	if e.initialized {
		return
	}

	e.Vm = otto.New()

	e.initializeExtensions()

	// CreateScripNotificationsObject(workflow, e)
	e.initialized = true

	e.CreateAppVariables(e.App().GetVars())
	e.CreateEnvironmentVariables(e.App().GetEnviron())
}

func (e JavaScriptEngine) GetFindTaskById(id string) (any, bool) { //GetFindTaskById(id string) (any, error) {
	return e.App().GetWorkflow().FindTaskById(id)
}

func (e *JavaScriptEngine) initializeExtensions() {
	engine := e.App().GetJsEngine()

	devextension.Create().OnInstall(engine)
	varsextension.Create(engine).OnInstall(engine)
	netextension.Create(e.App().GetGateway()).OnInstall(engine)
	appextension.Create().OnInstall(engine)
	fsextension.Create().OnInstall(engine)
	functionsextension.Create(engine).OnInstall(engine)
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

func (e *JavaScriptEngine) ResultType(v any) (reflect.Kind, interface{}, error) {
	var value any = reflect.ValueOf(v).Interface()
	valueOf := reflect.ValueOf(v)
	kind := reflect.TypeOf(v).Kind()

	if kind == reflect.String {
		value = valueOf.String()
	} else if kind == reflect.Int {
		value = valueOf.Int()
	} else if kind == reflect.Bool {
		value = valueOf.Bool()
	} else if kind == reflect.Uint {
		value = valueOf.Uint()
	}

	var err error = nil

	if kind == reflect.Invalid {
		err = fmt.Errorf("invalid type")
	}

	return kind, value, err
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

	if strings.TrimSpace(script) == "" {
		return ""
	}

	return "{{ " + script + " }}"
}

func (e *JavaScriptEngine) GetVm() *otto.Otto {
	return e.Vm
}
