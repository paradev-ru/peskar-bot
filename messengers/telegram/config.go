package telegram

const (
	DefaultTelegramURL                 = "https://api.telegram.org/bot"
	DefaultTelegramLinksPreviewDisable = false
	DefaultTelegramNotificationDisable = false
)

type Config struct {
	Enabled               bool   `toml:"enabled"`
	URL                   string `toml:"url"`
	Token                 string `toml:"token"`
	ChatId                string `toml:"chat_id"`
	ParseMode             string `toml:"parse_mode"`
	DisableWebPagePreview bool   `toml:"disable_web_page_preview"`
	DisableNotification   bool   `toml:"disable_notification"`
}

func NewConfig() Config {
	return Config{
		URL: DefaultTelegramURL,
		DisableWebPagePreview: DefaultTelegramLinksPreviewDisable,
		DisableNotification:   DefaultTelegramNotificationDisable,
	}
}
