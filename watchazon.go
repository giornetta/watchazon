package watchazon

import "time"

// A Domain represents a top-level-domain (e.g. com, es, it...) for an Amazon website.
type Domain string

// A Product represents the data collected for each Amazon product.
type Product struct {
	Title     string
	Image     string
	Link      string
	Price     float64
	CheckedAt time.Time
}

// Formatted time returns the time of the last check for the product, formatted in a nice way.
func (p Product) FormattedTime() string {
	return p.CheckedAt.Format("2 Jan 2006 at 15:04")
}

// Notification is used to represent which user has to receive a notification on a specific product.
type Notification struct {
	Product *Product
	UserID  int64
}

// Service defines the required methods of the Bot.
type Service interface {
	AddToWatchList(link string, userID int64) error
	RemoveFromWatchList(link string, userID int64) error
	GetUserWatchList(user int64) ([]*Product, error)
	Search(query string, domain Domain) ([]*Product, error)
	Update() error
	Listen() <-chan *Notification
}

// Locator is used to decide which local version of the Amazon website must be scraped based on the user's location.
type Locator interface {
	Locate(lat, long float32) (Domain, error)
}
