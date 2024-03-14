package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AppUrl      string `env:"APP_URL" env-default:"http://localhost:3000"`
	LogFilePath string `env:"LOG_FILE_PATH" env-default:"/var/log/"`
	LogFileName string `env:"LOG_FILE_NAME" env-default:"gogammu-checker.log"`

	BotToken  string `env:"BOT_TOKEN"`
	BotChatId string `env:"BOT_CHAT_ID"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}
