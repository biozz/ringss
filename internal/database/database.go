package database

import (
	"fmt"
	"git.mills.io/prologic/bitcask"
	"log"
	"strconv"
)

type Database struct {
	db                           *bitcask.Bitcask
	userStateKeyBuilder          func(int) []byte
	userMinifluxAPIKeyKeyBuilder func(int) []byte
	feedKeyBuilder               func(int64) []byte
}

func New(dbPath string) (*Database, error) {
	db, _ := bitcask.Open(dbPath)
	return &Database{
		db: db,
		userStateKeyBuilder: func(userID int) []byte {
			return []byte(fmt.Sprintf("user:%d:state", userID))
		},
		userMinifluxAPIKeyKeyBuilder: func(userID int) []byte {
			return []byte(fmt.Sprintf("user:%d:miniflux", userID))
		},
		feedKeyBuilder: func(feedID int64) []byte {
			return []byte(fmt.Sprintf("feed:%d", feedID))
		},
	}, nil
}

func (d *Database) DeferredAction() {
	d.db.Close()
}

type State string

const (
	StateUnknown               State = "unknown"
	StatePending               State = "pending"
	StateWaitingFeedRef        State = "waiting_feed_ref"
	StateWaitingFeedID        State = "waiting_feed_id"
	StateWaitingMinifluxAPIKey State = "waiting_miniflux_api_key"
)

func (d *Database) SetUserState(userID int, state State) {
	key := d.userStateKeyBuilder(userID)
	err := d.db.Put(key, []byte(state))
	if err != nil {
		log.Printf("Unknown db.Put error for user state: %v", err)
	}
}

func (d *Database) GetUserState(userID int) State {
	key := d.userStateKeyBuilder(userID)
	state, err := d.db.Get(key)
	if err != nil {
		return StateUnknown
	}
	return State(state)
}

func (d *Database) ClearUserState(userID int) {
	key := d.userStateKeyBuilder(userID)
	err := d.db.Delete(key)
	if err != nil {
		log.Printf("Unknown db.Delete error for user clear state: %v", err)
	}
}

func (d *Database) SetPollerEnabled(enabled string) {
	err := d.db.Put([]byte("poller:enabled"), []byte(enabled))
	if err != nil {
		log.Printf("Unknown db.Put error for poller state: %v", err)
	}
}

func (d *Database) GetPollerState() string {
	pollerState, err := d.db.Get([]byte("poller:enabled"))
	if err != nil {
		log.Printf("Unknown db.Get error for poller state: %v", err)
	}
	return string(pollerState)
}

func (d *Database) ScanFeeds(callable func(key []byte) error) {
	err := d.db.Scan([]byte("feed:"), callable)
	if err != nil {
		log.Printf("Unknown Scan error for feeds: %v", err)
	}
}

func (d *Database) SetFeed(feedID int64, userID int) {
	key := d.feedKeyBuilder(feedID)
	value := []byte(strconv.Itoa(userID))
	err := d.db.Put(key, value)
	if err != nil {
		log.Printf("Unknown db.Put error for feed: %v", err)
	}
}

func (d *Database) GetFeed(feedID int64) int {
	key := d.feedKeyBuilder(feedID)
	telegramUserIDRaw, _ := d.db.Get(key)
	if telegramUserIDRaw == nil {
		log.Printf("No telegram user id for feed %d", feedID)
		return 0
	}
	telegramUserID, _ := strconv.Atoi(string(telegramUserIDRaw))
	return telegramUserID
}

func (d *Database) ClearFeed(feedID int64) {
	key := d.feedKeyBuilder(feedID)
	err := d.db.Delete(key)
	if err != nil {
		log.Printf("Unable to delete feed: %v\n", err)
	}
}

func (d *Database) GetMinifluxAPIKey(userID int) string {
	key := d.userMinifluxAPIKeyKeyBuilder(userID)
	apiKey, err := d.db.Get(key)
	if err != nil {
		log.Printf("Unknown db.Get error for miniflux api key: %v", err)
		return ""
	}
	return string(apiKey)
}

func (d *Database) SetMinifluxAPIKey(userID int, minifluxAPIKey string) {
	key := d.userMinifluxAPIKeyKeyBuilder(userID)
	err := d.db.Put(key, []byte(minifluxAPIKey))
	if err != nil {
		log.Printf("Unknown db.Put error for miniflux api key: %v", err)
	}
}
