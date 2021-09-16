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

var reservedChars = []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}

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
	    keys := p.db.GetKeysWIthPrefix("feed:")
		for _, key := range keys {
			parts := strings.Split(key, ":")
			// we are interested in feed:123123123 -> 123123123 (telegram user id
			//                        0       1
			if len(parts) != 2 {
				log.Printf("skipping %s, length is %d\n", key, len(parts))
				continue
			}
			feedID, _ := strconv.ParseInt(parts[1], 10, 64)
			telegramUserID := p.db.GetFeed(feedID)
			minifluxAPIKey := p.db.GetMinifluxAPIKey(telegramUserID)
			if minifluxAPIKey == "" {
				log.Println("Unable to get miniflux key from the database")
				continue
			}
			client := miniflux.New(p.minifluxBaseURL, minifluxAPIKey)
			entries, err := client.FeedEntries(feedID, &miniflux.Filter{Status: miniflux.EntryStatusUnread})
			if err != nil {
				log.Printf("unable to get unread entries: %v\n", err)
				continue
			}
			telegramUser := tb.User{ID: telegramUserID}
			if len(entries.Entries) == 0 {
				log.Println("No unread entries, exiting early")
				continue
			}
			entriesToUpdate := make([]int64, entries.Total)
			for i, entry := range entries.Entries {
				_, err := p.b.Send(
					&telegramUser,
					fmt.Sprintf(
						"*%s*\n%s",
						escapeTelegramString(entry.Title),
						escapeTelegramString(entry.URL),
					),
					&tb.SendOptions{
						ParseMode: tb.ModeMarkdownV2,
						ReplyMarkup: &tb.ReplyMarkup{
							InlineKeyboard: [][]tb.InlineButton{
								{
									tb.InlineButton{
										// TODO: add prettier markup?
										Text: escapeTelegramString(entry.Feed.Title), 
										URL: fmt.Sprintf("%sfeed/%d/entries/all", p.minifluxBaseURL, entry.FeedID),
									},
								},
							},
						},
					},
				)
				if err != nil {
					log.Printf("Telegram send error: %v\n", err)
					continue
				}
				entriesToUpdate[i] = entry.ID
			}
			err = client.UpdateEntries(entriesToUpdate, miniflux.EntryStatusRead)
			if err != nil {
				fmt.Printf("Error updating entries: %v \n", err)
				continue
			}
			continue
		}
		log.Println("Poller cycle ended, waiting...")
		time.Sleep(time.Duration(p.pollerIntervalSeconds*1000) * time.Millisecond)
	}
}

func escapeTelegramString(s string) string {
	result := s
	for _, rc := range reservedChars {
		result = strings.ReplaceAll(result, rc, fmt.Sprintf("\\%s", rc))
	}
	return result
}
