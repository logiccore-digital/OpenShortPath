# Build stage
FROM golang:1.23-alpine AS builder

# Install Node.js and npm for dashboard build
RUN apk add --no-cache nodejs npm

WORKDIR /build

# Copy dashboard files
COPY dashboard/ ./dashboard/

# Build dashboard
WORKDIR /build/dashboard
RUN npm install && npm run build

# Copy server go mod files first for dependency caching
WORKDIR /build
COPY server/go.mod server/go.sum ./server/

# Download Go dependencies
WORKDIR /build/server
RUN go mod download

# Copy all server source code
WORKDIR /build
COPY server/ ./server/

# Copy dashboard build to server directory (matching Makefile behavior)
RUN rm -rf server/dashboard-dist && \
    mkdir -p server/dashboard-dist && \
    cp -r dashboard/dist/* server/dashboard-dist/ 2>/dev/null || true && \
    echo "placeholder" > server/dashboard-dist/placeholder.txt

# Build the server application
WORKDIR /build/server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/server/server .

# Copy entrypoint script from builder
COPY --from=builder /build/server/docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Expose default port
EXPOSE 3000

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"]

