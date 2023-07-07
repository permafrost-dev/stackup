package main

import (
	"flag"
	"os"
	"path/filepath"
	"runtime"
)

var (
	IsPlatformWindows         = runtime.GOOS == "windows"
	FrontendProjectPath, _    = filepath.Abs("./../acd-pos-frontend")
	BackendProjectPath, _     = os.Getwd()
	SeedDatabase, DisplayHelp bool
	// startupTasks              []lib.Task
	//processDefinitions        []lib.ProcessDefinition
)

func init() {
	flag.BoolVar(&SeedDatabase, "seed", false, "Seed database")
	flag.BoolVar(&DisplayHelp, "help", false, "Display help")
	flag.Parse()

	// Tasks to run after the containers start but before the server processes are started
	// startupTasks = []lib.Task{
	// 	{
	// 		Binary: "php",
	// 		Args: []string{
	// 			"artisan",
	// 			"migrate:fresh",
	// 			"--seed",
	// 			"--no-interaction",
	// 		},
	// 	},
	// 	{
	// 		Binary: "php",
	// 		Args: []string{
	// 			"artisan",
	// 			"qb:create-test-token",
	// 		},
	// 	},
	// }

	// Server processes to start after the containers are started and the startup tasks are run
	// processDefinitions = []lib.ProcessDefinition{
	// 	{
	// 		Name:      "frontend",
	// 		Binary:    "node",
	// 		Args:      []string{FrontendProjectPath + "/node_modules/.bin/next", "dev"},
	// 		Cwd:       FrontendProjectPath,
	// 		RunsOnWin: true,
	// 	},
	// 	{
	// 		Name:      "httpd",
	// 		Binary:    "php",
	// 		Args:      []string{"artisan", "serve"},
	// 		Cwd:       BackendProjectPath,
	// 		RunsOnWin: true,
	// 	},
	// 	{
	// 		Name:      "horizon",
	// 		Binary:    "php",
	// 		Args:      []string{"artisan", "horizon"},
	// 		Cwd:       BackendProjectPath,
	// 		RunsOnWin: false,
	// 	},
	// 	{
	// 		Name:      "horizon-info",
	// 		Binary:    "php",
	// 		Args:      []string{"artisan", "horizon:supervisors"},
	// 		Cwd:       BackendProjectPath,
	// 		RunsOnWin: false,
	// 		Delay:     5000 * time.Millisecond,
	// 	},
	// }
}

type PodmanRunningContainerInfo struct {
	AutoRemove bool              `json:"AutoRemove"`
	Command    interface{}       `json:"Command"`
	CreatedAt  string            `json:"CreatedAt"`
	Exited     bool              `json:"Exited"`
	ExitedAt   int64             `json:"ExitedAt"`
	ExitCode   int               `json:"ExitCode"`
	ID         string            `json:"Id"`
	Image      string            `json:"Image"`
	ImageID    string            `json:"ImageID"`
	IsInfra    bool              `json:"IsInfra"`
	Labels     map[string]string `json:"Labels"`
	Mounts     []interface{}     `json:"Mounts"`
	Names      []string          `json:"Names"`
	Namespaces struct {
	} `json:"Namespaces"`
	Networks []string `json:"Networks"`
	Pid      int      `json:"Pid"`
	Pod      string   `json:"Pod"`
	PodName  string   `json:"PodName"`
	Ports    []struct {
		HostIP        string `json:"host_ip"`
		ContainerPort int    `json:"container_port"`
		HostPort      int    `json:"host_port"`
		Range         int    `json:"range"`
		Protocol      string `json:"protocol"`
	} `json:"Ports"`
	Size      interface{} `json:"Size"`
	StartedAt int         `json:"StartedAt"`
	State     string      `json:"State"`
	Status    string      `json:"Status"`
	Created   int         `json:"Created"`
}
