package app

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"

	lla "github.com/emirpasic/gods/lists/arraylist"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/checksums"
	"github.com/stackup-app/stackup/lib/downloader"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
	"gopkg.in/yaml.v2"
)

type StackupWorkflow struct {
	Name          string            `yaml:"name"`
	Description   string            `yaml:"description"`
	Version       string            `yaml:"version"`
	Settings      *WorkflowSettings `yaml:"settings"`
	Env           []string          `yaml:"env"`
	Init          string            `yaml:"init"`
	Preconditions []*Precondition   `yaml:"preconditions"`
	Tasks         []*Task           `yaml:"tasks"`
	TaskList      *lla.List
	Startup       []TaskReference    `yaml:"startup"`
	Shutdown      []TaskReference    `yaml:"shutdown"`
	Servers       []TaskReference    `yaml:"servers"`
	Scheduler     []ScheduledTask    `yaml:"scheduler"`
	Includes      []*WorkflowInclude `yaml:"includes"`
	State         *StackupWorkflowState
	Cache         *cache.Cache
}

type WorkflowInclude struct {
	Url             string   `yaml:"url"`
	Headers         []string `yaml:"headers"`
	File            string   `yaml:"file"`
	ChecksumUrl     string   `yaml:"checksum-url"`
	VerifyChecksum  *bool    `yaml:"verify,omitempty"`
	AccessKey       string   `yaml:"access-key"`
	SecretKey       string   `yaml:"secret-key"`
	Secure          bool     `yaml:"secure"`
	ChecksumIsValid *bool
	ValidationState string
	Contents        string
	Hash            string
	//Workflow        *StackupWorkflow
}

type WorkflowSettings struct {
	Defaults               *WorkflowSettingsDefaults `yaml:"defaults"`
	ExitOnChecksumMismatch bool                      `yaml:"exit-on-checksum-mismatch"`
	DotEnvFiles            []string                  `yaml:"dotenv"`
	Cache                  *WorkflowSettingsCache    `yaml:"cache"`
	Domains                struct {
		Allowed []string `yaml:"allowed"`
	} `yaml:"domains"`
}

type WorkflowSettingsCache struct {
	TtlMinutes int `yaml:"ttl-minutes"`
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
	OnFail     string `yaml:"on-fail"`
	FromRemote bool
	Attempts   int
	MaxRetries *int `yaml:"max-retries,omitempty"`
}

func GetState() *StackupWorkflowState {
	return App.Workflow.State
}

func (p *Precondition) Initialize() {
	p.Attempts = 0
	if p.MaxRetries == nil {
		p.MaxRetries = new(int)
		*p.MaxRetries = 0
	}
}

func (p *Precondition) HandleOnFailure() bool {
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

func (wi *WorkflowInclude) getChecksumFromContents(contents string) string {
	checksumsArr, _ := checksums.ParseChecksumFileContents(contents)
	checksum := checksums.FindChecksumForFileFromUrl(checksumsArr, wi.FullUrl())

	if checksum != nil {
		// fmt.Printf("found checksum: %s\n", checksum.Hash)
		return checksum.Hash
	}

	//try to match a hash using regex
	regex := regexp.MustCompile("^([a-fA-F0-9]{48,})$")
	matches := regex.FindAllString(contents, -1)

	if len(matches) > 0 {
		return matches[0]
	}

	return strings.TrimSpace(contents)
}

func (wi *WorkflowInclude) ValidateChecksum(contents string) (bool, error) {
	if wi.VerifyChecksum != nil && *wi.VerifyChecksum == false {
		return true, nil
	}

	checksumUrls := []string{
		wi.FullUrlPath() + "/checksums.sha256.txt",
		wi.FullUrlPath() + "/checksums.sha512.txt",
		wi.FullUrl() + ".sha256",
		wi.FullUrl() + ".sha512",
	}

	algorithm := ""
	storedChecksum := ""
	checksumContents := ""

	for _, url := range checksumUrls {

		if !App.Workflow.Cache.IsExpired(url) {
			wi.ChecksumUrl = url
			checksumContents = App.Workflow.Cache.Get(url)
			// fmt.Printf("using cached checksum file %s\n", url)
			break
		}

		checksumContents, err := utils.GetUrlContents(url)
		if err != nil {
			continue
		}

		wi.ChecksumUrl = url

		if checksumContents != "" {
			// fmt.Printf("using checksum file %s\n", wi.ChecksumUrl)
			App.Workflow.Cache.Set(url, checksumContents, 3)
			// fmt.Printf("using non-cached checksum file %s\n", url)
			break
		}
	}

	if checksumContents != "" {
		storedChecksum = wi.getChecksumFromContents(checksumContents)

		// wi.ChecksumUrl = hashUrl
		// fmt.Println("checksum url: " + wi.ChecksumUrl)
		algorithm = wi.GetChecksumAlgorithm()
	}

	// algorithm = "sha256"

	if algorithm == "unknown" || algorithm == "" {
		return false, fmt.Errorf("unable to find valid checksum file for %s", wi.DisplayUrl())
	}

	var hash string

	switch algorithm {
	case "sha256":
		hash = checksums.CalculateSha256Hash(contents)
		break
	case "sha512":
		hash = checksums.CalculateSha512Hash(contents)
		break
	default:
		wi.SetChecksumIsValid(false)
		return false, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	if !strings.EqualFold(hash, storedChecksum) {
		wi.SetChecksumIsValid(false)
		return false, nil
	}

	wi.SetChecksumIsValid(true)
	return true, nil
}

func (wi *WorkflowInclude) IsLocalFile() bool {
	return wi.File != "" && utils.IsFile(wi.File)
}

func (wi *WorkflowInclude) IsRemoteUrl() bool {
	return wi.FullUrl() != "" && strings.HasPrefix(wi.FullUrl(), "http")
}

func (wi *WorkflowInclude) IsS3Url() bool {
	return wi.FullUrl() != "" && strings.HasPrefix(wi.FullUrl(), "s3:")
}

func (wi *WorkflowInclude) Filename() string {
	return utils.AbsoluteFilePath(wi.File)
}

func (wi *WorkflowInclude) FullUrl() string {
	if strings.HasPrefix(strings.TrimSpace(wi.Url), "gh:") {
		return "https://raw.githubusercontent.com/" + strings.TrimPrefix(wi.Url, "gh:")
	}

	return wi.Url
}

func (wi *WorkflowInclude) FullUrlPath() string {
	temp, _ := utils.ReplaceFilenameInUrl(wi.FullUrl(), "")

	return strings.TrimSuffix(temp, "/")
}

func (wi *WorkflowInclude) DisplayUrl() string {
	displayUrl := strings.Replace(wi.FullUrl(), "https://", "", -1)
	displayUrl = strings.Replace(displayUrl, "github.com/", "", -1)
	displayUrl = strings.Replace(displayUrl, "raw.githubusercontent.com/", "", -1)
	// displayUrl = strings.Replace(displayUrl, "s3://", "", -1)

	return displayUrl
}

func (wi *WorkflowInclude) DisplayName() string {
	if wi.IsRemoteUrl() {
		return wi.DisplayUrl()
	}

	if wi.IsLocalFile() {
		return wi.Filename()
	}

	if wi.IsS3Url() {
		return wi.DisplayUrl()
	}

	return "<unknown>"
}

func (wi *WorkflowInclude) GetChecksumAlgorithm() string {
	if strings.HasSuffix(wi.ChecksumUrl, ".sha256") || strings.HasSuffix(wi.ChecksumUrl, ".sha256.txt") {
		return "sha256"
	}
	if strings.HasSuffix(wi.ChecksumUrl, ".sha512") || strings.HasSuffix(wi.ChecksumUrl, ".sha512.txt") {
		return "sha512"
	}

	return "unknown"
}

func (wi *WorkflowInclude) SetChecksumIsValid(value bool) {
	wi.ChecksumIsValid = &value
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

func (workflow *StackupWorkflow) reversePreconditions(items []*Precondition) []*Precondition {
	length := len(items)
	for i := 0; i < length/2; i++ {
		items[i], items[length-i-1] = items[length-i-1], items[i]
	}

	return items
}

func (workflow *StackupWorkflow) Initialize() {
	workflow.Cache = cache.CreateCache("")

	// generate uuids for each task as the initial step, as other code below relies on a uuid existing
	for _, task := range workflow.Tasks {
		task.Uuid = utils.GenerateTaskUuid()
	}

	if len(workflow.Env) > 0 {
		for _, def := range workflow.Env {
			key, value, _ := strings.Cut(def, "=")
			os.Setenv(key, value)
		}
	}

	// no default settings were provided, so create sensible defaults
	if workflow.Settings == nil {
		workflow.Settings = &WorkflowSettings{
			DotEnvFiles: []string{".env"},
			Cache:       &WorkflowSettingsCache{TtlMinutes: 5},
			Defaults: &WorkflowSettingsDefaults{
				Tasks: &WorkflowSettingsDefaultsTasks{
					Silent:    false,
					Path:      App.JsEngine.MakeStringEvaluatable("getCwd()"),
					Platforms: []string{"windows", "linux", "darwin"},
				},
			},
		}
	}

	if workflow.Settings.Cache.TtlMinutes <= 0 {
		workflow.Settings.Cache.TtlMinutes = 5
	}

	if len(workflow.Settings.DotEnvFiles) == 0 {
		workflow.Settings.DotEnvFiles = []string{".env"}
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

	// initialize the includes
	for _, inc := range workflow.Includes {
		inc.Initialize()
	}

	workflow.ProcessIncludes()

	if len(workflow.Init) > 0 {
		App.JsEngine.Evaluate(workflow.Init)
	}

	for _, pc := range workflow.Preconditions {
		pc.Initialize()
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

// func (workflow *StackupWorkflow) ProcessIncludes() {
// 	for _, include := range workflow.Includes {
// 		workflow.ProcessInclude(include)
// 	}
// }

func (workflow *StackupWorkflow) ProcessIncludes() {
	var wg sync.WaitGroup
	for _, include := range workflow.Includes {
		wg.Add(1)
		go func(inc *WorkflowInclude) {
			defer wg.Done()
			workflow.ProcessInclude(inc)
		}(include)
	}
	wg.Wait()
}

func (workflow *StackupWorkflow) ProcessInclude(include *WorkflowInclude) bool {
	if !include.IsLocalFile() && !include.IsRemoteUrl() && !include.IsS3Url() {
		return false
	}

	var err error

	if workflow.Cache.Has(include.DisplayName()) && !workflow.Cache.IsExpired(include.DisplayName()) {
		include.Contents = workflow.Cache.Get(include.DisplayName())
		include.Hash = workflow.Cache.GetHash(include.DisplayName())
		// fmt.Println("loaded from cache")
	}

	if !workflow.Cache.Has(include.DisplayName()) || workflow.Cache.IsExpired(include.DisplayName()) {
		// fmt.Println("not loaded from cache")
		if include.IsLocalFile() {
			include.Contents, err = utils.GetFileContents(include.Filename())
		} else if include.IsRemoteUrl() {
			include.Contents, err = utils.GetUrlContentsEx(include.FullUrl(), include.Headers)
		} else if include.IsS3Url() {
			include.AccessKey = os.ExpandEnv(include.AccessKey)
			include.SecretKey = os.ExpandEnv(include.SecretKey)
			include.Contents = downloader.ReadS3FileContents(include.FullUrl(), include.AccessKey, include.SecretKey, include.Secure)
		} else {
			return false
		}

		include.Hash = checksums.CalculateSha256Hash(include.Contents)
		workflow.Cache.Set(include.DisplayName(), include.Contents, 3)
	}

	// fmt.Printf("value: %v\n", workflow.Cache.Get(include.DisplayName()))
	// fmt.Printf("expires: %v\n", workflow.Cache.GetExpiresAt(include.DisplayName()))
	// fmt.Printf("hash: %v\n", workflow.Cache.GetHash(include.DisplayName()))

	// fmt.Printf("include: %s\n", include.DisplayName())
	// fmt.Printf("include: %s\n", include.FullUrl())
	// fmt.Printf("contents: %s\n", include.Contents)

	if err != nil {
		fmt.Println(err)
		return false
	}

	include.ValidationState = "checksum not validated"

	if include.IsRemoteUrl() {
		if *include.VerifyChecksum == true || include.VerifyChecksum == nil {
			// support.StatusMessage("Validating checksum for remote include: "+include.DisplayUrl(), false)
			validated, _ := include.ValidateChecksum(include.Contents)

			if include.ChecksumIsValid != nil && *include.ChecksumIsValid == true {
				include.ValidationState = "checksum validated"
			}

			if include.ChecksumIsValid != nil && *include.ChecksumIsValid == false {
				include.ValidationState = "checksum mismatch"
			}

			// if err != nil {
			// 	fmt.Println(err)
			// 	return false
			// }

			if !validated {
				if App.Workflow.Settings.ExitOnChecksumMismatch {
					support.FailureMessageWithXMark("Exiting due to checksum mismatch.")
					App.exitApp()
					return false
				}
			}
		}
	}

	template := &IncludedTemplate{}
	err = yaml.Unmarshal([]byte(include.Contents), template)

	if err != nil {
		fmt.Println(err)
		return false
	}

	if len(template.Init) > 0 {
		App.Workflow.Init += "\n" + template.Init
	}

	// prepend the included preconditions; we reverse the order of the preconditions in the included file,
	// then reverse the existing preconditions, append them, then reverse the workflow preconditions again
	// to achieve the correct order.
	App.Workflow.Preconditions = App.Workflow.reversePreconditions(App.Workflow.Preconditions)
	template.Preconditions = App.Workflow.reversePreconditions(template.Preconditions)

	for _, p := range template.Preconditions {
		p.FromRemote = true
		App.Workflow.Preconditions = append(App.Workflow.Preconditions, p)
	}

	App.Workflow.Preconditions = App.Workflow.reversePreconditions(App.Workflow.Preconditions)

	for _, t := range template.Tasks {
		t.FromRemote = true
		t.Uuid = utils.GenerateTaskUuid()
		App.Workflow.Tasks = append(App.Workflow.Tasks, t)
	}

	support.SuccessMessageWithCheck("Included file (" + include.ValidationState + "): " + include.DisplayName())

	return true
}

func (wi *WorkflowInclude) Initialize() {
	//wi.Workflow = workflow

	// expand environment variables in the include headers
	for i, v := range wi.Headers {
		if App.JsEngine.IsEvaluatableScriptString(v) {
			wi.Headers[i] = App.JsEngine.Evaluate(v).(string)
		}
		wi.Headers[i] = os.ExpandEnv(v)
	}

	// set some default values
	if wi.VerifyChecksum == nil {
		boolValue := true //wi.ChecksumUrl != ""
		wi.VerifyChecksum = &boolValue
	}
	wi.ValidationState = "not validated"
	wi.ChecksumIsValid = nil
}
