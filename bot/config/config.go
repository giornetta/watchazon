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
	Here           struct {
		AppID   string
		AppCode string
	}
}

func FromDotEnv() (*Config, error) {
	config := &Config{}

	if err := godotenv.Load(); err != nil {
		return nil, err
	}

	config.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	config.AllowedDomains = strings.Split(os.Getenv("ALLOWED_DOMAINS"), ",")
	config.BadgerPath = os.Getenv("BADGER_PATH")

	config.Here.AppCode = os.Getenv("HERE_APP_CODE")
	config.Here.AppID = os.Getenv("HERE_APP_ID")

	return config, nil
}
