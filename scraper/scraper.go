package scraper

import (
	"fmt"
	"log"
	"net/url"
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

func (s *Scraper) Scrape(link string) (*watchazon.Product, error) {
	domain, err := domainFrom(link)
	if err != nil {
		return nil, err
	}

	product := &watchazon.Product{
		Link: link,
	}

	c := colly.NewCollector(
		colly.AllowedDomains(s.AllowedDomains...),
	)

	c.OnHTML("#productTitle", func(e *colly.HTMLElement) {
		product.Title = strings.TrimSpace(e.Text)
	})

	c.OnHTML("#corePriceDisplay_desktop_feature_div span.a-price span.a-offscreen", func(e *colly.HTMLElement) {
		fmt.Println(e.Text)
		product.Price, err = convertPrice(e.Text, domain)
	})

	// Gets correct pricing for books and items providing various buying options
	c.OnHTML("#price", func(e *colly.HTMLElement) {
		fmt.Println("p ", e.Text)
		product.Price, err = convertPrice(e.Text, domain)
	})

	if err := c.Visit(link); err != nil {
		return nil, err
	}

	return product, err
}

func (s *Scraper) Search(query string, domain watchazon.Domain) ([]*watchazon.Product, error) {
	query = strings.Join(strings.Split(query, " "), "+")
	link := fmt.Sprintf("https://www.amazon.%s/s?k=%s", domain, query)
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
		price, err := convertPrice(p, domain)
		if err != nil {
			return
		}

		img := e.ChildAttr(".s-image", "src")
		title := e.ChildText("span.a-color-base.a-text-normal")

		foundUrl := e.ChildAttr("a.a-link-normal.a-text-normal", "href")
		link := fmt.Sprintf("https://www.amazon.%s%s", domain, foundUrl)

		products = append(products, &watchazon.Product{
			Title: title,
			Image: img,
			Link:  link,
			Price: price,
		})
	})

	if err := c.Visit(link); err != nil {
		log.Println(err)
		return nil, err
	}

	return products, nil
}

func convertPrice(text string, domain watchazon.Domain) (float64, error) {
	text = strings.Replace(text, "\u00a0", "", -1)
	text = strings.Replace(text, " ", "", -1)

	switch domain {
	case "com":
		text = strings.Replace(text, "$", "", 1)
		text = strings.Replace(text, ",", "", -1)
	case "it", "es", "de":
		text = strings.Replace(text, "€", "", 1)
		text = strings.Replace(text, ".", "", -1)
		text = strings.Replace(text, ",", ".", 1)
	}

	return strconv.ParseFloat(text, 64)
}

func domainFrom(link string) (watchazon.Domain, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	d := strings.Split(u.Host, ".")
	return watchazon.Domain(d[len(d)-1]), nil
}
