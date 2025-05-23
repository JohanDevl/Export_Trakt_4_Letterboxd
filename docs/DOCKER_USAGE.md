# Docker Usage Guide

This document provides detailed information about using the Export Trakt 4 Letterboxd application with Docker.

## Prerequisites

- Docker installed on your system
- Docker Compose (optional, but recommended)
- For multi-architecture builds: Docker Buildx

## Using Docker Compose (Recommended)

### Quick Start

1. Clone the repository:

   ```bash
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Build and start the container:

   ```bash
   docker compose up
   ```

   This will build and run the container once, which will execute the script and exit.

### Initial Setup

For first-time setup, use the setup profile:

```bash
docker compose --profile setup up
```

This will launch an interactive container to set up your Trakt authentication.

### Running on a Schedule

To run the exporter on a schedule:

```bash
docker compose --profile scheduled up -d
```

This will start the container in the background and execute the export script according to the cron schedule defined in the docker-compose.yml file.

## Docker Compose Profiles

The docker-compose.yml file includes several profiles for different use cases:

- Default (no profile): Run once and exit
- `setup`: Run the initial setup script
- `scheduled`: Run with a cron schedule
- `env-config`: Run with all configuration via environment variables

Example usage:

```bash
# Run the initial setup
docker compose --profile setup up

# Run with a cron schedule
docker compose --profile scheduled up -d

# Run with configuration via environment variables
docker compose --profile env-config up -d
```

## Building Multi-Architecture Images

You can build multi-architecture Docker images using the provided build script:

```bash
# Build and push multi-arch images (amd64, arm64, armv7)
./build-docker.sh --tag v1.0.0

# Build for local platform only
./build-docker.sh --local

# Build but don't push
./build-docker.sh --no-push

# See all options
./build-docker.sh --help
```

## Environment Variables

| Variable        | Description                                             | Default  |
| --------------- | ------------------------------------------------------- | -------- |
| `TZ`            | Timezone                                                | `UTC`    |
| `CRON_SCHEDULE` | Cron schedule expression (empty to run once)            | Empty    |
| `EXPORT_OPTION` | Export option (`normal`, `initial`, `complete`)         | `normal` |
| `API_KEY`       | Trakt API key (from config file unless specified)       | Empty    |
| `API_SECRET`    | Trakt API secret (from config file unless specified)    | Empty    |
| `ACCESS_TOKEN`  | Trakt access token (from config file unless specified)  | Empty    |
| `REFRESH_TOKEN` | Trakt refresh token (from config file unless specified) | Empty    |
| `USERNAME`      | Trakt username (from config file unless specified)      | Empty    |

## Docker Volumes

The Docker container uses the following volumes to persist data:

- `/app/config`: Contains the configuration file
- `/app/logs`: Contains log files
- `/app/copy`: Contains the exported Letterboxd CSV file
- `/app/backup`: Contains Trakt API backup data

## Docker Healthchecks

The Docker container includes built-in health checks that verify:

- Required directories are present and writable
- Required files are present and readable
- Required commands are available
- Trakt API connectivity (if credentials are configured)

You can check the container health status using:

```bash
docker inspect --format "{{.State.Health.Status}}" trakt-export
```

## Using Docker Secrets

For production deployments, you can use Docker secrets to manage sensitive configuration:

```yaml
version: "3.8"

services:
  trakt-export:
    image: johandevl/export-trakt-4-letterboxd:latest
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./copy:/app/copy
      - ./backup:/app/backup
    environment:
      - TZ=Europe/Paris
      - CRON_SCHEDULE=0 3 * * *
      - EXPORT_OPTION=normal
    secrets:
      - trakt_api_key
      - trakt_api_secret
      - trakt_access_token
      - trakt_refresh_token

secrets:
  trakt_api_key:
    file: ./secrets/api_key.txt
  trakt_api_secret:
    file: ./secrets/api_secret.txt
  trakt_access_token:
    file: ./secrets/access_token.txt
  trakt_refresh_token:
    file: ./secrets/refresh_token.txt
```

## Advanced Docker Compose Examples

### Production Deployment with Resource Limits

```yaml
version: "3.8"

services:
  trakt-export:
    image: johandevl/export-trakt-4-letterboxd:latest
    container_name: trakt-export
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./copy:/app/copy
      - ./backup:/app/backup
    environment:
      - TZ=Europe/Paris
      - CRON_SCHEDULE=0 3 * * *
      - EXPORT_OPTION=normal
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/app/docker-entrypoint.sh", "healthcheck"]
      interval: 1m
      timeout: 10s
      retries: 3
    deploy:
      resources:
        limits:
          cpus: "0.5"
          memory: 256M
        reservations:
          cpus: "0.1"
          memory: 128M
```

### Integration with Traefik Reverse Proxy

```yaml
version: "3.8"

services:
  trakt-export:
    image: johandevl/export-trakt-4-letterboxd:latest
    container_name: trakt-export
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./copy:/app/copy
      - ./backup:/app/backup
    environment:
      - TZ=Europe/Paris
      - CRON_SCHEDULE=0 3 * * *
      - EXPORT_OPTION=normal
    restart: unless-stopped
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.trakt.rule=Host(`trakt.example.com`)"
      - "traefik.http.routers.trakt.entrypoints=websecure"
      - "traefik.http.routers.trakt.tls.certresolver=myresolver"
      - "traefik.http.services.trakt.loadbalancer.server.port=8000"
    networks:
      - traefik

networks:
  traefik:
    external: true
```

## Using Without Docker Compose

If you prefer to use Docker directly without Docker Compose:

```bash
docker run -it --name trakt-export \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/copy:/app/copy \
  -v $(pwd)/backup:/app/backup \
  -e TZ=Europe/Paris \
  -e EXPORT_OPTION=complete \
  johandevl/export-trakt-4-letterboxd:latest
```

## Troubleshooting

If you encounter issues with the Docker container:

1. Check the container logs:

   ```bash
   docker logs trakt-export
   ```

2. Check the container health:

   ```bash
   docker inspect --format "{{.State.Health.Status}}" trakt-export
   ```

3. Enter the container to investigate:

   ```bash
   docker exec -it trakt-export bash
   ```

4. If the container is not starting, check the Docker daemon logs:
   ```bash
   docker system events
   ```
