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

func (a *Application) loadWorkflowFile(filename string) *workflow.StackupWorkflow {
	var result workflow.StackupWorkflow

	contents, err := os.ReadFile(filename)
	if err != nil {
		return &workflow.StackupWorkflow{
			CommandStartCb: a.CmdStartCallback,
			ExitAppFunc:    a.exitApp,
			Gateway:        a.Gateway,
			JsEngine:       a.JsEngine,
			ProcessMap:     a.ProcessMap,
		}
	}

	err = yaml.Unmarshal(contents, &result)
	if err != nil {
		return &workflow.StackupWorkflow{
			CommandStartCb: a.CmdStartCallback,
			ExitAppFunc:    a.exitApp,
			Gateway:        a.Gateway,
			JsEngine:       a.JsEngine,
			ProcessMap:     a.ProcessMap,
		}
	}

	result.State = &workflow.StackupWorkflowState{
		CurrentTask: nil,
		Stack:       linkedliststack.New(),
		History:     linkedliststack.New(),
	}

	result.CommandStartCb = a.CmdStartCallback
	result.ExitAppFunc = a.exitApp
	result.Gateway = a.Gateway
	result.JsEngine = a.JsEngine
	result.ProcessMap = a.ProcessMap

	return &result
}

func (a *Application) init() {
	a.Gateway = gateway.New([]string{}, []string{})
	a.ConfigFilename = support.FindExistingFile([]string{"stackup.dist.yaml", "stackup.yaml"}, "stackup.yaml")

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

	a.scheduledTaskMap = &sync.Map{}
	a.ProcessMap = &sync.Map{}
	a.Vars = &sync.Map{}

	a.Workflow = a.loadWorkflowFile(a.ConfigFilename)
	a.JsEngine = scripting.CreateNewJavascriptEngine(a.Vars, a.Gateway, a.GetWorkflowContract, a.GetApplicationIconPath)
	a.cronEngine = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))
	a.DownloadApplicationIcon()
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
	// 	// task := i.(*Task)
	// 	// runs := fmt.Sprintf("%v", task.RunCount)
	// 	//support.StatusMessageLine("[task history] task ran: "+task.GetDisplayName()+" ("+runs+" executions)", true)
	// }

	os.Exit(1)
}

func (a *Application) createScheduledTasks() {
	for _, def := range a.Workflow.Scheduler {
		task := a.Workflow.FindTaskById(def.TaskId())

		if task == nil {
			support.FailureMessageWithXMark("Task " + def.TaskId() + " not found.")
			continue
		}

		cron := def.Cron
		taskId := def.TaskId()

		a.cronEngine.AddFunc(cron, func() {
			task := a.Workflow.FindTaskById(taskId)
			task.Run(true)
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
		task := a.Workflow.FindTaskById(def.TaskId())

		if task == nil {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		a.Workflow.State.CurrentTask = task

		//GetState().Stack.Push(task)
		task.Run(true)
		//GetState().Stack.Pop()
	}
}

func (a *Application) runShutdownTasks() {
	for _, def := range a.Workflow.Shutdown {
		task := a.Workflow.FindTaskById(def.TaskId())

		if task == nil {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		a.Workflow.State.CurrentTask = task
		task.Run(true)
	}
}

func (a *Application) runServerTasks() {
	for _, def := range a.Workflow.Servers {
		task := a.Workflow.FindTaskById(def.TaskId())

		if task == nil {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		a.Workflow.State.CurrentTask = task
		task.Run(false)
	}
}

func (a *Application) runPrecondition(c *workflow.WorkflowPrecondition) bool {
	result := true

	if c.Check != "" {
		if (c.Attempts - 1) > *c.MaxRetries {
			support.FailureMessageWithXMark(c.Name)
			return false
		}

		c.Attempts++

		result = a.JsEngine.Evaluate(c.Check).(bool)

		if !result && len(c.OnFail) > 0 {
			support.FailureMessageWithXMark(c.Name)
			rerunCheck := c.HandleOnFailure()

			if rerunCheck {
				return a.runPrecondition(c)
			}

			return false
		}

		if !result {
			support.FailureMessageWithXMark(c.Name)
			return false
		}
	}

	return result
}

func (a *Application) runPreconditions() {
	for _, c := range a.Workflow.Preconditions {
		if !a.runPrecondition(c) {
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
          - '*.github.com'
      gateway:
        content-types:
          blocked:
          allowed:
            - application/json
            - text/*

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
	updateAvailable, release := updater.IsLatestApplicationReleaseNewerThanCurrent(a.Workflow.Cache, version.APP_VERSION, "permafrost-dev/stackup")

	if updateAvailable {
		support.WarningMessage(fmt.Sprintf("A new version of StackUp is available, released %s.", release.TimeSinceRelease))
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

func (a *Application) GetWorkflowContract() interface{} {
	return a.Workflow
}

func (a *Application) GetApplicationIconPath() string {
	return path.Join(a.GetConfigurationPath(), "/stackup-icon.png")
}

func (a *Application) Run() {
	godotenv.Load()
	a.init()
	a.handleFlagOptions()

	a.Analytics = telemetry.New(true, a.Gateway)
	a.Workflow.Initialize()
	a.Gateway.SetAllowedDomains(a.Workflow.Settings.Domains.Allowed)

	if *a.Workflow.Settings.AnonymousStatistics {
		a.Analytics.EventOnly("app.start")
	}

	if len(a.Workflow.Settings.DotEnvFiles) > 0 {
		godotenv.Load(a.Workflow.Settings.DotEnvFiles...)
	}
	a.JsEngine.CreateEnvironmentVariables()
	a.JsEngine.CreateAppVariables(a.Vars)

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
