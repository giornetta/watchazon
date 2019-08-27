package service

import (
	"errors"
	"log"
	"sync"

	"github.com/giornetta/watchazon"
	"github.com/giornetta/watchazon/database"
	"github.com/giornetta/watchazon/scraper"
)

type Service struct {
	scraper       *scraper.Scraper
	database      *database.Database
	notifications chan *watchazon.Notification
}

var (
	ErrInvalidLink = errors.New("invalid link")
	ErrInternal    = errors.New("internal server error")
)

func New(sc *scraper.Scraper, db *database.Database) *Service {
	return &Service{
		scraper:       sc,
		database:      db,
		notifications: make(chan *watchazon.Notification),
	}
}

func (s *Service) AddToWatchList(link string, userID int) error {
	link, err := watchazon.SanitizeURL(link)
	if err != nil {
		return ErrInvalidLink
	}

	scraped, err := s.scraper.Scrape(link)
	if err != nil {
		log.Printf("could not scrape %s: %v", link, err)
		return ErrInternal
	}

	stored, err := s.database.Get(link)
	if err != nil {
		err := s.database.Insert(scraped, userID)
		if err != nil {
			log.Printf("could not insert product: %v", err)
			return ErrInternal
		}
		return nil
	}

	if stored.Price != scraped.Price {
		for _, u := range stored.Users {
			s.Notify(scraped, u)
		}
	}

	err = s.database.Update(scraped, userID)
	if err != nil {
		log.Printf("could not update product: %v", err)
		return ErrInternal
	}

	return nil
}

func (s *Service) GetUserWatchList(user int) ([]*watchazon.Product, error) {
	prods, err := s.database.GetUserWatchList(user)
	if err != nil {
		log.Printf("could not get watchlist")
		return nil, ErrInternal
	}

	products := make([]*watchazon.Product, len(prods))
	for i, p := range prods {
		products[i] = p.Product
	}

	return products, nil
}

func (s *Service) Update() error {
	products, err := s.database.GetAll()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(len(products))

	for _, p := range products {
		go func(p *database.Record) {
			defer wg.Done()

			scraped, err := s.scraper.Scrape(p.Link)
			if err != nil {
				return
			}

			if p.Price == scraped.Price {
				return
			}

			for _, u := range p.Users {
				s.Notify(scraped, u)
			}

			err = s.database.Update(scraped, 0)
			if err != nil {
				return
			}
		}(p)
	}

	wg.Wait()

	return nil
}

func (s *Service) RemoveFromWatchList(link string, userID int) error {
	return s.database.RemoveFromWatchList(link, userID)
}

func (s *Service) Notify(product *watchazon.Product, userID int) {
	s.notifications <- &watchazon.Notification{
		Product: product,
		UserID:  userID,
	}
}

func (s *Service) Listen() <-chan *watchazon.Notification {
	return s.notifications
}

func (s *Service) Search(query string) ([]*watchazon.Product, error) {
	if query == "" {
		return nil, errors.New("empty query")
	}
	return s.scraper.Search(query)
}
