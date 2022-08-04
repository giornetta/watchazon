package main

import (
	"fmt"
	"log"
	"os"

	"github.com/giornetta/watchazon"
	"github.com/giornetta/watchazon/config"
	"github.com/giornetta/watchazon/scraper"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatalf("usage: ./cli <query> <domain>")
	}

	c := config.FromDotEnv()

	scr := scraper.New(c.AllowedDomains...)

	prods, err := scr.Search(os.Args[1], watchazon.Domain(os.Args[2]))
	if err != nil {
		log.Fatalf("could not search: %v", err)
	}

	for _, p := range prods {
		fmt.Println(p.Title, p.Price)
	}
}
