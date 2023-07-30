package app

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type IncludedTemplate struct {
	Name         string  `yaml:"name"`
	Version      string  `yaml:"version"`
	Checksum     string  `yaml:"checksum"`
	LastModified string  `yaml:"last-modified"`
	Author       string  `yaml:"author"`
	Description  string  `yaml:"description"`
	Init         string  `yaml:"init"`
	Tasks        []*Task `yaml:"tasks"`
}

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

func (task *Task) ProcessInclude() bool {
	if !strings.HasPrefix(strings.TrimSpace(task.Include), "https") || task.Include == "" {
		return false
	}

	contents, err := utils.GetUrlContents(task.Include)

	if err != nil {
		fmt.Println(err)
		return false
	}

	if App.Workflow.RemoteTemplateIndex.Loaded {
		support.StatusMessage("Validating checksum for remote template: "+task.Include, false)
		remoteMeta := App.Workflow.RemoteTemplateIndex.GetTemplate(task.Include)
		validated, err := remoteMeta.ValidateChecksum(contents)

		if err != nil {
			support.PrintXMarkLine()
			fmt.Println(err)
			return false
		}

		if !validated {
			support.PrintXMarkLine()
			support.WarningMessage("Checksum mismatch for remote template: " + task.Include)
			task.Include = ""
			return false
		}

		support.PrintCheckMarkLine()
	}

	template := &IncludedTemplate{}
	err = yaml.Unmarshal([]byte(contents), template)

	if err != nil {
		fmt.Println(err)
		return false
	}

	// if App.Workflow.RemoteTemplateIndex.Loaded {
	// 	remoteTemplate := App.Workflow.RemoteTemplateIndex.GetTemplate(task.Include)

	// 	if remoteTemplate != nil {
	// 		if remoteTemplate.Checksum != template.Checksum {
	// 			support.WarningMessage("Checksum mismatch for remote template: " + task.Include)
	// 			task.Include = ""
	// 			return false
	// 		}
	// 	}
	// }

	if len(template.Init) > 0 {
		App.JsEngine.Evaluate(template.Init)
	}

	// TODO: validate checksum for each task file
	for _, t := range template.Tasks {
		t.Initialize()
		t.FromRemote = true
		App.Workflow.Tasks = append(App.Workflow.Tasks, t)
	}

	displayUrl := task.Include
	displayUrl = strings.Replace(displayUrl, "https://", "", -1)
	displayUrl = strings.Replace(displayUrl, "github.com/", "", -1)
	displayUrl = strings.Replace(displayUrl, "raw.githubusercontent.com/", "", -1)

	support.SuccessMessageWithCheck("Included remote task file: " + displayUrl)

	return true
}

func (task *Task) Initialize() {

	task.RunCount = 0

	if task.MaxRuns <= 0 {
		task.MaxRuns = 999999999
	}

	if len(task.Path) == 0 {
		task.Path = App.JsEngine.MakeStringEvaluatable("getCwd()")
	}

	task.If = App.JsEngine.MakeStringEvaluatable(task.If)

	if App.JsEngine.IsEvaluatableScriptString(task.Name) {
		task.Name = App.JsEngine.Evaluate(task.Name).(string)
	}
}

func (task *Task) runWithStatusMessagesSync(runningSilently bool) {
	command := task.Command

	if App.JsEngine.IsEvaluatableScriptString(command) {
		command = App.JsEngine.Evaluate(command).(string)
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

func (task *Task) GetDisplayName() string {
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
	App.Workflow.State.History.Push(task)
	App.Workflow.State.CurrentTask = task

	defer func() {
		App.Workflow.State.CurrentTask = nil
	}()

	if task.RunCount >= task.MaxRuns && task.MaxRuns > 0 {
		support.SkippedMessageWithSymbol(task.GetDisplayName())
		return
	}

	task.RunCount++

	// allow the path property to be an environment variable reference without wrapping it in `{{ }}`
	if utils.MatchesPattern(task.Path, "^\\$[\\w_]+$") {
		task.Path = App.JsEngine.MakeStringEvaluatable(task.Path)
	}

	if App.JsEngine.IsEvaluatableScriptString(task.Path) {
		tempCwd := App.JsEngine.Evaluate(task.Path)
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

	if App.JsEngine.IsEvaluatableScriptString(command) {
		command = App.JsEngine.Evaluate(command).(string)
	}

	cmd, _ := utils.StartCommand(command, task.Path)
	App.CmdStartCallback(cmd)
	cmd.Start()

	support.PrintCheckMarkLine()

	App.ProcessMap.Store(task.Uuid, cmd)
}
