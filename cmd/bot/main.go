package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/giornetta/watchazon/config"

	"github.com/giornetta/watchazon/database"
	"github.com/giornetta/watchazon/scraper"
	"github.com/giornetta/watchazon/service"
	"github.com/giornetta/watchazon/telegram"
)

func main() {
	c, err := config.FromDotEnv()
	if err != nil {
		log.Fatalf("could not read config from .env: %v\n", err)
	}

	scr := scraper.New(c.AllowedDomains...)

	db, err := database.Open(c.BadgerPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	svc := service.New(scr, db)

	bot, err := telegram.New(c.TelegramToken, svc)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Service running...")
	go func() {
		for {
			log.Println("Running products update...")
			err := svc.Update()
			if err != nil {
				log.Printf("could not update products: %v", err)
			}

			time.Sleep(3 * time.Hour)
		}
	}()

	log.Println("Bot running...")
	go bot.Run()
	defer bot.Stop()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	<-ch
}
