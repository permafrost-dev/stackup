package lib

import (
	"os"
	"os/exec"

	"github.com/antonmedv/expr"
)

type Engine struct {
	env map[string]interface{}
}

func NewEngine(app *Application) *Engine {
	e := &Engine{
		env: make(map[string]interface{}),
	}

	e.env["exists"] = func(filename string) bool {
		_, err := os.Stat(filename)
		return !os.IsNotExist(err)
	}

	e.env["hasFlag"] = func(name string) bool {
		result, err := app.CurrentCommand.Flags().GetBool(name)

		if err != nil {
			return false
		}

		return result
	}

	e.env["binariesInPath"] = func(names ...string) bool {
		for _, name := range names {
			_, err := exec.LookPath(name)
			if err != nil {
				return false
			}
		}
		return true
	}

	e.env["env"] = func(name string) string {
		value, ok := os.LookupEnv(name)

		if !ok {
			return ""
		}

		return value
	}

	return e
}

func (e *Engine) AddStepResult(step string, result interface{}) {
	e.env[step+".result"] = result
}

func (e *Engine) Evaluate(expression string) (interface{}, error) {
	return expr.Eval(expression, e.env)
}
