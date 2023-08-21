package app

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"sync"
	"syscall"

	"github.com/eiannone/keyboard"
	"github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/telemetry"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/updater"
	"github.com/stackup-app/stackup/lib/utils"
	"github.com/stackup-app/stackup/lib/version"
	"github.com/stackup-app/stackup/lib/workflow"
	"gopkg.in/yaml.v2"
)

var App *Application

type AppFlags struct {
	DisplayHelp    *bool
	DisplayVersion *bool
	NoUpdateCheck  *bool
	ConfigFile     *string
}

type Application struct {
	Workflow            *workflow.StackupWorkflow
	JsEngine            *scripting.JavaScriptEngine
	cronEngine          *cron.Cron
	scheduledTaskMap    *sync.Map
	ProcessMap          *sync.Map
	Vars                *sync.Map
	flags               AppFlags
	CmdStartCallback    types.CommandCallback
	KillCommandCallback types.CommandCallback
	ConfigFilename      string
	Gateway             *gateway.Gateway
	Analytics           *telemetry.Telemetry
}

func NewApplication() *Application {
	result := &Application{}
	result.init()

	return result
}

func (a *Application) GetWorkflow() workflow.StackupWorkflow {
	return *a.Workflow
}

func (a *Application) loadWorkflowFile(filename string, wf *workflow.StackupWorkflow) {
	wf.CommandStartCb = a.CmdStartCallback
	wf.ExitAppFunc = a.exitApp
	wf.Gateway = a.Gateway
	wf.JsEngine = scripting.CreateNewJavascriptEngine(a.Vars, a.Gateway, *(a.Workflow), a.GetApplicationIconPath)
	wf.ProcessMap = a.ProcessMap

	contents, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, wf)
	if err != nil {
		fmt.Printf("error loading configuration file: %v", err)
		return
	}

	for _, task := range wf.Tasks {
		task.Workflow = a.Workflow
	}

	wf.State = workflow.StackupWorkflowState{
		CurrentTask: nil,
		Stack:       linkedliststack.New(),
		History:     linkedliststack.New(),
	}
}

func (a *Application) init() {
	a.scheduledTaskMap = &sync.Map{}
	a.ProcessMap = &sync.Map{}
	a.Vars = &sync.Map{}
	a.ConfigFilename = support.FindExistingFile([]string{"stackup.dist.yaml", "stackup.yaml"}, "stackup.yaml")
	a.Workflow = workflow.CreateWorkflow(a.Gateway)
	a.Gateway = gateway.New()

	a.flags = AppFlags{
		DisplayHelp:    flag.Bool("help", false, "Display help"),
		DisplayVersion: flag.Bool("version", false, "Display version"),
		NoUpdateCheck:  flag.Bool("no-update-check", false, "Disable update check"),
		ConfigFile:     flag.String("config", "", "Load a specific config file"),
	}

	flag.Parse()
	if a.flags.ConfigFile != nil && *a.flags.ConfigFile != "" {
		a.ConfigFilename = *a.flags.ConfigFile
	}

	a.loadWorkflowFile(a.ConfigFilename, a.Workflow)
	a.Workflow.ConfigureDefaultSettings()
	a.JsEngine = scripting.CreateNewJavascriptEngine(a.Vars, a.Gateway, *(a.Workflow), a.GetApplicationIconPath)
	a.cronEngine = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))

	for _, task := range a.Workflow.Tasks {
		// var temp interface{} = a.Workflow
		// ref := temp.(types.AppWorkflowContract)
		task.Workflow = a.Workflow

		task.Initialize()
	}
}

func (a *Application) hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		<-c
		a.exitApp()
	}()
}

func (a *Application) hookKeyboard() {
	go func() {
		for {
			char, key, err := keyboard.GetSingleKey()

			if err != nil {
				return
			}

			if key == keyboard.KeyCtrlC || char == 'q' {
				a.exitApp()
			}
		}
	}()
}

func (a *Application) exitApp() {
	a.cronEngine.Stop()
	a.stopServerProcesses()
	support.StatusMessageLine("Running shutdown tasks...", true)
	a.runShutdownTasks()

	// for _, i := range a.Workflow.State.History.Values() {
	// 	// task := i.task
	// 	// runs := fmt.Sprintf("%v", task.RunCount)
	// 	//support.StatusMessageLine("[task history] task ran: "+task.GetDisplayName()+" ("+runs+" executions)", true)
	// }

	os.Exit(1)
}

func (a *Application) createScheduledTasks() {
	for _, def := range a.Workflow.Scheduler {
		_, found := a.Workflow.FindTaskById(def.TaskId())

		if !found {
			support.FailureMessageWithXMark("Task " + def.TaskId() + " not found.")
			continue
		}

		cron := def.Cron
		taskId := def.TaskId()

		a.cronEngine.AddFunc(cron, func() {
			task, found := a.Workflow.FindTaskById(taskId)
			if found {
				task.(*workflow.Task).Run(true)
			}
		})

		a.scheduledTaskMap.Store(def.TaskId(), &def)
	}

	a.cronEngine.Start()
}

func (a *Application) stopServerProcesses() {
	a.ProcessMap.Range(func(key any, value any) bool {
		t := a.Workflow.FindTaskByUuid(key.(string))
		support.StatusMessage("Stopping "+t.GetDisplayName()+"...", false)
		a.KillCommandCallback(value.(*exec.Cmd))
		support.PrintCheckMarkLine()

		return true
	})
}

func (a *Application) runEventLoop() {
	for {
		utils.WaitForStartOfNextMinute()
	}
}

func (a *Application) runStartupTasks() {
	for _, def := range a.Workflow.Startup {
		task, found := a.Workflow.FindTaskById(def.TaskId())

		if !found {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		// a.Workflow.State.CurrentTask = task
		task.(*workflow.Task).Run(true)
	}
}

func (a *Application) runShutdownTasks() {
	for _, def := range a.Workflow.Shutdown {
		task, found := a.Workflow.FindTaskById(def.TaskId())

		if !found {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		// a.Workflow.State.CurrentTask = task

		task.(*workflow.Task).Run(true)
	}
}

func (a *Application) runServerTasks() {
	for _, def := range a.Workflow.Servers {
		task, found := a.Workflow.FindTaskById(def.TaskId())

		if !found {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		task, _ = a.Workflow.FindTaskById(def.TaskId())

		// a.Workflow.State.CurrentTask = task
		task.(*workflow.Task).Run(false)
	}
}

func (a Application) runPreconditions() {
	for _, c := range a.Workflow.Preconditions {
		c.Initialize(a.Workflow, a.JsEngine)
		if !c.Run() {
			support.FailureMessageWithXMark(c.Name)
			os.Exit(1)
		}
		support.SuccessMessageWithCheck(c.Name)
	}
}

func (a *Application) createNewConfigFile() {
	if _, err := os.Stat("stackup.yaml"); err == nil {
		fmt.Println("stackup.yaml already exists.")
		return
	}

	dependencyBin := "php"

	if utils.IsFile("composer.json") {
		dependencyBin = "php"
	} else if utils.IsFile("package.json") {
		dependencyBin = "node"
	} else if utils.IsFile("requirements.txt") {
		dependencyBin = "python"
	}

	filename := "stackup.yaml"
	contents := `name: my stack
    description: application stack
    version: 1.0.0

    settings:
      anonymous-statistics: false
      exit-on-checksum-mismatch: false
      dotenv: ['.env', '.env.local']
      checksum-verification: true
      cache:
        ttl-minutes: 15
      domains:
        allowed:
          - '*.githubusercontent.com'
        hosts:
          - hostname: '*.github.com'
            gateway: allow
            headers:
              - 'Accept: application/vnd.github.v3+json'
      gateway:
        content-types:
          allowed:
            - '*'

    includes:
      - url: gh:permafrost-dev/stackup/main/templates/remote-includes/containers.yaml
      - url: gh:permafrost-dev/stackup/main/templates/remote-includes/` + dependencyBin + `.yaml

    # project type preconditions are loaded from included file above
    preconditions:

    startup:
      - task: start-containers

    shutdown:
      - task: stop-containers

    servers:

    scheduler:

    # tasks are loaded from included files above
    tasks:
    `
	os.WriteFile(filename, []byte(contents), 0644)
}

func (a *Application) checkForApplicationUpdates() {
	updateAvailable, release := updater.
		New(a.Gateway).IsLatestApplicationReleaseNewerThanCurrent(a.Workflow.Cache, version.APP_VERSION, "permafrost-dev/stackup")

	if updateAvailable {
		support.WarningMessage(fmt.Sprintf("A new version of StackUp is available, released %s.", release.TimeSinceRelease()))
	}
}

func (a *Application) handleFlagOptions() {
	if *a.flags.DisplayHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *a.flags.DisplayVersion {
		fmt.Println("StackUp version " + version.APP_VERSION)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "init" {
		a.createNewConfigFile()
		os.Exit(0)
	}
}

func (a *Application) GetConfigurationPath() string {
	pathname, _ := utils.EnsureConfigDirExists("stackup")

	return pathname
}

func (a *Application) DownloadApplicationIcon() {
	filename := a.GetApplicationIconPath()

	if utils.FileExists(filename) && utils.IsFile(filename) {
		return
	}

	utils.SaveUrlToFile("https://raw.githubusercontent.com/permafrost-dev/stackup/main/assets/stackup-app-512px.png", filename)
}

func (a Application) GetWorkflowContract() *types.AppWorkflowContract {
	var result interface{} = a.Workflow
	return result.(*types.AppWorkflowContract)
}

func (a *Application) GetApplicationIconPath() string {
	return path.Join(a.GetConfigurationPath(), "/stackup-icon.png")
}

func (a *Application) Run() {
	utils.EnsureConfigDirExists("stackup")
	godotenv.Load()
	a.handleFlagOptions()

	if len(a.Workflow.Settings.DotEnvFiles) > 0 {
		godotenv.Load(a.Workflow.Settings.DotEnvFiles...)
	}

	a.JsEngine.CreateEnvironmentVariables(os.Environ())
	a.JsEngine.CreateAppVariables(a.Vars)

	a.Workflow.Initialize(a.GetConfigurationPath())

	a.Gateway.Initialize(a.Workflow.Settings, a.JsEngine.AsContract(), nil)
	a.Analytics = telemetry.New(a.Workflow.Settings.AnonymousStatistics, a.Gateway)
	a.Gateway.AllowedDomains = []string{"*"}

	a.Workflow.ProcessIncludes()

	a.JsEngine.CreateEnvironmentVariables(os.Environ())
	a.JsEngine.CreateAppVariables(a.Vars)

	a.JsEngine.Evaluate(a.Workflow.Init)

	if a.Analytics.IsEnabled {
		a.Analytics.EventOnly("app.start")
	}

	if !*a.flags.NoUpdateCheck {
		a.checkForApplicationUpdates()
	}

	a.hookSignals()
	a.hookKeyboard()

	support.StatusMessageLine("Running precondition checks...", true)
	a.runPreconditions()

	support.StatusMessageLine("Running startup tasks...", true)
	a.runStartupTasks()

	support.StatusMessageLine("Starting server processes...", true)
	a.runServerTasks()

	support.StatusMessage("Creating scheduled tasks...", false)
	a.createScheduledTasks()
	support.PrintCheckMarkLine()

	utils.WaitForStartOfNextMinute()

	a.runEventLoop()
}
