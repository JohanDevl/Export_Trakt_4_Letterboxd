# Docker Guide for Export Trakt for Letterboxd

This guide explains how to use the Docker image for the Export Trakt for Letterboxd application.

## Quick Start

The simplest way to get started is using Docker Compose:

```bash
# Clone the repository (if you haven't already)
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd

# Create required directories
mkdir -p config logs exports

# Run the setup wizard
docker-compose --profile setup up

# After configuration is complete, run a normal export
docker-compose up
```

## Available Docker Images

The application is available on Docker Hub:

- `johandevl/export-trakt-4-letterboxd:latest` - Latest stable release
- `johandevl/export-trakt-4-letterboxd:2.0.0` - Specific version
- `johandevl/export-trakt-4-letterboxd:dev` - Development version

It's also available on GitHub Container Registry:

- `ghcr.io/johandevl/export-trakt-4-letterboxd:latest`
- `ghcr.io/johandevl/export-trakt-4-letterboxd:2.0.0`
- `ghcr.io/johandevl/export-trakt-4-letterboxd:dev`

## Docker Compose Profiles

The Docker Compose file includes several profiles for different use cases:

| Profile     | Description                 | Command                                    |
| ----------- | --------------------------- | ------------------------------------------ |
| `default`   | Normal export               | `docker-compose up`                        |
| `setup`     | Interactive setup wizard    | `docker-compose --profile setup up`        |
| `complete`  | Complete export (all lists) | `docker-compose --profile complete up`     |
| `initial`   | Initial export              | `docker-compose --profile initial up`      |
| `validate`  | Validate configuration      | `docker-compose --profile validate up`     |
| `scheduled` | Run as a scheduled job      | `docker-compose --profile scheduled up -d` |

## Volume Mounts

The Docker image uses the following volumes:

- `/app/config` - Configuration files
- `/app/logs` - Log files
- `/app/exports` - Exported CSV files

Mount these volumes to persist data between container runs:

```bash
docker run -v ./config:/app/config -v ./logs:/app/logs -v ./exports:/app/exports johandevl/export-trakt-4-letterboxd:latest
```

## Environment Variables

You can customize the container behavior with these environment variables:

| Variable          | Description                           | Default                  |
| ----------------- | ------------------------------------- | ------------------------ |
| `TZ`              | Timezone                              | `UTC`                    |
| `EXPORT_SCHEDULE` | Cron schedule (for scheduled profile) | `0 2 * * *` (2 AM daily) |

## Running Different Export Modes

### Interactive Setup

```bash
docker run -it -v ./config:/app/config johandevl/export-trakt-4-letterboxd:latest setup
```

### Normal Export (default)

```bash
docker run -v ./config:/app/config -v ./logs:/app/logs -v ./exports:/app/exports johandevl/export-trakt-4-letterboxd:latest export --mode normal
```

### Complete Export

```bash
docker run -v ./config:/app/config -v ./logs:/app/logs -v ./exports:/app/exports johandevl/export-trakt-4-letterboxd:latest export --mode complete
```

### Initial Export

```bash
docker run -v ./config:/app/config -v ./logs:/app/logs -v ./exports:/app/exports johandevl/export-trakt-4-letterboxd:latest export --mode initial
```

### Validate Configuration

```bash
docker run -v ./config:/app/config johandevl/export-trakt-4-letterboxd:latest validate
```

## Running as a Scheduled Job

To run the application on a schedule:

```bash
docker run -d --name export-trakt-scheduled \
  -v ./config:/app/config \
  -v ./logs:/app/logs \
  -v ./exports:/app/exports \
  -e EXPORT_SCHEDULE="0 2 * * *" \
  --restart unless-stopped \
  johandevl/export-trakt-4-letterboxd:latest \
  /bin/sh -c "echo \"$EXPORT_SCHEDULE\" > /tmp/crontab && \
  echo \"# Run Export Trakt for Letterboxd\" >> /tmp/crontab && \
  echo \"$EXPORT_SCHEDULE /app/export-trakt export >> /app/logs/cron.log 2>&1\" >> /tmp/crontab && \
  crond -f -d 8 -c /tmp"
```

This runs the export every day at 2 AM.

## Building the Docker Image

If you want to build the image yourself:

```bash
docker build -t export-trakt-4-letterboxd:custom .
```

You can pass build arguments:

```bash
docker build \
  --build-arg VERSION=2.0.0 \
  --build-arg COMMIT_SHA=$(git rev-parse --short HEAD) \
  --build-arg BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ") \
  -t export-trakt-4-letterboxd:custom .
```

## Troubleshooting

### Checking Logs

View container logs:

```bash
docker logs export-trakt
```

Check application logs:

```bash
docker exec export-trakt cat /app/logs/app.log
```

### Accessing the Container

To get a shell inside the container:

```bash
docker exec -it export-trakt /bin/sh
```

### Common Issues

1. **Permission denied errors**: Ensure the mounted volumes have appropriate permissions.
2. **Configuration not found**: Make sure you've run the setup wizard first.
3. **API rate limiting**: If you encounter rate limiting errors, try the `--delay` flag with the export command.
