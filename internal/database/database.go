package database

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/gomodule/redigo/redis"
)

type Database struct {
	Raw                          redis.Conn
	userStateKeyBuilder          func(int) []byte
	userMinifluxAPIKeyKeyBuilder func(int) []byte
	feedKeyBuilder               func(int64) []byte
}

func New(dbURL string) (*Database, error) {
	c, err := redis.DialURL(dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to the database: %v", err)
	}
	return &Database{
		Raw: c,
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
	d.Raw.Close()
}

type State string

const (
	StateUnknown               State = "unknown"
	StatePending               State = "pending"
	StateWaitingFeedRef        State = "waiting_feed_ref"
	StateWaitingFeedID         State = "waiting_feed_id"
	StateWaitingMinifluxAPIKey State = "waiting_miniflux_api_key"
)

func (d *Database) SetUserState(userID int, state State) {
	key := d.userStateKeyBuilder(userID)
	err := d.Raw.Send("SET", key, state)
	if err != nil {
		log.Printf("Unknown SET error for user state: %v", err)
	}
}

func (d *Database) GetUserState(userID int) State {
	key := d.userStateKeyBuilder(userID)
	state, err := redis.String(d.Raw.Do("GET", key))
	if err != nil {
		return StateUnknown
	}
	return State(state)
}

func (d *Database) ClearUserState(userID int) {
	key := d.userStateKeyBuilder(userID)
	_, err := redis.Bool(d.Raw.Do("DEL", key))
	if err != nil {
		log.Printf("Unknown db.Delete error for user clear state: %v", err)
	}
}

func (d *Database) SetPollerEnabled(enabled string) {
	result, err := redis.String(d.Raw.Do("SET", "poller:enabled", enabled))
	if err != nil {
		log.Printf("Unknown db.Put error for poller state: %v", err)
	}
	log.Println(result)
}

func (d *Database) GetPollerState() string {
	state, err := redis.String(d.Raw.Do("GET", "poller:enabled"))
	if err != nil {
		log.Printf("Unknown db.Get error for poller state: %v", err)
	}
	return string(state)
}

func (d *Database) GetKeysWIthPrefix(prefix string) []string {
	keys, err := redis.Strings(d.Raw.Do("KEYS", "*"))
	if err != nil {
		log.Printf("Unknown KEYS error: %v", err)
	}
	var result []string
	for _, key := range keys {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		result = append(result, key)
	}
	return result
}

func (d *Database) SetFeed(feedID int64, userID int) {
	key := d.feedKeyBuilder(feedID)
	value := []byte(strconv.Itoa(userID))
	_, err := redis.Bool(d.Raw.Do("SET", key, value))
	if err != nil {
		log.Printf("Unknown db.Put error for feed: %v", err)
	}
}

func (d *Database) GetFeed(feedID int64) int {
	key := d.feedKeyBuilder(feedID)
	telegramUserID, _ := redis.Int(d.Raw.Do("GET", key))
	if telegramUserID == 0 {
		log.Printf("No telegram user id for feed %d", feedID)
		return 0
	}
	return telegramUserID
}

func (d *Database) ClearFeed(feedID int64) {
	key := d.feedKeyBuilder(feedID)
	_, err := redis.Bool(d.Raw.Do("DEL", key))
	if err != nil {
		log.Printf("Unable to delete feed: %v\n", err)
	}
}

func (d *Database) GetMinifluxAPIKey(userID int) string {
	key := d.userMinifluxAPIKeyKeyBuilder(userID)
	result, err := redis.String(d.Raw.Do("GET", key))
	if err != nil {
		return ""
	}
	return result
}

func (d *Database) SetMinifluxAPIKey(userID int, minifluxAPIKey string) {
	key := d.userMinifluxAPIKeyKeyBuilder(userID)
	err := d.Raw.Send("SET", key, minifluxAPIKey)
	if err != nil {
		log.Printf("Unknown SET error for miniflux api key: %v", err)
	}
}
