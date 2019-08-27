package watchazon

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Product struct {
	Title string
	Image string
	Link  string
	Price float64
}

type Notification struct {
	Product *Product
	UserID  int
}

type Service interface {
	AddToWatchList(link string, userID int) error
	RemoveFromWatchList(link string, userID int) error
	GetUserWatchList(user int) ([]*Product, error)
	Search(query string) ([]*Product, error)
	Update() error
	Notify(product *Product, userID int)
	Listen() <-chan *Notification
}

func SanitizeURL(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	var productID string
	splitPath := strings.Split(u.Path, "/")
	for i, p := range splitPath {
		if p == "dp" || p == "product" {
			productID = splitPath[i+1]
			break
		} else if p == "gp" {
			return SanitizeURL(fmt.Sprintf("%s://%s/%s", u.Scheme, u.Host, u.Query()["url"][0]))
		}
	}

	if productID == "" {
		return "", errors.New("could not find productID")
	}

	s := fmt.Sprintf("%s://%s/dp/%s", u.Scheme, u.Host, productID)
	return s, nil
}
