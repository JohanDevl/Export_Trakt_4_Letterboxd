# Docker Usage Guide

This document provides detailed information about using the Export Trakt 4 Letterboxd application with Docker.

## Prerequisites

- Docker installed on your system
- Docker Compose (optional, but recommended)

## Using Docker Compose (Recommended)

1. Clone the repository:

   ```bash
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Build and start the container:

   ```bash
   docker compose up -d
   ```

3. Enter the container:

   ```bash
   docker compose exec trakt-export bash
   ```

4. Configure Trakt authentication:

   ```bash
   ./setup_trakt.sh
   ```

5. Run the export script:
   ```bash
   ./Export_Trakt_4_Letterboxd.sh [option]
   ```

## Using Docker Directly

1. Build the Docker image:

   ```bash
   docker build -t trakt-export .
   ```

2. Run the container:

   ```bash
   docker run -it --name trakt-export \
     -v $(pwd)/config:/app/config \
     -v $(pwd)/logs:/app/logs \
     -v $(pwd)/copy:/app/copy \
     -v $(pwd)/brain_ops:/app/brain_ops \
     -v $(pwd)/backup:/app/backup \
     trakt-export
   ```

3. Configure Trakt authentication:

   ```bash
   ./setup_trakt.sh
   ```

4. Run the export script:
   ```bash
   ./Export_Trakt_4_Letterboxd.sh [option]
   ```

## Using Pre-built Images

You can pull the pre-built image from GitHub Container Registry:

```bash
# Pull the latest stable version (from main branch)
docker pull ghcr.io/johandevl/export_trakt_4_letterboxd:latest

# Pull the latest development version (from develop branch)
docker pull ghcr.io/johandevl/export_trakt_4_letterboxd:develop

# Or pull a specific version
docker pull ghcr.io/johandevl/export_trakt_4_letterboxd:v1.0.0
```

Example docker-compose.yml using the pre-built image:

```yaml
version: "3"

services:
  trakt-export:
    # For stable production use:
    image: ghcr.io/johandevl/export_trakt_4_letterboxd:latest

    # For testing the latest development version:
    # image: ghcr.io/johandevl/export_trakt_4_letterboxd:develop

    container_name: trakt-export
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./copy:/app/copy
      - ./brain_ops:/app/brain_ops
      - ./backup:/app/backup
    environment:
      - TZ=Europe/Paris
      - CRON_SCHEDULE=0 3 * * * # Run daily at 3:00 AM
      - EXPORT_OPTION=normal # Use the normal export option
```

## Docker Volumes

The Docker container uses the following volumes to persist data:

- `/app/config`: Contains the configuration file
- `/app/logs`: Contains log files
- `/app/copy`: Contains the exported Letterboxd CSV file
- `/app/brain_ops`: Contains additional export data
- `/app/backup`: Contains Trakt API backup data

## Automated Exports with Cron

You can configure the Docker container to automatically run the export script on a schedule using cron. To enable this feature, set the following environment variables:

- `CRON_SCHEDULE`: The cron schedule expression (e.g., `0 3 * * *` for daily at 3:00 AM)
- `EXPORT_OPTION`: The export option to use (`normal`, `initial`, or `complete`)

### Example with Docker Compose:

```yaml
version: "3"

services:
  trakt-export:
    build: .
    container_name: trakt-export
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./copy:/app/copy
      - ./brain_ops:/app/brain_ops
      - ./backup:/app/backup
    environment:
      - TZ=Europe/Paris
      - CRON_SCHEDULE=0 3 * * * # Run daily at 3:00 AM
      - EXPORT_OPTION=normal # Use the normal export option
    stdin_open: true
    tty: true
```

### Example with Docker Run:

```bash
docker run -it --name trakt-export \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/copy:/app/copy \
  -v $(pwd)/brain_ops:/app/brain_ops \
  -v $(pwd)/backup:/app/backup \
  -e CRON_SCHEDULE="0 3 * * *" \
  -e EXPORT_OPTION="normal" \
  trakt-export
```

## Cron Job Logging and Monitoring

The cron job provides comprehensive logging to help you monitor the export process:

1. **Container Logs**:

   - User-friendly messages with emojis appear in the container logs
   - Start and completion notifications with timestamps
   - Progress indicators and success confirmations

   View these logs with:

   ```bash
   docker logs trakt-export
   ```

2. **Detailed Export Logs**:

   - Complete export details are saved to `/app/logs/cron_export.log`
   - Includes API responses, processing steps, and any warnings or errors
   - Timestamped entries for easier troubleshooting

   View these logs with:

   ```bash
   docker exec trakt-export cat /app/logs/cron_export.log
   ```

The cron job is configured to provide clear visual feedback about the export process, making it easy to confirm that your exports are running successfully.

## Docker Implementation Notes

The Docker implementation includes several optimizations:

1. **Modified `sed` commands**: The `sed` commands in the scripts have been adapted to work in Alpine Linux by removing the empty string argument (`''`) which is specific to macOS/BSD versions of `sed`.

2. **Configuration file handling**: The Docker setup uses a dedicated configuration directory (`/app/config`) with proper symlinks to ensure scripts can find and modify the configuration file.

3. **Permissions management**: The Docker entrypoint script ensures all files and directories have the correct permissions for read/write operations.

4. **Path handling**: All scripts have been updated to use absolute paths with the `SCRIPT_DIR` variable to ensure consistent file access regardless of the current working directory.

If you encounter any issues with the Docker implementation, please check the logs and ensure your configuration file is properly set up.
