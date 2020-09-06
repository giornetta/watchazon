package config

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// Config contains the required configuration variables for the program.
type Config struct {
	TelegramToken  string
	AllowedDomains []string
	BadgerPath     string
	Here           struct {
		AppID   string
		AppCode string
	}
}

// FromDotEnv loads the required configuration variables from a .env file.
func FromDotEnv() *Config {
	config := &Config{}

	if err := godotenv.Load(); err != nil {
		log.Println("could not find .env file, loading env variables")
	}

	config.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	config.AllowedDomains = strings.Split(os.Getenv("ALLOWED_DOMAINS"), ",")
	config.BadgerPath = os.Getenv("BADGER_PATH")

	config.Here.AppCode = os.Getenv("HERE_APP_CODE")
	config.Here.AppID = os.Getenv("HERE_APP_ID")

	return config
}
