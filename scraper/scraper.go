package scraper

import (
	"log"
	"strconv"
	"strings"

	"github.com/giornetta/watchazon"
	"github.com/gocolly/colly"
)

type Scraper struct {
	AllowedDomains []string
}

func New(domains ...string) *Scraper {
	return &Scraper{
		domains,
	}
}

func (s *Scraper) Scrape(url string) (*watchazon.Product, error) {
	product := &watchazon.Product{
		Link: url,
	}
	var err error

	c := colly.NewCollector(
		colly.AllowedDomains(s.AllowedDomains...),
	)

	c.OnHTML("#productTitle", func(e *colly.HTMLElement) {
		product.Title = strings.TrimSpace(e.Text)
	})

	c.OnHTML("#priceblock_ourprice, #priceblock_dealprice", func(e *colly.HTMLElement) {
		product.Price, err = convertPrice(e.Text)
	})

	c.OnHTML("span.a-size-medium.a-color-price.offer-price.a-text-normal", func(e *colly.HTMLElement) {
		product.Price, err = convertPrice(e.Text)
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}

	return product, err
}

func (s *Scraper) Search(query string) ([]*watchazon.Product, error) {
	url := "https://www.amazon.it/s?k=" + strings.Join(strings.Split(query, " "), "+")
	products := make([]*watchazon.Product, 0)

	c := colly.NewCollector(
		colly.AllowedDomains(s.AllowedDomains...),
	)

	c.OnHTML(".s-result-item", func(e *colly.HTMLElement) {
		// Price
		var p string
		e.ForEach("span.a-price-whole", func(i int, element *colly.HTMLElement) {
			if i > 0 {
				return
			}
			p = element.Text
		})
		// If the price cannot be converted, it probably is because the product is out of stock. We skip it.
		price, err := convertPrice(p)
		if err != nil {
			return
		}

		img := e.ChildAttr(".s-image", "src")
		title := e.ChildText("span.a-color-base.a-text-normal")

		foundUrl := e.ChildAttr("a.a-link-normal.a-text-normal", "href")
		link, err := watchazon.SanitizeURL("https://www.amazon.it" + foundUrl)
		if err != nil {
			log.Printf("%s (%s): %v", title, foundUrl, err)
			return
		}

		products = append(products, &watchazon.Product{
			Title: title,
			Image: img,
			Link:  link,
			Price: price,
		})
	})

	if err := c.Visit(url); err != nil {
		return nil, err
	}

	return products, nil
}

func convertPrice(text string) (float64, error) {
	p := strings.Replace(text, " €", "", 1)
	p = strings.Replace(p, ".", "", -1)
	p = strings.Replace(p, ",", ".", 1)

	return strconv.ParseFloat(p, 64)
}
