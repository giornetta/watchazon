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
	bot     *telebot.Bot
	service watchazon.Service
}

func New(token string, s watchazon.Service) (*Bot, error) {
	b, err := telebot.NewBot(telebot.Settings{
		Token:  token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		return nil, err
	}

	return &Bot{
		bot:     b,
		service: s,
	}, nil
}

func (b *Bot) Stop() {
	b.bot.Stop()
}

func (b *Bot) Run() {
	b.bot.Handle("/start", b.handleStart)
	b.bot.Handle("/list", b.handleList)
	b.bot.Handle(telebot.OnText, b.handleWatch)
	b.bot.Handle(telebot.OnCallback, func(c *telebot.Callback) {
		if strings.Contains(c.Data, "DELETE") {
			link := c.Data[8:]

			err := b.service.RemoveFromWatchList(link, c.Sender.ID)
			if err != nil {
				log.Printf("could not remove user %d from product %s: %v", c.Sender.ID, link, err)
			}

			_ = b.bot.Delete(c.Message)

			_ = b.bot.Respond(c, &telebot.CallbackResponse{
				Text: "Successfully Removed!",
			})
		} else if strings.Contains(c.Data, "OPEN") {
			fmt.Println("ciao")
		}

		_ = b.bot.Respond(c, &telebot.CallbackResponse{})
	})
	b.bot.Handle(telebot.OnQuery, b.handleQuery)

	go b.bot.Start()

	for n := range b.service.Listen() {
		msg := fmt.Sprintf("%s has changed price! New price is %.2f €!\n%s", n.Product.Title, n.Product.Price, n.Product.Link)
		_, _ = b.bot.Send(sendableUser(n.UserID), msg)
	}
}

func (b *Bot) handleStart(m *telebot.Message) {
	_, _ = b.bot.Send(m.Sender, "Welcome! Send an Amazon link to add it to your watchlist, or use the inline keyboard to search for products!")
}

func (b *Bot) handleWatch(m *telebot.Message) {
	substr := "https://www.amazon"
	if !strings.HasPrefix(m.Text, substr) {
		_, _ = b.bot.Send(m.Sender, "That's not a valid Amazon link! Click the button below to start searching!", &telebot.ReplyMarkup{
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

	err := b.service.AddToWatchList(m.Text, m.Sender.ID)
	if err != nil {
		_, _ = b.bot.Send(m.Sender, err.Error())
		return
	}

	_, _ = b.bot.Send(m.Sender, "Product successfully added to the watchlist!")
}

func (b *Bot) handleList(m *telebot.Message) {
	products, err := b.service.GetUserWatchList(m.Sender.ID)
	if err != nil {
		_, _ = b.bot.Send(m.Sender, "An error occurred! Sorry!")
		return
	}

	if len(products) == 0 {
		_, _ = b.bot.Send(m.Sender, "There are no products in your watchlist!")
		return
	}

	for _, p := range products {
		_, err := b.bot.Send(m.Sender, fmt.Sprintf("%s: %.2f €", p.Title, p.Price), &telebot.SendOptions{
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
		})
		if err != nil {
			log.Println(err)
		}
	}
}

type recipient int

func (r recipient) Recipient() string {
	return strconv.Itoa(int(r))
}
func sendableUser(user int) recipient {
	return recipient(user)
}

func (b *Bot) handleQuery(q *telebot.Query) {
	products, err := b.service.Search(q.Text)
	if err != nil {
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
	if err = b.bot.Answer(q, &telebot.QueryResponse{
		Results:   tgRes,
		CacheTime: 60, // a minute
	}); err != nil {
		fmt.Println(err)
		return
	}

	log.Printf("User %d looked for %s: %d results sent", q.From.ID, q.Text, l)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
