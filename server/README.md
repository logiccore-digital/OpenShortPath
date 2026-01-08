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

## Docker

The server can be run in a Docker container with configuration provided via environment variables. The Dockerfile uses a multi-stage build to create a minimal production image.

### Building the Docker Image

From the project root directory:

```bash
docker build -t openshortpath-server .
```

The Dockerfile follows the same build process as the Makefile `server` target: it builds the dashboard first, copies it to `server/dashboard-dist`, then builds the server.

### Running with Docker

The Docker container automatically generates `config.yaml` from environment variables at startup. All environment variables are optional and will only be included in the generated config if set.

#### Basic Example (SQLite)

```bash
docker run -d \
  -p 3000:3000 \
  -v $(pwd)/data:/app/data \
  -e SQLITE_PATH=/app/data/db.sqlite \
  openshortpath-server
```

#### PostgreSQL Example

```bash
docker run -d \
  -p 3000:3000 \
  -e POSTGRES_URI="postgres://user:password@host:5432/dbname?sslmode=disable" \
  -e PORT=3000 \
  -e AUTH_PROVIDER=local \
  -e JWT_ALGORITHM=HS256 \
  -e JWT_SECRET_KEY="your-secret-key-here" \
  openshortpath-server
```

#### With JWT Authentication (HS256)

```bash
docker run -d \
  -p 3000:3000 \
  -e PORT=3000 \
  -e AUTH_PROVIDER=local \
  -e JWT_ALGORITHM=HS256 \
  -e JWT_SECRET_KEY="your-secret-key-here" \
  -e ADMIN_PASSWORD="your-super-long-admin-password-here" \
  openshortpath-server
```

#### With JWT Authentication (RS256)

For RS256, you need to provide multi-line public/private keys. You can do this by:

1. **Using a file with environment variable substitution:**

```bash
# Create a file with the key
cat > /tmp/jwt_public_key.txt << 'EOF'
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----
EOF

# Pass it as an environment variable
docker run -d \
  -p 3000:3000 \
  -e AUTH_PROVIDER=local \
  -e JWT_ALGORITHM=RS256 \
  -e JWT_PUBLIC_KEY="$(cat /tmp/jwt_public_key.txt)" \
  -e JWT_PRIVATE_KEY="$(cat /tmp/jwt_private_key.txt)" \
  openshortpath-server
```

2. **Using docker-compose with environment file:**

```yaml
version: '3.8'
services:
  server:
    build: .
    ports:
      - "3000:3000"
    environment:
      - PORT=3000
      - AUTH_PROVIDER=local
      - JWT_ALGORITHM=RS256
      - JWT_PUBLIC_KEY=${JWT_PUBLIC_KEY}
      - JWT_PRIVATE_KEY=${JWT_PRIVATE_KEY}
    volumes:
      - ./data:/app/data
```

Then create a `.env` file:

```bash
JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
-----END PUBLIC KEY-----"
JWT_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----"
```

### Environment Variables

The following environment variables map to configuration keys:

| Environment Variable | YAML Key | Type | Required | Notes |
|---------------------|----------|------|----------|-------|
| `PORT` | `port` | int | No | Server port (default: 3000) |
| `POSTGRES_URI` | `postgres_uri` | string | No | PostgreSQL connection URI. If set, uses Postgres instead of SQLite |
| `SQLITE_PATH` | `sqlite_path` | string | No | SQLite database path (default: `db.sqlite`) |
| `AVAILABLE_SHORT_DOMAINS` | `available_short_domains` | list | No | Comma-separated list of domains (e.g., `localhost:3000,example.com`) |
| `AUTH_PROVIDER` | `auth_provider` | string | Yes* | `"local"` or `"external_jwt"` |
| `JWT_ALGORITHM` | `jwt.algorithm` | string | No | `"HS256"` or `"RS256"` |
| `JWT_SECRET_KEY` | `jwt.secret_key` | string | No* | Required if using HS256 |
| `JWT_PUBLIC_KEY` | `jwt.public_key` | string | No* | Required if using RS256 (supports multi-line) |
| `JWT_PRIVATE_KEY` | `jwt.private_key` | string | No* | Required for RS256 with local auth (supports multi-line) |
| `ADMIN_PASSWORD` | `admin_password` | string | No | Admin API password |
| `DASHBOARD_DEV_SERVER_URL` | `dashboard_dev_server_url` | string | No | Development server URL for dashboard |

\* Required based on `AUTH_PROVIDER` and `JWT_ALGORITHM` settings. See the Configuration section for validation rules.

### Multi-line Values

For JWT keys (`JWT_PUBLIC_KEY` and `JWT_PRIVATE_KEY`), you can pass multi-line PEM-formatted keys. The entrypoint script will properly format them in the generated YAML. When using docker-compose or environment files, ensure the newlines are preserved in the environment variable value.

### Docker Compose Example

```yaml
version: '3.8'
services:
  openshortpath-server:
    build: .
    ports:
      - "3000:3000"
    environment:
      - PORT=3000
      - POSTGRES_URI=postgres://user:password@postgres:5432/openshortpath?sslmode=disable
      - AUTH_PROVIDER=local
      - JWT_ALGORITHM=HS256
      - JWT_SECRET_KEY=your-secret-key-here
      - ADMIN_PASSWORD=your-super-long-admin-password-here
      - AVAILABLE_SHORT_DOMAINS=localhost:3000,example.com
    volumes:
      - ./data:/app/data
    depends_on:
      - postgres

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
      - POSTGRES_DB=openshortpath
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
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
