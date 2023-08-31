package app

import (
	"reflect"

	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/support"
)

type WorkflowPrecondition struct {
	Name       string `yaml:"name"`
	Check      string `yaml:"check"`
	OnFail     string `yaml:"on-fail,omitempty"`
	MaxRetries int    `yaml:"max-retries,omitempty"`
	FromRemote bool
	Attempts   int
	JsEngine   *scripting.JavaScriptEngine
	Workflow   *StackupWorkflow
}

func (p *WorkflowPrecondition) Initialize(workflow *StackupWorkflow) {
	p.Workflow = workflow
	p.JsEngine = workflow.JsEngine
	p.Attempts = 0
	p.MaxRetries = consts.MAX_TASK_RUNS
}

func (p *WorkflowPrecondition) HandleOnFailure() bool {
	if p.JsEngine.IsEvaluatableScriptString(p.OnFail) {
		return p.JsEngine.Evaluate(p.OnFail).(bool)
	}

	if task, found := p.Workflow.GetTaskById(p.OnFail); found {
		return task.RunSync()
	}

	return true
}

func (wp *WorkflowPrecondition) CanRun() bool {
	return wp.Attempts < wp.MaxRetries
}

func (wp *WorkflowPrecondition) Run() bool {
	if wp.Check == "" {
		return true
	}

	result := wp.CanRun()
	if !result {
		support.FailureMessageWithXMark(wp.Name)
		return result
	}

	wp.Attempts++

	scriptResult := wp.JsEngine.Evaluate(wp.Check).(any)
	resultType, resultValue, _ := wp.JsEngine.ResultType(scriptResult)

	if resultType == reflect.String && resultValue != "" {
		return wp.JsEngine.Evaluate(resultValue.(string)).(bool)
	}

	if resultType == reflect.Bool && resultValue == false {
		if wp.handleOnFail() {
			return result
		}
		support.FailureMessageWithXMark(wp.Name)
	}

	// if result.(bool) == false {
	// 	if result = wp.handleOnFail(); result {
	// 		return result
	// 	}
	// 	support.FailureMessageWithXMark(wp.Name)
	// }
	return result
}

func (wp *WorkflowPrecondition) handleOnFail() bool {
	if len(wp.OnFail) == 0 {
		return false
	}

	support.FailureMessageWithXMark(wp.Name)

	if wp.JsEngine.IsEvaluatableScriptString(wp.OnFail) {
		return wp.JsEngine.Evaluate(wp.OnFail).(bool)
	}

	if task, found := wp.Workflow.GetTaskById(wp.OnFail); found {
		return task.RunSync()
	}

	if wp.HandleOnFailure() {
		return wp.Run()
	}

	return false
}
