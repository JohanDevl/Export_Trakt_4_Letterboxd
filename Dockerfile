# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Set build arguments
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_DATE=unknown

# Set working directory
WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with version information
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -ldflags "-s -w \
    -X github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/version.Version=${VERSION} \
    -X github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/version.CommitSHA=${COMMIT_SHA} \
    -X github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/version.BuildDate=${BUILD_DATE}" \
    -o export-trakt ./cmd/export_trakt

# Runtime stage
FROM alpine:3.19

# Install CA certificates for HTTPS
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create directories and set permissions
RUN mkdir -p /app/config /app/logs /app/exports /app/web \
    && chown -R appuser:appgroup /app

# Copy binary from builder stage
COPY --from=builder /app/export-trakt /app/export-trakt

# Copy locales
COPY --from=builder /app/locales /app/locales

# Copy web assets (templates, CSS, JS)
COPY --from=builder /app/web /app/web

# Set environment variables
ENV EXPORT_TRAKT_EXPORT_OUTPUT_DIR=/app/exports
ENV EXPORT_TRAKT_LOGGING_FILE=/app/logs/export.log

# Switch to non-root user
USER appuser

# Create volumes for persistent data
VOLUME ["/app/config", "/app/logs", "/app/exports"]

# Set entrypoint
ENTRYPOINT ["/app/export-trakt"]

# Default command if none is provided
CMD ["--help"]

# Metadata
LABEL org.opencontainers.image.title="Export Trakt for Letterboxd"
LABEL org.opencontainers.image.description="Tool to export Trakt.tv data for Letterboxd import"
LABEL org.opencontainers.image.authors="JohanDevl"
LABEL org.opencontainers.image.url="https://github.com/JohanDevl/Export_Trakt_4_Letterboxd"
LABEL org.opencontainers.image.source="https://github.com/JohanDevl/Export_Trakt_4_Letterboxd"
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.revision="${COMMIT_SHA}"
LABEL org.opencontainers.image.licenses="MIT"