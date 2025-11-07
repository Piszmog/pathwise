# Pathwise

A modern job application tracking system with both web interface and programmatic access. Built with Go, templ, HTMX, and SQLite, Pathwise helps you organize and track your job search with features like status updates, notes, salary tracking, timeline views, and export capabilities.

## Components

- **Web Application**: Interactive web interface for managing job applications
- **MCP Server**: Model Context Protocol server providing programmatic access to your job data
- **Jobs Processor**: Background service that scrapes job postings from Hacker News and processes them with AI/LLM integration

> **Note**: This follows standard Go project layout with applications in `cmd/` and private code in `internal/`. Commands should be run from the project root unless otherwise specified.

## Features

- **Application Tracking**: Track where you've applied, position details, application dates, and current status
- **Status Management**: Monitor applications through stages (applied, interviewing, offered, rejected, etc.)
- **Notes & Timeline**: Add notes and view a complete timeline of your application history
- **Salary Tracking**: Record salary ranges and currency for each position
- **Archive System**: Archive old applications to keep your active list focused
- **Export Functionality**: Export your data in various formats
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **User Authentication**: Secure login system with session management
- **HN Job Scraping**: Automated scraping of job postings from Hacker News with AI-powered processing
- **MCP Integration**: Programmatic access via Model Context Protocol for AI assistants and automation

## MCP Server

The MCP (Model Context Protocol) server provides programmatic access to your job application data for AI assistants and automation tools.

### Authentication
Generate an API key through the web application settings, then configure your MCP client with the key and server URL.

## Tech Stack

- **Backend**: Go 1.25.0+ with standard library HTTP server
- **Frontend**: [templ](https://templ.guide) templates with [HTMX](https://htmx.org) for dynamic interactions
- **Database**: SQLite with [sqlc](https://sqlc.dev) for type-safe queries
- **Styling**: Tailwind CSS processed with go-tw
- **Testing**: Go testing with Playwright for E2E tests
- **Development**: Air for hot reloading

## Prerequisites

- **Go 1.25.0+** - [Download Go](https://golang.org/dl/)
- **Air** (optional, for development) - `go install github.com/air-verse/air@latest`
- **golangci-lint** (optional, for linting) - [Installation guide](https://golangci-lint.run/welcome/install/)

## Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/Piszmog/pathwise.git
   cd pathwise
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Generate code and build assets**
   ```bash
   go tool templ generate -path ./internal/ui/components
   go tool go-tw -i ./internal/ui/styles/input.css -o ./internal/ui/dist/assets/css/output@dev.css
   go tool sqlc generate
   ```

4. **Run the application**
   ```bash
   # Web application with hot reload
   cd cmd/ui && air

   # MCP server (optional, for programmatic access)
   cd cmd/mcp && air

   # Jobs processor (optional, for HN scraping - requires GEMINI_API_KEY)
   export GEMINI_API_KEY=your_api_key_here
   cd cmd/jobs && air
   ```

5. **Open your browser**
   Navigate to `http://localhost:8080`

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `LOG_OUTPUT` | Log output (stdout or file path) | `stdout` |
| `DB_URL` | Database URL | `./db.sqlite3` |
| `DB_TOKEN` | Database token (for remote databases) | - |
| `DB_PRIMARY_URL` | Primary database URL (used by jobs and mcp for write operations) | - |
| `DB_TOKEN_READONLY` | Read-only database token (used by mcp) | - |
| `ENC_KEY` | Encryption key for sensitive data | - |
| `URL_SEARCH` | Search service URL (used by ui) | - |
| `GEMINI_API_KEY` | Google Gemini API key (required for jobs processor) | - |
| `VERSION` | Application version (used by ui and mcp) | - |

## Development

### Commands

```bash
# Development servers with hot reload
cd cmd/ui && air                   # Web application
cd cmd/mcp && air                  # MCP server
cd cmd/jobs && air                 # Jobs processor (requires GEMINI_API_KEY)

# Build applications
go build -o ./tmp/main ./cmd/ui    # Web application
go build -o ./tmp/mcp ./cmd/mcp    # MCP server
go build -o ./tmp/jobs ./cmd/jobs  # Jobs processor

# Run tests
go test ./...                                             # All tests
go test -tags=e2e ./internal/ui/e2e/...                  # E2E tests (requires Playwright)
go test ./internal/mcp/tool/... -tags=integration        # MCP integration tests

# Lint code
golangci-lint run

# Generate code (templates, CSS, SQL)
go tool templ generate -path ./internal/ui/components
go tool go-tw -i ./internal/ui/styles/input.css -o ./internal/ui/dist/assets/css/output@dev.css
go tool sqlc generate
```

### Project Structure

```
pathwise/
├── cmd/               # Application entry points
│   ├── ui/            # Web application
│   │   ├── main.go    # Web application entry point
│   │   ├── .air.toml  # Hot reload configuration
│   │   └── Dockerfile # Container configuration
│   ├── mcp/           # MCP server
│   │   ├── main.go    # MCP server entry point
│   │   └── Dockerfile # Container configuration
│   └── jobs/          # Jobs processor
│       ├── main.go    # Jobs processor entry point
│       ├── .air.toml  # Hot reload configuration
│       └── Dockerfile # Container configuration
├── internal/          # Private application code
│   ├── db/
│   │   ├── migrations/# Database schema migrations
│   │   └── queries/   # SQL queries for sqlc
│   ├── jobs/          # Jobs processor implementation
│   │   ├── hn/        # Hacker News scraper
│   │   └── llm/       # LLM client (Gemini integration)
│   ├── mcp/           # MCP server implementation
│   │   ├── server/    # MCP server setup and middleware
│   │   │   └── middleware/# Authentication middleware
│   │   └── tool/      # MCP tool implementations
│   ├── search/        # Search client implementation
│   ├── server/        # Shared server utilities
│   │   ├── health/    # Health check handlers
│   │   ├── middleware/# Common HTTP middleware
│   │   └── mux/       # HTTP router setup
│   ├── ui/            # Frontend code and assets
│   │   ├── components/# Templ templates (.templ files)
│   │   ├── dist/      # Static assets (CSS, JS, images)
│   │   ├── e2e/       # End-to-end tests
│   │   ├── server/
│   │   │   ├── handler/   # HTTP request handlers
│   │   │   ├── middleware/# HTTP middleware
│   │   │   └── router/    # Route definitions
│   │   ├── styles/    # Tailwind CSS source files
│   │   ├── types/     # Domain types and business logic
│   │   └── utils/     # Utility functions
│   ├── logger/        # Structured logging setup
│   ├── context_key/   # Context key definitions
│   ├── testutil/      # Test utilities
│   └── version/       # Application version management
├── .github/           # GitHub workflows
├── go.mod             # Go module with tools
├── README.md
└── LICENSE
```

### Code Generation

The project uses several code generation tools:

- **templ**: Compiles `.templ` files to Go code for type-safe HTML templates
- **sqlc**: Generates type-safe Go code from SQL queries
- **go-tw**: Processes Tailwind CSS for styling

## Testing

```bash
# Unit tests
go test ./...

# E2E tests (requires Playwright setup)
go test -tags=e2e ./internal/ui/e2e/...

# MCP integration tests
go test ./internal/mcp/tool/... -tags=integration

# Test specific package
go test ./internal/ui/server/handler -run TestJobHandler
```

## Deployment

### Docker

```bash
# Build application images
docker build -f cmd/ui/Dockerfile -t pathwise-ui .
docker build -f cmd/mcp/Dockerfile -t pathwise-mcp .
docker build -f cmd/jobs/Dockerfile -t pathwise-jobs .

# Run containers
docker run -p 8080:8080 pathwise-ui                              # Web application
docker run -p 8081:8080 pathwise-mcp                             # MCP server
docker run -e GEMINI_API_KEY=your_key pathwise-jobs              # Jobs processor
```

### Manual Deployment

1. Build the applications:
   ```bash
   go build -o pathwise-ui ./cmd/ui      # Web application
   go build -o pathwise-mcp ./cmd/mcp    # MCP server
   go build -o pathwise-jobs ./cmd/jobs  # Jobs processor
   ```
2. Set environment variables as needed
3. Run the binaries:
   ```bash
   ./pathwise-ui    # Web application on port 8080
   ./pathwise-mcp   # MCP server on port 8080 (or different port)
   ./pathwise-jobs  # Jobs processor (requires GEMINI_API_KEY)
   ```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Run tests and linting: `go test ./... && golangci-lint run`
5. Commit your changes: `git commit -am 'Add feature'`
6. Push to the branch: `git push origin feature-name`
7. Submit a pull request

## Related Projects

- [Desktop version](https://github.com/Piszmog/job-app-tracker) - A desktop application with similar functionality

## License

See [LICENSE](./LICENSE) for details.
