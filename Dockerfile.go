FROM alpine:3.18

LABEL maintainer="Johan Devlaminck <info@johandevlaminck.com>"
LABEL org.opencontainers.image.source=https://github.com/JohanDevl/Export_Trakt_4_Letterboxd
LABEL org.opencontainers.image.description="Export your Trakt.tv history to Letterboxd format"
LABEL org.opencontainers.image.licenses=MIT

# Create necessary directories
RUN mkdir -p /app/config /app/logs /app/temp_locales /app/exports

# Set working directory
WORKDIR /app

# Copy the executable and required files
COPY build/export_trakt /app/
COPY temp_locales/ /app/temp_locales/
COPY config/config.example.toml /app/config/config.toml

# Make the binary executable
RUN chmod +x /app/export_trakt

# Set environment variables
ENV CONFIG_PATH=/app/config/config.toml
ENV GO_ENV=production

# Volumes
VOLUME ["/app/config", "/app/logs", "/app/exports"]

# Set the entrypoint
ENTRYPOINT ["/app/export_trakt", "--config", "/app/config/config.toml"] 