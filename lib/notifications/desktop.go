package notifications

import (
	"github.com/gen2brain/beeep"
)

type DesktopNotification struct {
}

// NewDesktopNotification creates a new instance of the DesktopNotification
// struct.
func NewDesktopNotification() *DesktopNotification {
	return &DesktopNotification{}
}

func (tn *DesktopNotification) Send(title, message string) error {
    return beeep.Notify(title, message, "")
}
