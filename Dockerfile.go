FROM alpine:3.18

LABEL maintainer="Johan Devlaminck <info@johandevlaminck.com>"
LABEL org.opencontainers.image.source=https://github.com/JohanDevl/Export_Trakt_4_Letterboxd
LABEL org.opencontainers.image.description="Export your Trakt.tv history to Letterboxd format"
LABEL org.opencontainers.image.licenses=MIT

# Create app directories
RUN mkdir -p /app/config /app/logs /app/locales /app/exports

# Set working directory
WORKDIR /app

# Copy pre-built binary
COPY build/export_trakt /app/

# Copy translation files
COPY locales/ /app/locales/

# Default config file
COPY config/config.toml /app/config/

# Ensure the binary is executable
RUN chmod +x /app/export_trakt

# Set environment variables
ENV CONFIG_PATH=/app/config/config.toml
ENV GO_ENV=production

# Volumes
VOLUME ["/app/config", "/app/logs", "/app/exports"]

# Run the application
ENTRYPOINT ["/app/export_trakt", "--config", "/app/config/config.toml"] 