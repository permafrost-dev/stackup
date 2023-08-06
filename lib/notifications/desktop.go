package notifications

import (
	"github.com/gen2brain/beeep"
)

type DesktopNotification struct {
    IconPath string
}

func NewDesktopNotification(iconPath string) *DesktopNotification {
	return &DesktopNotification{
        IconPath: iconPath,
    }
}

func (dn *DesktopNotification) Send(title, message string) error {
    beeep.DefaultDuration = 8

    return beeep.Notify(title, message, dn.IconPath)
}
