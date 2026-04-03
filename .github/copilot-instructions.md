# GitHub Copilot Instructions — UserService Go Backend

## What This Project Is
A Go REST API backend built for a backend engineering interview. It manages **Estates** (rectangular grids) and **Trees** (positioned within estates), with endpoints for statistics and drone survey path planning.

- **Module**: `github.com/SawitProRecruitment/UserService`
- **Stack**: Go 1.21, Echo v4, PostgreSQL 14, oapi-codegen, golang/mock, testify
- **Ports**: App → `1323` (internal) / `8080` (Docker)

---

## Architecture

```
api.yml              ← OpenAPI 3.0 spec (edit this FIRST)
generated/           ← Auto-generated (DO NOT EDIT)
cmd/main.go          ← Entry point
handler/
  server.go          ← Server struct
  endpoints.go       ← HTTP handler implementations
  endpoints_test.go  ← Unit tests with mocks
repository/
  interfaces.go      ← DB interface (declare methods here first)
  implementations.go ← SQL implementations
  repository.go      ← DB connection setup
  types.go           ← Input/Output structs
database.sql         ← PostgreSQL schema
tests/api_test.go    ← Integration tests
```

---

## Suggestions Copilot Should Prioritize

### Always do these when adding features:
1. Update `api.yml` first (OpenAPI spec is source of truth)
2. Add DB method signature to `repository/interfaces.go`
3. Add Input/Output types to `repository/types.go`
4. Implement SQL in `repository/implementations.go` using `r.Db.QueryRowContext` or `r.Db.QueryContext`
5. Implement HTTP handler in `handler/endpoints.go`
6. Write unit tests using generated mock in `handler/endpoints_test.go`

### Error response pattern (ALWAYS use this shape):
```go
return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
    Message: "description of error",
})
```

### UUID generation (always server-side):
```go
import "github.com/google/uuid"
id := uuid.New().String()
```

### Repository method pattern:
```go
// interface (interfaces.go)
CreateEstate(ctx context.Context, input CreateEstateInput) (CreateEstateOutput, error)

// implementation (implementations.go)
func (r *Repository) CreateEstate(ctx context.Context, input CreateEstateInput) (output CreateEstateOutput, err error) {
    err = r.Db.QueryRowContext(ctx,
        `INSERT INTO estate (id, length, width, created_at, updated_at)
         VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`,
        input.Id, input.Length, input.Width,
    ).Scan(&output.Id)
    return
}
```

### Handler validation pattern:
```go
// Validate bounds for tree insertion
if req.X < 1 || req.X > estate.Width || req.Y < 1 || req.Y > estate.Length {
    return ctx.JSON(http.StatusBadRequest, generated.ErrorResponse{
        Message: "tree coordinates out of estate bounds",
    })
}
```

---

## Domain Rules
- Estate dimensions: `length` = Y-axis, `width` = X-axis
- Valid tree positions: `1 ≤ x ≤ width`, `1 ≤ y ≤ length`
- All lengths/widths/heights must be positive integers
- Stats: `count`, `min`, `max`, `median` of tree heights in an estate
- Drone plan: total Manhattan-style flight distance over all plots

---

## What NOT to suggest
- Editing anything in `generated/` — it's auto-generated
- Writing SQL inside `handler/` files
- Accepting client-provided IDs for new entity creation
- Using `panic` anywhere except `NewRepository` (startup DB failure)
- Skipping `make generate` after api.yml changes
