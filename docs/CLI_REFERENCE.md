# Command-Line Interface Reference

This document provides a detailed reference for the Export Trakt for Letterboxd command-line interface, including all commands, options, and usage examples.

## Table of Contents

- [Basic Usage](#basic-usage)
- [Global Options](#global-options)
- [Commands](#commands)
  - [Export Command](#export-command)
  - [Setup Command](#setup-command)
  - [Auth Command](#auth-command)
  - [Validate Command](#validate-command)
  - [Info Command](#info-command)
- [Usage Examples](#usage-examples)
- [Environment Variables](#environment-variables)
- [Exit Codes](#exit-codes)

## Basic Usage

The basic syntax for using the command-line interface is:

```
export-trakt [global options] command [command options] [arguments...]
```

If no command is specified, the default action is to run the export command.

## Global Options

These options can be used with any command:

| Option             | Description                                      | Default                |
| ------------------ | ------------------------------------------------ | ---------------------- |
| `--config`, `-c`   | Path to the configuration file                   | `./config/config.toml` |
| `--log-level`      | Set the logging level (debug, info, warn, error) | `info`                 |
| `--language`, `-l` | Set the interface language                       | `en`                   |
| `--help`, `-h`     | Show help                                        |                        |
| `--version`, `-v`  | Show version information                         |                        |

## Commands

### Export Command

The `export` command exports your Trakt.tv data to Letterboxd-compatible CSV files.

#### Syntax

```
export-trakt export [options]
```

#### Options

| Option                  | Description                                | Default     |
| ----------------------- | ------------------------------------------ | ----------- |
| `--mode`, `-m`          | Export mode (normal, complete, initial)    | `normal`    |
| `--output-dir`, `-o`    | Directory where export files will be saved | `./exports` |
| `--from`, `-f`          | Start date for export range (YYYY-MM-DD)   |             |
| `--to`, `-t`            | End date for export range (YYYY-MM-DD)     |             |
| `--include-watchlist`   | Include watchlist items                    | `true`      |
| `--include-collections` | Include collection items                   | `false`     |
| `--include-ratings`     | Include ratings                            | `true`      |
| `--min-rating`          | Minimum rating to include (0-10)           | `0`         |

### Setup Command

The `setup` command helps you create a new configuration file through an interactive prompt.

#### Syntax

```
export-trakt setup [options]
```

#### Options

| Option              | Description                                       | Default |
| ------------------- | ------------------------------------------------- | ------- |
| `--force`, `-f`     | Overwrite existing configuration file             | `false` |
| `--non-interactive` | Run in non-interactive mode (requires all params) | `false` |

### Auth Command

The `auth` command manages authentication with the Trakt.tv API.

#### Syntax

```
export-trakt auth [subcommand]
```

#### Subcommands

- `init`: Initialize authentication for a new device
- `refresh`: Refresh the authentication token
- `status`: Check authentication status
- `revoke`: Revoke authentication

### Validate Command

The `validate` command validates your configuration file without performing an export.

#### Syntax

```
export-trakt validate [options]
```

#### Options

| Option  | Description                                | Default |
| ------- | ------------------------------------------ | ------- |
| `--fix` | Attempt to fix common configuration issues | `false` |

### Info Command

The `info` command displays information about your Trakt.tv account and Export Trakt for Letterboxd.

#### Syntax

```
export-trakt info [subcommand]
```

#### Subcommands

- `account`: Display Trakt.tv account information
- `stats`: Display statistics about your Trakt.tv data
- `system`: Display system information
- `version`: Display version information

## Usage Examples

### Basic Export

Export your watched movies with the default settings:

```bash
export-trakt
```

### Complete Export

Export your entire Trakt.tv history:

```bash
export-trakt export --mode complete
```

### Date Range Export

Export movies watched in 2023:

```bash
export-trakt export --from 2023-01-01 --to 2023-12-31
```

### Custom Output Directory

Export to a specific directory:

```bash
export-trakt export --output-dir ~/Documents/letterboxd_exports
```

### Filtering by Rating

Export only movies you've rated 7 or higher:

```bash
export-trakt export --min-rating 7
```

### Export Without Watchlist

Export only watched movies, not your watchlist:

```bash
export-trakt export --include-watchlist=false
```

### Complete Export with Collections

Export everything, including your movie collection:

```bash
export-trakt export --mode complete --include-collections
```

### Export with Verbose Logging

Export with detailed debug information:

```bash
export-trakt --log-level debug
```

### Using a Different Configuration File

Use an alternative configuration file:

```bash
export-trakt --config ~/custom_config.toml
```

### Interactive Setup

Run the interactive setup to create a new configuration file:

```bash
export-trakt setup
```

### Validate Configuration

Check if your configuration file is valid:

```bash
export-trakt validate
```

### Check Account Information

Display information about your Trakt.tv account:

```bash
export-trakt info account
```

### Get Export Statistics

View statistics about your exported data:

```bash
export-trakt info stats
```

## Environment Variables

All command-line options can also be set using environment variables. The environment variables follow the pattern `EXPORT_TRAKT_<OPTION>` where `<OPTION>` is the uppercase name of the option with dashes replaced by underscores.

Examples:

```bash
# Set configuration file path
export EXPORT_TRAKT_CONFIG="/path/to/config.toml"

# Set logging level
export EXPORT_TRAKT_LOG_LEVEL="debug"

# Set export mode
export EXPORT_TRAKT_MODE="complete"
```

## Exit Codes

The application uses the following exit codes:

| Code | Description          |
| ---- | -------------------- |
| 0    | Success              |
| 1    | General error        |
| 2    | Configuration error  |
| 3    | Authentication error |
| 4    | API error            |
| 5    | File system error    |
| 6    | Network error        |
| 7    | User input error     |
