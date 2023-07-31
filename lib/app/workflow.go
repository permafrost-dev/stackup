package app

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"path"
	"strings"

	lla "github.com/emirpasic/gods/lists/arraylist"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name                string            `yaml:"name"`
	Description         string            `yaml:"description"`
	Version             string            `yaml:"version"`
	Settings            *WorkflowSettings `yaml:"settings"`
	Init                string            `yaml:"init"`
	Preconditions       []*Precondition   `yaml:"preconditions"`
	Tasks               []*Task           `yaml:"tasks"`
	TaskList            *lla.List
	Startup             []TaskReference    `yaml:"startup"`
	Shutdown            []TaskReference    `yaml:"shutdown"`
	Servers             []TaskReference    `yaml:"servers"`
	Scheduler           []ScheduledTask    `yaml:"scheduler"`
	Includes            []*WorkflowInclude `yaml:"includes"`
	State               *StackupWorkflowState
	RemoteTemplateIndex *RemoteTemplateIndex
}

type WorkflowInclude struct {
	Url            string `yaml:"url"`
	ChecksumUrl    string `yaml:"checksum-url"`
	VerifyChecksum *bool  `yaml:"verify,omitempty"`
}

type WorkflowSettings struct {
	Defaults               *WorkflowSettingsDefaults `yaml:"defaults"`
	RemoteIndexUrl         string                    `yaml:"remote-index-url"`
	ExitOnChecksumMismatch bool                      `yaml:"exit-on-checksum-mismatch"`
}

type WorkflowSettingsDefaults struct {
	Tasks *WorkflowSettingsDefaultsTasks `yaml:"tasks"`
}

type WorkflowSettingsDefaultsTasks struct {
	Silent    bool     `yaml:"silent"`
	Path      string   `yaml:"path"`
	Platforms []string `yaml:"platforms"`
}

type StackupWorkflowState struct {
	CurrentTask *Task
	Stack       *lls.Stack
	History     *lls.Stack
}

type Precondition struct {
	Name       string `yaml:"name"`
	Check      string `yaml:"check"`
	FromRemote bool
}

type TaskReference struct {
	Task string `yaml:"task"`
}

type ScheduledTask struct {
	Task string `yaml:"task"`
	Cron string `yaml:"cron"`
}

func GetState() *StackupWorkflowState {
	return App.Workflow.State
}

func (ws *WorkflowSettings) FullRemoteIndexUrl() string {
	if strings.HasPrefix(strings.TrimSpace(ws.RemoteIndexUrl), "gh:") {
		return "https://raw.githubusercontent.com/" + strings.TrimPrefix(ws.RemoteIndexUrl, "gh:")
	}

	return ws.RemoteIndexUrl
}

func (wi *WorkflowInclude) getChecksumFromContents(contents string) string {
	lines := strings.Split(contents, "\n")
	filename := path.Base(wi.FullUrl())

	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) != 2 {
			continue
		}

		if strings.TrimSpace(parts[1]) != filename {
			continue
		}

		return strings.TrimSpace(parts[0])
	}

	return strings.TrimSpace(contents)
}

func (wi *WorkflowInclude) ValidateChecksum(contents string) (bool, error) {
	if wi.VerifyChecksum != nil && *wi.VerifyChecksum == false {
		return true, nil
	}

	checksumUrls := []string{
		wi.FullUrl() + ".sha256",
		wi.FullUrl() + ".sha512",
	}

	algorithm := ""
	checksumContents := ""

	for _, url := range checksumUrls {
		checksumContents, err := utils.GetUrlContents(url)
		if err != nil {
			continue
		}

		if checksumContents != "" {
			checksumContents = wi.getChecksumFromContents(checksumContents)
			fmt.Println(checksumContents)

			wi.ChecksumUrl = url
			if strings.HasSuffix(url, ".sha256") {
				algorithm = "sha256"
			}
			if strings.HasSuffix(url, ".sha512") {
				algorithm = "sha512"
			}
			break
		}
	}

	if algorithm == "" {
		// return false, fmt.Errorf("unable to find valid checksum file for %s", wi.DisplayUrl())
	}

	var hash []byte

	switch algorithm {
	case "sha256":
		h := sha256.New()
		h.Write([]byte(wi.getChecksumFromContents(checksumContents)))
		hash = h.Sum(nil)
		break
	case "sha512":
		h := sha512.New()
		h.Write([]byte(wi.getChecksumFromContents(checksumContents)))
		hash = h.Sum(nil)
		break
	default:
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	fmt.Printf("checksum url: %s\n", wi.ChecksumUrl)
	fmt.Printf("algorithm: %s\n", algorithm)
	fmt.Printf("hash: %x\n", hash)
	fmt.Printf("checksum: %s\n", checksumContents)

	checksumBytes, err := hex.DecodeString(string(checksumContents))
	if err != nil {
		return false, err
	}
	if !hmac.Equal(hash, checksumBytes) {
		return false, nil
	}

	return true, nil
}

func (wi *WorkflowInclude) FullUrl() string {
	if strings.HasPrefix(strings.TrimSpace(wi.Url), "gh:") {
		return "https://raw.githubusercontent.com/" + strings.TrimPrefix(wi.Url, "gh:")
	}

	return wi.Url
}

func (wi *WorkflowInclude) DisplayUrl() string {
	displayUrl := strings.Replace(wi.FullUrl(), "https://", "", -1)
	displayUrl = strings.Replace(displayUrl, "github.com/", "", -1)
	displayUrl = strings.Replace(displayUrl, "raw.githubusercontent.com/", "", -1)

	return displayUrl
}

func (workflow *StackupWorkflow) FindTaskById(id string) *Task {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Id, id) && len(task.Id) > 0 {
			return task
		}
	}

	return nil
}

func (workflow *StackupWorkflow) FindTaskByUuid(uuid string) *Task {
	for _, task := range workflow.Tasks {
		if strings.EqualFold(task.Uuid, uuid) && len(uuid) > 0 {
			return task
		}
	}

	return nil
}

func (workflow *StackupWorkflow) TaskIdToUuid(id string) string {
	task := workflow.FindTaskById(id)

	if task == nil {
		return ""
	}

	return task.Uuid
}

func (workflow *StackupWorkflow) Initialize() {
	// generate uuids for each task as the initial step, as other code below relies on a uuid existing
	for _, task := range workflow.Tasks {
		task.Uuid = utils.GenerateTaskUuid()
	}

	// no default settings were provided, so create sensible defaults
	if workflow.Settings == nil {
		workflow.Settings = &WorkflowSettings{
			Defaults: &WorkflowSettingsDefaults{
				Tasks: &WorkflowSettingsDefaultsTasks{
					Silent:    false,
					Path:      App.JsEngine.MakeStringEvaluatable("getCwd()"),
					Platforms: []string{"windows", "linux", "darwin"},
				},
			},
		}
	}

	// copy the default settings into each task if appropriate
	for _, task := range workflow.Tasks {
		if task.Path == "" && len(workflow.Settings.Defaults.Tasks.Path) > 0 {
			task.Path = workflow.Settings.Defaults.Tasks.Path
		}

		if !task.Silent && workflow.Settings.Defaults.Tasks.Silent {
			task.Silent = workflow.Settings.Defaults.Tasks.Silent
		}

		if (task.Platforms == nil || len(task.Platforms) == 0) && len(workflow.Settings.Defaults.Tasks.Platforms) > 0 {
			task.Platforms = workflow.Settings.Defaults.Tasks.Platforms
		}
	}

	workflow.ProcessIncludes()

	if len(workflow.Init) > 0 {
		App.JsEngine.Evaluate(workflow.Init)
	}

	for _, task := range workflow.Tasks {
		task.Initialize()
	}
}

func (workflow *StackupWorkflow) RemoveTasks(uuidsToRemove []string) {
	// Create a map of UUIDs to remove for faster lookup
	uuidMap := make(map[string]bool)
	for _, uuid := range uuidsToRemove {
		uuidMap[uuid] = true
	}

	// Remove tasks with UUIDs in the uuidMap
	var newTasks []*Task
	for _, task := range workflow.Tasks {
		if !uuidMap[task.Uuid] {
			newTasks = append(newTasks, task)
		}
	}
	workflow.Tasks = newTasks
}

func (workflow *StackupWorkflow) ProcessIncludes() {
	workflow.RemoteTemplateIndex = &RemoteTemplateIndex{Loaded: false}

	// set default value for verify checksum to true
	for _, wi := range workflow.Includes {
		if wi.VerifyChecksum == nil {
			boolValue := wi.ChecksumUrl != ""
			wi.VerifyChecksum = &boolValue
		}
	}

	// if workflow.Settings.RemoteIndexUrl != "" {
	// 	remoteIndex, err := LoadRemoteTemplateIndex(workflow.Settings.FullRemoteIndexUrl())

	// 	remoteIndex.Loaded = err == nil
	// 	if !remoteIndex.Loaded {
	// 		support.WarningMessage("Unable to load remote template index.")
	// 	} else {
	// 		workflow.RemoteTemplateIndex = remoteIndex
	// 		support.SuccessMessageWithCheck("Downloaded remote template index file.")
	// 	}
	// }

	for _, include := range workflow.Includes {
		workflow.ProcessInclude(include)
	}
}

func (workflow *StackupWorkflow) ProcessInclude(include *WorkflowInclude) bool {
	if !strings.HasPrefix(strings.TrimSpace(include.FullUrl()), "https") || include.Url == "" {
		return false
	}

	contents, err := utils.GetUrlContents(include.FullUrl())

	if err != nil {
		fmt.Println(err)
		return false
	}

	if *include.VerifyChecksum == true {
		support.StatusMessage("Validating checksum for remote include: "+include.DisplayUrl(), false)
		validated, err := include.ValidateChecksum(contents)

		if err != nil {
			support.PrintXMarkLine()
			fmt.Println(err)
			return false
		}

		if !validated {
			support.PrintXMarkLine()

			if App.Workflow.Settings.ExitOnChecksumMismatch {
				support.FailureMessageWithXMark("Exiting due to checksum mismatch.")
				App.exitApp()
				return false
			}
		} else {
			support.PrintCheckMarkLine()
		}
	}

	template := &IncludedTemplate{}
	err = yaml.Unmarshal([]byte(contents), template)

	if err != nil {
		fmt.Println(err)
		return false
	}

	if len(template.Init) > 0 {
		workflow.Init += "\n\n" + template.Init
	}

	for _, p := range template.Preconditions {
		p.FromRemote = true
		App.Workflow.Preconditions = append(App.Workflow.Preconditions, p)
	}

	for _, t := range template.Tasks {
		t.FromRemote = true
		App.Workflow.Tasks = append(App.Workflow.Tasks, t)
	}

	support.SuccessMessageWithCheck("Included remote file (checksum verified): " + include.DisplayUrl())

	return true
}

func (tr *TaskReference) TaskId() string {
	if App.JsEngine.IsEvaluatableScriptString(tr.Task) {
		return App.JsEngine.Evaluate(tr.Task).(string)
	}

	return tr.Task
}

func (st *ScheduledTask) TaskId() string {
	if App.JsEngine.IsEvaluatableScriptString(st.Task) {
		return App.JsEngine.Evaluate(st.Task).(string)
	}

	return st.Task
}
