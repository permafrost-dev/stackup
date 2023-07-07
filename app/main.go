package main

import (
	"bufio"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/stackup-app/stackup/lib"
	"github.com/stackup-app/stackup/support"
	"github.com/stackup-app/stackup/utils"
	"github.com/stackup-app/stackup/workflows"
)

var (
	seedDatabase = flag.Bool("seed", false, "Seed the database")
	displayHelp  = flag.Bool("help", false, "Display help")

	frontendProjectPath = filepath.Join("..", "acd-pos-frontend")
	backendProjectPath  = filepath.Dir(os.Args[0])

	workflow = workflows.LoadWorkflowFile("stack-supervisor.config.dev.yaml")

	jsengine = lib.CreateNewJavascriptEngine()

	processes sync.Map
)

func hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)

	go func() {
		<-c
		stopServerProcesses()
		os.Exit(0)
	}()
}

func startServerProcesses(serverDefs []workflows.Server) {
	for _, def := range serverDefs {
		go func(def workflows.Server) {
			// time.Sleep(def.delay)

			def.Cwd = jsengine.Evaluate(def.Cwd).(string)
			if def.Message != "" {
				support.StatusMessageLine(def.Message, false)
			} else {
				support.StatusMessage("Starting "+def.Name+"...", false)
			}

			cmd, _ := utils.StartCommand(def.Command)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Start()

			if cmd.ProcessState.Exited() && !cmd.ProcessState.Success() {
				support.FailureMessageWithXMark("Failed to start process " + def.Name)
				stopServerProcesses()
				os.Exit(1)
			}

			processes.Store(def.Name, cmd)
		}(def)
	}
}

func stopServerProcesses() {
	processes.Range(func(key, value interface{}) bool {
		support.StatusMessage("Stopping "+key.(string)+"...", false)
		value.(*os.Process).Kill()
		support.PrintCheckMark()
		return true
	})

	support.StatusMessage("Stopping containers...", true)
	utils.RunCommand("podman-compose down")
	support.PrintCheckMark()
}

func main() {
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

func runStartupTasks(tasks []workflows.Task) {
	for _, task := range tasks {

		if task.If != "" {
			result := jsengine.Evaluate(task.If)

			if result != nil && !result.(bool) {
				support.StatusMessageLine("Task "+task.Name+" skipped", true)
				continue
			}
		}

		cmd := utils.RunCommand(task.Command)
		if cmd.ProcessState.Success() {
			support.SuccessMessageWithCheck(task.Name)
		} else {
			support.FailureMessageWithXMark(task.Name)
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
