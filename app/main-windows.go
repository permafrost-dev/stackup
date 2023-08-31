//go:build WINDOWS

package main

import (
	"os/exec"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stackup-app/stackup/lib/utils"
)

func main() {
	app.App = app.NewApplication()

	app.App.CmdStartCallback = func(cmd *exec.Cmd) {}
	app.App.KillCommandCallback = func(cmd *exec.Cmd) {
		utils.KillProcessOnWindows(cmd)
	}

	app.App.Run()
}
