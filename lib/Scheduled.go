package lib

import (
	"fmt"

	"github.com/permafrost-dev/stack-supervisor/utils"
)

type Scheduled struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
	Cron    string `yaml:"cron"`
	Running bool
}

func (s *Scheduled) Init(app *Application) {
	app.CronEngine.AddFunc(s.Cron, s.Execute)
}

func (s *Scheduled) Execute() {
	fmt.Println("Executing scheduled task '" + s.Name + "'...")
	utils.RunCommand(s.Command)
}

func (s *Scheduled) Stop() {
	// s.CronEngine.Stop().Deadline()
}
