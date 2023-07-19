package main

import (
	"bufio"
	"flag"
	"os"
	"os/exec"
	"os/signal"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/stackup-app/stackup/config"
	"github.com/stackup-app/stackup/lib"
	"github.com/stackup-app/stackup/support"
	"github.com/stackup-app/stackup/utils"
	"github.com/stackup-app/stackup/workflows"
)

var (
	displayHelp = flag.Bool("help", false, "Display help")
	configFile  = flag.String("config", "stackup.yaml", "Load a specific config file")
	cfg         = config.NewConfiguration()
	workflow    = workflows.LoadWorkflowFile(config.FindExistingConfigurationFile(*configFile))
	jsengine    = lib.CreateNewJavascriptEngine()

	cronEngine = cron.New(
		cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)),
		cron.WithParser(cron.NewParser(cron.Second|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow)),
	)

	scheduledTaskMap = sync.Map{}
	processMap       = sync.Map{}
)

func hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		<-c
		cronEngine.Stop()
		stopServerProcesses()
		support.StatusMessageLine("Running shutdown tasks...", true)
		runShutdownTasks(workflow.Tasks)
		os.Exit(1)
	}()
}

func createScheduledTasks(defs []workflows.ScheduledTask) {
	for _, def := range defs {
		_, found := scheduledTaskMap.Load(def.Name)

		if found {
			continue
		}

		cronEngine.AddFunc(def.Cron, func() {
			if jsengine.IsEvaluatableScriptString(def.Command) {
				jsengine.Evaluate(jsengine.GetEvaluatableScriptString(def.Command))
			} else {
				utils.RunCommand(def.Command)
			}
			support.SuccessMessageWithCheck(def.Name)
		})

		scheduledTaskMap.Store(def.Name, &def)
	}
}

func startServerProcesses(serverDefs []workflows.Server) {
	for _, def := range serverDefs {
		// time.Sleep(def.delay)

		if jsengine.IsEvaluatableScriptString(def.Cwd) {
			script := jsengine.GetEvaluatableScriptString(def.Cwd)
			tempCwd := jsengine.Evaluate(script)
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

		cmd, _ := utils.StartCommand(def.Command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = def.Cwd
		cmd.Start()

		support.PrintCheckMarkLine()

		// if cmd.ProcessState.Exited() && !cmd.ProcessState.Success() {
		// 	support.FailureMessageWithXMark("Failed to start process " + def.Name)
		// 	stopServerProcesses()
		// 	os.Exit(1)
		// }

		processMap.Store(def.Name, cmd)
	}
}

func stopServerProcesses() {
	var stopServer = func(key any, value any) {
		support.StatusMessage("Stopping "+key.(string)+"...", false)
		value.(*exec.Cmd).Process.Kill()
		support.PrintCheckMarkLine()
	}

	processMap.Range(func(key any, value any) bool {
		stopServer(key, value)
		return true
	})

	// run shutdown commands
	// for _, cmd := range workflow.Commands {
	// 	if cmd.On != "shutdown" {
	// 		continue
	// 	}

	// 	support.StatusMessageLine("Running "+cmd.Name+"...", true)

	// 	if cmd.Silent {
	// 		utils.RunCommandSilent(cmd.Command)
	// 	} else {
	// 		utils.RunCommand(cmd.Command)
	// 	}

	// 	support.SuccessMessageWithCheck(cmd.Description)
	// }
}

func main() {
	godotenv.Load()

	flag.Parse()

	if *displayHelp {
		flag.Usage()
		os.Exit(0)
	}

	hookSignals()
	hookKeyboard()

	support.StatusMessageLine("Running precondition checks...", true)
	runPreconditions(workflow.Preconditions)

	createScheduledTasks(workflow.Scheduler)

	support.StatusMessageLine("Running startup tasks...", true)
	runStartupTasks(workflow.Tasks)

	support.StatusMessageLine("Starting server processes...", true)
	startServerProcesses(workflow.Servers)

	support.StatusMessageLine("Waiting for the start of the next minute to begin event loop...", true)
	waitForStartOfNextMinute()

	cronEngine.Start()

	runEventLoop(true, &workflow.EventLoop)
}

func runEventLoop(showStatusMessages bool, eventLoop *workflows.EventLoop) {
	if showStatusMessages {
		support.StatusMessageLine("Event loop executing at 1 min intervals, at the start of each minute.", true)
	}

	interval, _ := time.ParseDuration(eventLoop.Interval)

	for {
		for _, job := range eventLoop.Jobs {
			support.StatusMessageLine("Running job "+job.Name, true)
			job.Cwd = jsengine.Evaluate(job.Cwd).(string)
			utils.RunCommandEx(job.Command, job.Cwd)
		}

		time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(interval)))
	}
}

func hookKeyboard() {
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			reader.ReadByte()
		}
	}()
}

func waitForStartOfNextMinute() {
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
}

func runTask(task *workflows.Task) {
	if jsengine.IsEvaluatableScriptString(task.Cwd) {
		script := jsengine.GetEvaluatableScriptString(task.Cwd)
		tempCwd := jsengine.Evaluate(script)
		task.Cwd = tempCwd.(string)
	}

	if task.If != "" {
		result := jsengine.Evaluate(task.If)

		if result != nil && !result.(bool) {
			support.SkippedMessageWithSymbol(task.Name)
			return
		}
	}

	if jsengine.IsEvaluatableScriptString(task.Command) {
		jsengine.Evaluate(jsengine.GetEvaluatableScriptString(task.Command))
		support.SuccessMessageWithCheck(task.Name)
	} else {
		runningSilently := reflect.TypeOf(task.Silent).Kind() == reflect.Bool && task.Silent == true

		support.StatusMessage(task.Name+"...", false)
		cmd := utils.RunCommandInPath(task.Command, task.Cwd, runningSilently)

		if cmd != nil {
			if runningSilently {
				support.PrintCheckMarkLine()
				return
			}
			support.SuccessMessageWithCheck(task.Name)
		} else {
			if runningSilently {
				support.PrintXMarkLine()
				return
			}
			support.FailureMessageWithXMark(task.Name)
		}
	}
}

func runStartupTasks(tasks []workflows.Task) {
	for _, task := range tasks {
		if task.On == "startup" {
			runTask(&task)
		}
	}
}

func runShutdownTasks(tasks []workflows.Task) {
	for _, task := range tasks {
		if task.On == "shutdown" {
			runTask(&task)
		}
	}
}

func runPreconditions(checks []workflows.Precondition) {
	for _, c := range checks {
		if c.Check != "" {
			result := jsengine.Evaluate(c.Check)

			if result != nil && !result.(bool) {
				support.FailureMessageWithXMark(c.Name)
				os.Exit(1)

			}
		}

		support.SuccessMessageWithCheck(c.Name)
	}
}
