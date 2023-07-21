package app

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/eiannone/keyboard"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/stackup-app/stackup/lib/config"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/utils"
	"github.com/stackup-app/stackup/lib/version"
	"github.com/stackup-app/stackup/lib/workflows"
)

type CommandCallback func(cmd *exec.Cmd)

type AppFlags struct {
	DisplayHelp    *bool
	DisplayVersion *bool
	ConfigFile     *string
}

type App struct {
	cfg                 config.Configuration
	workflow            workflows.StackupWorkflow
	jsEngine            scripting.JavaScriptEngine
	cronEngine          *cron.Cron
	scheduledTaskMap    sync.Map
	processMap          sync.Map
	flags               AppFlags
	CmdStartCallback    CommandCallback
	KillCommandCallback CommandCallback
}

func (a *App) init() {
	a.flags = AppFlags{
		DisplayHelp:    flag.Bool("help", false, "Display help"),
		DisplayVersion: flag.Bool("version", false, "Display version"),
		ConfigFile:     flag.String("config", "stackup.yaml", "Load a specific config file"),
	}

	flag.Parse()

	a.scheduledTaskMap = sync.Map{}
	a.processMap = sync.Map{}

	a.cfg = config.NewConfiguration()
	a.workflow = workflows.LoadWorkflowFile(config.FindExistingConfigurationFile(*a.flags.ConfigFile))
	a.jsEngine = scripting.CreateNewJavascriptEngine()
	a.cronEngine = cron.New(
		cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)),
		//cron.WithParser(cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)),
		//cron.WithParser(cron.NewParser(cron.Second|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow)),
	)
}

func (a *App) hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		<-c
		a.exitApp()
	}()
}

func (a *App) hookKeyboard() {
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

func (a *App) exitApp() {
	a.cronEngine.Stop()
	a.stopServerProcesses()
	support.StatusMessageLine("Running shutdown tasks...", true)
	a.runShutdownTasks()
	os.Exit(1)
}

func (a *App) createScheduledTasks() {
	for _, def := range a.workflow.Scheduler {
		cron := def.Cron
		command := def.Command
		cwd := def.Cwd
		name := def.Name

		a.cronEngine.AddFunc(cron, func() {
			if cwd != "" && a.jsEngine.IsEvaluatableScriptString(cwd) {
				tempCwd := a.jsEngine.Evaluate(a.jsEngine.GetEvaluatableScriptString(cwd))
				cwd = tempCwd.(string)
			}

			if a.jsEngine.IsEvaluatableScriptString(command) {
				a.jsEngine.Evaluate(a.jsEngine.GetEvaluatableScriptString(command))
			} else {
				utils.RunCommandInPath(command, cwd, false)
			}

			support.SuccessMessageWithCheck(name)
		})

		a.scheduledTaskMap.Store(def.Name, &def)
	}
}

func (a *App) startServerProcesses() {
	for _, def := range a.workflow.Servers {

		if a.jsEngine.IsEvaluatableScriptString(def.Cwd) {
			script := a.jsEngine.GetEvaluatableScriptString(def.Cwd)
			tempCwd := a.jsEngine.Evaluate(script)
			def.Cwd = tempCwd.(string)
		}

		if def.Platforms != nil {
			foundPlatform := false
			for _, name := range def.Platforms {
				if strings.EqualFold(runtime.GOOS, name) {
					foundPlatform = true
					break
				}
			}

			if !foundPlatform {
				support.WarningMessage("Skipping " + def.Name + ", it is not supported on this operating system.")
				continue
			}
		}

		support.StatusMessage(def.Name+"...", false)

		cmd, _ := utils.StartCommand(def.Command, def.Cwd)
		a.CmdStartCallback(cmd)
		cmd.Start()

		support.PrintCheckMarkLine()

		a.processMap.Store(def.Name, cmd)
	}
}

func (a *App) stopServerProcesses() {
	var stopServer = func(key any, value any) {
		support.StatusMessage("Stopping "+key.(string)+"...", false)

		a.KillCommandCallback(value.(*exec.Cmd))

		support.PrintCheckMarkLine()
	}

	a.processMap.Range(func(key any, value any) bool {
		stopServer(key, value)
		return true
	})
}

func (a *App) runEventLoop() {
	support.StatusMessageLine("Event loop executing.", true)

	for {
		utils.WaitForStartOfNextMinute()
	}
}

func (a *App) runTask(task *workflows.Task) {
	if a.jsEngine.IsEvaluatableScriptString(task.Cwd) {
		script := a.jsEngine.GetEvaluatableScriptString(task.Cwd)
		tempCwd := a.jsEngine.Evaluate(script)
		task.Cwd = tempCwd.(string)
	}

	if task.If != "" {
		result := a.jsEngine.Evaluate(task.If)

		if result != nil && !result.(bool) {
			support.SkippedMessageWithSymbol(task.Name)
			return
		}
	}

	if a.jsEngine.IsEvaluatableScriptString(task.Command) {
		a.jsEngine.Evaluate(a.jsEngine.GetEvaluatableScriptString(task.Command))
		support.SuccessMessageWithCheck(task.Name)
		return
	}

	runningSilently := reflect.TypeOf(task.Silent).Kind() == reflect.Bool && task.Silent == true

	support.StatusMessage(task.Name+"...", false)

	cmd := utils.RunCommandInPath(task.Command, task.Cwd, runningSilently)

	if cmd != nil {
		if runningSilently {
			support.PrintCheckMarkLine()
			return
		}
		support.SuccessMessageWithCheck(task.Name)
		return
	}

	if runningSilently {
		support.PrintXMarkLine()
		return
	}
	support.FailureMessageWithXMark(task.Name)
}

func (a *App) runStartupTasks() {
	for _, task := range a.workflow.Tasks {
		if strings.EqualFold(task.On, "startup") {
			a.runTask(&task)
		}
	}
}

func (a *App) runShutdownTasks() {
	for _, task := range a.workflow.Tasks {
		if strings.EqualFold(task.On, "shutdown") {
			a.runTask(&task)
		}
	}
}

func (a *App) runPreconditions() {
	for _, c := range a.workflow.Preconditions {
		if c.Check != "" {
			result := a.jsEngine.Evaluate(c.Check)

			if result != nil && !result.(bool) {
				support.FailureMessageWithXMark(c.Name)
				os.Exit(1)
			}
		}

		support.SuccessMessageWithCheck(c.Name)
	}
}

func (a *App) Run() {
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
	a.startServerProcesses()

	support.StatusMessageLine("Waiting for the start of the next minute to begin event loop...", true)
	utils.WaitForStartOfNextMinute()

	support.StatusMessage("Creating scheduled jobs...", false)
	a.createScheduledTasks()
	a.cronEngine.Start()
	support.PrintCheckMarkLine()

	a.runEventLoop()
}