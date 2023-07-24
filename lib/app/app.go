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
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
	"github.com/stackup-app/stackup/lib/version"
	"gopkg.in/yaml.v2"
)

var App *Application

type CommandCallback func(cmd *exec.Cmd)
type AppFlags struct {
	DisplayHelp    *bool
	DisplayVersion *bool
	ConfigFile     *string
}

type Application struct {
	Workflow            *StackupWorkflow
	JsEngine            *JavaScriptEngine
	cronEngine          *cron.Cron
	scheduledTaskMap    sync.Map
	ProcessMap          sync.Map
	flags               AppFlags
	CmdStartCallback    CommandCallback
	KillCommandCallback CommandCallback
}

func (a *Application) loadWorkflowFile(filename string) StackupWorkflow {
	var result StackupWorkflow

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return StackupWorkflow{}
	}

	err = yaml.Unmarshal(contents, &result)
	if err != nil {
		return StackupWorkflow{}
	}

	result.State.CurrentTask = nil

	return result
}

func (a *Application) init() {
	a.flags = AppFlags{
		DisplayHelp:    flag.Bool("help", false, "Display help"),
		DisplayVersion: flag.Bool("version", false, "Display version"),
		ConfigFile:     flag.String("config", "stackup.yaml", "Load a specific config file"),
	}

	flag.Parse()

	a.scheduledTaskMap = sync.Map{}
	a.ProcessMap = sync.Map{}

	workflow := a.loadWorkflowFile(*a.flags.ConfigFile)
	a.Workflow = &workflow
	jsEngine := CreateNewJavascriptEngine()
	a.JsEngine = &jsEngine
	a.cronEngine = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))

	for _, task := range a.Workflow.Tasks {
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
	os.Exit(1)
}

func (a *Application) createScheduledTasks() {
	for _, def := range a.Workflow.Scheduler {
		task := a.Workflow.FindTaskById(def.Task)

		if task == nil {
			support.FailureMessageWithXMark("Task " + def.Task + " not found.")
			continue
		}

		cron := def.Cron
		taskId := def.Task

		a.cronEngine.AddFunc(cron, func() {
			task := a.Workflow.FindTaskById(taskId)
			task.Run(true)
		})

		a.scheduledTaskMap.Store(def.Task, &def)
	}

	a.cronEngine.Start()
}

func (a *Application) stopServerProcesses() {
	a.ProcessMap.Range(func(key any, value any) bool {
		support.StatusMessage("Stopping "+key.(string)+"...", false)
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
		task := a.Workflow.FindTaskById(def.Task)

		if task == nil {
			support.SkippedMessageWithSymbol("Task " + def.Task + " not found.")
			continue
		}

		task.Run(true)
	}
}

func (a *Application) runShutdownTasks() {
	for _, def := range a.Workflow.Shutdown {
		task := a.Workflow.FindTaskById(def.Task)

		if task == nil {
			support.SkippedMessageWithSymbol("Task " + def.Task + " not found.")
			continue
		}

		task.Run(true)
	}
}

func (a *Application) runServerTasks() {
	for _, def := range a.Workflow.Servers {
		task := a.Workflow.FindTaskById(def.Task)

		if task == nil {
			support.SkippedMessageWithSymbol("Task " + def.Task + " not found.")
			continue
		}

		task.Run(false)
	}
}

func (a *Application) runPreconditions() {
	for _, c := range a.Workflow.Preconditions {
		if c.Check != "" {
			result := a.JsEngine.Evaluate(c.Check)

			if result != nil && !result.(bool) {
				support.FailureMessageWithXMark(c.Name)
				os.Exit(1)
			}
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

func (a *Application) Run() {
	godotenv.Load()
	a.init()

	if *a.flags.DisplayHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *a.flags.DisplayVersion {
		fmt.Println("StackUp version " + version.APP_VERSION)
		os.Exit(0)
	}

	if os.Args[1] == "init" {
		a.createNewConfigFile()
		os.Exit(0)
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
