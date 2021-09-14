package config

type EnvConfig struct {
	TelegramBotToken      string `required:"true" split_words:"true" desc:"API Token obtained from @BotFather"`
	DatabasePath          string `required:"true" split_words:"true" default:"db" desc:"Absolute path to the database or relative to /app inside the container"`
	MinifluxBaseURL       string `required:"true" split_words:"true" desc:"A Miniflux instance URL without trailing slash"`
	DefaultPollerState    string `split_words:"true" default:"1" desc:"Enables or disables poller on start. Can be re-enabled at runtime with poller:enable key set to 1"`
	AdminUserIds          []int  `split_words:"true" desc:"A list of Telegram User IDs who can access special commands"`
	PollerIntervalSeconds int    `split_words:"true" default:"10" desc:"Time between poller cycles"`
}
