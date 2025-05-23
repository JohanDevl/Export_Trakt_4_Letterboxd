# Logs Directory

This directory contains application log files generated during runtime.

## Log Files

The application generates several types of log files:

- `export.log` - Main application log
- `app.log` - General application events
- `cron.log` - Scheduled export logs
- `Export_Trakt_4_Letterboxd_YYYY-MM-DD_HH-MM-SS.log` - Timestamped execution logs

## Log Levels

The application supports multiple log levels:

- `ERROR` - Error messages
- `WARN` - Warning messages
- `INFO` - Informational messages
- `DEBUG` - Debug messages (verbose mode)

## Configuration

Log level can be configured in the `config.toml` file:

```toml
[logging]
level = "info"
```

## Note

Log files are automatically ignored by git as they contain runtime information and can become large over time.
