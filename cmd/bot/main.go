package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/giornetta/watchazon/config"
	"github.com/giornetta/watchazon/locator"

	"github.com/giornetta/watchazon/database"
	"github.com/giornetta/watchazon/scraper"
	"github.com/giornetta/watchazon/service"
	"github.com/giornetta/watchazon/telegram"
)

func main() {
	// Load configuration
	c := config.FromDotEnv()

	// Initialize Amazon scraper
	scr := scraper.New(c.AllowedDomains...)

	// Open badger database
	db, err := database.Open(c.BadgerPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize Locator service
	loc := locator.New(c.Here.AppID, c.Here.AppCode)

	// Initialize the Service
	svc := service.New(scr, db)

	// Create the bot
	bot, err := telegram.New(c.TelegramToken, svc, loc)
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

			time.Sleep(25 * time.Minute)
		}
	}()

	log.Println("Bot running...")
	go bot.Run()
	defer bot.Stop()

	go func() {
		fs := http.FileServer(http.Dir("./web"))
		http.Handle("/", fs)

		log.Println("Listening on :80...")
		port := os.Getenv("PORT")
		err = http.ListenAndServe(":"+port, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	<-ch
}
