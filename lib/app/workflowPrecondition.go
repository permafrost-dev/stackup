package app

type WorkflowPrecondition struct {
	Name       string `yaml:"name"`
	Check      string `yaml:"check"`
	OnFail     string `yaml:"on-fail"`
	FromRemote bool
	Attempts   int
	MaxRetries *int `yaml:"max-retries,omitempty"`
}

func (p *WorkflowPrecondition) Initialize() {
	p.Attempts = 0
	if p.MaxRetries == nil {
		p.MaxRetries = new(int)
		*p.MaxRetries = 0
	}
}

func (p *WorkflowPrecondition) HandleOnFailure() bool {
	result := true

	if App.JsEngine.IsEvaluatableScriptString(p.OnFail) {
		App.JsEngine.Evaluate(p.OnFail)
	} else {
		task := App.Workflow.FindTaskById(p.OnFail)
		if task != nil {
			task.Run(true)
		}
	}

	return result
}
