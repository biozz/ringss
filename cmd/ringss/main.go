package main

import (
	"flag"
	"fmt"
	"github.com/biozz/ringss/internal/config"
	"github.com/biozz/ringss/internal/database"
	"github.com/biozz/ringss/internal/poller"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
	tb "gopkg.in/tucnak/telebot.v2"
	miniflux "miniflux.app/client"
)

const (
	Start       = "/start"
	AddFeed     = "/addfeed"
	RemoveFeed  = "/removefeed"
	Cancel      = "/cancel"
	KillPoller  = "/killpoller"
	StartPoller = "/startpoller"
	Test        = "/test"
	DB          = "/db"
)

var (
	showEnv = flag.Bool("env", false, "Display env vars")
)

func main() {
	flag.Parse()

	var c config.EnvConfig

	if *showEnv {
		envconfig.Usage("", &c)
		return
	}

	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal(err.Error())
	}

	db, _ := database.New(c.DatabasePath)
	defer db.DeferredAction()

	b, err := tb.NewBot(tb.Settings{
		Token:  c.TelegramBotToken,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(Start, func(m *tb.Message) {
		apiKey := db.GetMinifluxAPIKey(m.Sender.ID)
		if apiKey != "" {
			b.Send(m.Sender, "Твой профиль готов к работе. Можно начать с добавления фида /addfeed")
			return
		}
		b.Send(m.Sender, "Отправь API ключ Miniflux")
		db.SetUserState(m.Sender.ID, database.StateWaitingMinifluxAPIKey)
	})

	b.Handle(Test, func(m *tb.Message) {
		fmt.Println(m.Sender.ID)
		b.Send(m.Sender, "This is a test command, it might do something magical sometimes.")
	})

	b.Handle(DB, func(m *tb.Message) {
		if !isAdmin(m.Sender.ID, c.AdminUserIds) {
			return
		}
		parts := strings.Split(m.Text, " ")
		cmd := parts[1]
		var result []byte
		switch cmd {
		case "get":
			result, err = db.Raw.Get([]byte(parts[2]))
		case "put":
			err = db.Raw.Put([]byte(parts[2]), []byte(parts[3]))
		case "delete":
			err = db.Raw.Delete([]byte(parts[2]))
		default:
			b.Send(m.Sender, "Invalid command (get, put, delete)")
		}
		if err != nil {
			b.Send(m.Sender, fmt.Sprintf("NOK\n%v", err))
			return
		}
		msg := "OK"
		if len(result) > 0 {
			msg += fmt.Sprintf("\n%s", string(result))
		}
		b.Send(m.Sender, msg)
	})

	b.Handle(Cancel, func(m *tb.Message) {
		db.ClearUserState(m.Sender.ID)
		b.Send(m.Sender, "Операция отменена, можно начать заново")
	})

	b.Handle(AddFeed, func(m *tb.Message) {
		b.Send(m.Sender, "Отправь ссылку на фид или id фида (можно отменить это действие с помощью /cancel)")
		db.SetUserState(m.Sender.ID, database.StateWaitingFeedRef)
	})

	b.Handle(RemoveFeed, func(m *tb.Message) {
		b.Send(m.Sender, "Отправь id фида для отключения (можно отменить это действие с помощью /cancel)")
		db.SetUserState(m.Sender.ID, database.StateWaitingFeedID)
	})

	b.Handle(KillPoller, func(m *tb.Message) {
		db.SetPollerEnabled("0")
	})

	b.Handle(StartPoller, func(m *tb.Message) {
		db.SetPollerEnabled("1")
	})

	b.Handle(tb.OnText, func(m *tb.Message) {
		state := db.GetUserState(m.Sender.ID)
		defer func() {
			db.SetUserState(m.Sender.ID, database.StatePending)
		}()
		switch state {
		case database.StateWaitingMinifluxAPIKey:
			db.SetMinifluxAPIKey(m.Sender.ID, m.Text)
			db.SetUserState(m.Sender.ID, database.StatePending)
			b.Send(m.Sender, "Ключ сохранён. Теперь можно добавить фид с помощью /addfeed.")
		case database.StateWaitingFeedRef:
			minifluxAPIKey := db.GetMinifluxAPIKey(m.Sender.ID)
			if err != nil {
				b.Send(m.Sender, fmt.Sprintf("Не нашел ключ клиента miniflux: %v", err))
				return
			}
			client := miniflux.New(c.MinifluxBaseURL, minifluxAPIKey)
			markAsRead := true
			feedID, _ := strconv.ParseInt(m.Text, 10, 64)
			if feedID != 0 {
				feed, _ := client.Feed(feedID)
				if feed == nil {
					b.Send(m.Sender, "Фид с таким id не найден")
					return
				}
				markAsRead = false
			} else {
				categories, err := client.Categories()
				if err != nil {
					b.Send(m.Sender, "Категории не найдены")
					return
				}
				feedID, err = client.CreateFeed(&miniflux.FeedCreationRequest{
					FeedURL:    m.Text,
					CategoryID: categories[0].ID,
				})
				if strings.Contains(err.Error(), "already exists") {
					b.Send(m.Sender, "Этот фид уже добавлен, попробуй другой")
					return
				}
				if err != nil {
					b.Send(m.Sender, fmt.Sprintf("Что-то пошло не так: %v", err))
					return
				}
			}
			b.Send(m.Sender, fmt.Sprintf("Фид добавлен, id - %d. Убрать оповещения можно с помощью /removefeed", feedID))
			db.SetFeed(feedID, m.Sender.ID)
			if markAsRead {
				client.MarkFeedAsRead(feedID)
				b.Send(m.Sender, "Я пометил все новости как прочитанные, теперь будут приходить оповещения только о новых")
				return
			}
			b.Send(m.Sender, "Сейчас могут придти оповещения из только что добавленного фида")
		case database.StateWaitingFeedID:
			feedID, _ := strconv.ParseInt(m.Text, 10, 64)
			if feedID == 0 {
				b.Send(m.Sender, "Неправильный id фида")
				return
			}
			minifluxAPIKey := db.GetMinifluxAPIKey(m.Sender.ID)
			if err != nil {
				b.Send(m.Sender, fmt.Sprintf("Не нашел ключ клиента miniflux: %v", err))
				return
			}
			client := miniflux.New(c.MinifluxBaseURL, minifluxAPIKey)
			feed, _ := client.Feed(feedID)
			if feed == nil {
				b.Send(m.Sender, "Фид не найден")
				return
			}
			db.ClearFeed(feedID)
			b.Send(m.Sender, fmt.Sprintf("Оповещения для фида %d отключены", feedID))
		}
	})

	db.SetPollerEnabled(c.DefaultPollerState)
	p := poller.New(db, b, c.MinifluxBaseURL, c.PollerIntervalSeconds)
	go p.Run()

	b.Start()
}

func isAdmin(userID int, adminUserIDs []int) bool {
	for _, id := range adminUserIDs {
		if userID == id {
			return true
		}
	}
	return false
}
