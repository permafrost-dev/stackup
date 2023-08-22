package app

import (
	"runtime"
	"strings"
	"sync"

	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/scripting"
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
	ProcessMap     *sync.Map
	CommandStartCb types.CommandCallback
	//Workflow   *StackupWorkflow //*types.AppWorkflowContract
	JsEngine *scripting.JavaScriptEngine
	// types.AppWorkflowTaskContract
}

type TaskReferenceContract interface {
	TaskId() string
	Initialize(wf *StackupWorkflow)
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

func (task *Task) Initialize() {
	task.RunCount = 0
	if task.MaxRuns <= 0 {
		task.MaxRuns = 999999999
	}

	engine := task.JsEngine

	if len(task.Path) == 0 {
		task.Path = engine.MakeStringEvaluatable(consts.DEFAULT_CWD_SETTING)
	}

	task.If = engine.MakeStringEvaluatable(task.If)

	if engine.IsEvaluatableScriptString(task.Name) {
		task.Name = engine.Evaluate(task.Name).(string)
	}
}

func (task *Task) runWithStatusMessagesSync(runningSilently bool) {
	command := task.Command
	engine := task.JsEngine

	if engine.IsEvaluatableScriptString(command) {
		command = engine.Evaluate(command).(string)
	}

	cmd, err := utils.RunCommandInPath(command, task.Path, runningSilently)

	if err != nil {
		support.FailureMessageWithXMark(task.GetDisplayName())
		return
	}

	if cmd != nil && runningSilently {
		support.PrintCheckMarkLine()
	} else if cmd != nil {
		support.SuccessMessageWithCheck(task.GetDisplayName())
	}

	if cmd == nil && runningSilently {
		support.PrintXMarkLine()
	} else if cmd == nil {
		support.FailureMessageWithXMark(task.GetDisplayName())
	}
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

func (task *Task) Run(synchronous bool) {
	//cleanup :=
	// task.Workflow.State.SetCurrent(task)
	// if cleanup != nil {
	// 	defer cleanup()
	// }

	if task.RunCount >= task.MaxRuns && task.MaxRuns > 0 {
		support.SkippedMessageWithSymbol(task.GetDisplayName())
		return
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
		return
	}

	if !task.CanRunOnCurrentPlatform() {
		support.SkippedMessageWithSymbol("Task '" + task.GetDisplayName() + "' is not supported on this operating system.")
		return
	}

	command := task.Command
	runningSilently := task.Silent == true

	support.StatusMessage(task.GetDisplayName()+"...", false)

	if synchronous {
		task.runWithStatusMessagesSync(runningSilently)
		return
	}

	if task.JsEngine.IsEvaluatableScriptString(command) {
		command = task.JsEngine.Evaluate(command).(string)
	}

	cmd := utils.StartCommand(command, task.Path, false)
	task.CommandStartCb(cmd)
	cmd.Start()

	support.PrintCheckMarkLine()

	task.ProcessMap.Store(task.Uuid, cmd)
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

func (st *ScheduledTask) Initialize(workflow *StackupWorkflow, engine *scripting.JavaScriptEngine) {
	st.Workflow = workflow
	st.JsEngine = engine

	if workflow.JsEngine.IsEvaluatableScriptString(st.Task) {
		st.Task = workflow.JsEngine.Evaluate(st.Task).(string)
	}
}
