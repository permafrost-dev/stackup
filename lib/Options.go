package lib

import (
	"strconv"

	"github.com/stackup-app/stackup/utils"
)

type Options struct {
	Dotenv    bool             `yaml:"dotenv,omitempty"`
	Webui     bool             `yaml:"webui,omitempty"`
	EventLoop EventLoopOptions `yaml:"event-loop"`
}

func (o *Options) Init(config *StackConfig) {
	dotenvStr := strconv.FormatBool(o.Dotenv)
	webuiStr := strconv.FormatBool(o.Webui)

	o.Dotenv, _ = strconv.ParseBool(utils.ReplaceConfigurationKeyVariablesInStruct(dotenvStr, config, "config"))
	o.Webui, _ = strconv.ParseBool(utils.ReplaceConfigurationKeyVariablesInStruct(webuiStr, config, "config"))

	o.EventLoop.IntervalMilliseconds = utils.ParseDurationString(o.EventLoop.Interval)
}
