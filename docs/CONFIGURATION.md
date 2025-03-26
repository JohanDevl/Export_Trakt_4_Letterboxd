# Configuration Guide

This document provides detailed information about configuring Export Trakt for Letterboxd to suit your needs.

## Table of Contents

- [Configuration File](#configuration-file)
- [Configuration Options](#configuration-options)
- [Environment Variables](#environment-variables)
- [Command-Line Overrides](#command-line-overrides)
- [Configuration Examples](#configuration-examples)
- [Advanced Configuration](#advanced-configuration)

## Configuration File

Export Trakt for Letterboxd uses a TOML configuration file located at `config/config.toml` by default. This file contains all the settings needed to run the application.

### Basic Structure

The configuration file is organized into sections:

```toml
[trakt]
# Trakt.tv API settings

[export]
# Export settings

[logging]
# Logging configuration

[i18n]
# Internationalization settings
```

### Creating a Configuration File

You can create a configuration file manually or use the application's interactive setup:

```bash
# Interactive setup
export-trakt setup

# Or manually create the file
mkdir -p config
touch config/config.toml
```

## Configuration Options

### Trakt Section

The `[trakt]` section contains settings for connecting to the Trakt.tv API:

```toml
[trakt]
# Required: Your Trakt.tv API client ID
client_id = "YOUR_CLIENT_ID"

# Required: Your Trakt.tv API client secret
client_secret = "YOUR_CLIENT_SECRET"

# Optional: Redirect URI for OAuth authentication
# Default: "urn:ietf:wg:oauth:2.0:oob"
redirect_uri = "urn:ietf:wg:oauth:2.0:oob"

# Optional: Token file path for storing authentication tokens
# Default: "./config/token.json"
token_file = "./config/token.json"

# Optional: API request timeout in seconds
# Default: 30
timeout = 30

# Optional: Maximum number of retries for failed API requests
# Default: 3
max_retries = 3

# Optional: API request rate limit per minute
# Default: 60
rate_limit = 60
```

### Export Section

The `[export]` section controls how data is exported:

```toml
[export]
# Optional: Directory where export files will be saved
# Default: "./exports"
output_dir = "./exports"

# Optional: Export file naming format
# Default: "trakt_{{type}}_{{timestamp}}.csv"
file_format = "trakt_{{type}}_{{timestamp}}.csv"

# Optional: Default export mode
# Options: "normal", "complete", "initial"
# Default: "normal"
mode = "normal"

# Optional: Include watchlist items in export
# Default: true
include_watchlist = true

# Optional: Include collection items in export
# Default: false
include_collections = false

# Optional: Include ratings in export
# Default: true
include_ratings = true

# Optional: Minimum rating to include (0-10 scale)
# Default: 0 (include all ratings)
min_rating = 0

# Optional: Convert Trakt's 10-point scale to Letterboxd's 5-star scale
# Default: true
convert_ratings = true

# Optional: Keep temporary files after export
# Default: false
keep_temp_files = false
```

### Logging Section

The `[logging]` section configures logging behavior:

```toml
[logging]
# Optional: Log level
# Options: "debug", "info", "warn", "error"
# Default: "info"
level = "info"

# Optional: Log file path
# Default: "./logs/export.log"
file = "./logs/export.log"

# Optional: Maximum log file size in MB before rotation
# Default: 10
max_size = 10

# Optional: Maximum number of log files to keep
# Default: 5
max_files = 5

# Optional: Enable console logging
# Default: true
console = true

# Optional: Enable color in console output
# Default: true
color = true
```

### Internationalization Section

The `[i18n]` section configures language settings:

```toml
[i18n]
# Optional: UI language
# Options: "en", "fr", etc. (depends on available translations)
# Default: "en"
language = "en"

# Optional: Path to directory containing locale files
# Default: "./locales"
locales_dir = "./locales"
```

## Environment Variables

All configuration options can also be set using environment variables, which is particularly useful in containerized environments. Environment variables take precedence over the configuration file.

Variables follow the pattern: `EXPORT_TRAKT_<SECTION>_<OPTION>` (uppercase, with underscores).

Examples:

```bash
# Set Trakt.tv client ID
export EXPORT_TRAKT_TRAKT_CLIENT_ID="your-client-id"

# Set log level to debug
export EXPORT_TRAKT_LOGGING_LEVEL="debug"

# Enable collection export
export EXPORT_TRAKT_EXPORT_INCLUDE_COLLECTIONS="true"
```

In Docker, you can set these in your docker-compose.yml:

```yaml
services:
  export-trakt:
    image: johandevl/export-trakt-4-letterboxd:latest
    environment:
      - EXPORT_TRAKT_TRAKT_CLIENT_ID=your-client-id
      - EXPORT_TRAKT_TRAKT_CLIENT_SECRET=your-client-secret
      - EXPORT_TRAKT_EXPORT_OUTPUT_DIR=/app/exports
```

## Command-Line Overrides

Most configuration options can be overridden using command-line flags, which take highest precedence:

```bash
# Override output directory
export-trakt --output-dir /custom/path/to/exports

# Override log level
export-trakt --log-level debug

# Use a different configuration file
export-trakt --config /path/to/custom-config.toml
```

Run `export-trakt --help` to see all available command-line options.

## Configuration Examples

### Basic Configuration

A minimal configuration file with just the required settings:

```toml
[trakt]
client_id = "your-client-id"
client_secret = "your-client-secret"
```

### Complete Export Configuration

Configuration optimized for a complete export of all Trakt.tv data:

```toml
[trakt]
client_id = "your-client-id"
client_secret = "your-client-secret"
max_retries = 5
timeout = 60

[export]
mode = "complete"
output_dir = "./exports/complete"
include_watchlist = true
include_collections = true
include_ratings = true
file_format = "trakt_complete_{{type}}_{{timestamp}}.csv"

[logging]
level = "info"
file = "./logs/complete_export.log"
```

### Regular Update Configuration

Configuration for regular updates to Letterboxd:

```toml
[trakt]
client_id = "your-client-id"
client_secret = "your-client-secret"

[export]
mode = "normal"
output_dir = "./exports/updates"
include_watchlist = false
min_rating = 0

[logging]
level = "info"
file = "./logs/updates.log"
max_files = 10
```

### Development Configuration

Configuration useful during development:

```toml
[trakt]
client_id = "your-client-id"
client_secret = "your-client-secret"

[export]
mode = "normal"
output_dir = "./dev_exports"
keep_temp_files = true

[logging]
level = "debug"
file = "./logs/dev.log"
console = true
color = true

[i18n]
language = "en"
```

## Advanced Configuration

### Multiple Configuration Files

You can maintain multiple configuration files for different purposes and select which to use at runtime:

```bash
# Use complete export configuration
export-trakt --config ./config/complete.toml

# Use update configuration
export-trakt --config ./config/update.toml
```

### Using jq to Modify Configuration

You can programmatically modify your configuration using tools like `jq`:

```bash
# Update the output directory
cat config/config.toml | jq '.export.output_dir = "./new_exports"' > config/config.toml.new
mv config/config.toml.new config/config.toml
```

### Secure Storage of API Keys

For improved security, consider:

1. Using environment variables for sensitive data
2. Using a secrets manager
3. Setting restricted file permissions on your config file:

```bash
chmod 600 config/config.toml
```

### Configuration Validation

The application validates your configuration at startup. For manual validation:

```bash
export-trakt validate --config ./config/config.toml
```

This will check your configuration for errors without running the export.
