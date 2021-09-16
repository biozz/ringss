package config

type EnvConfig struct {
	TelegramBotToken      string `required:"true" split_words:"true" desc:"API Token obtained from @BotFather"`
	DatabaseURL           string `required:"true" split_words:"true" default:"redis://localhost:6379" desc:"Redis connection URL"`
	MinifluxBaseURL       string `required:"true" split_words:"true" desc:"A Miniflux instance URL without trailing slash"`
	DefaultPollerState    string `split_words:"true" default:"1" desc:"Enables or disables poller on start. Can be re-enabled at runtime with poller:enable key set to 1"`
	PollerIntervalSeconds int    `split_words:"true" default:"10" desc:"Time between poller cycles"`
}
