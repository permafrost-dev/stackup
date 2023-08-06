package notifications

import (
	"context"
	"os"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/telegram"
)

type TelegramNotification struct {
	ApiToken string
	ChatIds  []int64
	Service  *telegram.Telegram
	Notifier *notify.Notify
}

// NewTelegramNotification creates a new instance of the TelegramNotification struct with the provided notifier,
// API token, and chat IDs.
func NewTelegramNotification(apiToken string, chatIds ...int64) *TelegramNotification {
	service, _ := telegram.New(os.ExpandEnv(apiToken))

	return &TelegramNotification{
		ApiToken: apiToken,
		ChatIds:  chatIds,
		Service:  service,
	}
}

func (tn *TelegramNotification) Send(title, message string) error {
	tn.Service.AddReceivers(tn.ChatIds...)
	notify.UseServices(tn.Service)

	err := notify.Send(
		context.Background(),
		title,
		message,
	)

	return err
}
