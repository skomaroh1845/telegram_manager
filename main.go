package main

import (
	"log"
	"os"
	"telegram_bot/internal/bot"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Telegram struct {
		Token          string `yaml:"token"`
		MenuServiceURL string `yaml:"menu_service_url"`
	} `yaml:"telegram"`
	Users map[string]string `yaml:"users"`
}

func main() {
	// Read config file
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Error reading config file:", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatal("Error parsing config file:", err)
	}

	if config.Telegram.Token == "" {
		log.Fatal("Telegram bot token is not set in config file")
	}

	menuServiceURL := config.Telegram.MenuServiceURL
	if menuServiceURL == "" {
		menuServiceURL = "http://localhost:8080"
	}

	bot, err := bot.New(config.Telegram.Token, menuServiceURL, config.Users)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Bot started...")
	if err := bot.Start(); err != nil {
		log.Fatal(err)
	}
}
