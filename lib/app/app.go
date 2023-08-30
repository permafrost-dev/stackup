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
	"github.com/stackup-app/stackup/lib/cache"
	"github.com/stackup-app/stackup/lib/consts"
	"github.com/stackup-app/stackup/lib/debug"
	"github.com/stackup-app/stackup/lib/gateway"
	"github.com/stackup-app/stackup/lib/scripting"
	"github.com/stackup-app/stackup/lib/support"
	"github.com/stackup-app/stackup/lib/telemetry"
	"github.com/stackup-app/stackup/lib/types"
	"github.com/stackup-app/stackup/lib/updater"
	"github.com/stackup-app/stackup/lib/utils"
	"github.com/stackup-app/stackup/lib/version"
	"gopkg.in/yaml.v2"
)

var App *Application

type Application struct {
	Workflow   *StackupWorkflow
	JsEngine   *scripting.JavaScriptEngine
	cronEngine *cron.Cron
	// scheduledTaskMap    *sync.Map
	ProcessMap          *sync.Map
	Vars                *sync.Map
	flags               AppFlags
	CmdStartCallback    types.CommandCallback
	KillCommandCallback types.CommandCallback
	ConfigFilename      string
	Gateway             *gateway.Gateway
	Analytics           *telemetry.Telemetry
}

func NewApplication() *Application {
	result := &Application{
		ProcessMap: &sync.Map{},
		Vars:       &sync.Map{},
		flags: AppFlags{
			DisplayHelp:    flag.Bool("help", false, "Display help"),
			DisplayVersion: flag.Bool("version", false, "Display version"),
			NoUpdateCheck:  flag.Bool("no-update-check", false, "Disable update check"),
			ConfigFile:     flag.String("config", "", "Load a specific config file"),
		},
		ConfigFilename: support.FindExistingFile([]string{"stackup.dist.yaml", "stackup.yaml"}, "stackup.yaml"),
		Gateway:        gateway.New(nil),
		cronEngine:     cron.New(cron.WithChain(cron.SkipIfStillRunning(cron.DiscardLogger))),
	}
	result.flags.app = result
	result.Workflow = CreateWorkflow(result.Gateway, result.ProcessMap)

	return result
}

func (a *Application) loadWorkflowFile(filename string, wf *StackupWorkflow) {
	wf.ExitAppFunc = a.exitApp
	wf.Gateway = a.Gateway
	wf.ProcessMap = a.ProcessMap
	wf.CommandStartCb = a.CmdStartCallback

	contents, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(contents, wf)
	if err != nil {
		fmt.Printf("error loading configuration file: %v", err)
		return
	}

	wf.State = WorkflowState{
		CurrentTask: nil,
		Stack:       linkedliststack.New(),
		History:     linkedliststack.New(),
	}

	wf.ConfigureDefaultSettings()

	if !wf.Debug {
		wf.Debug = os.Getenv("DEBUG") == "true" || os.Getenv("DEBUG") == "1"
	}
}

// parse command-line flags, load the workflow file, load .env files,
// initialize the workflow, gateway and js engine
func (a *Application) Initialize() {
	utils.EnsureConfigDirExists(utils.GetDefaultConfigurationBasePath("~", "."), consts.APP_CONFIG_PATH_BASE_NAME)
	a.flags.Parse()
	a.JsEngine = scripting.CreateNewJavascriptEngine(
		a.Vars,
		a.Gateway,
		func(id string) (any, error) {
			result, _ := a.Workflow.FindTaskById(id)
			return result, nil
		},
		a.GetApplicationIconPath,
	)
	a.loadWorkflowFile(a.ConfigFilename, a.Workflow)
	debug.Dbg.SetEnabled(a.Workflow.Debug)
	godotenv.Load(a.Workflow.Settings.DotEnvFiles...)

	a.Analytics = telemetry.New(a.Workflow.Settings.AnonymousStatistics, a.Gateway)
	a.Gateway.Initialize(a.Workflow.Settings, a.JsEngine.AsContract(), nil)
	a.initializeCache()
	a.Workflow.Initialize(a.JsEngine, a.GetConfigurationPath())
	a.JsEngine.Initialize(a.Vars, os.Environ())

	a.Analytics.EventOnly("app.start")
	a.checkForApplicationUpdates(!*a.flags.NoUpdateCheck)
}

func (a *Application) initializeCache() {
	a.Workflow.Cache = cache.New("stackup", a.GetConfigurationPath(), a.Workflow.Settings.Cache.TtlMinutes)
	a.Gateway.Cache = a.Workflow.Cache
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

	for _, uid := range a.Workflow.State.History.Values() {
		task := a.Workflow.FindTaskByUuid(uid.(string))
		if task == nil {
			continue
		}

		runs := fmt.Sprintf("runs: %v", task.RunCount)
		support.StatusMessageLine("[task history] task: "+task.GetDisplayName()+" ("+runs+")", true)
	}

	os.Exit(1)
}

func (a *Application) createScheduledTasks() {
	support.StatusMessage("Creating scheduled tasks...", false)

	for _, def := range a.Workflow.Scheduler {
		def.Workflow = a.Workflow
		def.JsEngine = a.JsEngine

		_, found := a.Workflow.FindTaskById(def.TaskId())

		if !found {
			support.FailureMessageWithXMark("Task " + def.TaskId() + " not found.")
			continue
		}

		cron := def.Cron
		taskId := def.TaskId()

		a.cronEngine.AddFunc(cron, func() {
			task, found := a.Workflow.FindTaskById(taskId)
			if found {
				task.RunSync()
			}
		})

		// a.scheduledTaskMap.Store(def.TaskId(), &def)
	}

	a.cronEngine.Start()
	support.PrintCheckMarkLine()
}

func (a *Application) stopServerProcesses() {
	a.ProcessMap.Range(func(key any, value any) bool {
		t := a.Workflow.FindTaskByUuid(key.(string))
		if t != nil {
			support.StatusMessage("Stopping "+t.GetDisplayName()+"...", false)
		}

		if value != nil {
			a.KillCommandCallback(value.(*exec.Cmd))
		}

		support.PrintCheckMarkLine()

		return true
	})
}

func (a *Application) runEventLoop() {
	support.StatusMessageLine("Running event loop...", true)

	for {
		utils.WaitForStartOfNextMinute()
	}
}

func (a *Application) runTaskReferences(refs []*TaskReference) {
	for _, def := range refs {
		def.Workflow = a.Workflow
		def.JsEngine = a.JsEngine

		task, found := a.Workflow.FindTaskById(def.TaskId())
		if !found {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		task.RunSync()
	}
}

func (a *Application) runStartupTasks() {
	support.StatusMessageLine("Running startup tasks...", true)

	a.runTaskReferences(a.Workflow.Startup)
}

func (a *Application) runShutdownTasks() {
	a.runTaskReferences(a.Workflow.Shutdown)
}

func (a *Application) runServerTasks() {
	support.StatusMessageLine("Starting server processes...", true)

	for _, def := range a.Workflow.Servers {
		task, found := a.Workflow.FindTaskById(def.TaskId())

		if !found {
			support.SkippedMessageWithSymbol("Task " + def.TaskId() + " not found.")
			continue
		}

		task.RunAsync()
	}
}

func (a Application) runPreconditions() {
	support.StatusMessageLine("Running precondition checks...", true)

	for _, c := range a.Workflow.Preconditions {

		if !c.Run() {
			support.FailureMessageWithXMark(c.Name)
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

	var dependencyBin string = "php"

	if utils.IsFile("composer.json") {
		dependencyBin = "php"
	} else if utils.IsFile("package.json") {
		dependencyBin = "node"
	} else if utils.IsFile("requirements.txt") {
		dependencyBin = "python"
	}

	filename := "stackup.yaml"
	contents := fmt.Sprintf(consts.INIT_CONFIG_FILE_CONTENTS, dependencyBin)
	os.WriteFile(filename, []byte(contents), 0644)
}

func (a *Application) checkForApplicationUpdates(canCheck bool) {
	if !canCheck {
		return
	}

	if hasUpdate, release := updater.New(a.Gateway).IsUpdateAvailable(consts.APP_REPOSITORY, version.APP_VERSION); hasUpdate {
		support.WarningMessage("A new version of StackUp is available, released " + release.TimeSinceRelease())
	}
}

func (a *Application) GetConfigurationPath() string {
	pathname, _ := utils.EnsureConfigDirExists(
		utils.GetDefaultConfigurationBasePath("~", "."),
		consts.APP_CONFIG_PATH_BASE_NAME,
	)

	return pathname
}

func (a *Application) DownloadApplicationIcon() {
	filename := a.GetApplicationIconPath()

	if utils.IsFile(filename) {
		return
	}

	a.Gateway.SaveUrlToFile("https://raw.githubusercontent.com/"+consts.APP_REPOSITORY+"/main/assets/stackup-app-512px.png", filename)
}

func (a Application) GetWorkflowContract() *types.AppWorkflowContract {
	var result interface{} = a.Workflow
	return result.(*types.AppWorkflowContract)
}

func (a *Application) GetApplicationIconPath() string {
	return path.Join(a.GetConfigurationPath(), "/stackup-icon.png")
}

func (a *Application) runInitScript() {
	support.StatusMessageLine("Running init script...", true)

	a.JsEngine.Evaluate(a.Workflow.Init)
}

func (a *Application) Run() {
	a.Initialize()
	defer a.Workflow.Cache.Cleanup(false)

	a.hookSignals()
	a.hookKeyboard()

	a.runInitScript()
	a.runPreconditions()
	a.runStartupTasks()
	a.runServerTasks()
	a.createScheduledTasks()

	a.runEventLoop()
}
