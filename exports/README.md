# Exports Directory

This directory contains the exported data from Trakt.tv in Letterboxd-compatible format.

## Structure

Export files are automatically generated with timestamps:

- `export_YYYY-MM-DD_HH-MM/` - Timestamped export directories
- Each export contains CSV files compatible with Letterboxd import

## Usage

Export files are created when running the application:

```bash
./export_trakt --config ./config/config.toml
```

## Note

Export files are automatically ignored by git as they contain personal data and are meant to be used locally or uploaded to Letterboxd manually.
