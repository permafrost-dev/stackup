//go:build WINDOWS

package main

import (
	"os/exec"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stackup-app/stackup/lib/utils"
)

func main() {
	app.App = &app.Application{
		CmdStartCallback: func(cmd *exec.Cmd) {},
		KillCommandCallback: func(cmd *exec.Cmd) {
			utils.KillProcessOnWindows(cmd)
		},
	}

	app.App.Run()
}
