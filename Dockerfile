# Build stage
FROM alpine:3.19 as builder

# Install build dependencies
RUN apk add --no-cache bash curl jq sed git

# Set working directory
WORKDIR /build

# Copy only necessary build files first
COPY lib/ /build/lib/
COPY Export_Trakt_4_Letterboxd.sh setup_trakt.sh install.sh /build/

# Make scripts executable
RUN chmod +x /build/*.sh
RUN find /build/lib -name "*.sh" -exec chmod +x {} \;

# Final stage
FROM alpine:3.19

# Define build arguments for version
ARG APP_VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

# Add metadata labels using OCI standard
LABEL org.opencontainers.image.version=$APP_VERSION \
      org.opencontainers.image.created=$BUILD_DATE \
      org.opencontainers.image.revision=$VCS_REF \
      org.opencontainers.image.title="Export Trakt 4 Letterboxd" \
      org.opencontainers.image.description="Tool for exporting Trakt.tv history to Letterboxd compatible format" \
      org.opencontainers.image.url="https://github.com/JohanDevl/Export_Trakt_4_Letterboxd" \
      org.opencontainers.image.documentation="https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/blob/main/README.md" \
      maintainer="JohanDevl"

# Install required runtime packages (minimal set)
RUN apk add --no-cache bash curl jq sed ca-certificates tzdata \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy application files from builder
COPY --from=builder /build /app/
COPY docker-entrypoint.sh /app/

# Create necessary directories with proper permissions
RUN mkdir -p /app/backup /app/logs /app/copy /app/TEMP /app/config \
    && chmod +x /app/*.sh \
    && chmod -R 755 /app/lib \
    && chown -R appuser:appgroup /app/backup /app/logs /app/copy /app/TEMP /app/config

# Set environment variables
ENV DOSLOG=/app/logs \
    DOSCOPY=/app/copy \
    BACKUP_DIR=/app/backup \
    CONFIG_DIR=/app/config \
    CRON_SCHEDULE="" \
    EXPORT_OPTION="normal" \
    APP_VERSION=$APP_VERSION \
    TZ=UTC

# Set volume for persistent data
VOLUME ["/app/logs", "/app/copy", "/app/backup", "/app/config"]

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=1m --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"] 