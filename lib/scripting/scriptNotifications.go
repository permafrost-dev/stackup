package scripting

import (
	"github.com/stackup-app/stackup/lib/settings"
	"github.com/stackup-app/stackup/lib/types"
)

type ScriptNotifications struct {
	engine                 *JavaScriptEngine
	settings               *settings.Settings
	telegramObj            *ScriptNotificationsTelegram
	slackObj               *ScriptNotificationsSlack
	desktopObj             *DesktopNotification
	GetApplicationIconPath func() string
}

// type DesktopNotification struct {
// 	state struct {
// 		title   string
// 		message string
// 	}
// 	sn *ScriptNotifications
// }

// type ScriptNotificationsTelegram struct {
// 	APIToken string
// 	state    struct {
// 		chatIds []int64
// 		title   string
// 		message string
// 	}
// 	sn *ScriptNotifications
// }

// type ScriptNotificationsSlack struct {
// 	WebhookUrl string
// 	state      struct {
// 		channelIds []string
// 		title      string
// 		message    string
// 	}
// 	sn *ScriptNotifications
// }

func CreateScripNotificationsObject(wf *types.AppWorkflowContract, e *JavaScriptEngine) {
	return
}

// 	return

// 	// wc := e.GetWorkflowContract()
// 	// fmt.Printf("wc: %v\n", (*wc).GetSettings())

// 	// if wf == nil {
// 	//     return
// 	// }

// 	// obj := &ScriptNotifications{
// 	// 	engine:   e,
// 	// 	settings: (*wf).GetSettings(),
// 	// 	telegramObj: &ScriptNotificationsTelegram{
// 	// 		APIToken: "",
// 	// 	},
// 	// 	slackObj: &ScriptNotificationsSlack{
// 	// 		WebhookUrl: "",
// 	// 	},
// 	// 	desktopObj:             &DesktopNotification{},
// 	// 	GetApplicationIconPath: e.GetApplicationIconPath,
// 	// }
// 	// obj.desktopObj.sn = obj
// 	// obj.telegramObj.sn = obj
// 	// obj.slackObj.sn = obj

// 	// e.Vm.Set("notifications", obj)
// }

// func (dn *DesktopNotification) create() *DesktopNotification {
// 	dn.resetState()

// 	return dn
// }

// func (dn *DesktopNotification) Send() bool {
// 	result := notifications.NewDesktopNotification(dn.sn.GetApplicationIconPath()).
// 		Send(dn.state.title, dn.state.message)

// 	dn.resetState()
// 	return result == nil
// }

// func (dn *DesktopNotification) resetState() {
// 	dn.state.title = ""
// 	dn.state.message = ""
// }

// func (dn *DesktopNotification) Message(message string, title ...string) *DesktopNotification {
// 	dn.state.message = message
// 	dn.state.title = "notification"

// 	if len(title) > 0 {
// 		dn.state.title = title[0]
// 	}

// 	return dn
// }

// func (sn *ScriptNotifications) Desktop() *DesktopNotification {
// 	return sn.desktopObj
// }

// func (sn *ScriptNotifications) Telegram() *ScriptNotificationsTelegram {
// 	return nil
// 	// token := sn.settings().Notifications.Telegram.APIKey
// 	// return sn.telegramObj.create(token)
// }

// func (sn *ScriptNotifications) Slack() *ScriptNotificationsSlack {
// 	return nil
// 	// webhookUrl := sn.settings().Notifications.Slack.WebhookUrl
// 	// return sn.slackObj.create(webhookUrl)
// }

// func (sns *ScriptNotificationsSlack) create(webhookUrl string) *ScriptNotificationsSlack {
// 	sns.WebhookUrl = webhookUrl

// 	return sns
// }

// func (sns *ScriptNotificationsSlack) Message(message string) *ScriptNotificationsSlack {
// 	sns.state.message = message
// 	sns.state.title = "notification"

// 	return sns
// }

// func (sns *ScriptNotificationsSlack) To(channelIds ...string) *ScriptNotificationsSlack {
// 	for _, channelIdStr := range channelIds {
// 		sns.state.channelIds = append(sns.state.channelIds, channelIdStr)
// 	}

// 	return sns
// }

// func (sns *ScriptNotificationsSlack) Send() bool {
// 	// temp := (*sns.sn.engine).toInterface()

// 	// sns.To(temp.(JavaScriptEngine).GetWorkflowContract()).GetSettings().Notifications.Slack)
// 	// webhookUrl := sns.sn.engine.Evaluate((*temp).GetSettings().Notifications.Slack.WebhookUrl).(string)

// 	// result := notifications.NewSlackNotification(webhookUrl, sns.state.channelIds...).
// 	// 	Send(sns.state.title, sns.state.message)

// 	// sns.resetState()
// 	return false
// 	// return result == nil
// }

// func (sns *ScriptNotificationsSlack) resetState() {
// 	sns.state.channelIds = []string{}
// 	sns.state.title = ""
// 	sns.state.message = ""
// }

// func (snt *ScriptNotificationsTelegram) resetState() {
// 	snt.state.chatIds = []int64{}
// 	snt.state.title = ""
// 	snt.state.message = ""
// }

// // create is a method of the `ScriptNotificationsTelegram` struct. It takes an
// // `apiToken` string as a parameter and returns a pointer to a `ScriptNotificationsTelegram` object.
// func (snt *ScriptNotificationsTelegram) create(apiToken string) *ScriptNotificationsTelegram {
// 	snt.APIToken = apiToken
// 	snt.resetState()

// 	return snt
// }

// func (snt *ScriptNotificationsTelegram) Message(message string) *ScriptNotificationsTelegram {
// 	snt.state.message = message
// 	snt.state.title = "notification"
// 	return snt
// }

// func (snt *ScriptNotificationsTelegram) To(chatIDs ...string) *ScriptNotificationsTelegram {
// 	for _, chatIdStr := range chatIDs {
// 		id32, _ := strconv.Atoi(chatIdStr)
// 		snt.state.chatIds = append(snt.state.chatIds, int64(id32))
// 	}

// 	return snt
// }

// // Send is a method of the `ScriptNotificationsTelegram` struct. It takes three
// // parameters: `chatId` of type `int64`, `title` of type `string`, and `message` of type `string`.
// func (snt *ScriptNotificationsTelegram) Send() bool {
// 	return false
// 	// if len(snt.state.chatIds) == 0 {
// 	// 	snt.To(snt.sn.settings().Notifications.Telegram.ChatIds...)
// 	// }

// 	// apiKey := snt.sn.engine.Evaluate(snt.sn.settings().Notifications.Telegram.APIKey).(string)
// 	// result := notifications.
// 	// 	NewTelegramNotification(apiKey, snt.state.chatIds...).
// 	// 	Send(snt.state.title, snt.state.message)

// 	// snt.resetState()

// 	// return result == nil
// }
