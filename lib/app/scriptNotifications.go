package app

import (
	"strconv"

	"github.com/robertkrimen/otto"
	"github.com/stackup-app/stackup/lib/notifications"
)

type ScriptNotifications struct {
	Vm          *otto.Otto
	telegramObj *ScriptNotificationsTelegram
    slackObj *ScriptNotificationsSlack
}

type ScriptNotificationsTelegram struct {
	APIToken string
	state    struct {
		chatIds    []int64
		title      string
		message    string
	}
}

type ScriptNotificationsSlack struct {
    WebhookUrl string
    state    struct {
		channelIds    []string
		title      string
		message    string
	}
}

func CreateScripNotificationsObject(vm *otto.Otto) {
	obj := &ScriptNotifications{
		Vm: vm,
		telegramObj: &ScriptNotificationsTelegram{
			APIToken: "",
		},
        slackObj: &ScriptNotificationsSlack{
            WebhookUrl: "",
        },
	}
	vm.Set("notifications", obj)
}

func (sn *ScriptNotifications) Telegram() *ScriptNotificationsTelegram {
	token := App.Workflow.Settings.Notifications.Telegram.APIKey
	return sn.telegramObj.create(token)
}

func (sn *ScriptNotifications) Slack() *ScriptNotificationsSlack {
    webhookUrl := App.Workflow.Settings.Notifications.Slack.WebhookUrl
    return sn.slackObj.create(webhookUrl)
}

func (sns *ScriptNotificationsSlack) create(webhookUrl string) *ScriptNotificationsSlack {
    sns.WebhookUrl = webhookUrl

    return sns
}

func (sns *ScriptNotificationsSlack) Message(message string) *ScriptNotificationsSlack {
    sns.state.message = message
	sns.state.title = "notification"

    return sns
}

func (sns *ScriptNotificationsSlack) To(channelIds ...string) *ScriptNotificationsSlack {
    for _, channelIdStr := range channelIds {
        sns.state.channelIds = append(sns.state.channelIds, channelIdStr)
    }

    return sns
}

func (sns *ScriptNotificationsSlack) Send() bool {
    sns.To(App.Workflow.Settings.Notifications.Slack.ChannelIds...)
    webhookUrl := App.JsEngine.Evaluate(App.Workflow.Settings.Notifications.Slack.WebhookUrl).(string)

    result := notifications.NewSlackNotification(webhookUrl, sns.state.channelIds...).
        Send(sns.state.title, sns.state.message)

    sns.resetState()
    return result == nil
}

func (sns *ScriptNotificationsSlack) resetState() {
    sns.state.channelIds = []string{}
    sns.state.title = ""
    sns.state.message = ""
}

func (snt *ScriptNotificationsTelegram) resetState() {
	snt.state.chatIds = []int64{}
	snt.state.title = ""
	snt.state.message = ""
}

// create is a method of the `ScriptNotificationsTelegram` struct. It takes an
// `apiToken` string as a parameter and returns a pointer to a `ScriptNotificationsTelegram` object.
func (snt *ScriptNotificationsTelegram) create(apiToken string) *ScriptNotificationsTelegram {
	snt.APIToken = apiToken
	snt.resetState()

	return snt
}

func (snt *ScriptNotificationsTelegram) Message(message string) *ScriptNotificationsTelegram {
	snt.state.message = message
	snt.state.title = "notification"
	return snt
}

func (snt *ScriptNotificationsTelegram) To(chatIDs ...string) *ScriptNotificationsTelegram {
	for _, chatIdStr := range chatIDs {
		id32, _ := strconv.Atoi(chatIdStr)
		snt.state.chatIds = append(snt.state.chatIds, int64(id32))
	}

	return snt
}

// Send is a method of the `ScriptNotificationsTelegram` struct. It takes three
// parameters: `chatId` of type `int64`, `title` of type `string`, and `message` of type `string`.
func (snt *ScriptNotificationsTelegram) Send() bool {
	if len(snt.state.chatIds) == 0 {
		snt.To(App.Workflow.Settings.Notifications.Telegram.ChatIds...)
	}

	apiKey := App.JsEngine.Evaluate(App.Workflow.Settings.Notifications.Telegram.APIKey).(string)
	result := notifications.
		NewTelegramNotification(apiKey, snt.state.chatIds...).
		Send(snt.state.title, snt.state.message)

	snt.resetState()

	return result == nil
}
