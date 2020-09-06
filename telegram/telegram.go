package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/giornetta/watchazon"

	"github.com/giornetta/telebot"
)

type Bot struct {
	telegram *telebot.Bot
	service  watchazon.Service
	locator  watchazon.Locator
}

func New(token string, svc watchazon.Service, loc watchazon.Locator) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	return &Bot{
		telegram: b,
		service:  svc,
		locator:  loc,
	}, nil
}

func (b *Bot) Stop() {
	b.telegram.Stop()
}

func (b *Bot) Run() {
	b.telegram.Handle("/start", b.handleStart)
	b.telegram.Handle("/list", b.handleList)
	b.telegram.Handle(telebot.OnText, b.handleWatch)
	b.telegram.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		if strings.Contains(c.Data, "DELETE") {
			link := c.Data[8:]

			err := b.service.RemoveFromWatchList(link, c.Sender.ID)
			if err != nil {
				log.Printf("could not remove user %d from product %s: %v", c.Sender.ID, link, err)
			}

			_ = b.telegram.Delete(c.Message)

			_ = b.telegram.Respond(c, &telebot.CallbackResponse{
				Text: "Successfully Removed!",
			})
		} else if strings.Contains(c.Data, "OPEN") {
			fmt.Println("ciao")
		}

		_ = b.telegram.Respond(c, &telebot.CallbackResponse{})
	})
	b.telegram.Handle(telebot.OnQuery, b.handleQuery)

	go b.telegram.Start()

	format := "🔥 A product in your watchlist has changed price!\n\n<b>📦 Product:</b> %s\n<b>💵 Price:</b> %.2f €\n<b>🕛 Last check:</b> %s"
	for n := range b.service.Listen() {
		msg := fmt.Sprintf(format, n.Product.Title, n.Product.Price, n.Product.FormattedTime())
		_, _ = b.telegram.Send(sendableUser(n.UserID), msg, &telebot.SendOptions{
			ReplyMarkup: &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						telebot.InlineButton{
							Text: "✔️ Go to Amazon! ✔️",
							URL:  n.Product.Link,
						},
					},
				},
			},
			ParseMode: "HTML",
		})
	}
}

func (b *Bot) handleStart(m *telebot.Message) {
	_, _ = b.telegram.Send(m.Sender, "Welcome! Send an Amazon link to add it to your watchlist, or use the inline keyboard to search for products!", &telebot.ReplyMarkup{
		InlineKeyboard: [][]telebot.InlineButton{
			{
				{
					Text:                   "Search...",
					InlineQueryCurrentChat: "",
				},
			},
		},
	})
	return
}

func (b *Bot) handleWatch(m *telebot.Message) {
	substr := "https://www.amazon"
	if !strings.HasPrefix(m.Text, substr) {
		_, _ = b.telegram.Send(m.Sender, "That's not a valid Amazon link! Click the button below to start searching!", &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					{
						Text:                   "Search...",
						InlineQueryCurrentChat: "",
					},
				},
			},
		})
		return
	}
	_, _ = b.telegram.Send(m.Sender, "🔄 Adding your product...")

	err := b.service.AddToWatchList(m.Text, m.Sender.ID)
	if err != nil {
		_, _ = b.telegram.Send(m.Sender, err.Error())
		return
	}

	_, _ = b.telegram.Send(m.Sender, "✅ Product successfully added to the watchlist!")
}

func (b *Bot) handleList(m *telebot.Message) {
	products, err := b.service.GetUserWatchList(m.Sender.ID)
	if err != nil {
		_, _ = b.telegram.Send(m.Sender, "An error occurred! Sorry!")
		return
	}

	if len(products) == 0 {
		_, _ = b.telegram.Send(m.Sender, "There are no products in your watchlist!")
		return
	}

	msgFormat := "<b>📦 Product:</b> %s\n<b>💵 Price:</b> %.2f €\n<b>🕛 Last check:</b> %s"
	for _, p := range products {
		_, err := b.telegram.Send(m.Sender, fmt.Sprintf(msgFormat, p.Title, p.Price, p.FormattedTime()), &telebot.SendOptions{
			ReplyMarkup: &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						telebot.InlineButton{
							Unique: "DELETE",
							Text:   "❌ Delete! ❌",
							Data:   p.Link,
						},
					},
					{
						telebot.InlineButton{
							Text: "✔️ Go to Amazon! ✔️",
							URL:  p.Link,
						},
					},
				},
			},
			ParseMode: "HTML",
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func (b *Bot) handleQuery(q *telebot.Query) {
	var loc watchazon.Domain
	var err error
	if q.Location != nil {
		loc, err = b.locator.Locate(q.Location.Lat, q.Location.Lng)
	} else {
		loc = "com"
	}
	if err != nil {
		log.Println(err)
		loc = "com"
	}
	fmt.Println(loc)
	products, err := b.service.Search(q.Text, loc)
	if err != nil {
		log.Println(err)
		return
	}

	l := min(50, len(products))
	tgRes := make(telebot.Results, l)
	for i, p := range products {
		if i == 50 {
			break
		}

		result := &telebot.ArticleResult{
			Title:       p.Title,
			Text:        p.Link,
			URL:         p.Link,
			HideURL:     false,
			Description: fmt.Sprintf("%.2f€", p.Price),
			ThumbURL:    p.Image,
		}

		tgRes[i] = result
		tgRes[i].SetResultID(strconv.Itoa(i)) // It's needed to set a unique string ID for each result
	}
	if err = b.telegram.Answer(q, &telebot.QueryResponse{
		Results:   tgRes,
		CacheTime: 60, // a minute
	}); err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("User %d looked for %s: %d results sent", q.From.ID, q.Text, l)
}

type recipient int

func (r recipient) Recipient() string {
	return strconv.Itoa(int(r))
}

func sendableUser(user int) recipient {
	return recipient(user)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
