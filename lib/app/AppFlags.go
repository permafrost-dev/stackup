package app

import (
	"flag"
	"fmt"
	"os"

	"github.com/stackup-app/stackup/lib/version"
)

type AppFlags struct {
	DisplayHelp    *bool
	DisplayVersion *bool
	NoUpdateCheck  *bool
	ConfigFile     *string
	app            *Application
}

func (af *AppFlags) Parse() {
	flag.Parse()

	if af.ConfigFile != nil && *af.ConfigFile != "" {
		af.app.ConfigFilename = *af.ConfigFile
	}

	af.handle()
}

func (af *AppFlags) handle() {
	if *af.DisplayHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *af.DisplayVersion {
		fmt.Println("StackUp version " + version.APP_VERSION)
		os.Exit(0)
	}

	if len(os.Args) > 1 && os.Args[1] == "init" {
		af.app.createNewConfigFile()
		os.Exit(0)
	}
}
