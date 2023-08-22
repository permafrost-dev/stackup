package app

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
)

type Task struct {
	Name       string   `yaml:"name"`
	Command    string   `yaml:"command"`
	If         string   `yaml:"if,omitempty"`
	Id         string   `yaml:"id,omitempty"`
	Silent     bool     `yaml:"silent"`
	Path       string   `yaml:"path"`
	Platforms  []string `yaml:"platforms,omitempty"`
	MaxRuns    int      `yaml:"maxRuns,omitempty"`
	Include    string   `yaml:"include,omitempty"`
	RunCount   int
	Uuid       string
	FromRemote bool
	Workflow   *StackupWorkflow //*types.AppWorkflowContract
	JsEngine   *scripting.JavaScriptEngine
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
	TaskReferenceContract
}

func (task *Task) engine() *scripting.JavaScriptEngine {
	return task.Workflow.JsEngine
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

	return task.Workflow.JsEngine.Evaluate(task.If).(bool)
}

func (task *Task) Initialize() {
	task.RunCount = 0

	if task.MaxRuns <= 0 {
		task.MaxRuns = 999999999
	}

	// task.JsEngine = task.Workflow.JsEngine
	// enginePtr := task.engine()
	//.(scripting.JavaScriptEngine)

	// if enginePtr == nil {
	// 	return
	// }

	engine := task.JsEngine

	// *(*task.Workflow).GetJsEngine()

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
	engine := *task.engine()

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

	enginePtr := task.engine()
	if enginePtr == nil {
		return
	}

	if task.RunCount >= task.MaxRuns && task.MaxRuns > 0 {
		support.SkippedMessageWithSymbol(task.GetDisplayName())
		return
	}

	task.RunCount++

	// allow the path property to be an environment variable reference without wrapping it in `{{ }}`
	if utils.MatchesPattern(task.Path, "^\\$[\\w_]+$") {
		task.Path = (*task.engine()).MakeStringEvaluatable(task.Path)
	}

	if (*task.engine()).IsEvaluatableScriptString(task.Path) {
		tempCwd := (*task.engine()).Evaluate(task.Path)
		task.Path = tempCwd.(string)
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

	if (*task.engine()).IsEvaluatableScriptString(command) {
		command = (*task.engine()).Evaluate(command).(string)
	}

	cmd := utils.StartCommand(command, task.Path, false)
	task.Workflow.CommandStartCb(cmd)
	cmd.Start()

	support.PrintCheckMarkLine()

	task.Workflow.ProcessMap.Store(task.Uuid, cmd)
}

func (tr *TaskReference) Initialize(workflow *StackupWorkflow) {
	tr.Workflow = workflow
	tr.JsEngine = workflow.JsEngine
}

func (tr *TaskReference) TaskId() string {
	fmt.Printf("st == %v\n", tr)

	if tr.JsEngine.IsEvaluatableScriptString(tr.Task) {
		return tr.JsEngine.Evaluate(tr.Task).(string)
	}

	return tr.Task
}

func (st *ScheduledTask) TaskId() string {
	fmt.Printf("st == %v\n", st)
	if st.Workflow.JsEngine.IsEvaluatableScriptString(st.Task) {
		return st.Workflow.JsEngine.Evaluate(st.Task).(string)
	}

	return st.Task
}

func (st *ScheduledTask) Initialize(workflow *StackupWorkflow) {
	st.Workflow = workflow

	if workflow.JsEngine.IsEvaluatableScriptString(st.Task) {
		st.Task = workflow.JsEngine.Evaluate(st.Task).(string)
	}
}
