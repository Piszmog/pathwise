# Pathwise

A modern job application tracking web application built with Go, templ, HTMX, and SQLite. Pathwise helps you organize and track your job search with features like status updates, notes, salary tracking, timeline views, and export capabilities.

## Features

- **Application Tracking**: Track where you've applied, position details, application dates, and current status
- **Status Management**: Monitor applications through stages (applied, interviewing, offered, rejected, etc.)
- **Notes & Timeline**: Add notes and view a complete timeline of your application history
- **Salary Tracking**: Record salary ranges and currency for each position
- **Archive System**: Archive old applications to keep your active list focused
- **Export Functionality**: Export your data in various formats
- **Responsive Design**: Works seamlessly on desktop and mobile devices
- **User Authentication**: Secure login system with session management

## Tech Stack

- **Backend**: Go 1.24+ with standard library HTTP server
- **Frontend**: [templ](https://templ.guide) templates with [HTMX](https://htmx.org) for dynamic interactions
- **Database**: SQLite with [sqlc](https://sqlc.dev) for type-safe queries
- **Styling**: Tailwind CSS processed with go-tw
- **Testing**: Go testing with Playwright for E2E tests
- **Development**: Air for hot reloading

## Prerequisites

- **Go+** - [Download Go](https://golang.org/dl/)
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
   go tool templ generate -path ./components
   go tool go-tw -i ./styles/input.css -o ./dist/assets/css/output@dev.css
   go tool sqlc generate
   ```

4. **Run the application**
   ```bash
   # Development with hot reload
   air

   # Or build and run manually
   go build -o ./tmp/main .
   ./tmp/main
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
| `VERSION` | Application version | - |

## Development

### Commands

```bash
# Development server with hot reload
air

# Build application
go build -o ./tmp/main .

# Run tests
go test ./...

# Run E2E tests (requires Playwright)
go test -tags=e2e ./e2e/...

# Lint code
golangci-lint run

# Generate code (templates, CSS, SQL)
go tool templ generate -path ./components
go tool go-tw -i ./styles/input.css -o ./dist/assets/css/output@dev.css
go tool sqlc generate
```

### Project Structure

```
pathwise/
├── components/          # Templ templates (.templ files)
├── db/
│   ├── migrations/      # Database schema migrations
│   └── queries/         # SQL queries for sqlc
├── dist/               # Static assets (CSS, JS, images)
├── e2e/                # End-to-end tests
├── server/
│   ├── handler/        # HTTP request handlers
│   ├── middleware/     # HTTP middleware
│   └── router/         # Route definitions
├── styles/             # Tailwind CSS source files
├── types/              # Domain types and business logic
├── utils/              # Utility functions
└── main.go            # Application entry point
```

### Database

Pathwise uses SQLite with migrations managed by golang-migrate. The database schema includes:

- **Users**: Authentication and user management
- **Job Applications**: Core application data with status tracking
- **Notes**: Timeline notes for applications
- **Status History**: Audit trail of status changes
- **Sessions**: User session management

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
go test -tags=e2e ./e2e/...

# Test specific package
go test ./server/handler -run TestJobHandler
```

## Deployment

### Docker

```bash
# Build image
docker build -t pathwise .

# Run container
docker run -p 8080:8080 pathwise
```

### Manual Deployment

1. Build the application: `go build -o pathwise .`
2. Set environment variables as needed
3. Run the binary: `./pathwise`

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
