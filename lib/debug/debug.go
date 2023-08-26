package debug

import (
	"fmt"

	"github.com/stackup-app/stackup/lib/utils"
)

type DebugContract interface {
	IsEnabled() bool
	SetEnabled(enabled bool)
	Logf(format string, a ...any)
}

type Debug struct {
	Enabled bool
	DebugContract
}

var Dbg = &Debug{Enabled: false}

func (d *Debug) SetEnabled(enabled bool) {
	d.Enabled = enabled
}

func (d *Debug) IsEnabled() bool {
	return d.Enabled
}

func (d *Debug) Log(msg ...string) {
	if len(msg) == 0 {
		d.Logf("")
		return
	}

	for _, m := range msg {
		d.Logf("%s", m)
	}
}

func (d *Debug) Logf(format string, a ...any) {
	if !d.IsEnabled() {
		return
	}

	fmt.Printf(" [debug] "+utils.EnforceSuffix(format, "\n"), a)
}

func Logf(format string, a ...any) {
	Dbg.Logf(format, a)
}

func Log(msg ...string) {
	Dbg.Log(msg...)
}
