# AGENTS.md - Development Guide for Pathwise

## Project Overview
Pathwise is a job application tracking web application built with Go, templ, HTMX, and SQLite. It helps users track job applications with features like status updates, notes, salary tracking, and timeline views.

## Build & Test Commands
- **Build**: `go build -o ./tmp/main .`
- **Run dev**: `air` (uses .air.toml config with hot reload on port 8080)
- **Test all**: `go test ./...`
- **Test single**: `go test ./path/to/package -run TestName`
- **E2E tests**: `go test -tags=e2e ./e2e/...` (requires Playwright)
- **Generate code**: `go tool templ generate -path ./components && go tool go-tw -i ./styles/input.css -o ./dist/assets/css/output@dev.css && go tool sqlc generate`

## Environment Variables
- **PORT**: Server port (default: 8080)
- **LOG_LEVEL**: Logging level (debug, info, warn, error)
- **DB_URL**: Database URL (default: ./db.sqlite3 for local SQLite)
- **DB_TOKEN**: Database token (for remote databases like Turso)
- **VERSION**: Application version string

## Code Style & Conventions
- **Imports**: Standard library first, then external packages, then local packages (separated by blank lines)
- **Naming**: Use camelCase for variables/functions, PascalCase for types/constants, snake_case for SQL
- **Types**: Define custom types for domain concepts (e.g., `JobApplicationStatus`), use `sql.Null*` for nullable DB fields
- **Error handling**: Always check errors, use structured logging with slog, return early on errors
- **Structs**: Group related fields, use receiver methods for behavior, embed time fields (CreatedAt, UpdatedAt)
- **Constants**: Group related constants with `const ()` blocks, use typed constants for enums
- **Database**: Use sqlc for queries, enable foreign keys with `PRAGMA foreign_keys = ON`, migrations in db/migrations/
- **SQLC**: Write SQL queries in .sql files, generate type-safe Go code with `go tool sqlc generate`
- **HTTP handlers**: Use templ for templates, set proper content types, use structured logging with slog.Logger
- **Testing**: Use testify for assertions, build tag `//go:build e2e` for E2E tests, use Playwright for browser tests
- **Logging**: Use structured logging with `h.Logger.Error("message", "key", value)` pattern in handlers

## Domain Model
- **JobApplication**: Core entity with Company, Title, URL, Status, AppliedAt, Salary fields
- **JobApplicationStatus**: Enum with values: applied, watching, interviewing, offered, accepted, rejected, declined, withdrawn, canceled, closed
- **JobApplicationNote**: Timeline notes attached to applications
- **JobApplicationStatusHistory**: Timeline of status changes
- **User**: Authentication with email/password, sessions with tokens
- **Timeline Interface**: Both notes and status history implement JobApplicationTimelineEntry

## Templ Guidelines (https://templ.guide)
- **Components**: Define in `components/` package, use PascalCase for component names (e.g., `JobDetails`)
- **Parameters**: Pass Go types as parameters, use struct types for complex data (e.g., `types.JobApplication`)
- **Composition**: Use `@componentName()` syntax to call other components, compose small reusable components
- **Conditionals**: Use Go syntax: `if condition { ... }`, `for _, item := range items { ... }`
- **Attributes**: Use `attr?={ condition }` for conditional attributes, `{ expression }` for dynamic values
- **HTML**: Write standard HTML inside templ blocks, use Tailwind CSS classes for styling
- **Integration**: Call `h.html(ctx, w, status, component)` in handlers to render templ components
- **Generation**: Run `go tool templ generate -path ./components` to compile .templ files to .go files
- **HTMX**: Use hx-* attributes for dynamic behavior, hx-target for DOM updates, hx-swap for replacement strategy
- **Existing Components**: alert, archives, drawer, filter, footer, head, header, input_select, jobs_reload, jobs, main, modal, note, pagination, settings, signin, signup, stats, status_badge, timeline, update_job

## HTMX Guidelines (https://htmx.org/docs/)
- **Core Attributes**: `hx-get`, `hx-post`, `hx-patch`, `hx-delete` for HTTP requests; `hx-target` for DOM targeting; `hx-swap` for replacement strategy
- **Triggers**: Use `hx-trigger` for custom events (e.g., `hx-trigger="load"`, `hx-trigger="click"`), default is click for buttons/submit for forms
- **Trigger Modifiers**: `once`, `changed`, `delay:500ms`, `throttle:1s`, `from:closest form` for advanced trigger behavior
- **Targeting**: `hx-target="#element-id"` to specify where response HTML goes, supports extended CSS selectors (`closest`, `next`, `previous`, `find`)
- **Swapping**: `hx-swap="outerHTML|innerHTML|afterbegin|beforeend|afterend|beforebegin"` for different replacement strategies
- **Extensions**: Use `hx-ext="response-targets"` with `hx-target-error` for error handling, enables different targets for success/error responses
- **Events**: Use `hx-on::after-request` for post-request JavaScript, `hx-on::before-request` for pre-request actions
- **Out-of-Band**: Use `hx-swap-oob="true"` in response HTML to update multiple page sections simultaneously
- **Forms**: HTMX automatically serializes form data, use hidden inputs for additional data, forms submit on change with `hx-trigger="change"`
- **Loading States**: Use CSS transitions with `.htmx-request` class for loading indicators, disable forms during requests with `disabled?={ condition }`
- **Synchronization**: Use `hx-sync="closest form:abort"` to coordinate requests between elements
- **Validation**: Integrates with HTML5 validation API, use `hx-validate="true"` for non-form elements
- **History**: Use `hx-push-url="true"` for browser history integration, `hx-history="false"` to disable caching
- **Integration**: Handlers return templ components via `h.html()`, use URL query parameters for filtering/pagination

## Project Structure
- `components/`: Templ templates (.templ files compiled to .go)
- `db/`: Database migrations (up/down SQL), sqlc queries, connection logic
- `server/`: HTTP handlers, middleware (auth, cache, logging), routing
- `types/`: Domain types and business logic (JobApplication, User, etc.)
- `e2e/`: End-to-end tests with Playwright (requires `//go:build e2e` tag)
- `dist/`: Static assets (CSS, JS, images) served by the application
- `styles/`: Tailwind CSS input files processed by go-tw
- `logger/`: Structured logging setup with slog
- `utils/`: Utility functions (crypto, ID generation, number formatting)
- `version/`: Application version management

## Database Schema
- **users**: id, email, password, created_at, updated_at
- **sessions**: id, user_id, token, expires_at, user_agent, created_at, updated_at
- **job_applications**: id, user_id, company, title, url, status, applied_at, archived, salary_min, salary_max, salary_currency, created_at, updated_at
- **job_application_notes**: id, job_application_id, note, created_at, updated_at
- **job_application_status_history**: id, job_application_id, status, created_at, updated_at
- **user_ips**: id, user_id, ip_address, created_at, updated_at

## SQLC Guidelines (https://docs.sqlc.dev/en/latest/tutorials/getting-started-sqlite.html)
- **Configuration**: sqlc.yml defines engine (sqlite), queries path (db/queries/), schema (db/migrations), and output (db/queries)
- **Query Files**: Write SQL queries in .sql files in db/queries/ directory with special comments for code generation
- **Query Annotations**: Use `-- name: QueryName :one|many|exec` to define query name and return type
  - `:one` - Returns single row (GetUserByID)
  - `:many` - Returns multiple rows (GetJobApplicationsByUserID)  
  - `:exec` - Returns sql.Result for INSERT/UPDATE/DELETE (DeleteUserByID)
- **Parameters**: Use `?` placeholders for parameters, sqlc generates type-safe function signatures
- **Generated Code**: Run `go tool sqlc generate` to create .go files with type-safe query functions
- **Models**: sqlc generates Go structs in models.go that match database schema
- **Database Interface**: Generated Queries struct provides all query methods, accepts any sql.DB-compatible interface
- **Usage Pattern**: `queries := db.New(database)` then call `queries.GetUserByID(ctx, userID)`
- **Transactions**: Pass sql.Tx to queries.WithTx(tx) for transactional operations
- **Example Query Structure**:
  ```sql
  -- name: GetJobApplicationByID :one
  SELECT applied_at, company, title, status, url, id, user_id 
  FROM job_applications 
  WHERE id = ?;
  ```

## Handler Patterns
- **Base Handler**: `Handler` struct with `Logger *slog.Logger` and `Database db.Database`
- **Rendering**: Use `h.html(ctx, w, status, component)` for templ components
- **User ID**: Extract from `USER-ID` header via `getUserID(r *http.Request)`
- **Client IP**: Extract via `getClientIP(r *http.Request)` with X-Forwarded-For support
- **Error Handling**: Log errors with structured logging, return appropriate HTTP status codes
- **Database Queries**: Use sqlc-generated queries via `h.Database.Queries()` for type-safe database operations