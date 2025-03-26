# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o export_trakt ./cmd/export_trakt

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/export_trakt .

# Copy configuration
COPY config/config.toml /app/config/

# Create necessary directories
RUN mkdir -p /app/exports /app/logs \
    && chown -R appuser:appgroup /app

# Set environment variables
ENV CONFIG_PATH=/app/config/config.toml \
    TZ=UTC

# Switch to non-root user
USER appuser

# Set volume for persistent data
VOLUME ["/app/exports", "/app/logs"]

# Set entrypoint
ENTRYPOINT ["/app/export_trakt"]