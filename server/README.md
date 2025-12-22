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
- `available_short_domains` (list of strings): List of domains used to shorten URLs (default: `["localhost:3000"]`)
- `jwt` (object, optional): JWT authentication configuration
  - `algorithm` (string): JWT signing algorithm - `"HS256"` for symmetric (HMAC) or `"RS256"` for asymmetric (RSA)
  - `secret_key` (string): Secret key for HS256 algorithm (required if using HS256)
  - `public_key` (string): Public key for RS256 algorithm in PEM format (required if using RS256)

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
available_short_domains:
  - localhost:3000
```

**Default SQLite (minimal config):**

```yaml
port: 3000
```

If no config file is provided, the server uses:

- Port: 3000
- Database: SQLite at `db.sqlite`
- Available short domains: `["localhost:3000"]`

**JWT Authentication (optional):**

```yaml
jwt:
  algorithm: HS256
  secret_key: your-secret-key-here
```

Or for RS256:

```yaml
jwt:
  algorithm: RS256
  public_key: |
    -----BEGIN PUBLIC KEY-----
    MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
    -----END PUBLIC KEY-----
```

## JWT Authentication

The server supports optional JWT authentication for all API routes. When JWT configuration is provided, the server will validate Bearer tokens in the `Authorization` header.

### Features

- **Optional Authentication**: Routes work with or without JWT tokens. Invalid or missing tokens are silently ignored.
- **Supported Algorithms**: HS256 (symmetric) and RS256 (asymmetric)
- **User ID Extraction**: The `sub` claim from the JWT token is extracted and used as the User ID
- **Automatic User Association**: When a valid JWT token is provided to the shorten API, the `user_id` field is automatically populated

### Usage

1. **Configure JWT** in your `config.yaml` file (see Configuration section above)

2. **Include Bearer Token** in API requests:

```bash
# Without JWT token (works as before)
curl -X POST http://localhost:3000/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"domain": "localhost:3000", "url": "https://example.com"}'

# With JWT token (user_id will be populated)
curl -X POST http://localhost:3000/api/v1/shorten \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{"domain": "localhost:3000", "url": "https://example.com"}'
```

### JWT Token Requirements

- **Format**: Bearer token in `Authorization` header: `Authorization: Bearer <token>`
- **Required Claim**: The token must contain a `sub` claim (string) representing the User ID
- **Validation**: Only signature validation is performed. Standard claims like `exp`, `iat`, `iss`, and `aud` are not validated.

### Example JWT Token Structure

```json
{
  "sub": "user123",
  "iat": 1234567890,
  "exp": 1234571490
}
```

The `sub` field value (`"user123"`) will be stored in the `user_id` field when creating short URLs via the shorten API.

## Project Structure

```
server/
├── main.go              # Application entry point
├── config/              # Configuration package
│   └── config.go       # Config loading logic
├── handlers/            # HTTP handlers
│   ├── hello.go        # Hello world handler
│   ├── shorten.go      # Shorten URL handler
│   └── redirect.go     # Redirect handler
├── middleware/          # Middleware package
│   └── jwt.go          # JWT authentication middleware
├── models/              # Data models
│   └── short_url.go    # ShortURL model
├── go.mod              # Go module file
└── README.md           # This file
```

## Technologies

- **Gin**: HTTP web framework
- **GORM**: ORM library for Go
- **SQLite**: Default database (`db.sqlite`)
- **PostgreSQL**: Optional database (via `postgres_uri` config)
- **JWT**: JSON Web Token authentication (via `github.com/golang-jwt/jwt/v5`)
