# Windsurf Rules — UserService Go Backend

## Project Summary
Go REST API: Estate + Tree management with stats and drone-plan endpoints.
- **Module**: `github.com/SawitProRecruitment/UserService`
- **Stack**: Go 1.21 · Echo v4 · PostgreSQL 14 · oapi-codegen · mockgen · testify
- **App listens on**: `:1323` (Docker maps to `:8080`)

---

## Project Structure

```
api.yml                     # OpenAPI 3.0 spec — SOURCE OF TRUTH
generated/                  # Auto-generated from api.yml — READ ONLY
cmd/main.go                 # Entry point
handler/
  server.go                 # Server struct (holds RepositoryInterface)
  endpoints.go              # HTTP handlers — implements generated.ServerInterface
  endpoints_test.go         # Unit tests (mock repository)
repository/
  interfaces.go             # RepositoryInterface — add new methods here first
  implementations.go        # SQL queries implementing the interface
  repository.go             # PostgreSQL connection via database/sql
  types.go                  # Input/Output types for all repository methods
  interfaces.mock.gen.go    # Auto-generated mocks — READ ONLY
database.sql                # DDL schema (PostgreSQL)
tests/api_test.go           # End-to-end tests against live server at :8080
Makefile                    # Build/generate/test targets
docker-compose.yml          # App + PostgreSQL services
Dockerfile                  # Multi-stage Go build
```

---

## Workflow: Adding a New Endpoint

> Follow this order strictly:

1. **`api.yml`** — Define path, parameters, request body, and response schemas
2. **`make generate`** — Regenerate `generated/api.gen.go`
3. **`repository/interfaces.go`** — Declare the new DB method on `RepositoryInterface`
4. **`repository/types.go`** — Define `*Input` and `*Output` structs
5. **`repository/implementations.go`** — Implement the SQL logic
6. **`make generate_mocks`** — Regenerate mock file
7. **`handler/endpoints.go`** — Implement the HTTP handler
8. **`handler/endpoints_test.go`** — Unit test with mock
9. **`tests/api_test.go`** — Add integration test steps

---

## Coding Standards

### HTTP Handlers
- Bind request body with `ctx.Bind(&req)`
- Return `400 BadRequest` for invalid/missing params
- Return `404 NotFound` when entity doesn't exist, `500` for unexpected DB errors
- Always use `generated.ErrorResponse{Message: "..."}` for error responses
- Generate UUIDs server-side: `uuid.New()` (never accept ID from client)

### Repository
- All methods accept `context.Context` as first arg
- Use named return values for clean error handling
- Use `$1, $2, ...` placeholders (PostgreSQL-style)
- Only `database/sql` queries — no ORM

### Testing
- Unit tests: `make test` — uses mocks, no real DB
- Integration tests: `make test_api` — requires `docker compose up --build`
- New features need both unit and integration test coverage

---

## Key Commands

```bash
make init                              # Setup (first run)
make generate                          # Regen from api.yml and interfaces.go
make generate_mocks                    # Regen only mocks
make test                              # Unit tests + coverage
make test_api                          # Integration tests
docker compose up --build              # Launch app + DB
docker compose down --volumes          # Wipe & reset DB
```

---

## Hard Rules
| Rule | Enforcement |
|------|------------|
| Never edit `generated/` | Auto-generated — changes will be overwritten |
| Never edit `interfaces.mock.gen.go` | Auto-generated |
| No SQL in `handler/` | Strict layer boundary |
| No panic outside startup | Only in `NewRepository` on DB connect failure |
| IDs are server-generated | Always use `uuid.New()`, never trust client IDs |
| OpenAPI first | `api.yml` changes before implementation |
