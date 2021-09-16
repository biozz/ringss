# ringss

A simple Telegram bot, which wraps around Miniflux API for RSS feeds notifications.

## Environment variables

```
KEY                        TYPE       DEFAULT                   REQUIRED    DESCRIPTION
TELEGRAM_BOT_TOKEN         String                               true        API Token obtained from @BotFather
DATABASE_URL               String     redis://localhost:6379    true        Redis connection URL
MINIFLUX_BASE_URL          String                               true        A Miniflux instance URL without trailing slash
DEFAULT_POLLER_STATE       String     1                                     Enables or disables poller on start. Can be re-enabled at runtime with poller:enable key set to 1
POLLER_INTERVAL_SECONDS    Integer    10                                    Time between poller cycles
```
