FROM alpine:3.18

# Install required packages
RUN apk add --no-cache \
    bash \
    curl \
    jq \
    sed \
    git

# Set working directory
WORKDIR /app

# Copy application files
COPY . /app/

# Create necessary directories
RUN mkdir -p /app/backup /app/logs /app/copy /app/TEMP /app/config

# Make scripts executable
RUN chmod +x /app/Export_Trakt_4_Letterboxd.sh /app/setup_trakt.sh /app/docker-entrypoint.sh
RUN find /app/lib -name "*.sh" -exec chmod +x {} \;
RUN [ -f /app/install.sh ] && chmod +x /app/install.sh || echo "install.sh not found"

# Set proper permissions for volume directories
RUN chmod -R 777 /app/backup /app/logs /app/copy /app/config
RUN chmod -R 755 /app/lib

# Set environment variables
ENV DOSLOG=/app/logs \
    DOSCOPY=/app/copy \
    BACKUP_DIR=/app/backup \
    CONFIG_DIR=/app/config \
    CRON_SCHEDULE="" \
    EXPORT_OPTION="normal"

# Set volume for persistent data
VOLUME ["/app/logs", "/app/copy", "/app/backup", "/app/config"]

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"] 