package workflow

import (
	"runtime"
	"strings"

	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/types"
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
	// types.AppWorkflowTaskContract
}

type TaskReferenceContract interface {
	TaskId() string
	Initialize(wf *StackupWorkflow)
}

type TaskReference struct {
	Task     string `yaml:"task"`
	Workflow *StackupWorkflow
	TaskReferenceContract
}

type ScheduledTask struct {
	Task     string `yaml:"task"`
	Cron     string `yaml:"cron"`
	Workflow *StackupWorkflow
	TaskReferenceContract
}

func (task *Task) engine() types.JavaScriptEngineContract {
	engine := (*task.Workflow).GetJsEngine()

	return *engine
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

	jse := (*task.Workflow).GetJsEngine()
	result := (*jse).Evaluate(task.If).(bool)

	return result
}

func (task *Task) Initialize() {
	task.RunCount = 0

	if task.MaxRuns <= 0 {
		task.MaxRuns = 999999999
	}

	engine := *(*task.Workflow).GetJsEngine()

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

	if task.engine().IsEvaluatableScriptString(command) {
		command = task.engine().Evaluate(command).(string)
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
	cleanup := task.Workflow.State.SetCurrent(task)
	defer cleanup()

	if task.RunCount >= task.MaxRuns && task.MaxRuns > 0 {
		support.SkippedMessageWithSymbol(task.GetDisplayName())
		return
	}

	task.RunCount++

	// allow the path property to be an environment variable reference without wrapping it in `{{ }}`
	if utils.MatchesPattern(task.Path, "^\\$[\\w_]+$") {
		task.Path = task.engine().MakeStringEvaluatable(task.Path)
	}

	if task.engine().IsEvaluatableScriptString(task.Path) {
		tempCwd := task.engine().Evaluate(task.Path)
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

	if task.engine().IsEvaluatableScriptString(command) {
		command = task.engine().Evaluate(command).(string)
	}

	cmd := utils.StartCommand(command, task.Path, false)
	task.Workflow.CommandStartCb(cmd)
	cmd.Start()

	support.PrintCheckMarkLine()

	task.Workflow.ProcessMap.Store(task.Uuid, cmd)
}

func (tr *TaskReference) Initialize(workflow *StackupWorkflow) {
	tr.Workflow = workflow
}

func (tr *TaskReference) TaskId() string {
	if tr.Workflow.JsEngine.IsEvaluatableScriptString(tr.Task) {
		return tr.Workflow.JsEngine.Evaluate(tr.Task).(string)
	}

	return tr.Task
}

func (st *ScheduledTask) TaskId() string {
	if st.Workflow.JsEngine.IsEvaluatableScriptString(st.Task) {
		return st.Workflow.JsEngine.Evaluate(st.Task).(string)
	}

	return st.Task
}

func (st *ScheduledTask) Initialize(workflow *StackupWorkflow) {
	st.Workflow = workflow
}
