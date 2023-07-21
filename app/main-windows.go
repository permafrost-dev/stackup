//go:build WINDOWS

package main

import (
	"os/exec"

	"github.com/stackup-app/stackup/lib/app"
	"github.com/stackup-app/stackup/lib/utils"
)

func main() {
	a := app.App{
		CmdStartCallback: func(cmd *exec.Cmd) {},
		KillCommandCallback: func(cmd *exec.Cmd) {
			utils.KillProcessOnWindows(cmd)
		},
	}
	a.Run()
}
