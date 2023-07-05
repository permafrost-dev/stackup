package lib

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/permafrost-dev/stack-supervisor/state"
	"github.com/permafrost-dev/stack-supervisor/utils"
	"github.com/robertkrimen/otto"
	"github.com/robfig/cron/v3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type Application struct {
	//Name string
	Config             *StackConfig
	Processes          map[string]*exec.Cmd
	ProcessDefinitions []ProcessDefinition
	State              *state.AppState
	CurrentCommand     *cobra.Command
	CronEngine         *cron.Cron
	Jsvm               *otto.Otto
}

func (app *Application) LoadStackConfig(filename string) error {
	var result StackConfig

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(contents, &result)
	if err != nil {
		return err
	}

	tempMap := make(map[string]interface{})
	if err := yaml.Unmarshal(contents, &tempMap); err != nil {
		return err
	}
	result.Props = tempMap

	app.Config = &result

	return nil
}

func (app *Application) getContainerComposerBinaryPath() string {
	return app.Config.Applications.Orchestrator
}

func (app *Application) hasRunningPodmanContainersForProject() bool {
	// cmd := utils.RunCommand(app.getContainerManagerBinaryPath() + " ps --format json")
	// output, _ := cmd.CombinedOutput()

	// var items []PodmanRunningContainerInfo
	// json.Unmarshal(output, &items)

	// fmt.Printf("%v", output)

	// var podmanContainers []string

	// for _, data := range items {
	// 	for k, v := range data.Labels {
	// 		if k == "com.project.stack" && v == "acd-pos-stack" {
	// 			podmanContainers = append(podmanContainers, data.ID)
	// 		}
	// 	}
	// }

	// var activeContainers []PodmanRunningContainerInfo = make([]PodmanRunningContainerInfo, 0)

	// for _, container := range items {
	// 	if container.Exited == false {
	// 		activeContainers = append(activeContainers, container)
	// 	}
	// }

	// return len(activeContainers) > 0
	return false
}

func (app *Application) addProcess(name string, cmd *exec.Cmd) {
	app.Processes[name] = cmd
}

func (app *Application) stopProcesses() {
	for name, p := range app.Processes {
		fmt.Print(aurora.White("Killing " + name + "..."))
		p.Process.Kill()
		p.Wait()
		// p.Process.Signal(os.Interrupt)

		// if p.ProcessState.Exited() {
		fmt.Println(aurora.Green("✓"))
		delete(app.Processes, name)
	}

}

func (app *Application) stopContainers() {
	fmt.Println("Stopping containers...")
	utils.RunCommand(app.getContainerComposerBinaryPath() + " down")
}

func (app *Application) startProcesses() {

	for _, def := range app.Config.Stack.Processes {
		// if def.RunsOnWin == false {
		// 	continue
		// }

		cmd := exec.Command(def.Bin, strings.Split(def.Command, " ")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = def.Cwd
		cmd.Dir = strings.ReplaceAll(cmd.Dir, "FRONTEND_PATH", app.Config.FindDefinition("FRONTEND_PATH").Value)
		cmd.Dir = strings.ReplaceAll(cmd.Dir, "BACKEND_PATH", app.Config.FindDefinition("BACKEND_PATH").Value)

		if def.Delay > 0 {
			time.Sleep(time.Until(time.Now().Add(time.Millisecond * def.Delay)))
		}

		err := cmd.Start()

		if err != nil {
			fmt.Println(err)
			fmt.Println(`Failed while spawning process for "` + def.Name + `".`)
			fmt.Println(`Stopping all processes and exiting.`)
			app.stopProcesses()
			app.stopContainers()
			os.Exit(1)
		}

		app.addProcess(def.Name, cmd)
	}
}

func (app *Application) InitJavascriptEngine() {
	if app.Jsvm != nil {
		return
	}

	app.Jsvm = otto.New()

	// Define a JavaScript function in Go
	app.Jsvm.Set("sayHello", func(call otto.FunctionCall) otto.Value {
		fmt.Printf("Hello, %s.\n", call.Argument(0).String())
		return otto.Value{}
	})

	// // Call the JavaScript function from Go
	// vm.Run(`
	// 	sayHello("World");
	// `)
}

func (app *Application) Init() {
	app.InitJavascriptEngine()

	app.CronEngine = cron.New(cron.WithParser(
		cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow),
	))

	for idx := range app.Config.Stack.Tasks {
		app.Config.Stack.Tasks[idx].Init(app.Config)
	}
	for idx := range app.Config.Definitions {
		app.Config.Definitions[idx].Init(app.Config)
	}
	for idx := range app.Config.Stack.Checks {
		app.Config.Stack.Checks[idx].Init(app.Config)
	}

}

func (app *Application) runStartupTasks() {
	for _, task := range app.Config.Stack.Tasks {

		task.Evaluate(app)

		if !task.Result {
			fmt.Println(aurora.BrightYellow("Task '" + task.Name + "' skipped!"))
		} else {

			cmd := utils.RunCommand(task.Bin + " " + task.Command)

			if cmd.ProcessState.Success() {
				fmt.Println(aurora.BrightGreen(fmt.Sprintf("Task '%s' succeeded.", task.Name)))
			} else {
				fmt.Println(aurora.BrightYellow(fmt.Sprintf("Task '%s' failed.", task.Name)))
			}
		}
	}
}

func (app *Application) hookSignals() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGSEGV)

	go func() {
		<-c
		app.stopScheduledTasks()
		app.stopProcesses()
		app.stopContainers()
		os.Exit(1)
	}()
}

func (app *Application) CurrentCommandHasFlag(flag string) bool {
	result, err := app.CurrentCommand.Flags().GetBool("skip-checks")

	if err != nil {
		result = false
	}

	return result
}

func (app *Application) startScheduledTasks() {
	if len(app.Config.Stack.Scheduled) > 0 {
		fmt.Print(aurora.White("Initializing scheduled tasks..."))
		for _, scheduled := range app.Config.Stack.Scheduled {
			scheduled.Init(app)
		}
		fmt.Println(aurora.Green("done ✓"))
	}

	app.CronEngine.Start()
}

func (app *Application) stopScheduledTasks() {
	if len(app.Config.Stack.Scheduled) > 0 {
		fmt.Print(aurora.White("Stopping scheduled tasks..."))
		for _, scheduled := range app.Config.Stack.Scheduled {
			scheduled.Stop()
		}
		fmt.Println(aurora.Green("done ✓"))
	}
}

func (app *Application) Run(cmd *cobra.Command) {
	app.Processes = make(map[string]*exec.Cmd)
	app.State = state.NewAppState("1.0")
	app.CurrentCommand = cmd

	app.hookSignals()

	if app.Config.Options.Dotenv {
		utils.LoadEnv()
	}

	// evaluate any scripts in the value props
	for _, def := range app.Config.Definitions {
		def.Evaluate()
	}

	skipChecks := app.CurrentCommandHasFlag("skip-checks")

	if !skipChecks {
		for _, check := range app.Config.Stack.Checks {
			for _, def := range app.Config.Definitions {
				check.Check = strings.ReplaceAll(check.Check, def.Name, def.Value)
			}
			check.Evaluate(app)

			if !check.Result {
				fmt.Println(aurora.BrightYellow("Check '" + check.Name + "' failed, exiting."))
				os.Exit(1)
			}

			fmt.Print(check.Name + " ")
			fmt.Println(aurora.Green("✓"))
		}
	}

	if app.hasRunningPodmanContainersForProject() {
		fmt.Println(aurora.BrightYellow("There are running containers for this project, stopping containers..."))
		app.stopContainers()
	}

	fmt.Println(aurora.BrightGreen("Starting project containers..."))

	app.Init()
	app.runStartupTasks()
	app.startProcesses()
	app.startScheduledTasks()

	fmt.Printf(aurora.Sprintf(aurora.White(" > Waiting for the start of the next minute before starting event loop... (%v sec)"), time.Until(time.Now().Truncate(time.Minute).Add(time.Minute))))
	time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
	fmt.Println(" > Starting event loop for artisan scheduled tasks, executing once per minute at the start of each minute.")

	for {
		for _, elt := range app.Config.Stack.EventLoop {
			utils.RunCommand(strings.TrimSpace(elt.Bin + " " + elt.Command))
		}
		time.Sleep(time.Until(time.Now().Truncate(time.Minute).Add(time.Minute)))
	}
}
