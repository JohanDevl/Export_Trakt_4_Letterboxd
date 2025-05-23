# Docker Compose Changes - New Scheduling Features

This document summarizes the changes made to the Docker Compose configuration to support the new `--run` and `--schedule` functionality.

## Overview of Changes

The Docker Compose configuration has been significantly enhanced to support three execution modes:

1. **Immediate Execution (`--run`)**: Execute once and exit
2. **Scheduled Execution (`--schedule`)**: Run on a cron schedule
3. **Legacy Mode**: Traditional command-based approach (for backward compatibility)

## New Services Added

### Immediate Execution Services (`--run`)

| Service Name                  | Profile          | Command                                   | Purpose                       |
| ----------------------------- | ---------------- | ----------------------------------------- | ----------------------------- |
| `export-trakt-run-watched`    | `run-watched`    | `--run --export watched --mode normal`    | Export watched movies only    |
| `export-trakt-run-all`        | `run-all`        | `--run --export all --mode complete`      | Export all data (recommended) |
| `export-trakt-run-collection` | `run-collection` | `--run --export collection --mode normal` | Export collection only        |
| `export-trakt-run-ratings`    | `run-ratings`    | `--run --export ratings --mode complete`  | Export ratings only           |
| `export-trakt-run-watchlist`  | `run-watchlist`  | `--run --export watchlist --mode normal`  | Export watchlist only         |
| `export-trakt-run-shows`      | `run-shows`      | `--run --export shows --mode complete`    | Export shows only             |

### Scheduled Execution Services (`--schedule`)

| Service Name                   | Profile           | Schedule           | Command                                                    | Purpose                |
| ------------------------------ | ----------------- | ------------------ | ---------------------------------------------------------- | ---------------------- |
| `export-trakt-schedule-6h`     | `schedule-6h`     | Every 6 hours      | `--schedule "0 */6 * * *" --export all --mode complete`    | Production scheduler   |
| `export-trakt-schedule-daily`  | `schedule-daily`  | Daily at 2:30 AM   | `--schedule "30 2 * * *" --export all --mode complete`     | Daily backup           |
| `export-trakt-schedule-weekly` | `schedule-weekly` | Sundays at 3:00 AM | `--schedule "0 3 * * 0" --export all --mode complete`      | Weekly backup          |
| `export-trakt-schedule-15min`  | `schedule-15min`  | Every 15 minutes   | `--schedule "*/15 * * * *" --export watched --mode normal` | High-frequency testing |
| `export-trakt-schedule-custom` | `schedule-custom` | Configurable       | Custom via env vars                                        | Custom schedule        |

## Profile Organization

### New Profiles

- **Immediate Execution**: `run`, `run-watched`, `run-all`, `run-collection`, `run-ratings`, `run-watchlist`, `run-shows`
- **Scheduled Execution**: `schedule`, `schedule-6h`, `schedule-daily`, `schedule-weekly`, `schedule-15min`, `schedule-test`, `schedule-custom`
- **Legacy Compatibility**: `legacy`, `legacy-scheduled`

### Updated Profiles

- **Default/Export**: Maintained for backward compatibility, now also tagged as `legacy`
- **Complete/Initial**: Now also tagged as `legacy`
- **Scheduled**: Now tagged as `legacy-scheduled`

## Configuration Changes

### Removed Obsolete Version

```yaml
# REMOVED
version: "3.8"
```

The `version` attribute is no longer needed in modern Docker Compose.

### Enhanced Custom Scheduler

The `schedule-custom` service now properly handles environment variables:

```yaml
export-trakt-schedule-custom:
  <<: *export-trakt-base
  profiles: ["schedule-custom"]
  container_name: export-trakt-schedule-custom
  restart: unless-stopped
  environment:
    - TZ=UTC
    - CUSTOM_SCHEDULE=${SCHEDULE:-0 */6 * * *}
    - CUSTOM_EXPORT_TYPE=${EXPORT_TYPE:-all}
    - CUSTOM_EXPORT_MODE=${EXPORT_MODE:-complete}
  entrypoint: ["/bin/sh", "-c"]
  command:
    - |
      /app/export-trakt --schedule "$${CUSTOM_SCHEDULE}" --export "$${CUSTOM_EXPORT_TYPE}" --mode "$${CUSTOM_EXPORT_MODE}"
```

## Usage Examples

### Quick Commands

```bash
# Test configuration
docker compose --profile run-watched up

# Export all data once
docker compose --profile run-all up

# Start production scheduler
docker compose --profile schedule-6h up -d

# Custom schedule
SCHEDULE="0 */4 * * *" EXPORT_TYPE="watched" EXPORT_MODE="normal" \
docker compose --profile schedule-custom up -d
```

### Migration from Legacy

| Old Command                                | New Equivalent                               |
| ------------------------------------------ | -------------------------------------------- |
| `docker compose up`                        | `docker compose --profile run-watched up`    |
| `docker compose --profile complete up`     | `docker compose --profile run-all up`        |
| `docker compose --profile scheduled up -d` | `docker compose --profile schedule-6h up -d` |

## Documentation Added

### New Files

1. **`docker/README.md`**: Comprehensive Docker usage guide
2. **`docker/test-docker.sh`**: Test script for Docker functionality
3. **`DOCKER_CHANGES.md`**: This file documenting changes

### Updated Files

1. **`docker-compose.yml`**: Complete restructure with new services
2. **Comment section**: Detailed usage examples in the compose file

## Benefits

### For Users

- **Simplified Usage**: Clear profiles for different use cases
- **Flexible Scheduling**: Multiple pre-configured schedules
- **Better Testing**: Dedicated test profiles
- **Backward Compatibility**: Legacy services still work

### For Development

- **Modular Design**: Each service has a specific purpose
- **Easy Extension**: Adding new schedules is straightforward
- **Clear Separation**: Different modes are clearly separated
- **Maintainable**: Well-documented and organized

## Testing

The changes have been tested with:

- ✅ Docker Compose syntax validation (`docker compose config`)
- ✅ Profile listing (`docker compose config --profiles`)
- ✅ Service validation with test script
- ✅ Environment variable handling
- ✅ Backward compatibility

## Backward Compatibility

All existing commands continue to work:

- `docker compose up` (default export)
- `docker compose --profile setup up` (setup)
- `docker compose --profile complete up` (complete export)
- `docker compose --profile scheduled up -d` (legacy scheduler)

## Future Enhancements

The new structure makes it easy to add:

- Additional schedule presets
- Different export configurations
- Health checks
- Monitoring integrations
- Resource limits

This Docker Compose update provides a solid foundation for both current users and future feature development!
