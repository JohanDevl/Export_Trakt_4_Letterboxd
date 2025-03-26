# Export Features Guide

This document details the export features of Export Trakt for Letterboxd, explaining the different export modes, data formats, and how to use them effectively.

## Table of Contents

- [Export Formats](#export-formats)
- [Export Modes](#export-modes)
- [Data Sources](#data-sources)
- [Command-Line Options](#command-line-options)
- [Customizing Exports](#customizing-exports)
- [Working with Exported Data](#working-with-exported-data)
- [Importing to Letterboxd](#importing-to-letterboxd)
- [Automation](#automation)

## Export Formats

Export Trakt for Letterboxd generates CSV files compatible with Letterboxd's import format. Two primary export formats are supported:

### 1. Watched Movies List

The watched movies export includes:

- **Title**: Movie title
- **Year**: Release year
- **WatchedDate**: Date when you watched the movie (YYYY-MM-DD format)
- **Rating**: Your rating (0-10 scale, converted to Letterboxd's 0.5-5 star scale)
- **TMDb ID**: Movie identifier from The Movie Database (used for accurate matching)

Example:

```csv
Title,Year,WatchedDate,Rating,TMDbID
The Shawshank Redemption,1994,2023-05-15,5,278
Pulp Fiction,1994,2023-04-22,4.5,680
```

### 2. Watchlist

The watchlist export includes:

- **Title**: Movie title
- **Year**: Release year
- **TMDb ID**: Movie identifier from The Movie Database

Example:

```csv
Title,Year,TMDbID
The Godfather,1972,238
Citizen Kane,1941,15
```

## Export Modes

The application supports several export modes to accommodate different needs:

### 1. Normal Mode (Default)

Exports movies watched and rated since the last export. This is ideal for regular updates to your Letterboxd account.

```bash
export-trakt
```

### 2. Complete Mode

Exports your entire watch history, regardless of when the movies were watched. Use this for a full export of your Trakt.tv history.

```bash
export-trakt --mode complete
```

### 3. Initial Mode

Similar to complete mode, but optimized for first-time exports. Includes additional data enrichment and verification steps.

```bash
export-trakt --mode initial
```

### 4. Date Range Mode

Exports only movies watched within a specific date range.

```bash
export-trakt --from 2023-01-01 --to 2023-12-31
```

## Data Sources

The application can pull data from various sources in your Trakt.tv profile:

### 1. Watch History

Movies you've marked as watched on Trakt.tv, including watch dates and rewatches.

### 2. Ratings

Your movie ratings from Trakt.tv, which are included with the corresponding watched entries.

### 3. Watchlist

Movies you've added to your Trakt.tv watchlist for future viewing.

### 4. Collections (Optional)

Movies in your Trakt.tv collection can be exported separately.

```bash
export-trakt --include-collections
```

## Command-Line Options

The application provides several command-line options to customize your exports:

```
Usage: export-trakt [options]

Options:
  --mode <mode>              Export mode: normal, complete, initial (default: normal)
  --from <date>              Start date for export range (YYYY-MM-DD)
  --to <date>                End date for export range (YYYY-MM-DD)
  --output-dir <path>        Directory to save export files (default: ./exports)
  --include-collections      Include collection items in export
  --include-watchlist        Include watchlist items (default: true)
  --include-ratings          Include ratings (default: true)
  --min-rating <rating>      Only include movies with rating >= value
  --language <lang>          Set UI language (default: en)
  --log-level <level>        Log level: debug, info, warn, error (default: info)
  --config <path>            Path to config file (default: ./config/config.toml)
  --version                  Show version information
  --help                     Show this help message
```

## Customizing Exports

### Filtering by Rating

You can filter exports to include only movies with ratings above a certain threshold:

```bash
export-trakt --min-rating 7
```

This would only export movies you've rated 7 or higher on Trakt's 10-point scale.

### Customizing Output Directory

Change where export files are saved:

```bash
export-trakt --output-dir /path/to/exports
```

### Excluding Data Types

You can exclude certain data types from your export:

```bash
export-trakt --include-watchlist=false
```

## Working with Exported Data

### File Naming Convention

Exported files follow this naming convention:

- `trakt_watched_YYYYMMDD_HHMMSS.csv`: Watched movies
- `trakt_watchlist_YYYYMMDD_HHMMSS.csv`: Watchlist movies
- `trakt_collection_YYYYMMDD_HHMMSS.csv`: Collection items (if enabled)

### Backup and History

All exports are preserved in the export directory, allowing you to:

- Track how your watch history grows over time
- Recover previous exports if needed
- Compare changes between exports

### Manual Editing

You may want to manually edit your exported files before importing to Letterboxd:

1. Open the CSV file in a spreadsheet application like Excel or Google Sheets
2. Make your changes (fix titles, adjust dates, change ratings)
3. Save as CSV with UTF-8 encoding

## Importing to Letterboxd

Once you have your export files, follow these steps to import them to Letterboxd:

1. Go to [Letterboxd.com](https://letterboxd.com/) and log in
2. Go to your profile settings
3. Click on "Import & Export"
4. Select "Import your data"
5. Choose the CSV file from your exports directory
6. Follow the on-screen instructions to complete the import

### Tips for Successful Imports

- Letterboxd limits the number of films per import (currently 1,800)
- If you have more films, split your CSV file into smaller batches
- Imports with TMDb IDs have higher success rates for matching
- Check for any import errors reported by Letterboxd

## Automation

You can automate regular exports using cron jobs or scheduled tasks.

### Linux/macOS Cron Example

To export your data every day at 2 AM:

```bash
# Add to your crontab with 'crontab -e'
0 2 * * * /usr/local/bin/export-trakt --mode normal >> /path/to/logs/export.log 2>&1
```

### Docker Scheduled Example

Using Docker Compose with a cron schedule:

```yaml
version: "3"
services:
  export-trakt:
    image: johandevl/export-trakt-4-letterboxd:latest
    container_name: trakt-export
    volumes:
      - ./config:/app/config
      - ./logs:/app/logs
      - ./exports:/app/exports
    environment:
      - EXPORT_SCHEDULE=0 2 * * * # Run at 2 AM daily
    restart: unless-stopped
```

### Windows Task Scheduler

1. Open Task Scheduler
2. Create a new Basic Task
3. Set the trigger to daily at your preferred time
4. Set the action to start a program
5. Enter the path to `export-trakt.exe`
6. Add any command-line arguments you need

## Best Practices

- Run exports regularly to keep your Letterboxd account in sync with Trakt.tv
- Consider using the `--from` option to only export recent watches if you update frequently
- Keep backups of your export files
- Check the logs if you encounter any issues

With these export features, you can maintain a synchronized history between Trakt.tv and Letterboxd, ensuring your movie watching data is consistent across platforms.
