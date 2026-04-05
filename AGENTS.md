# AGENTS.md — UserService Go Backend

> This file is read by AI coding agents (OpenAI Codex, Devin, etc.) to understand
> project conventions before taking any actions.

## Project Identity
- **Name**: UserService — Go REST API Backend
- **Module**: `github.com/SawitProRecruitment/UserService`
- **Language**: Go 1.21+
- **Framework**: Echo v4 (HTTP), oapi-codegen (OpenAPI → Go), golang/mock, testify
- **Database**: PostgreSQL 14 via `database/sql` + `lib/pq`
- **Container**: Docker + Docker Compose

---

## Repository Layout

```
api.yml                       ← OpenAPI 3.0 spec (edit first — source of truth)
generated/                    ← Auto-generated from api.yml (READ ONLY — never edit)
cmd/main.go                   ← Entry: wires Echo, handler, repository
handler/
  server.go                   ← Server struct with RepositoryInterface field
  endpoints.go                ← HTTP handlers (generated.ServerInterface implementation)
  endpoints_test.go           ← Unit tests using mocks
repository/
  interfaces.go               ← RepositoryInterface (add new methods here first)
  implementations.go          ← SQL implementations of the interface
  repository.go               ← DB connection (NewRepository)
  types.go                    ← Input/Output structs for repo methods
  interfaces.mock.gen.go      ← Auto-generated mocks (READ ONLY — never edit)
database.sql                  ← PostgreSQL DDL schema
tests/api_test.go             ← Integration tests (live server at :8080)
Makefile                      ← Build/generate/test targets
Dockerfile                    ← Multi-stage Go build
docker-compose.yml            ← App (:8080→:1323) + PostgreSQL (:5432)
.env                          ← Local DATABASE_URL (not committed)
```

---

## Domain Model

### Tables
```sql
estate (id UUID PK, length INT, width INT, created_at, updated_at)
tree   (id UUID PK, estate_id UUID FK→estate, x INT, y INT, height INT, created_at, updated_at)
```

### Business Rules
- Estate is a rectangular grid: columns `1..width` (x-axis), rows `1..length` (y-axis)
- Tree coordinates must satisfy: `1 ≤ x ≤ width` AND `1 ≤ y ≤ length`
- All numeric fields (length, width, height, x, y) must be positive integers
- All entity IDs are UUID, generated server-side

### Expected API Surface
| Method | Route | Description |
|--------|-------|-------------|
| POST | `/estate` | Create estate, returns `{id}` |
| POST | `/estate/{id}/tree` | Add tree to estate, returns `{id}` |
| GET | `/estate/{id}/stats` | Returns `{count, min, max, median}` of tree heights |
| GET | `/estate/{id}/drone-plan` | Returns `{distance}` for full drone survey |

---

## Mandatory Agent Workflow for Adding Features

Agents MUST follow this sequence when adding any new feature:

```
1. Edit api.yml            → define endpoint, request/response schemas
2. make generate           → regenerate generated/api.gen.go
3. repository/interfaces.go → add method signature to RepositoryInterface
4. repository/types.go     → add Input/Output struct pair
5. repository/implementations.go → write SQL implementation
6. make generate_mocks     → regenerate interfaces.mock.gen.go
7. handler/endpoints.go    → implement handler method
8. handler/endpoints_test.go → add unit test (using mock)
9. tests/api_test.go       → add integration test case
```

---

## Code Rules

### Forbidden
- ❌ Edit any file inside `generated/`
- ❌ Edit `repository/interfaces.mock.gen.go`
- ❌ Write SQL inside `handler/` package
- ❌ Use `panic` except in `NewRepository` startup
- ❌ Accept entity UUIDs from client request bodies
- ❌ Skip `make generate` after changing `api.yml`

### Required Patterns

**UUID generation**
```go
import "github.com/google/uuid"
id := uuid.New()  // always server-side
```

**Error response**
```go
return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
    Message: "human-readable reason",
})
```

**Repository method signature**
```go
// In interfaces.go
MethodName(ctx context.Context, input MethodNameInput) (MethodNameOutput, error)
```

**Repository SQL implementation**
```go
func (r *Repository) MethodName(ctx context.Context, input MethodNameInput) (output MethodNameOutput, err error) {
    err = r.Db.QueryRowContext(ctx, `SELECT ... FROM ...`, input.Field).Scan(&output.Field)
    return
}
```

**Input validation in handlers**
```go
if err := ctx.Bind(&req); err != nil {
    return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{Message: "invalid request body"})
}
// validate all fields before calling repository
```

---

## Environment & Running

```bash
# Setup (first time)
make init

# Code generation (after api.yml or interfaces.go changes)
make generate

# Run all unit tests
make test

# Run integration tests (requires docker compose up first)
make test_api

# Start full environment
docker compose up --build

# Reset database schema
docker compose down --volumes && docker compose up --build
```

**Environment variables** (set in `.env` or Docker Compose):
```
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=disable
```

---

## Testing Notes
- Unit tests run against mocks — no real DB needed: `make test`
- Integration tests hit live API at `http://localhost:8080`: `make test_api`
- Coverage report: `coverage.out` (generated by `make test`)
- The test suite validates: estate creation, tree planting (valid + OOB), stats accuracy, drone distance
