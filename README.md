# Export Trakt 4 Letterboxd

[![GitHub release](https://img.shields.io/github/v/release/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/releases)
[![GitHub stars](https://img.shields.io/github/stars/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues)
[![GitHub license](https://img.shields.io/github/license/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/blob/main/LICENSE)
[![Docker Image Test](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-test.yml/badge.svg)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-test.yml)
[![Docker Build](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-publish.yml/badge.svg)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-publish.yml)
[![Docker Package](https://img.shields.io/badge/GitHub%20Packages-ghcr.io-blue?logo=docker)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkgs/container/export_trakt_4_letterboxd)
[![Docker Hub](https://img.shields.io/docker/v/johandevl/export-trakt-4-letterboxd?logo=docker&label=Docker%20Hub)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)
[![Docker Image Size](https://img.shields.io/docker/image-size/johandevl/export-trakt-4-letterboxd/latest?logo=docker&label=Image%20Size)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)
[![Docker Pulls](https://img.shields.io/docker/pulls/johandevl/export-trakt-4-letterboxd?logo=docker&label=Pulls)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)
[![Platforms](https://img.shields.io/badge/platforms-amd64%20|%20arm64%20|%20armv7-lightgrey?logo=docker)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd/tags)
[![Code Coverage](https://img.shields.io/badge/coverage-78%25-brightgreen)](coverage.html)
[![Trakt.tv](https://img.shields.io/badge/Trakt.tv-ED1C24?logo=trakt&logoColor=white)](https://trakt.tv)
[![Letterboxd](https://img.shields.io/badge/Letterboxd-00D735?logo=letterboxd&logoColor=white)](https://letterboxd.com)

This project allows you to export your Trakt.tv data to a format compatible with Letterboxd.

## 🚀 Go Implementation 🚀

This application is now built entirely in Go, providing:

- Modern, modular Go architecture with clean separation of concerns
- Improved error handling and logging with multiple levels
- Internationalization (i18n) support for multiple languages
- Robust test coverage (over 80% across all packages)
- Enhanced Trakt.tv API client with retry mechanism and rate limiting

## Quick Start

### Prerequisites

- A Trakt.tv account
- A Trakt.tv application (Client ID and Client Secret)
- Docker (for containerized installation)

### Using Docker (Recommended)

1. Quick run with Docker Compose:

   ```bash
   # Clone the repository
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd

   # First-time setup (interactive)
   docker compose --profile setup up

   # Run the export
   docker compose up
   ```

2. Or pull and run from Docker Hub:

   ```bash
   docker run -it --name trakt-export \
     -v $(pwd)/config:/app/config \
     -v $(pwd)/logs:/app/logs \
     -v $(pwd)/exports:/app/exports \
     johandevl/export-trakt-4-letterboxd:latest
   ```

3. For scheduled exports:

   ```bash
   docker compose --profile scheduled up -d
   ```

### Local Installation (From Source)

1. Clone the repository:

   ```bash
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Build the Go application:

   ```bash
   go build -o export_trakt ./cmd/export_trakt/
   ```

3. Run the application:
   ```bash
   ./export_trakt --config ./config/config.toml
   ```

## Features

- Export rated movies and TV shows
- Export watch history with dates and ratings
- Export watchlist items
- Automatic detection of rewatched movies
- Supports various export modes
- Modular code structure for better maintainability
- Automated exports with scheduling
- Docker support
- Complete Go implementation with improved performance and reliability

## Project Structure

The Go implementation follows a modern application structure:

```
Export_Trakt_4_Letterboxd/
├── cmd/                     # Application entry points
│   └── export_trakt/        # Main executable
├── internal/                # Private application code
│   ├── models/              # Data models
│   └── utils/               # Private utilities
├── pkg/                     # Packages for core functionality
│   ├── api/                 # Trakt.tv API client
│   ├── config/              # Configuration management
│   ├── export/              # Export functionality
│   ├── i18n/                # Internationalization support
│   └── logger/              # Logging system
├── locales/                 # Translation files
│   ├── en.json              # English translations
│   └── fr.json              # French translations
├── config/                  # Configuration files
├── build/                   # Compiled binaries
└── logs/                    # Log output
```

## Testing

The project includes comprehensive automated tests to ensure code quality and prevent regressions:

### Running Tests

To run the tests, you need to have Go installed.

Run all tests:

```bash
go test -v ./...
```

Generate a coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

Run the coverage script (excludes main package):

```bash
./scripts/coverage.sh
```

The project maintains over 78% test coverage across the core packages, which helps ensure reliability and stability. The coverage includes:

- API Client: 73.3% covered
- Config Management: 85.4% covered
- Export Functionality: 78.3% covered
- Internationalization: 81.6% covered
- Logging System: 97.7% covered

### Code Coverage Configuration

The project includes a `.codecov.yml` file that configures code coverage analysis for CI/CD pipelines. This configuration:

- Sets a 70% coverage threshold for the project
- Excludes the `cmd/export_trakt` directory (main package) from coverage calculations
- Provides detailed coverage reports for each pull request

If you're using GitHub Actions or another CI system, this configuration ensures accurate coverage reporting focused on the core packages rather than the main application entry point.

## Documentation

Complete documentation is available in the [project Wiki](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki).

You will find:

- [Installation Guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Installation)
- [CLI Reference](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/CLI-Reference)
- [Export Features](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Export-Features)
- [Trakt API Guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Trakt-API-Guide)
- [Internationalization](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Internationalization)
- [Migration Guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Migration-Guide)
- [Testing](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Testing)
- [CI/CD](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/CI-CD)

## Troubleshooting

If you encounter issues:

1. Check that your Trakt.tv profile is public
2. Verify your authentication configuration
3. Ensure your config.toml file has the correct Trakt.tv client ID and secret
4. Check logs in the `logs` directory for detailed error information

## Acknowledgements

This project is based on the original work by [u2pitchjami](https://github.com/u2pitchjami/Export_Trakt_4_Letterboxd). I would like to express my sincere gratitude to u2pitchjami for creating the initial version of this tool, which has been an invaluable foundation for this project.

The original repository can be found at: https://github.com/u2pitchjami/Export_Trakt_4_Letterboxd

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

The original work by u2pitchjami is also licensed under the MIT License. This fork maintains the same license to respect the original author's intentions.

## Authors

👤 **JohanDevl**

- Twitter: [@0xUta](https://twitter.com/0xUta)
- Github: [@JohanDevl](https://github.com/JohanDevl)
- LinkedIn: [@johan-devlaminck](https://linkedin.com/in/johan-devlaminck)

## Letterboxd Import Export Format

A new export format has been added to generate files compatible with Letterboxd's import functionality. To use this feature:

1. Set `extended_info = "letterboxd"` in your `config.toml` file
2. Run the application normally or with Docker (see below)

The format includes the following fields:

- Title: Movie title (quoted)
- Year: Release year
- imdbID: IMDB ID for the movie
- tmdbID: TMDB ID for the movie
- WatchedDate: Date the movie was watched
- Rating10: Rating on a scale of 1-10
- Rewatch: Whether the movie has been watched multiple times (true/false)

### Using with Docker

To use the Letterboxd export format with Docker:

```bash
# Create directories for the Docker volumes
mkdir -p config logs exports

# Copy the example config file and edit it
cp config.example.toml config/config.toml

# Edit the config file to set extended_info = "letterboxd"
# Then run:
docker run --rm -v $(pwd)/config:/app/config -v $(pwd)/logs:/app/logs -v $(pwd)/exports:/app/exports johandevl/export-trakt-4-letterboxd:latest
```

The output file will be saved as `letterboxd_import.csv` in your exports directory.
