package lib

import (
	"os"
	"path"
)

var MyGlobal = "hello world"
var dir, _ = os.Getwd()

var ConfigFileName = path.Join(dir, "stack-supervisor.config.yaml")

type globalDefs struct {
	ConfigFileName string
}

func (g *globalDefs) init() {
	g.ConfigFileName = ConfigFileName
}

var Globals globalDefs

func InitGlobals() {
	Globals.init()
}
