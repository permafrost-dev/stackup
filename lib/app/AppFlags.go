package app

import (
	"flag"
	"fmt"
	"os"
	"reflect"

	"github.com/stackup-app/stackup/lib/version"
)

type AppFlags struct {
	DisplayHelp    *bool
	DisplayVersion *bool
	NoUpdateCheck  *bool
	ConfigFile     *string
	app            *Application
}

func (af *AppFlags) Get(name string) any {
	var result *any = reflect.ValueOf(af).FieldByName(name).Interface().(*any)
	if result == nil {
		return nil
	}
	to := reflect.TypeOf(*result)
	if to.Kind() == reflect.Bool {
		return (*result).(bool)
	}
	if to.Kind() == reflect.String {
		return (*result).(string)
	}

	return *result
}

func (af *AppFlags) GetString(name string) string {
	if result := af.Get(name); result != nil {
		return result.(string)
	}
	return ""
}

func (af *AppFlags) GetBool(name string) bool {
	if result := af.Get(name); result != nil {
		return result.(bool)
	}
	return false
}

func (af *AppFlags) Parse() {
	flag.Parse()

	if af.ConfigFile != nil && *af.ConfigFile != "" {
		af.app.ConfigFilename = af.Get("ConfigFile").(string)
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
