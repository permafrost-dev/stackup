package config

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/stackup-app/stackup/utils"
)

type Configuration struct {
	Filename string `short:"c" long:"config" default:"stackup.yaml" description:"Specify the configuration filename to use"`
	Seed     bool   `short:"s" long:"seed" default:"false" description:"Seed the database"`
}

func NewConfiguration() Configuration {
	result := Configuration{}
	result.Init()

	return result
}

func (cfg *Configuration) Init() {
	flags.ParseArgs(cfg, os.Args)
}

func FindExistingConfigurationFile(defaultFn string) string {
	configFilenames := []string{defaultFn, "stackup.dev.yaml", "stackup.dist.yaml"}
	configFilename, _ := utils.FindFirstExistingFile(configFilenames)

	return configFilename
}
