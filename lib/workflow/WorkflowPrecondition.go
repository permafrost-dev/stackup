package workflow

type WorkflowPrecondition struct {
	Name       string `yaml:"name"`
	Check      string `yaml:"check"`
	OnFail     string `yaml:"on-fail"`
	FromRemote bool
	Attempts   int
	MaxRetries *int `yaml:"max-retries,omitempty"`
    Workflow *StackupWorkflow
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
