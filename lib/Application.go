package lib

import (
	"fmt"
	"os/exec"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	jsengine "github.com/stackup-app/stackup/lib/javascriptEngine"
	"github.com/stackup-app/stackup/workflows"
)

type Application struct {
	//Name string
	Config             workflows.StackupWorkflow
	Processes          map[string]*exec.Cmd
	ProcessDefinitions []ProcessDefinition
	State              workflows.WorkflowState
	CurrentCommand     *cobra.Command
	CronEngine         *cron.Cron
	Jsvm               JavaScriptEngine
	JsFunctions        jsengine.JavascriptFunctions
}

func GetApplication() interface{} {
	result := ""
	fmt.Printf("%v", result)

	// result.Config = result.LoadStackConfig(utils.WorkingDir("/stackup.config.dev.yaml"))
	// utils.LoadEnv(".env")
	//result.hookSignals()
	//result.Init()

	return result
}

// func (app *Application) LoadStackConfig(filename string) workflows.StackupWorkflow {
// 	var result workflows.StackupWorkflow

// 	contents, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		return workflows.StackupWorkflow{}
// 	}

// 	err = yaml.Unmarshal(contents, &result)
// 	if err != nil {
// 		return workflows.StackupWorkflow{}
// 	}

// 	return result
// }

// func (app *Application) hasRunningPodmanContainersForProject() bool {
// 	// cmd := utils.RunCommand(app.getContainerManagerBinaryPath() + " ps --format json")
// 	// output, _ := cmd.CombinedOutput()

// 	// var items []PodmanRunningContainerInfo
// 	// json.Unmarshal(output, &items)

// 	// fmt.Printf("%v", output)

// 	// var podmanContainers []string

// 	// for _, data := range items {
// 	// 	for k, v := range data.Labels {
// 	// 		if k == "com.project.stack" && v == "acd-pos-stack" {
// 	// 			podmanContainers = append(podmanContainers, data.ID)
// 	// 		}
// 	// 	}
// 	// }

// 	// var activeContainers []PodmanRunningContainerInfo = make([]PodmanRunningContainerInfo, 0)

// 	// for _, container := range items {
// 	// 	if container.Exited == false {
// 	// 		activeContainers = append(activeContainers, container)
// 	// 	}
// 	// }

// 	// return len(activeContainers) > 0
// 	return false
// }

// func (app *Application) addProcess(name string, cmd *exec.Cmd) {
// 	app.Processes[name] = cmd
// }

// func (app *Application) stopProcesses() {
// 	for name, p := range app.Processes {
// 		fmt.Print(aurora.White("Killing " + name + "..."))
// 		p.Process.Kill()
// 		p.Wait()
// 		// p.Process.Signal(os.Interrupt)

// 		// if p.ProcessState.Exited() {
// 		fmt.Println(aurora.Green("✓"))
// 		delete(app.Processes, name)
// 	}

// }

// func (app *Application) startContainers() {
// 	fmt.Println("Starting containers...")
// 	utils.RunCommand("podman-compose up -d")
// }

// func (app *Application) stopContainers() {
// 	fmt.Println("Stopping containers...")
// 	utils.RunCommand("podman-compose down")
// }

// func (app *Application) startProcesses() {

// 	for _, def := range app.Config.Servers {
// 		parts := strings.Fields(def.Command)
// 		cmd := exec.Command(parts[0], parts[:1]...) //strings.SplitN(def.Command, " ", 1)[0], strings.SplitAfter(def.Command, " ")...)
// 		cmd.Stdout = os.Stdout
// 		cmd.Stderr = os.Stderr
// 		cmd.Dir = app.Jsvm.Evaluate(def.Cwd).(string)

// 		// if def.Delay > 0 {
// 		// 	time.Sleep(time.Until(time.Now().Add(time.Millisecond * def.Delay)))
// 		//
// 		err := cmd.Start()

// 		if err != nil {
// 			// def.StoppedAt = time.Now()
// 			fmt.Println(err)
// 			fmt.Println(`Failed while spawning process for "` + def.Name + `".`)
// 			fmt.Println(`Stopping all processes and exiting.`)
// 			app.stopProcesses()
// 			app.stopContainers()
// 			os.Exit(1)
// 		}

// 		app.addProcess(def.Name, cmd)
// 	}
// }

// func (app *Application) InitJavascriptEngine() *otto.Otto {
// 	engine := otto.New()
// 	app.JsFunctions = jsengine.NewJavascriptFunctions(engine)

// 	return engine
// }

// func (app *Application) Init() {
// 	//app.InitializeDefinitions()

// 	// for idx := range app.Config.Stack.Definitions {
// 	// 	app.Config.Definitions[idx].Init(app.Config)
// 	// }
// 	// for idx := range app.Config.Stack.Tasks {
// 	// 	app.Config.Stack.Tasks[idx].Init(app.Config)
// 	// }
// 	// for _, check := range app.Config.Stack.Checks {
// 	// 	check.Init(app.Config)
// 	// }
// }

// func (app *Application) runStartupTasks() {
// 	for _, task := range app.Config.Tasks {
// 		fmt.Printf("%v\n", task)

// 		result := app.Jsvm.Evaluate(task.If).(bool)

// 		if !result {
// 			fmt.Println(aurora.BrightYellow("Task '" + task.Name + "' skipped!"))
// 		} else {
// 			cmd := utils.RunCommand(task.Command)

// 			if cmd.ProcessState.Success() {
// 				fmt.Println(aurora.BrightGreen(fmt.Sprintf("Task '%s' succeeded.", task.Name)))
// 			} else {
// 				fmt.Println(aurora.BrightYellow(fmt.Sprintf("Task '%s' failed.", task.Name)))
// 			}
// 		}
// 	}
// }

// func (app *Application) hookSignals() {
// 	// c := make(chan os.Signal, 1)
// 	// signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGSEGV)

// 	// // Define a channel to signal when the cleanup is finished.
// 	// done := make(chan struct{})

// 	// go func() {
// 	// 	for sig := range c {
// 	// 		log.Printf("received signal: %v, stopping...", sig)

// 	// 		// Run the cleanup operations in a separate goroutine, and signal when they're finished.
// 	// 		go func() {
// 	// 			app.stopScheduledTasks()
// 	// 			app.stopProcesses()
// 	// 			app.stopContainers()
// 	// 			close(done)
// 	// 		}()

// 	// 		// Wait for the cleanup to finish, then exit.
// 	// 		<-done
// 	// 		os.Exit(1)
// 	// 	}
// 	// }()
// }

// func (app *Application) CurrentCommandHasFlag(flag string) bool {
// 	result, err := app.CurrentCommand.Flags().GetBool("skip-checks")

// 	if err != nil {
// 		result = false
// 	}

// 	return result
// }

// func (app *Application) startScheduledTasks() {
// 	app.CronEngine.Stop()

// 	if len(app.Config.Scheduler) > 0 {
// 		fmt.Print(aurora.White("Initializing scheduled tasks..."))
// 		for _, scheduled := range app.Config.Scheduler {
// 			app.CronEngine.AddFunc(scheduled.Cron, func() {
// 				utils.RunCommand(scheduled.Command)
// 			})
// 		}
// 		fmt.Println(aurora.Green("done ✓"))
// 	}

// 	app.CronEngine.Start()
// }

// func (app *Application) stopScheduledTasks() {
// 	if len(app.Config.Scheduler) > 0 {
// 		fmt.Print(aurora.White("Stopping scheduled tasks..."))
// 		// for _, scheduled := range app.Config.Scheduler {
// 		// 	//
// 		// }
// 		fmt.Println(aurora.Green("done ✓"))
// 	}

// 	for _, entry := range app.CronEngine.Entries() {
// 		app.CronEngine.Remove(entry.ID)
// 	}

// 	app.CronEngine.Stop()
// }

// func (app *Application) Run(cmd *cobra.Command) {
// 	// skipChecks := app.CurrentCommandHasFlag("skip-checks")

// 	// if !skipChecks {
// 	// 	for _, check := range app.Config.Stack.Checks {
// 	// 		// for _, def := range app.Config.Stack.Definitions {
// 	// 		// 	//def.Init(app.Config)
// 	// 		// 	check.Check = strings.ReplaceAll(check.Check, def.Name, def.Value)
// 	// 		// }
// 	// 		fmt.Println(">>> check: ", check.Check)

// 	// 		checkResult := check.Evaluate(app.JsFunctions.Vm)

// 	// 		if checkResult != "true" {
// 	// 			fmt.Println(aurora.BrightYellow("Check '" + check.Name + "' failed, exiting."))
// 	// 			os.Exit(1)
// 	// 		}

// 	// 		fmt.Print(check.Name + " ")
// 	// 		fmt.Println(aurora.Green("✓"))
// 	// 	}
// 	// }

// 	// if app.hasRunningPodmanContainersForProject() {
// 	// 	fmt.Println(aurora.BrightYellow("There are running containers for this project, stopping containers..."))
// 	// 	app.stopContainers()
// 	// }

// 	fmt.Println(aurora.BrightGreen("Starting project containers..."))

// 	// app.startContainers()
// 	// app.runStartupTasks()
// 	// app.startProcesses()
// 	// app.startScheduledTasks()

// 	// ctrs := containers.GetActivePodmanContainers()

// 	// for _, ctr := range ctrs {
// 	// 	fmt.Printf("%s -- %s \n", ctr.Name, ctr.ID)
// 	// }

// 	fmt.Printf(aurora.Sprintf(aurora.White(" > Waiting for the start of the next minute before starting event loop... (%v sec)"), time.Until(time.Now().Truncate(time.Minute).Add(time.Minute))))
// 	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
// 	fmt.Println(" > Starting event loop for artisan scheduled tasks, executing once per minute at the start of each minute.")

// 	// for {
// 	// 	for _, elt := range app.Config.Stack.EventLoop {
// 	// 		utils.RunCommand(strings.TrimSpace(elt.Bin + " " + elt.Command))
// 	// 	}
// 	// 	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
// 	// }

// 	fmt.Println("done!")
// }
