package settings

import "reflect"

type Settings struct {
	Defaults               WorkflowSettingsDefaults      `yaml:"defaults"`
	ExitOnChecksumMismatch bool                          `yaml:"exit-on-checksum-mismatch"`
	ChecksumVerification   bool                          `yaml:"checksum-verification"`
	DotEnvFiles            []string                      `yaml:"dotenv"`
	Cache                  WorkflowSettingsCache         `yaml:"cache"`
	Domains                WorkflowSettingsDomains       `yaml:"domains"`
	AnonymousStatistics    bool                          `yaml:"anonymous-stats"`
	Gateway                WorkflowSettingsGateway       `yaml:"gateway"`
	Notifications          WorkflowSettingsNotifications `yaml:"notifications"`
	Debug                  bool                          `yaml:"debug"`
}

type GatewayBlockAllowListsContract interface {
	GetAllowed() []string
	GetBlocked() []string
	SetAllowed(items []string)
	SetBlocked(items []string)
}

type GatewayBlockAllowLists struct {
	Blocked []string `yaml:"blocked"`
	Allowed []string `yaml:"allowed"`

	GatewayBlockAllowListsContract
}

type GatewayContentTypes struct {
	GatewayBlockAllowLists
}

type WorkflowSettingsGateway struct {
	ContentTypes   *GatewayContentTypes                   `yaml:"content-types"`
	FileExtensions *WorkflowSettingsGatewayFileExtensions `yaml:"file-extensions"`
	Middleware     []string                               `yaml:"middleware"`
}

type WorkflowSettingsGatewayFileExtensions struct {
	Allow []string `yaml:"allow"`
	Block []string `yaml:"block"`
}

type WorkflowSettingsDomains struct {
	Allowed []string                      `yaml:"allowed"`
	Blocked []string                      `yaml:"blocked"`
	Hosts   []WorkflowSettingsDomainsHost `yaml:"hosts"`
}

type WorkflowSettingsDomainsHost struct {
	Hostname string   `yaml:"hostname"`
	Gateway  string   `yaml:"gateway"`
	Headers  []string `yaml:"headers"`
}

type WorkflowSettingsCache struct {
	TtlMinutes int `yaml:"ttl-minutes"`
}
type WorkflowSettingsDefaults struct {
	Tasks WorkflowSettingsDefaultsTasks `yaml:"tasks"`
}

type WorkflowSettingsDefaultsTasks struct {
	Silent    bool     `yaml:"silent"`
	Path      string   `yaml:"path"`
	Platforms []string `yaml:"platforms"`
}

type WorkflowSettingsNotifications struct {
	Telegram WorkflowSettingsNotificationsTelegram `yaml:"telegram"`
	Slack    WorkflowSettingsNotificationsSlack    `yaml:"slack"`
}

type WorkflowSettingsNotificationsTelegram struct {
	APIKey  string   `yaml:"api-key"`
	ChatIds []string `yaml:"chat-ids"`
}

type WorkflowSettingsNotificationsSlack struct {
	WebhookUrl string   `yaml:"webhook-url"`
	ChannelIds []string `yaml:"channel-ids"`
}

func arrayContains[T comparable](array1 []T, array2 any) bool {
	// Create a map to store the items in array1
	items := make(map[T]bool)
	for _, item := range array1 {
		items[item] = true
	}

	var arr2 []T
	if reflect.TypeOf(array2).Kind() != reflect.Slice {
		arr2 = []T{array2.(T)}
	} else {
		arr2 = array2.([]T)
	}

	for _, item := range arr2 {
		if !items[item] {
			return false
		}
	}

	return true
}

func getUniqueStrings(items []string) []string {
	result := []string{}
	for _, item := range items {
		if !arrayContains(result, item) {
			result = append(result, item)
		}
	}
	return result
}

func (gba *GatewayBlockAllowLists) GetAllowed() []string {
	return gba.Allowed
}

func (gba *GatewayBlockAllowLists) GetBlocked() []string {
	return gba.Blocked
}

func (gba *GatewayBlockAllowLists) SetAllowed(items []string) {
	gba.Allowed = getUniqueStrings(items)
}

func (gba *GatewayBlockAllowLists) SetBlocked(items []string) {
	gba.Blocked = getUniqueStrings(items)
}

func (gba *GatewayBlockAllowLists) Allow(item string) {
	gba.SetAllowed(append(gba.Allowed, item))
}

func (gba *GatewayBlockAllowLists) Block(item string) {
	gba.SetBlocked(append(gba.Blocked, item))
}
