package main

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
	"time"

	"github.com/eiannone/keyboard"
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
	jsEngine    = lib.CreateNewJavascriptEngine()

	cronEngine = cron.New(
	//cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger)),
	//cron.WithParser(cron.NewParser(cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)),
	//cron.WithParser(cron.NewParser(cron.Second|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow)),
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
		// _, found := scheduledTaskMap.Load(def.Name)

		// if found {
		// 	continue
		// }

		cronEngine.AddFunc(def.Cron, func() {
			if def.Cwd != "" && jsEngine.IsEvaluatableScriptString(def.Cwd) {
				tempCwd := jsEngine.Evaluate(jsEngine.GetEvaluatableScriptString(def.Cwd))
				def.Cwd = tempCwd.(string)
			}

			if jsEngine.IsEvaluatableScriptString(def.Command) {
				jsEngine.Evaluate(jsEngine.GetEvaluatableScriptString(def.Command))
			} else {
				utils.RunCommandCwd(def.Command, def.Cwd)
			}

			support.SuccessMessageWithCheck(def.Name)
		})

		// scheduledTaskMap.Store(def.Name, &def)
	}
}

func startServerProcesses(serverDefs []workflows.Server) {
	for _, def := range serverDefs {

		if jsEngine.IsEvaluatableScriptString(def.Cwd) {
			script := jsEngine.GetEvaluatableScriptString(def.Cwd)
			tempCwd := jsEngine.Evaluate(script)
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

	support.StatusMessageLine("Running startup tasks...", true)
	runStartupTasks(workflow.Tasks)

	support.StatusMessageLine("Starting server processes...", true)
	startServerProcesses(workflow.Servers)

	support.StatusMessageLine("Waiting for the start of the next minute to begin event loop...", true)
	waitForStartOfNextMinute()

	createScheduledTasks(workflow.Scheduler)
	cronEngine.Start()

	runEventLoop(true)
}

func runEventLoop(showStatusMessages bool) {
	if showStatusMessages {
		support.StatusMessageLine("Event loop executing at 1 min intervals, at the start of each minute.", true)
	}

	for {
		time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
	}
}

func hookKeyboard() {
	//reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			// b, _ := reader.ReadString(' ')

			// if b == 'r' {
			// 	fmt.Println("r pressed")
			// }
			char, key, err := keyboard.GetSingleKey()

			if err != nil {
				return
			}

			if char == 'r' {
				fmt.Println("r pressed")
			}

			if key == keyboard.KeyCtrlC {
				//
			}
		}
	}()
}

func waitForStartOfNextMinute() {
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
}

func runTask(task *workflows.Task) {
	if jsEngine.IsEvaluatableScriptString(task.Cwd) {
		script := jsEngine.GetEvaluatableScriptString(task.Cwd)
		tempCwd := jsEngine.Evaluate(script)
		task.Cwd = tempCwd.(string)
	}

	if task.If != "" {
		result := jsEngine.Evaluate(task.If)

		if result != nil && !result.(bool) {
			support.SkippedMessageWithSymbol(task.Name)
			return
		}
	}

	if jsEngine.IsEvaluatableScriptString(task.Command) {
		jsEngine.Evaluate(jsEngine.GetEvaluatableScriptString(task.Command))
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
		if strings.EqualFold(task.On, "startup") {
			runTask(&task)
		}
	}
}

func runShutdownTasks(tasks []workflows.Task) {
	for _, task := range tasks {
		if strings.EqualFold(task.On, "shutdown") {
			runTask(&task)
		}
	}
}

func runPreconditions(checks []workflows.Precondition) {
	for _, c := range checks {
		if c.Check != "" {
			result := jsEngine.Evaluate(c.Check)

			if result != nil && !result.(bool) {
				support.FailureMessageWithXMark(c.Name)
				os.Exit(1)
			}
		}

		support.SuccessMessageWithCheck(c.Name)
	}
}
