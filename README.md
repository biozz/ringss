# ringss

A simple Telegram bot, which wraps around Miniflux API for RSS feeds notifications.

## Environment variables

```
KEY                        TYPE                               DEFAULT    REQUIRED    DESCRIPTION
TELEGRAM_BOT_TOKEN         String                                                    
DATABASE_PATH              String                             db                     
MINIFLUX_BASE_URL          String                                                    
DEFAULT_POLLER_STATE       String                             1                      
ADMIN_USER_IDS             Comma-separated list of Integer                           
POLLER_INTERVAL_SECONDS    Integer                            10     
```
