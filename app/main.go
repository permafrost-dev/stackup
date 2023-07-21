package main

import (
	"os/exec"
	"syscall"

	"github.com/stackup-app/stackup/lib/app"
)

func main() {
	a := app.App{
		CmdStartCallback: func(cmd *exec.Cmd) {
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		},
		KillCommandCallback: func(cmd *exec.Cmd) {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		},
	}
	a.Run()
}
