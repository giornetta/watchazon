package watchazon

type Domain string

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
	Search(query string, domain Domain) ([]*Product, error)
	Update() error
	Listen() <-chan *Notification
}

type Locator interface {
	Locate(lat, long float32) (Domain, error)
}
