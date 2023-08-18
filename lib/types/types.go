package types

import "os/exec"

type CommandCallback func(cmd *exec.Cmd)
