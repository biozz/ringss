package config

type EnvConfig struct {
	TelegramBotToken      string   `split_words:"true"`
	DatabasePath          string   `split_words:"true" default:"db"`
	MinifluxBaseURL       string   `split_words:"true"`
	DefaultPollerState    string   `split_words:"true" default:"1"`
	AdminUserIDs          []string `split_words:"true"`
	PollerIntervalSeconds int      `split_words:"true" default:"10"`
}
