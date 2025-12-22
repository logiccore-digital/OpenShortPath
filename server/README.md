# OpenShortPath Server

A Go API server built with Gin and GORM.

## Prerequisites

- Go 1.21 or higher

## Getting Started

1. Install dependencies:

```bash
go mod download
```

2. Run the server:

```bash
# Without config file (uses defaults)
go run main.go

# With config file
go run main.go --config config.yaml
```

The server will start on port 3000 by default. You can change this via the config file or by setting the `PORT` environment variable.

3. Test the hello world endpoint:

```bash
curl http://localhost:3000/
```

## Configuration

The server supports an optional YAML configuration file specified with the `--config` flag.

### Configuration Options

- `port` (int): Server port (default: 3000)
- `postgres_uri` (string): PostgreSQL connection URI. If provided, the server will use Postgres instead of SQLite.
- `sqlite_path` (string): Path to SQLite database file (default: `db.sqlite`)

### Example Config Files

**SQLite with custom path:**

```yaml
port: 3000
sqlite_path: data/custom.db
```

**PostgreSQL:**

```yaml
port: 8080
postgres_uri: postgres://user:password@localhost:5432/dbname?sslmode=disable
```

**Default SQLite (minimal config):**

```yaml
port: 3000
```

If no config file is provided, the server uses:

- Port: 3000
- Database: SQLite at `db.sqlite`

## Project Structure

```
server/
├── main.go              # Application entry point
├── config/              # Configuration package
│   └── config.go       # Config loading logic
├── handlers/            # HTTP handlers
│   └── hello.go        # Hello world handler
├── go.mod              # Go module file
└── README.md           # This file
```

## Technologies

- **Gin**: HTTP web framework
- **GORM**: ORM library for Go
- **SQLite**: Default database (`db.sqlite`)
- **PostgreSQL**: Optional database (via `postgres_uri` config)
