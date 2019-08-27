package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken  string
	AllowedDomains []string
	BadgerPath     string
}

func FromDotEnv() (*Config, error) {
	config := &Config{}

	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	config.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	config.AllowedDomains = strings.Split(os.Getenv("ALLOWED_DOMAINS"), ",")
	config.BadgerPath = os.Getenv("BADGER_PATH")

	return config, nil
}
