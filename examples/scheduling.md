# Scheduling and Immediate Execution Examples

This document provides practical examples of using the new `--run` and `--schedule` options introduced in Export Trakt 4 Letterboxd.

## Overview

The application now supports three execution modes:

1. **Traditional Mode**: Using commands like `export`, `schedule`, etc.
2. **Immediate Execution Mode**: Using `--run` flag for one-time execution
3. **Scheduled Mode**: Using `--schedule` flag with cron expressions

## Immediate Execution (`--run`)

The `--run` flag executes the export immediately once and then exits. This is useful for:

- One-time exports
- Testing configurations
- Manual exports triggered by external scripts
- CI/CD pipeline integrations

### Examples

```bash
# Export all data immediately with complete mode
./export_trakt --run --export all --mode complete

# Export only watched movies with normal mode
./export_trakt --run --export watched --mode normal

# Export collection with custom config file
./export_trakt --run --export collection --config custom_config.toml

# Export ratings immediately
./export_trakt --run --export ratings --mode complete
```

## Scheduled Execution (`--schedule`)

The `--schedule` flag sets up continuous execution according to a cron schedule. This is perfect for:

- Automated backups
- Regular synchronization
- Unattended operation
- Server deployments

### Cron Format

The schedule uses standard cron format: `minute hour day-of-month month day-of-week`

```
*     *     *     *     *
|     |     |     |     |
|     |     |     |     +-- Day of week (0-7, Sunday=0 or 7)
|     |     |     +------- Month (1-12)
|     |     +------------- Day of month (1-31)
|     +------------------- Hour (0-23)
+------------------------- Minute (0-59)
```

### Schedule Examples

#### Frequent Updates

```bash
# Every 15 minutes (high-frequency monitoring)
./export_trakt --schedule "*/15 * * * *" --export watched --mode normal

# Every hour at minute 0
./export_trakt --schedule "0 * * * *" --export watched --mode normal

# Every 6 hours
./export_trakt --schedule "0 */6 * * *" --export all --mode complete
```

#### Daily Schedules

```bash
# Every day at 2:30 AM
./export_trakt --schedule "30 2 * * *" --export all --mode complete

# Every day at noon
./export_trakt --schedule "0 12 * * *" --export watchlist --mode normal

# Every day at 6:00 PM
./export_trakt --schedule "0 18 * * *" --export ratings --mode complete
```

#### Weekly Schedules

```bash
# Every Monday at 9:00 AM
./export_trakt --schedule "0 9 * * 1" --export all --mode complete

# Every Sunday at 3:00 AM (weekly backup)
./export_trakt --schedule "0 3 * * 0" --export all --mode complete

# Every Friday at 5:00 PM
./export_trakt --schedule "0 17 * * 5" --export collection --mode normal
```

#### Monthly Schedules

```bash
# First day of every month at midnight
./export_trakt --schedule "0 0 1 * *" --export all --mode complete

# 15th of every month at 3:30 AM
./export_trakt --schedule "30 3 15 * *" --export all --mode complete
```

## Use Cases and Scenarios

### Development and Testing

```bash
# Quick test of configuration
./export_trakt --run --export watched --mode normal

# Test all export types
./export_trakt --run --export all --mode complete
```

### Production Automation

```bash
# Daily backup at 2:00 AM
./export_trakt --schedule "0 2 * * *" --export all --mode complete

# Incremental updates every 4 hours
./export_trakt --schedule "0 */4 * * *" --export watched --mode normal
```

### Server Deployment

```bash
# Background scheduler (using nohup)
nohup ./export_trakt --schedule "0 */6 * * *" --export all --mode complete > scheduler.log 2>&1 &

# Systemd service with immediate start
./export_trakt --run --export all --mode complete && \
./export_trakt --schedule "0 4 * * *" --export all --mode complete
```

### Docker Integration

```bash
# Docker run with immediate execution
docker run --rm -v $(pwd)/config:/app/config \
  johandevl/export-trakt-4-letterboxd:latest \
  --run --export all --mode complete

# Docker run with scheduling
docker run -d --name trakt-scheduler \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/exports:/app/exports \
  johandevl/export-trakt-4-letterboxd:latest \
  --schedule "0 */6 * * *" --export all --mode complete
```

## Error Handling and Validation

### Invalid Cron Expressions

The application validates cron expressions and provides helpful error messages:

```bash
# Invalid format
./export_trakt --schedule "invalid" --export watched
# Output: Invalid cron schedule format: invalid
#         Error: expected exactly 5 fields, found 1: [invalid]
#         Example formats:
#           '0 */6 * * *'   - Every 6 hours
#           '0 9 * * 1'     - Every Monday at 9:00 AM
#           '30 14 * * *'   - Every day at 2:30 PM
```

### Configuration Validation

```bash
# Test configuration before scheduling
./export_trakt --run --export watched --mode normal
# If this succeeds, your configuration is valid for scheduling
```

## Monitoring and Logging

### Viewing Scheduler Status

When using `--schedule`, the application provides detailed logging:

```
INFO[2025-01-20T10:00:00Z] scheduler.started schedule="0 */6 * * *" next_run="2025-01-20T16:00:00Z"
INFO[2025-01-20T16:00:00Z] scheduler.executing_export export_type="all" export_mode="complete"
INFO[2025-01-20T16:05:00Z] export.completed_successfully export_type="all" export_mode="complete"
```

### Log Files

Configure logging in your `config.toml`:

```toml
[logging]
level = "info"
file = "logs/scheduler.log"
```

## Best Practices

### 1. Start with Immediate Execution

Test your configuration with `--run` before setting up scheduling:

```bash
./export_trakt --run --export watched --mode normal
```

### 2. Use Appropriate Export Modes

- `normal`: For frequent updates (every few hours)
- `complete`: For comprehensive backups (daily/weekly)

### 3. Consider Resource Usage

- More frequent exports consume more API calls
- Use `watched` type for frequent updates, `all` for comprehensive backups

### 4. Monitor Scheduler Health

- Check logs regularly
- Set up external monitoring if running in production
- Use process managers like systemd or Docker's restart policies

### 5. Backup Configurations

Always keep a backup of your working configuration files before making changes.

## Troubleshooting

### Common Issues

1. **Invalid Configuration**: Test with `--run` first
2. **Wrong Cron Format**: Use online cron validators
3. **Permission Issues**: Ensure write access to export directory
4. **API Rate Limits**: Avoid very frequent schedules (less than 15 minutes)

### Debugging Commands

```bash
# Test immediate execution
./export_trakt --run --export watched --mode normal

# Validate cron expression
./export_trakt --schedule "0 */6 * * *" --export watched --mode normal

# Check configuration
./export_trakt validate
```
