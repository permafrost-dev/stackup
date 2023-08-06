package notifications

import (
	slackapi "github.com/slack-go/slack"
)

type SlackNotification struct {
	WebhookUrl string
	ChannelIds  []string
}

// NewSlackNotification creates a new instance of the TelegramNotification struct with the provided notifier,
// API token, and chat IDs.
func NewSlackNotification(webhookUrl string, channelIds ...string) *SlackNotification {
	return &SlackNotification{
		WebhookUrl: webhookUrl,
		ChannelIds:  channelIds,
	}
}

func (tn *SlackNotification) Send(title, message string) error {
    for _, channelId := range tn.ChannelIds {
        msg := slackapi.WebhookMessage{Channel: channelId, Text: message}
        slackapi.PostWebhook(tn.WebhookUrl, &msg)
    }

	return nil
}
