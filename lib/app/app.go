package app

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eiannone/keyboard"
	"github.com/emirpasic/gods/stacks/linkedliststack"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/updater"
	"github.com/stackup-app/stackup/lib/utils"
	"github.com/stackup-app/stackup/lib/version"
	"gopkg.in/yaml.v2"
)

var App *Application

type CommandCallback func(cmd *exec.Cmd)
type AppFlags struct {
	DisplayHelp    *bool
	DisplayVersion *bool
	NoUpdateCheck  *bool
	ConfigFile     *string
}

type Application struct {
	Workflow            *StackupWorkflow
	JsEngine            *JavaScriptEngine
	cronEngine          *cron.Cron
	scheduledTaskMap    *sync.Map
	ProcessMap          *sync.Map
	Vars                *sync.Map
	flags               AppFlags
	CmdStartCallback    CommandCallback
	KillCommandCallback CommandCallback
	ConfigFilename      string
	Gatekeeper          *Gatekeeper
}

func (a *Application) loadWorkflowFile(filename string) *StackupWorkflow {
	var result StackupWorkflow

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return &StackupWorkflow{}
	}

	err = yaml.Unmarshal(contents, &result)
	if err != nil {
		return &StackupWorkflow{}
	}

	result.State = &StackupWorkflowState{
		CurrentTask: nil,
		Stack:       linkedliststack.New(),
		History:     linkedliststack.New(),
	}

	return &result
}

func (a *Application) init() {
	a.Gatekeeper = CreateGatekeeper()
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
	a.JsEngine = CreateNewJavascriptEngine()
	a.cronEngine = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))
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

		App.Workflow.State.CurrentTask = task

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

		App.Workflow.State.CurrentTask = task
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

		App.Workflow.State.CurrentTask = task
		task.Run(false)
	}
}

func (a *Application) runPrecondition(c *Precondition) bool {
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

	filename := "stackup.yaml"
	contents := `name: my stack
    description: application stack
    version: 1.0.0

    preconditions:
      - name: dependencies are installed
        check: binaryExists("php")

    startup:
      - task: start-containers

    shutdown:
      - task: stop-containers

    servers:

    scheduler:

    tasks:
      - name: spin up containers
        id: start-containers
        if: exists(getCwd() + "/docker-compose.yml")
        command: docker-compose up -d
        silent: true

      - name: stop containers
        id: stop-containers
        if: exists(getCwd() + "/docker-compose.yml")
        command: docker-compose down
        silent: true
    `
	ioutil.WriteFile(filename, []byte(contents), 0644)
}

func (a *Application) checkForApplicationUpdates() {
	if a.Gatekeeper.CanAccessUrl(updater.GetUpdateCheckUrlFormat()) {
		updateAvailable := updater.IsLatestApplicationReleaseNewerThanCurrent(version.APP_VERSION, "permafrost-dev/stackup")

		if updateAvailable {
			support.WarningMessage("A new version of StackUp is available. Please update to the latest version.")
		}
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

func (a *Application) Run() {
	godotenv.Load()
	a.init()
	a.handleFlagOptions()

	a.Workflow.Initialize()
	if len(a.Workflow.Settings.DotEnvFiles) > 0 {
		godotenv.Load(a.Workflow.Settings.DotEnvFiles...)
	}
	a.JsEngine.CreateEnvironmentVariables()
	a.JsEngine.CreateAppVariables()

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
