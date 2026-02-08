# Build stage
FROM golang:1.24.3-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Install swag and generate Swagger docs
RUN go install github.com/swaggo/swag/cmd/swag@latest && \
    swag init -g cmd/myapp/main.go -o docs

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o vpn-dev ./cmd/myapp

# Runtime stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# Create app user
RUN addgroup -S vpn && adduser -S vpn -G vpn

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/vpn-dev .

# Copy migrations
COPY --chown=vpn:vpn ./migrations ./migrations

# Copy Swagger docs
COPY --from=builder /app/docs ./docs

# Create necessary directories
RUN mkdir -p /app/logs && chown vpn:vpn /app/logs

# Switch to non-root user
USER vpn

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./vpn-dev"]
