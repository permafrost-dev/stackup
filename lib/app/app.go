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
	"github.com/stackup-app/stackup/lib/scripting"
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
	workflow            StackupWorkflow
	JsEngine            *scripting.JavaScriptEngine
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

	a.workflow = a.loadWorkflowFile(*a.flags.ConfigFile)
	jsEngine := scripting.CreateNewJavascriptEngine()
	a.JsEngine = &jsEngine
	a.cronEngine = cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)))

	for _, task := range a.workflow.Tasks {
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

			if char == 'r' {
				fmt.Println("r pressed")
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
	for _, def := range a.workflow.Scheduler {
		task := a.workflow.FindTaskById(def.Task)

		if task == nil {
			support.FailureMessageWithXMark("Task " + def.Task + " not found.")
			continue
		}

		cron := def.Cron
		taskId := def.Task

		a.cronEngine.AddFunc(cron, func() {
			task := a.workflow.FindTaskById(taskId)
			task.Run(true)
		})

		a.scheduledTaskMap.Store(def.Task, &def)
	}

	a.cronEngine.Start()
}

func (a *Application) stopServerProcesses() {
	var stopServer = func(key any, value any) {
		support.StatusMessage("Stopping "+key.(string)+"...", false)

		a.KillCommandCallback(value.(*exec.Cmd))

		support.PrintCheckMarkLine()
	}

	a.ProcessMap.Range(func(key any, value any) bool {
		stopServer(key, value)
		return true
	})
}

func (a *Application) runEventLoop() {
	for {
		utils.WaitForStartOfNextMinute()
	}
}

// func (a *Application) runTask(task *Task, synchronous bool) {
// 	if task.RunCount >= task.MaxRuns && task.MaxRuns > 0 {
// 		support.SkippedMessageWithSymbol(task.Name)
// 		return
// 	}

// 	task.RunCount++

// 	if a.JsEngine.IsEvaluatableScriptString(task.Path) {
// 		tempCwd := a.JsEngine.Evaluate(task.Path)
// 		task.Path = tempCwd.(string)
// 	}

// 	if !task.CanRunConditionally() {
// 		support.SkippedMessageWithSymbol(task.Name)
// 		return
// 	}

// 	if !task.CanRunOnCurrentPlatform() {
// 		support.WarningMessage("Skipping " + task.Name + ", it is not supported on this operating system.")
// 		return
// 	}

// 	command := task.Command
// 	runningSilently := task.Silent == true

// 	if a.JsEngine.IsEvaluatableScriptString(command) {
// 		command = a.JsEngine.Evaluate(command).(string)
// 	}

// 	support.StatusMessage(task.Name+"...", false)

// 	if synchronous {
// 		a.runTaskSyncWithStatusMessages(task, command, runningSilently)
// 		return
// 	}

// 	cmd, _ := utils.StartCommand(command, task.Path)
// 	a.CmdStartCallback(cmd)
// 	cmd.Start()

// 	support.PrintCheckMarkLine()

// 	a.ProcessMap.Store(task.Name, cmd)
// }

// func (a *Application) runTaskSyncWithStatusMessages(task *Task, command string, runningSilently bool) {
// 	cmd := utils.RunCommandInPath(command, task.Path, runningSilently)

// 	if cmd != nil && runningSilently {
// 		support.PrintCheckMarkLine()
// 	} else if cmd != nil {
// 		support.SuccessMessageWithCheck(task.Name)
// 	}

// 	if cmd == nil && runningSilently {
// 		support.PrintXMarkLine()
// 	} else if cmd == nil {
// 		support.FailureMessageWithXMark(task.Name)
// 	}
// }

func (a *Application) runStartupTasks() {
	for _, def := range a.workflow.Startup {
		task := a.workflow.FindTaskById(def.Task)

		if task == nil {
			support.FailureMessageWithXMark("Task " + def.Task + " not found.")
			continue
		}

		task.Run(true)
	}
}

func (a *Application) runShutdownTasks() {
	for _, def := range a.workflow.Shutdown {
		task := a.workflow.FindTaskById(def.Task)

		if task == nil {
			support.FailureMessageWithXMark("Task " + def.Task + " not found.")
			continue
		}

		task.Run(true)
	}
}

func (a *Application) runServerTasks() {
	for _, def := range a.workflow.Servers {
		task := a.workflow.FindTaskById(def.Task)

		if task == nil {
			support.FailureMessageWithXMark("Task " + def.Task + " not found.")
			continue
		}

		task.Run(false)
	}
}

func (a *Application) runPreconditions() {
	for _, c := range a.workflow.Preconditions {
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
