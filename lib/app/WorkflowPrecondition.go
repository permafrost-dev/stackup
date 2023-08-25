package app

import (
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

func (p *WorkflowPrecondition) Initialize(workflow *StackupWorkflow, engine *scripting.JavaScriptEngine) {
	p.Workflow = workflow
	p.JsEngine = engine
	p.Attempts = 0
	p.MaxRetries = 99999999999
}

func (p *WorkflowPrecondition) HandleOnFailure() bool {
	result := true

	if p.OnFail == "" {
		return result
	}

	if p.JsEngine.IsEvaluatableScriptString(p.OnFail) {
		return p.JsEngine.Evaluate(p.OnFail).(bool)
	}

	task, found := (*p.Workflow).FindTaskById(p.OnFail)

	if found {
		task.RunSync()
	}

	return result
}

func (wp *WorkflowPrecondition) Run() bool {
	result := true

	if wp.Check != "" {
		if wp.Attempts >= wp.MaxRetries {
			support.FailureMessageWithXMark(wp.Name)
			return false
		}

		wp.Attempts++

		result = wp.JsEngine.Evaluate(wp.Check).(bool)

		if !result && len(wp.OnFail) > 0 {
			support.FailureMessageWithXMark(wp.Name)

			if wp.HandleOnFailure() {
				return wp.Run()
			}

			result = false
		}

		if !result {
			support.FailureMessageWithXMark(wp.Name)
			return false
		}
	}

	return result
}
