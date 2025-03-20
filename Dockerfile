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
RUN mkdir -p /app/backup /app/logs /app/copy /app/brain_ops /app/TEMP /app/config

# Make scripts executable
RUN chmod +x /app/Export_Trakt_4_Letterboxd.sh /app/setup_trakt.sh /app/docker-entrypoint.sh

# Set proper permissions for volume directories
RUN chmod -R 777 /app/backup /app/logs /app/copy /app/brain_ops /app/config

# Set environment variables
ENV DOSLOG=/app/logs \
    DOSCOPY=/app/copy \
    BRAIN_OPS=/app/brain_ops \
    BACKUP_DIR=/app/backup \
    CRON_SCHEDULE="" \
    EXPORT_OPTION="complete"

# Set volume for persistent data
VOLUME ["/app/logs", "/app/copy", "/app/brain_ops", "/app/backup", "/app/config"]

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"] 