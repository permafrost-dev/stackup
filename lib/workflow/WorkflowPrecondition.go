package workflow

import (
	"fmt"

	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
)

type WorkflowPrecondition struct {
	Name       string `yaml:"name"`
	Check      string `yaml:"check"`
	OnFail     string `yaml:"on-fail,omitempty"`
	MaxRetries int    `yaml:"max-retries,omitempty"`
	FromRemote bool
	Attempts   int
	engine     *types.JavaScriptEngineContract
	Workflow   *StackupWorkflow
}

func (p *WorkflowPrecondition) Initialize(workflow *StackupWorkflow, engine *types.JavaScriptEngineContract) {
	p.Workflow = workflow
	p.engine = engine

	// var temp interface{} = workflow
	// p.Workflow = temp.(*StackupWorkflow)

	p.Attempts = 0
	p.MaxRetries = 99999999999
}

func (p *WorkflowPrecondition) HandleOnFailure() bool {
	result := true

	if p.OnFail == "" {
		return result
	}

	if (*p.engine).IsEvaluatableScriptString(p.OnFail) {
		return (*p.engine).Evaluate(p.OnFail).(bool)
	}

	task, found := (*p.Workflow).FindTaskById(p.OnFail)
	fmt.Printf("task: %v\n", task)
	if found {
		(*task).Run(true)
	}

	return result
}

func (wp *WorkflowPrecondition) Run() bool {
	result := true

	fmt.Printf("wp: %v\n", wp)

	if wp.Check != "" {
		if (wp.Attempts - 1) > wp.MaxRetries {
			support.FailureMessageWithXMark(wp.Name)
			return false
		}

		wp.Attempts++

		result = (*wp.engine).Evaluate(wp.Check).(bool)

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
