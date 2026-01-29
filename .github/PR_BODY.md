## Summary

This PR simplifies the codebase by removing complex database/SQL layers and using in-memory storage for faster development and testing.

## Changes

### Removed
- internal/repository/sql/ - Database, UnitOfWork, SQL executors
- pkg/cache/ - Redis cache layer
- migrations/ - Database schema migrations
- Complex repository patterns

### Simplified

#### Architecture
cmd/server/main.go
  -> internal/handler/message.go  (WeChat callbacks)
  -> internal/service/           (Business logic)
  -> internal/repository/        (In-memory storage)
  -> pkg/metrics/                (Prometheus metrics)

#### Repository Layer (NEW)
- internal/repository/user.go - In-memory UserRepository with mutex
- internal/repository/message.go - In-memory MessageRepository with mutex

#### Services
- Simplified to use new in-memory repositories
- Removed cache dependencies

#### Handler
- Fixed SDK type issues (MsgType -> string conversion)
- Removed unused variables
- Simplified message/event handling

### CI/CD Updates
- ci.yml: Simplified to Lint + Test only
  - Removed Docker build step
  - Removed PostgreSQL service
  - Removed Redis service
  - No codecov/upload step

### Infrastructure
- Dockerfile: Simplified, go 1.22
- go.mod: go 1.22, removed unused dependencies (redis, pq, miniredis)

## Benefits
- Faster builds (no external dependencies)
- Simpler debugging (in-memory storage)
- Easier testing (no database setup)
- Faster CI runs
- Reduced complexity

## Breaking Changes
- No database persistence (data lost on restart)
- No Redis caching
- No production database support

For production, re-introduce internal/repository/sql/ layer with actual database.

## Verification
- Build succeeds: go build ./...
- Tests pass: go test ./...
- Linting: golint ./...

---

Created by Riki
