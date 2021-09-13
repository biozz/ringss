package poller

import (
	"fmt"
	"github.com/biozz/ringss/internal/database"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	miniflux "miniflux.app/client"
	"strconv"
	"strings"
	"time"
)

type Poller struct {
	db                    *database.Database
	b                     *tb.Bot
	minifluxBaseURL       string
	pollerIntervalSeconds int
}

func New(db *database.Database, b *tb.Bot, minifluxBaseURL string, pollerIntervalSeconds int) *Poller {
	return &Poller{
		db:                    db,
		b:                     b,
		minifluxBaseURL:       minifluxBaseURL,
		pollerIntervalSeconds: pollerIntervalSeconds,
	}
}

func (p *Poller) Run() {
	log.Println("Poller started")
	for {
		pollerState := p.db.GetPollerState()
		if pollerState != "1" {
			log.Println("Poller is disabled, cycle skipped")
			time.Sleep(time.Duration(p.pollerIntervalSeconds*1000) * time.Millisecond)
			continue
		}
		log.Println("Poller cycle started")
		p.db.ScanFeeds(func(key []byte) error {
			strKey := string(key)
			parts := strings.Split(strKey, ":")
			// we are interested in feed:123123123 -> 123123123 (telegram user id
			//                        0       1
			if len(parts) != 2 {
				log.Printf("skipping %s, length is %d\n", strKey, len(parts))
				return nil
			}
			feedID, _ := strconv.ParseInt(parts[1], 10, 64)
			telegramUserID := p.db.GetFeed(feedID)
			minifluxAPIKey := p.db.GetMinifluxAPIKey(telegramUserID)
			if minifluxAPIKey == "" {
				log.Println("Unable to get miniflux key from the database")
				return nil
			}
			client := miniflux.New(p.minifluxBaseURL, minifluxAPIKey)
			entries, err := client.FeedEntries(feedID, &miniflux.Filter{Status: miniflux.EntryStatusUnread})
			if err != nil {
				log.Println("unable to get unread entries")
				return nil
			}
			telegramUser := tb.User{ID: telegramUserID}
			if len(entries.Entries) == 0 {
				log.Println("No unread entries, exiting early")
				return nil
			}
			entriesToUpdate := make([]int64, entries.Total)
			for i, entry := range entries.Entries {
				p.b.Send(
					&telegramUser,
					fmt.Sprintf("%s: %s\n**%s**", entry.Feed.Category.Title, entry.Feed.Title, entry.Title), &tb.SendOptions{
						ParseMode: tb.ModeMarkdownV2,
						ReplyMarkup: &tb.ReplyMarkup{
							InlineKeyboard: [][]tb.InlineButton{
								{
									tb.InlineButton{Text: "Внешняя ссылка", URL: entry.URL},
									tb.InlineButton{Text: "Ссылка на miniflux", URL: fmt.Sprintf("%s/feed/%d/entry/%d", p.minifluxBaseURL, entry.FeedID, entry.ID)},
								},
							},
						}})
				entriesToUpdate[i] = entry.ID
			}
			err = client.UpdateEntries(entriesToUpdate, miniflux.EntryStatusRead)
			if err != nil {
				fmt.Printf("Error updating entries: %v \n", err)
				return nil
			}
			return nil
		})
		log.Println("Poller cycle ended, waiting...")
		time.Sleep(time.Duration(p.pollerIntervalSeconds*1000) * time.Millisecond)
	}
}
