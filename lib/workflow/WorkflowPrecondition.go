package workflow

import "github.com/stackup-app/stackup/lib/support"

type WorkflowPrecondition struct {
	Name       string `yaml:"name"`
	Check      string `yaml:"check"`
	OnFail     string `yaml:"on-fail"`
	FromRemote bool
	Attempts   int
	MaxRetries *int `yaml:"max-retries,omitempty"`
	Workflow   *StackupWorkflow
}

func (p *WorkflowPrecondition) Initialize(workflow *StackupWorkflow) {
	p.Workflow = workflow

	p.Attempts = 0
	if p.MaxRetries == nil {
		p.MaxRetries = new(int)
		*p.MaxRetries = 0
	}
}

func (p *WorkflowPrecondition) HandleOnFailure() bool {
	result := true

	if p.Workflow.JsEngine.IsEvaluatableScriptString(p.OnFail) {
		p.Workflow.JsEngine.Evaluate(p.OnFail)
	} else {
		task := p.Workflow.FindTaskById(p.OnFail)
		if task != nil {
			task.Run(true)
		}
	}

	return result
}

func (wp *WorkflowPrecondition) Run() bool {
	result := true

	if wp.Check != "" {
		if (wp.Attempts - 1) > *wp.MaxRetries {
			support.FailureMessageWithXMark(wp.Name)
			return false
		}

		wp.Attempts++

		result = wp.Workflow.JsEngine.Evaluate(wp.Check).(bool)

		if !result && len(wp.OnFail) > 0 {
			support.FailureMessageWithXMark(wp.Name)

			if wp.HandleOnFailure() {
				return wp.Run()
			}

			return false
		}

		if !result {
			support.FailureMessageWithXMark(wp.Name)
			return false
		}
	}

	return result
}
