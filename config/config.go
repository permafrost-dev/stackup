package config

import (
	"os"

	"github.com/jessevdk/go-flags"
)

type Configuration struct {
	Filename string `short:"c" long:"config" default:"stack-supervisor.config.dev.yaml" description:"Specify the configuration filename to use"`
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
