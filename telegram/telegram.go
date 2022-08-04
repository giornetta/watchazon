package telegram

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/giornetta/watchazon"

	telebot "gopkg.in/telebot.v3"
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
	shouldSendPrompt := b.telegram.Group()

	shouldSendPrompt.Use(sendSearchButton)

	shouldSendPrompt.Handle("/start", b.handleStart)
	shouldSendPrompt.Handle("/list", b.handleList)
	shouldSendPrompt.Handle(telebot.OnText, b.handleWatch)

	b.telegram.Handle(telebot.OnCallback, b.handleCallback)

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

func (b *Bot) handleStart(ctx telebot.Context) error {
	return ctx.Send("Welcome to Watchazon Bot!")
}

func (b *Bot) handleWatch(ctx telebot.Context) error {
	substr := "https://www.amazon"
	if !strings.HasPrefix(ctx.Text(), substr) {
		return ctx.Send("That's not a valid Amazon link!")
	}

	_ = ctx.Send("🔄 Adding your product...")

	err := b.service.AddToWatchList(ctx.Text(), ctx.Sender().ID)
	if err != nil {
		return ctx.Send(err.Error())
	}

	return ctx.Send("✅ Product successfully added to the watchlist!")
}

func (b *Bot) handleList(ctx telebot.Context) error {
	products, err := b.service.GetUserWatchList(ctx.Sender().ID)
	if err != nil {
		return ctx.Send("An error occurred! Sorry!")
	}

	if len(products) == 0 {
		return ctx.Send("There are no products in your watchlist!")
	}

	msgFormat := "<b>📦 Product:</b> %s\n<b>💵 Price:</b> %.2f €\n<b>🕛 Last check:</b> %s"
	for _, p := range products {
		err := ctx.Send(fmt.Sprintf(msgFormat, p.Title, p.Price, p.FormattedTime()), &telebot.SendOptions{
			ReplyMarkup: &telebot.ReplyMarkup{
				InlineKeyboard: [][]telebot.InlineButton{
					{
						telebot.InlineButton{
							Unique: "DELETE",
							Text:   "❌ Delete!",
							Data:   p.Link,
						},
						telebot.InlineButton{
							Text: "✔️ Go to Amazon!",
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

	return nil
}

func (b *Bot) handleQuery(ctx telebot.Context) error {
	q := ctx.Query()

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

	products, err := b.service.Search(q.Text, loc)
	if err != nil {
		return err
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

	if err = ctx.Answer(&telebot.QueryResponse{
		Results:   tgRes,
		CacheTime: 60, // a minute
	}); err != nil {
		return err
	}

	log.Printf("User %d looked for %s: %d results sent", q.Sender.ID, q.Text, l)
	return nil
}

func (b *Bot) handleCallback(ctx telebot.Context) error {
	data := ctx.Data()

	if strings.Contains(data, "DELETE") {
		link := data[8:]

		err := b.service.RemoveFromWatchList(link, ctx.Sender().ID)
		if err != nil {
			return fmt.Errorf("could not remove user %d from product %s: %v", ctx.Sender().ID, link, err)
		}

		_ = b.telegram.Delete(ctx.Message())

		_ = ctx.Respond(&telebot.CallbackResponse{
			Text: "Successfully Removed!",
		})
	} else if strings.Contains(data, "OPEN") {
		fmt.Println("ciao")
	}

	return ctx.Respond(&telebot.CallbackResponse{})
}

func sendSearchButton(next telebot.HandlerFunc) telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		defer ctx.Send("Send an Amazon link to add it to your watchlist, or click below to search for products!", &telebot.ReplyMarkup{
			InlineKeyboard: [][]telebot.InlineButton{
				{
					{
						Text:            "Search...",
						InlineQueryChat: "",
					},
				},
			},
		})

		return next(ctx)
	}
}

type recipient int64

func (r recipient) Recipient() string {
	return strconv.FormatInt(int64(r), 10)
}

func sendableUser(user int64) recipient {
	return recipient(user)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
