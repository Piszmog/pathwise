# AGENTS.md - Development Guide for Pathwise

## Build & Test Commands
- **Test all**: `go test ./...`
- **Test single**: `go test ./path/to/package -run TestName`
- **Test with race detection**: `go test -race ./...`
- **Lint**: `golangci-lint run` (uses all linters except disabled ones in .golangci.yml)
- **Run dev UI**: `cd cmd/ui && air` (hot reload on port 8080)
- **Run dev Jobs**: `cd cmd/jobs && air` (background HN scraper)
- **Generate**: `go tool templ generate -path ./ui/components && go tool sqlc generate`
- **E2E tests**: `go test -tags=e2e ./ui/e2e/...`
- **Integration tests**: `go test ./mcp/tool/... -tags=integration`

## Code Style
- **NO COMMENTS**: Never add comments to code. Code must be self-documenting.
- **Imports**: Standard lib, external packages, local packages (blank line separated)
- **Naming**: camelCase variables/functions, PascalCase types, snake_case SQL
- **Error handling**: Always check errors, use structured logging with slog
- **Database**: Use sqlc queries, migrations in internal/db/migrations/
- **Templates**: Use templ components in ui/components/, call with `h.html(ctx, w, status, component)`
- **HTMX**: Use hx-* attributes, hx-target for DOM updates, hx-swap for strategy
- **Testing**: Use testify assertions, `//go:build e2e` tag for E2E tests
- **Linting**: Project uses all golangci-lint rules except disabled ones (no comments, no globals, etc.)

## Database (SQLite + sqlc)
- **Schema**: Migrations in internal/db/migrations/ (up/down SQL files)
- **Queries**: Write SQL in internal/db/queries/*.sql with `-- name: QueryName :one|many|exec`
- **Generate**: `go tool sqlc generate` creates type-safe Go functions
- **Usage**: `h.Database.Queries().QueryName(ctx, params)` in handlers
- **Transactions**: Use `h.Database.Queries().WithTx(tx)` for atomic operations

## Testing
- **Unit**: Standard Go tests with testify assertions
- **E2E**: Playwright tests in ui/e2e/, use `//go:build e2e` tag  
- **Integration**: MCP tests with `//go:build integration` tag
- **Run E2E**: Requires Playwright setup, tests web interface end-to-end

## Templ & HTMX
- **Templ**: Components in ui/components/*.templ, use `@componentName()` syntax, generate with `go tool templ generate`
- **Render**: Call components with `h.html(ctx, w, status, componentName(params))` in handlers
- **HTMX**: Use `hx-get/post/patch/delete` for requests, `hx-target="#id"` for DOM targeting
- **Swapping**: Use `hx-swap="innerHTML|outerHTML"` for replacement strategy
- **Forms**: Auto-serialize with HTMX, use `hx-trigger="change"` for live updates

## Architecture
- Go project layout: applications in `cmd/`, private code in `internal/`
- Web app: Go + templ + HTMX + SQLite in `cmd/ui/`
- MCP server: Model Context Protocol server in `cmd/mcp/`
- Jobs processor: Background HN scraper with LLM processing in `cmd/jobs/` (requires GEMINI_API_KEY)