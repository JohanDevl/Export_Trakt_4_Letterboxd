# Configuration and Basic Usage

This document provides detailed information about configuring and using the Export Trakt 4 Letterboxd application.

## Prerequisites

- A Trakt.tv account
- A Trakt.tv application (Client ID and Client Secret)
- jq (for JSON processing)
- curl (for API requests)

## Creating a Trakt.tv Application

1. Log in to your Trakt.tv account
2. Go to https://trakt.tv/oauth/applications
3. Click on "New Application"
4. Fill in the information:
   - Name: Export Trakt 4 Letterboxd
   - Redirect URL: urn:ietf:wg:oauth:2.0:oob
   - Description: (optional)
5. Save the application
6. Note your Client ID and Client Secret

## Setting Up the Configuration File

Copy the example configuration file to create your own:

```bash
cp .config.cfg.example .config.cfg
```

You can edit the configuration file manually if you prefer, but it's recommended to use the setup script in the next step.

## Authentication Configuration

Run the configuration script:

```bash
./setup_trakt.sh
```

This script will guide you through the following steps:

1. Enter your Client ID and Client Secret
2. Enter your Trakt username
3. Obtain an authorization code
4. Generate access tokens

## Basic Usage

### Export Your Data

```bash
./Export_Trakt_4_Letterboxd.sh [option]
```

Available options:

- `normal` (default): Exports rated movies, rated episodes, movie and TV show history, and watchlist
- `initial`: Exports only rated and watched movies
- `complete`: Exports all available data

### Result

The script generates a `letterboxd_import.csv` file that you can import on Letterboxd at the following address: https://letterboxd.com/import/

## Configuration File Options

The configuration file (`.config.cfg`) contains several options that you can customize:

```
# Trakt API credentials
CLIENT_ID="YOUR_TRAKT_CLIENT_ID"
CLIENT_SECRET="YOUR_TRAKT_CLIENT_SECRET"
TRAKT_USERNAME="YOUR_TRAKT_USERNAME"

# TMDB API key (optional, for better movie matching)
TMDB_API_KEY="YOUR_TMDB_API_KEY"

# Export options
EXPORT_RATINGS=true
EXPORT_HISTORY=true
EXPORT_WATCHLIST=true
EXPORT_EPISODES=true

# Date format for export (YYYY-MM-DD)
DATE_FORMAT="%Y-%m-%d"

# Minimum rating to export (1-10)
MIN_RATING=1

# Export path
EXPORT_PATH="/app/copy"

# Backup options
BACKUP_ENABLED=true
BACKUP_DIR="/app/backup"

# Log options
LOG_ENABLED=true
LOG_DIR="/app/logs"
LOG_LEVEL="info"

# Advanced options
USE_TMDB_FOR_MATCHING=true
INCLUDE_YEAR_IN_TITLE=true
INCLUDE_LETTERBOXD_TAGS=true
```

## Troubleshooting

### No Data is Exported

If the script runs without error but no data is exported:

1. Check that your Trakt.tv profile is public
2. Verify that you have correctly configured authentication
3. Run the configuration script again: `./setup_trakt.sh`

### Authentication Errors

If you encounter authentication errors:

1. Check that your Client ID and Client Secret are correct
2. Get a new access token by running `./setup_trakt.sh`

### File Permission Issues

If you encounter file permission issues:

1. Make sure the scripts are executable: `chmod +x *.sh`
2. Check that you have write permissions to the output directories
