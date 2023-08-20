package main

import (
	"os/exec"
	"syscall"

	"github.com/stackup-app/stackup/lib/app"
)

func main() {
	app.App = app.NewApplication()

	app.App.CmdStartCallback = func(cmd *exec.Cmd) {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}

	app.App.KillCommandCallback = func(cmd *exec.Cmd) {
		syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	app.App.Run()
}
