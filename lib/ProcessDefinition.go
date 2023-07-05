package lib

import "time"

type ProcessDefinition struct {
	Name      string
	Binary    string
	Args      []string
	Cwd       string
	RunsOnWin bool
	Delay     time.Duration
}
