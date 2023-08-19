package settings

type Settings struct {
	Defaults               *WorkflowSettingsDefaults      `yaml:"defaults"`
	ExitOnChecksumMismatch bool                           `yaml:"exit-on-checksum-mismatch"`
	ChecksumVerification   *bool                          `yaml:"checksum-verification"`
	DotEnvFiles            []string                       `yaml:"dotenv"`
	Cache                  *WorkflowSettingsCache         `yaml:"cache"`
	Domains                *WorkflowSettingsDomains       `yaml:"domains"`
	AnonymousStatistics    *bool                          `yaml:"anonymous-stats"`
	Gateway                *WorkflowSettingsGateway       `yaml:"gateway"`
	Notifications          *WorkflowSettingsNotifications `yaml:"notifications"`
}
type GatewayContentTypes struct {
	Blocked []string `yaml:"blocked"`
	Allowed []string `yaml:"allowed"`
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
	Allowed []string                       `yaml:"allowed"`
	Hosts   []WorkflowSettingsDomainsHosts `yaml:"hosts"`
}

type WorkflowSettingsDomainsHosts struct {
	Hostname string   `yaml:"hostname"`
	Gateway  *string  `yaml:"gateway"`
	Headers  []string `yaml:"headers"`
}

type WorkflowSettingsCache struct {
	TtlMinutes int `yaml:"ttl-minutes"`
}
type WorkflowSettingsDefaults struct {
	Tasks *WorkflowSettingsDefaultsTasks `yaml:"tasks"`
}

type WorkflowSettingsDefaultsTasks struct {
	Silent    bool     `yaml:"silent"`
	Path      string   `yaml:"path"`
	Platforms []string `yaml:"platforms"`
}

type WorkflowSettingsNotifications struct {
	Telegram *WorkflowSettingsNotificationsTelegram `yaml:"telegram"`
	Slack    *WorkflowSettingsNotificationsSlack    `yaml:"slack"`
}

type WorkflowSettingsNotificationsTelegram struct {
	APIKey  string   `yaml:"api-key"`
	ChatIds []string `yaml:"chat-ids"`
}

type WorkflowSettingsNotificationsSlack struct {
	WebhookUrl string   `yaml:"webhook-url"`
	ChannelIds []string `yaml:"channel-ids"`
}
