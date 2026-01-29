## Summary

Production-ready WeChat Service Account integration with Prometheus metrics, featuring clean architecture and simplified deployment.

## Features

- WeChat Integration: Message/event handling with silenceper/wechat SDK
- Prometheus Metrics: HTTP, message, event, and system metrics at /metrics
- Health Check: /health endpoint for liveness probes
- Clean Architecture: Handler -> Service -> Repository pattern
- Docker Support: Multi-stage build for containerized deployment
- CI Pipeline: Lint + Test workflow

## Architecture

cmd/server/main.go -> internal/handler/ -> internal/service/ -> internal/repository/ -> pkg/metrics/

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| /health | GET | Health check |
| /metrics | GET | Prometheus metrics |
| /wechat | GET/POST | WeChat webhook |

## Quick Start

```bash
# Configure
cp config.example.yaml config.yaml
cp .env.example .env

# Run locally
make run

# Docker
docker-compose up -d
```

## Metrics

Exposed at /metrics:
- HTTP request count/duration
- Messages received/sent by type
- Events received by type
- Go runtime metrics (goroutines, memory)

## Next Steps

- Add production database layer (PostgreSQL)
- Add Redis caching
- Enhance error handling
- Add unit tests

---

Squashed from 10 commits into single feature commit.
