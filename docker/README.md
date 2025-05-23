# Docker Usage Guide - Export Trakt 4 Letterboxd

This guide explains how to use the Docker Compose services with the new `--run` and `--schedule` functionality.

## Overview

The Docker Compose configuration now supports three execution modes:

1. **Immediate Execution (`--run`)**: Execute once and exit
2. **Scheduled Execution (`--schedule`)**: Run on a cron schedule
3. **Legacy Mode**: Traditional command-based approach (for backward compatibility)

## Quick Start

### Test Your Configuration

```bash
# Quick test to verify your configuration works
docker compose --profile run-watched up
```

### Production Scheduler

```bash
# Start a production scheduler (every 6 hours)
docker compose --profile schedule-6h up -d
```

### Check Status

```bash
# View scheduler logs
docker compose --profile schedule-6h logs -f
```

## Immediate Execution Services (`--run`)

These services execute once and then exit. Perfect for:

- Testing configurations
- Manual exports
- CI/CD integration

### Available Services

| Service                       | Profile          | Description                   | Command                                   |
| ----------------------------- | ---------------- | ----------------------------- | ----------------------------------------- |
| `export-trakt-run-watched`    | `run-watched`    | Export watched movies only    | `--run --export watched --mode normal`    |
| `export-trakt-run-all`        | `run-all`        | Export all data (recommended) | `--run --export all --mode complete`      |
| `export-trakt-run-collection` | `run-collection` | Export collection only        | `--run --export collection --mode normal` |
| `export-trakt-run-ratings`    | `run-ratings`    | Export ratings only           | `--run --export ratings --mode complete`  |
| `export-trakt-run-watchlist`  | `run-watchlist`  | Export watchlist only         | `--run --export watchlist --mode normal`  |
| `export-trakt-run-shows`      | `run-shows`      | Export shows only             | `--run --export shows --mode complete`    |

### Usage Examples

```bash
# Export all data immediately
docker compose --profile run-all up

# Export only watched movies
docker compose --profile run-watched up

# Export specific data types
docker compose --profile run-collection up
docker compose --profile run-ratings up
docker compose --profile run-watchlist up
docker compose --profile run-shows up

# Run multiple exports sequentially
docker compose --profile run-watched up && \
docker compose --profile run-ratings up
```

## Scheduled Execution Services (`--schedule`)

These services run continuously according to a cron schedule. Perfect for:

- Production automation
- Regular backups
- Unattended operation

### Available Services

| Service                        | Profile           | Schedule           | Description                  |
| ------------------------------ | ----------------- | ------------------ | ---------------------------- |
| `export-trakt-schedule-6h`     | `schedule-6h`     | Every 6 hours      | Recommended for production   |
| `export-trakt-schedule-daily`  | `schedule-daily`  | Daily at 2:30 AM   | Daily comprehensive export   |
| `export-trakt-schedule-weekly` | `schedule-weekly` | Sundays at 3:00 AM | Weekly backup                |
| `export-trakt-schedule-15min`  | `schedule-15min`  | Every 15 minutes   | High-frequency testing       |
| `export-trakt-schedule-custom` | `schedule-custom` | Configurable       | Custom schedule via env vars |

### Usage Examples

```bash
# Production scheduler (every 6 hours)
docker compose --profile schedule-6h up -d

# Daily backup at 2:30 AM
docker compose --profile schedule-daily up -d

# Weekly comprehensive backup
docker compose --profile schedule-weekly up -d

# High-frequency testing (every 15 minutes)
docker compose --profile schedule-15min up -d

# Custom schedule using environment variables
SCHEDULE="0 */4 * * *" EXPORT_TYPE="watched" EXPORT_MODE="normal" \
docker compose --profile schedule-custom up -d
```

## Custom Configuration

### Environment Variables for Custom Scheduler

The `schedule-custom` profile accepts these environment variables:

| Variable      | Default       | Description              | Example                                  |
| ------------- | ------------- | ------------------------ | ---------------------------------------- |
| `SCHEDULE`    | `0 */6 * * *` | Cron schedule expression | `"30 2 * * *"`                           |
| `EXPORT_TYPE` | `all`         | Type of export           | `watched`, `collection`, `ratings`, etc. |
| `EXPORT_MODE` | `complete`    | Export mode              | `normal`, `initial`, `complete`          |

### Custom Schedule Examples

```bash
# Export watched movies every 4 hours
SCHEDULE="0 */4 * * *" EXPORT_TYPE="watched" EXPORT_MODE="normal" \
docker compose --profile schedule-custom up -d

# Export all data every Monday at 9 AM
SCHEDULE="0 9 * * 1" EXPORT_TYPE="all" EXPORT_MODE="complete" \
docker compose --profile schedule-custom up -d

# Export ratings daily at noon
SCHEDULE="0 12 * * *" EXPORT_TYPE="ratings" EXPORT_MODE="complete" \
docker compose --profile schedule-custom up -d
```

## Legacy Services (Backward Compatibility)

These services use the traditional command-based approach:

| Service                  | Profile                         | Description              |
| ------------------------ | ------------------------------- | ------------------------ |
| `export-trakt`           | `default`, `legacy`             | Normal export (legacy)   |
| `export-trakt-complete`  | `complete`, `legacy`            | Complete export (legacy) |
| `export-trakt-initial`   | `initial`, `legacy`             | Initial export (legacy)  |
| `export-trakt-scheduled` | `scheduled`, `legacy-scheduled` | Legacy cron system       |

## Management Commands

### Start Services

```bash
# Start immediately and view logs
docker compose --profile run-all up

# Start in background (detached)
docker compose --profile schedule-6h up -d
```

### Monitor Services

```bash
# View logs (follow mode)
docker compose --profile schedule-6h logs -f

# View logs for specific time period
docker compose --profile schedule-6h logs --since="2h"

# Check service status
docker compose --profile schedule-6h ps
```

### Stop Services

```bash
# Stop specific service
docker compose --profile schedule-6h down

# Stop all services
docker compose down

# Stop and remove volumes
docker compose down -v
```

### Restart Services

```bash
# Restart scheduler
docker compose --profile schedule-6h restart

# Restart with new configuration
docker compose --profile schedule-6h down
docker compose --profile schedule-6h up -d
```

## Volume Management

The Docker Compose setup uses the following volumes:

| Volume        | Purpose             | Local Path  |
| ------------- | ------------------- | ----------- |
| Configuration | TOML config files   | `./config`  |
| Logs          | Application logs    | `./logs`    |
| Exports       | Generated CSV files | `./exports` |

### Backup Your Data

```bash
# Create backup of configuration and exports
tar -czf trakt-backup-$(date +%Y%m%d).tar.gz config/ exports/ logs/

# Restore from backup
tar -xzf trakt-backup-20240120.tar.gz
```

## Troubleshooting

### Common Issues

1. **Service won't start**

   ```bash
   # Check logs for errors
   docker compose --profile run-watched logs

   # Validate configuration first
   docker compose --profile validate up
   ```

2. **Invalid cron schedule**

   ```bash
   # Test with a simple schedule first
   SCHEDULE="*/5 * * * *" docker compose --profile schedule-custom up
   ```

3. **Permission issues**
   ```bash
   # Fix volume permissions
   sudo chown -R 1000:1000 config/ logs/ exports/
   ```

### Debug Commands

```bash
# Test immediate execution
docker compose --profile run-watched up

# Check if config is valid
docker compose --profile validate up

# View detailed logs
docker compose --profile schedule-6h logs --timestamps

# Connect to running container
docker compose --profile schedule-6h exec export-trakt-schedule-6h sh
```

## Best Practices

### 1. Start with Testing

```bash
# Always test your configuration first
docker compose --profile run-watched up
```

### 2. Use Appropriate Schedules

- **Production**: Every 6-12 hours (`schedule-6h`)
- **Development**: Every 15-30 minutes (`schedule-15min`)
- **Backup**: Weekly (`schedule-weekly`)

### 3. Monitor Resource Usage

```bash
# Check container resource usage
docker stats

# View disk usage
docker system df
```

### 4. Regular Maintenance

```bash
# Clean up old containers and images
docker system prune

# Update to latest image
docker compose pull
docker compose --profile schedule-6h up -d
```

### 5. Backup Configuration

Always backup your `config/` directory before making changes.

## Production Deployment

### Recommended Setup

```bash
# 1. Test configuration
docker compose --profile run-watched up

# 2. Start production scheduler
docker compose --profile schedule-6h up -d

# 3. Set up log rotation (optional)
# Add logrotate configuration for ./logs/*.log

# 4. Monitor with external tools
# Set up monitoring for container health
```

### Health Checks

```bash
# Check if scheduler is running
docker compose --profile schedule-6h ps

# View recent logs
docker compose --profile schedule-6h logs --tail=50

# Check export files are being created
ls -la exports/
```

This Docker setup provides a flexible and robust way to run Export Trakt 4 Letterboxd with the new scheduling capabilities!
