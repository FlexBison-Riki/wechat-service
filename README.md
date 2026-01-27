# WeChat Service Project

## Overview

Production-ready WeChat Service Account (服务号) integration service built with Go, featuring clean architecture, PostgreSQL persistence, and Redis caching.

## Quick Start

```bash
# Copy configuration
cp config.example.yaml config.yaml
cp .env.example .env

# Edit config.yaml and .env with your WeChat credentials

# Start with Docker
docker-compose up -d

# Or run locally
make deps
make run
```

## Project Structure

```
wechat-service/
├── cmd/
│   └── server/main.go           # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── handler/                 # HTTP/WeChat handlers
│   ├── model/                   # Data models (User, Message, Event)
│   ├── repository/              # Data access layer
│   │   └── sql/                 # PostgreSQL operations
│   └── service/                 # Business logic
├── pkg/
│   ├── cache/                   # Redis/Memory cache
│   └── logger/                  # Structured logging
├── migrations/                  # Database migrations
├── docs/                        # Documentation
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## Configuration

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `WECHAT_APPID` | Yes | WeChat AppID |
| `WECHAT_APPSECRET` | Yes | WeChat AppSecret |
| `WECHAT_TOKEN` | Yes | Server verification token |
| `WECHAT_ENCODING_AES_KEY` | No | Message encryption key |
| `REDIS_ADDR` | No | Redis address (default: localhost:6379) |
| `DB_HOST` | No | PostgreSQL host (default: localhost) |
| `DB_PORT` | No | PostgreSQL port (default: 5432) |
| `DB_NAME` | No | Database name (default: wechat_service) |

### WeChat Configuration

1. Register Service Account at [WeChat Official Platform](https://mp.weixin.qq.com/)
2. Configure server URL at [WeChat Developer Platform](https://developers.weixin.qq.com/platform/)
3. Set webhook URL: `https://your-domain.com/wechat`

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      API Layer                               │
│  cmd/server/main.go → internal/handler/                     │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                    Service Layer                             │
│         internal/service/ (MessageService, UserService)      │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                   Repository Layer                           │
│         internal/repository/ (DB + Cache access)             │
│                    ┌──────────┐                              │
│                    │ Interface │                              │
│                    └────┬─────┘                              │
│                         │                                    │
│         ┌───────────────┴───────────────┐                    │
│         ▼                               ▼                    │
│  ┌─────────────┐               ┌─────────────┐              │
│  │ DB Repository│               │   Cache     │              │
│  │ (PostgreSQL) │               │  (Redis)    │              │
│  └─────────────┘               └─────────────┘              │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│                     SQL Layer                                │
│              internal/repository/sql/                        │
│   (UserSQL, MessageSQL, UnitOfWork, Transactions)            │
└─────────────────────────────────────────────────────────────┘
```

## Database Schema

### Users Table
Stores WeChat user information with subscription status and location.

```sql
CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    openid          VARCHAR(64) NOT NULL UNIQUE,
    nickname        VARCHAR(256),
    sex             SMALLINT DEFAULT 0,
    city            VARCHAR(64),
    province        VARCHAR(64),
    subscribe_status SMALLINT DEFAULT 0,
    latitude        DECIMAL(10, 7),
    longitude       DECIMAL(10, 7),
    tags            JSONB DEFAULT '[]',
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Messages Table
Stores all messages exchanged with users.

```sql
CREATE TABLE messages (
    id              BIGSERIAL PRIMARY KEY,
    msg_id          BIGINT NOT NULL,
    from_user       VARCHAR(64) NOT NULL,
    to_user         VARCHAR(64) NOT NULL,
    direction       VARCHAR(4) NOT NULL DEFAULT 'in',
    msg_type        VARCHAR(32) NOT NULL,
    content         TEXT,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Events Table
Stores user events (subscribe, scan, click, location, etc.).

```sql
CREATE TABLE events (
    id              BIGSERIAL PRIMARY KEY,
    openid          VARCHAR(64) NOT NULL,
    event_type      VARCHAR(32) NOT NULL,
    event_key       VARCHAR(256),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

Run migrations:
```bash
psql -h localhost -U postgres -d wechat_service -f migrations/001_init_schema.sql
```

## API Endpoints

### WeChat Webhook
- `GET /wechat` - Server verification
- `POST /wechat` - Message/event callback

### REST API
- `GET /health` - Health check
- `GET /api/v1/users/:openid` - Get user by OpenID
- `GET /api/v1/messages` - List messages with filters

## Key Features

| Feature | Implementation |
|---------|---------------|
| **Message Handling** | Text, image, voice, video, events |
| **Custom Menus** | Click, view, scan, location, media |
| **User Management** | Subscribe, unsubscribe, tags, location |
| **Transactions** | Unit of Work pattern for atomic operations |
| **Caching** | Read-through cache with Redis |
| **Logging** | Structured JSON logging with file rotation |
| **Graceful Shutdown** | Signal handling, 5s timeout |

## Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Build binary
make build

# Run locally
make run

# Docker build
make docker-build
```

## SDK Integration

Uses [silenceper/wechat](https://github.com/silenceper/wechat) SDK:

```go
import "github.com/silenceper/wechat/v2"

wc := wechat.NewWechat()
officialAccount := wc.GetOfficialAccount(cfg)
server := officialAccount.GetServer(req, writer)
server.SetMessageHandler(func(msg *message.MixMessage) *message.Reply {
    // Handle message
})
server.Serve()
server.Send()
```

## Documentation

- [System Design: Message & Menu](docs/system-design-message-menu.md)
- [Integration Guide](docs/integration.md)
- [SDK Analysis](docs/sdk-analysis.md)

## Resources

- [WeChat Service Account Docs](https://developers.weixin.qq.com/doc/service/)
- [WeChat Developer Platform](https://developers.weixin.qq.com/platform/)
- [WeChat SDK (Go) Docs](https://silenceper.com/wechat)
- [SDK Examples](https://github.com/gowechat/example)

## License

Apache License 2.0
