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
[![Code Coverage](https://img.shields.io/badge/coverage-84%25-brightgreen)](coverage.html)
[![Trakt.tv](https://img.shields.io/badge/Trakt.tv-ED1C24?logo=trakt&logoColor=white)](https://trakt.tv)
[![Letterboxd](https://img.shields.io/badge/Letterboxd-00D735?logo=letterboxd&logoColor=white)](https://letterboxd.com)

This project allows you to export your Trakt.tv data to a format compatible with Letterboxd.

## 🚨 Important Update: Migration to Go 🚨

We're migrating the application from Bash to Go for better performance, maintainability, and extended features. The Go version is currently under development in the `feature/go-migration` branch and includes:

- Modern, modular Go architecture with clean separation of concerns
- Improved error handling and logging with multiple levels
- Internationalization (i18n) support for multiple languages
- Robust test coverage (over 80% across all packages)
- Enhanced Trakt.tv API client with retry mechanism and rate limiting

Stay tuned for the upcoming 2.0 release with these improvements!

## Quick Start

### Prerequisites

- A Trakt.tv account
- A Trakt.tv application (Client ID and Client Secret)
- jq and curl (for local installation)
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
     -v $(pwd)/copy:/app/copy \
     -v $(pwd)/backup:/app/backup \
     johandevl/export-trakt-4-letterboxd:latest
   ```

3. For scheduled exports:

   ```bash
   docker compose --profile scheduled up -d
   ```

See [Docker Usage Guide](docs/DOCKER_USAGE.md) for more details.

### Local Installation

1. Clone the repository:

   ```bash
   git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

2. Run the installation script:

   ```bash
   ./install.sh
   ```

3. Configure Trakt authentication:

   ```bash
   ./setup_trakt.sh
   ```

4. Export your data:
   ```bash
   ./Export_Trakt_4_Letterboxd.sh [option]
   ```
   Options: `normal` (default), `initial`, or `complete`

## Features

- Export rated movies and TV shows
- Export watch history with dates and ratings
- Export watchlist items
- Automatic detection of rewatched movies
- Supports various export modes (normal, initial, complete)
- Modular code structure for better maintainability
- Automated exports with cron
- Docker support
- Coming soon: Go implementation with improved performance and reliability

## Project Structure

### Current Bash Version

The codebase has been modularized for better maintenance and readability:

```
Export_Trakt_4_Letterboxd/
├── lib/                     # Library modules
│   ├── config.sh            # Configuration management
│   ├── utils.sh             # Utility functions and debugging
│   ├── trakt_api.sh         # API interaction functions
│   ├── data_processing.sh   # Data transformation functions
│   └── main.sh              # Main orchestration module
├── config/                  # Configuration files
├── logs/                    # Log output
├── backup/                  # Backup of API responses
├── TEMP/                    # Temporary processing files
├── copy/                    # Output CSV files
├── tests/                   # Automated tests
│   ├── unit/                # Unit tests for library modules
│   ├── integration/         # Integration tests
│   ├── mocks/               # Mock API responses
│   ├── run_tests.sh         # Test runner script
│   └── test_helper.bash     # Test helper functions
├── Export_Trakt_4_Letterboxd.sh # Main script (simplified)
├── setup_trakt.sh           # Authentication setup
└── install.sh               # Installation script
```

### Go Version (In Development)

The new Go implementation follows a modern application structure:

```
Export_Trakt_4_Letterboxd/
├── cmd/                     # Application entry points
│   └── export_trakt/        # Main executable
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
├── tests/                   # Test files
│   └── integration/         # Integration tests
└── coverage.html            # Test coverage report
```

## Testing

The project includes comprehensive automated tests to ensure code quality and prevent regressions:

### Running Tests

To run the tests, you need to have the following dependencies installed:

- jq
- bats-core (installed as Git submodule)
- Go (for Go version)

Run all tests for the Bash version:

```bash
./tests/run_tests.sh
```

Run all tests for the Go version:

```bash
go test -v ./...
```

Generate a coverage report for the Go version:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Testing Framework

The testing framework uses:

- Bats (Bash Automated Testing System) for testing the Bash version
- Go testing framework for the Go version
- Mock API responses to test without real API calls
- Integration tests to verify the complete export process
- Unit tests for core library functions

### Continuous Integration

Tests are automatically run in the CI/CD pipeline for every pull request to ensure code quality before merging.

## Documentation

For more detailed information, please refer to the documentation in the `docs` folder:

- [Configuration and Basic Usage](docs/CONFIGURATION.md)
- [Docker Usage Guide](docs/DOCKER_USAGE.md)
- [Docker Testing](docs/DOCKER_TESTING.md)
- [GitHub Actions](docs/GITHUB_ACTIONS.md)
- [Automatic Version Tagging](docs/AUTO_TAGGING.md)
- [Testing Framework](docs/TESTING.md)

## Troubleshooting

If you encounter issues:

1. Check that your Trakt.tv profile is public
2. Verify your authentication configuration
3. Run `./setup_trakt.sh` again to refresh your tokens
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
