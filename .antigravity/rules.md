# Antigravity Rules — UserService (Go Backend)

## Project Overview

This is a **Go 1.21+ REST API backend** (module: `github.com/SawitProRecruitment/UserService`) built for a backend engineering interview assignment.
It uses the **Echo** web framework with **oapi-codegen** for OpenAPI-first code generation. The database is **PostgreSQL** accessed via `database/sql` + `lib/pq`.

### Domain
- **Estate** management: create estates with `length` and `width` (UUID-keyed).
- **Tree** management: plant trees at `(x, y)` coordinates in an estate with a given `height`.
- **Stats** endpoint: return `count`, `min`, `max`, `median` tree heights per estate.
- **Drone plan** endpoint: return the total drone flight `distance` over an estate's tree survey path.

---

## Architecture

```
cmd/main.go          → Entry point; wires Echo, handler.Server, repository.Repository
handler/
  server.go          → Server struct definition (holds RepositoryInterface)
  endpoints.go       → HTTP handler implementations (implements generated.ServerInterface)
repository/
  interfaces.go      → RepositoryInterface (add new DB methods here)
  implementations.go → Actual SQL queries implementing RepositoryInterface
  repository.go      → NewRepository constructor (PostgreSQL via database/sql)
  types.go           → Input/Output structs for repository methods
api.yml              → OpenAPI 3.0 spec (source of truth for all endpoints)
generated/           → Auto-generated from api.yml — NEVER edit manually
database.sql         → PostgreSQL DDL schema (loaded at DB init time)
tests/api_test.go    → End-to-end API test suite against a running server at :8080
Makefile             → Build, generate, test targets
docker-compose.yml   → App + PostgreSQL services
Dockerfile           → Multi-stage Go build
```

---

## Coding Rules

### 1. OpenAPI-First Workflow
- **Always update `api.yml` first** before adding or modifying endpoints.
- Run `make generate` after any change to `api.yml` to regenerate `generated/api.gen.go`.
- Never manually edit anything inside `generated/`.

### 2. Layer Separation
- **Handler layer** (`handler/`): HTTP request parsing, validation, response formatting. No direct SQL.
- **Repository layer** (`repository/`): All SQL queries. Business logic belongs in the handler, not here.
- Always add new method signatures to `repository/interfaces.go` first, then implement in `implementations.go`.

### 3. Mock Generation
- After adding new methods to `repository/interfaces.go`, run `make generate_mocks` to regenerate `repository/interfaces.mock.gen.go`.
- Use mocks in unit tests (`handler/endpoints_test.go`), not real DB connections.

### 4. Request/Response Types
- Input/output types for repository methods live in `repository/types.go`.
- Request/response types for HTTP are auto-generated in `generated/api.gen.go`.

### 5. Error Handling
- Return proper HTTP status codes (`400 BadRequest`, `404 Not Found`, `500 Internal Server Error`).
- All error responses must use the `ErrorResponse` schema from the OpenAPI spec.
- Validate all request parameters. Bad/missing inputs must return `400`.

### 6. UUIDs
- All entity IDs use `github.com/google/uuid`. Generate them server-side on creation.
- Never accept IDs from clients; always assign them in the handler.

### 7. Database Schema (`database.sql`)
- After modifying `database.sql`, rebuild Docker volumes: `docker compose down --volumes && docker compose up --build`.
- Never use `INT(100)` in new tables — this is legacy syntax; prefer just `INT` or `BIGINT`.

### 8. Testing
- **Unit tests** in `handler/endpoints_test.go` use mocks and run with `make test`.
- **Integration/API tests** in `tests/api_test.go` hit a live `http://localhost:8080` server; run with `make test_api`.
- All new endpoints must have at least a happy-path and a bad-request test case in `tests/api_test.go`.

### 9. Environment Variables
- The only required env var is `DATABASE_URL` (PostgreSQL DSN).
- Load via `os.Getenv`. Do not use a `.env` loader in production code (godotenv is only for local dev convenience).

### 10. Docker
- The app listens on port **1323** internally; Docker Compose maps it to **8080** externally.
- API tests connect to `http://localhost:8080`.

---

## Makefile Quick Reference

| Target           | Description                                           |
|------------------|-------------------------------------------------------|
| `make init`      | Clean, generate, tidy, vendor                         |
| `make generate`  | Generate API types + server + mocks from api.yml      |
| `make test`      | Run all unit tests with coverage                      |
| `make test_api`  | Run integration tests (requires running server)       |
| `make build`     | Build binary to `build/main`                          |

---

## DO NOT
- Edit files in `generated/` manually.
- Write raw SQL in handler files.
- Skip adding mock regeneration when adding new repository interface methods.
- Use `panic` outside of critical startup failures (e.g., DB connection failure in `NewRepository`).
