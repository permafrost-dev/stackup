package app

import (
	"runtime"
	"strings"

	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/settings"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/utils"
)

type Task struct {
	Name           string   `yaml:"name"`
	Command        string   `yaml:"command"`
	If             string   `yaml:"if,omitempty"`
	Id             string   `yaml:"id,omitempty"`
	Silent         bool     `yaml:"silent"`
	Path           string   `yaml:"path"`
	Platforms      []string `yaml:"platforms,omitempty"`
	MaxRuns        int      `yaml:"maxRuns,omitempty"`
	Include        string   `yaml:"include,omitempty"`
	RunCount       int
	Uuid           string
	FromRemote     bool
	CommandStartCb types.CommandCallback
	//Workflow   *StackupWorkflow //*types.AppWorkflowContract
	JsEngine     *scripting.JavaScriptEngine
	setActive    SetActiveTaskCallback
	StoreProcess types.SetProcessCallback
	// types.AppWorkflowTaskContract
}

type TaskReferenceContract interface {
	TaskId() string
	Initialize(wf *StackupWorkflow, jse *scripting.JavaScriptEngine)
}

type TaskReference struct {
	Task     string `yaml:"task"`
	Workflow *StackupWorkflow
	JsEngine *scripting.JavaScriptEngine
	TaskReferenceContract
}

type ScheduledTask struct {
	Task     string `yaml:"task"`
	Cron     string `yaml:"cron"`
	Workflow *StackupWorkflow
	JsEngine *scripting.JavaScriptEngine
	TaskReferenceContract
}

func (task *Task) CanRunOnCurrentPlatform() bool {
	if task.Platforms == nil || len(task.Platforms) == 0 {
		return true
	}

	for _, name := range task.Platforms {
		if strings.EqualFold(runtime.GOOS, name) {
			return true
		}
	}

	return false
}

func (task *Task) CanRunConditionally() bool {
	if len(strings.TrimSpace(task.If)) == 0 {
		return true
	}

	return task.JsEngine.Evaluate(task.If).(bool)
}

func (task *Task) Initialize(workflow *StackupWorkflow) { //} *scripting.JavaScriptEngine, cmdStartCb types.CommandCallback, setActive SetActiveTaskCallback, storeProcess types.SetProcessCallback) {
	task.JsEngine = workflow.JsEngine
	task.setActive = workflow.State.CurrentTask.setActive
	task.CommandStartCb = workflow.CommandStartCb
	task.StoreProcess = workflow.ProcessMap.Store
	task.Uuid = utils.GenerateTaskUuid()

	task.RunCount = 0
	if task.MaxRuns <= 0 {
		task.MaxRuns = 999999999
	}

	if len(task.Path) == 0 {
		task.Path = task.JsEngine.MakeStringEvaluatable(consts.DEFAULT_CWD_SETTING)
	}

	if task.If != "" {
		task.If = task.JsEngine.MakeStringEvaluatable(task.If)
	}

	if task.JsEngine.IsEvaluatableScriptString(task.Name) {
		task.Name = task.JsEngine.Evaluate(task.Name).(string)
	}
}

func (task *Task) SetDefaultSetttings(s *settings.Settings) {
	if task.Path == "" && len(s.Defaults.Tasks.Path) > 0 {
		task.Path = s.Defaults.Tasks.Path
	}

	if len(task.Platforms) == 0 {
		copy(task.Platforms, s.Defaults.Tasks.Platforms)
	}

	task.Silent = s.Defaults.Tasks.Silent
}

func (task Task) GetDisplayName() string {
	if len(task.Include) > 0 {
		return strings.Replace(task.Include, "https://", "", -1)
	}

	if len(task.Name) > 0 {
		return task.Name
	}

	if len(task.Id) > 0 {
		return task.Id
	}

	return task.Uuid
}

func (task *Task) getCommand() string {
	result := task.Command

	if task.JsEngine.IsEvaluatableScriptString(result) {
		result = task.JsEngine.Evaluate(result).(string)
	}

	return result
}

func (task *Task) prepareRun() (bool, func()) {
	result := task.setActive(task)

	if task.Uuid == "" {
		task.Uuid = utils.GenerateTaskUuid()
	}

	if task.RunCount >= task.MaxRuns && task.MaxRuns > 0 {
		support.SkippedMessageWithSymbol(task.GetDisplayName())
		return false, nil
	}

	task.RunCount++

	// allow the path property to be an environment variable reference without wrapping it in `{{ }}`
	if utils.MatchesPattern(task.Path, "^\\$[\\w_]+$") {
		task.Path = task.JsEngine.MakeStringEvaluatable(task.Path)
	}

	if task.JsEngine.IsEvaluatableScriptString(task.Path) {
		task.Path = task.JsEngine.Evaluate(task.Path).(string)
	}

	if !task.CanRunConditionally() {
		support.SkippedMessageWithSymbol(task.GetDisplayName())
		return false, nil
	}

	if !task.CanRunOnCurrentPlatform() {
		support.SkippedMessageWithSymbol("Task '" + task.GetDisplayName() + "' is not supported on this operating system.")
		return false, nil
	}

	support.StatusMessage(task.GetDisplayName()+"...", false)

	return true, result
}

func (task *Task) RunSync() bool {
	var canRun bool
	var cleanup func()

	if canRun, cleanup = task.prepareRun(); !canRun {
		return false
	}

	defer cleanup()

	cmd, err := utils.RunCommandInPath(task.getCommand(), task.Path, task.Silent)
	if err != nil {
		support.FailureMessageWithXMark(task.GetDisplayName())
		return false
	}

	if cmd != nil && task.Silent {
		support.PrintCheckMarkLine()
	} else if cmd != nil {
		support.SuccessMessageWithCheck(task.GetDisplayName())
	}

	if cmd == nil && task.Silent {
		support.PrintXMarkLine()
	} else if cmd == nil {
		support.FailureMessageWithXMark(task.GetDisplayName())
	}

	return true
}

func (task *Task) RunAsync() {
	var canRun bool
	var cleanup func()

	if canRun, cleanup = task.prepareRun(); !canRun {
		return
	}

	defer cleanup()

	command := task.getCommand()
	cmd := utils.StartCommand(command, task.Path, false)

	if cmd == nil {
		support.FailureMessageWithXMark(task.GetDisplayName())
		return
	}

	task.CommandStartCb(cmd)
	err := cmd.Start()

	if err != nil {
		support.PrintXMarkLine()
	} else {
		support.PrintCheckMarkLine()
	}

	task.StoreProcess(task.Uuid, cmd)
}

func (tr *TaskReference) Initialize(workflow *StackupWorkflow) {
	tr.Workflow = workflow
	tr.JsEngine = workflow.JsEngine
}

func (tr *TaskReference) TaskId() string {
	if tr.JsEngine.IsEvaluatableScriptString(tr.Task) {
		return tr.JsEngine.Evaluate(tr.Task).(string)
	}

	return tr.Task
}

func (st *ScheduledTask) TaskId() string {
	if st.JsEngine.IsEvaluatableScriptString(st.Task) {
		return st.JsEngine.Evaluate(st.Task).(string)
	}

	return st.Task
}

func (st *ScheduledTask) Initialize(workflow *StackupWorkflow) {
	st.Workflow = workflow
	st.JsEngine = workflow.JsEngine

	if workflow.JsEngine.IsEvaluatableScriptString(st.Task) {
		st.Task = workflow.JsEngine.Evaluate(st.Task).(string)
	}
}
