package app

import (
	"runtime"
	"strings"

	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
)

type Task struct {
	Name      string   `yaml:"name"`
	Command   string   `yaml:"command"`
	If        string   `yaml:"if,omitempty"`
	Id        string   `yaml:"id,omitempty"`
	Silent    bool     `yaml:"silent"`
	Path      string   `yaml:"path"`
	Type      string   `yaml:"type"`
	Platforms []string `yaml:"platforms,omitempty"`
	MaxRuns   int      `yaml:"maxRuns,omitempty"`
	RunCount  int
}

func (task *Task) CanRunOnCurrentPlatform() bool {
	if task.Platforms == nil || len(task.Platforms) == 0 {
		return true
	}

	foundPlatform := false

	for _, name := range task.Platforms {
		if strings.EqualFold(runtime.GOOS, name) {
			foundPlatform = true
			break
		}
	}

	return foundPlatform
}

func (task *Task) CanRunConditionally() bool {
	if len(strings.TrimSpace(task.If)) == 0 {
		return true
	}

	result := App.JsEngine.Evaluate(task.If)

	if result.(bool) {
		return true
	}

	return false
}

func (task *Task) Initialize() {
	task.RunCount = 0

	if task.MaxRuns <= 0 {
		task.MaxRuns = 999999999
	}

	if len(task.Path) == 0 {
		task.Path = "{{ getCwd() }}"
	}
}

func (task *Task) runWithStatusMessagesSync(runningSilently bool) {
	command := task.Command

	if App.JsEngine.IsEvaluatableScriptString(command) {
		command = App.JsEngine.Evaluate(command).(string)
	}

	cmd := utils.RunCommandInPath(command, task.Path, runningSilently)

	if cmd != nil && runningSilently {
		support.PrintCheckMarkLine()
	} else if cmd != nil {
		support.SuccessMessageWithCheck(task.Name)
	}

	if cmd == nil && runningSilently {
		support.PrintXMarkLine()
	} else if cmd == nil {
		support.FailureMessageWithXMark(task.Name)
	}
}

func (task *Task) Run(synchronous bool) {
	if task.RunCount >= task.MaxRuns && task.MaxRuns > 0 {
		support.SkippedMessageWithSymbol(task.Name)
		return
	}

	task.RunCount++

	if App.JsEngine.IsEvaluatableScriptString(task.Path) {
		tempCwd := App.JsEngine.Evaluate(task.Path)
		task.Path = tempCwd.(string)
	}

	if !task.CanRunConditionally() {
		support.SkippedMessageWithSymbol(task.Name)
		return
	}

	if !task.CanRunOnCurrentPlatform() {
		support.SkippedMessageWithSymbol("Task '" + task.Name + "' is not supported on this operating system.")
		return
	}

	command := task.Command
	runningSilently := task.Silent == true

	support.StatusMessage(task.Name+"...", false)

	if synchronous {
		task.runWithStatusMessagesSync(runningSilently)
		return
	}

	if App.JsEngine.IsEvaluatableScriptString(command) {
		command = App.JsEngine.Evaluate(command).(string)
	}

	cmd, _ := utils.StartCommand(command, task.Path)
	App.CmdStartCallback(cmd)
	cmd.Start()

	support.PrintCheckMarkLine()

	App.ProcessMap.Store(task.Name, cmd)
}
