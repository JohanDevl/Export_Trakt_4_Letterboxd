# Changelog

All notable changes to the Export Trakt for Letterboxd project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **New execution modes**: `--run` flag for immediate one-time execution and `--schedule` flag for cron-based scheduling
- Comprehensive cron schedule validation with helpful error messages and examples
- Built-in scheduler with detailed logging and status reporting
- Support for immediate execution mode for testing and CI/CD integration
- Enhanced command-line interface with new scheduling options
- Comprehensive test suite with unit and integration tests
- Internationalization (i18n) support with English and French translations
- GitHub Actions CI/CD pipeline for automated testing
- Automated release workflow for cross-platform binary generation
- New issue templates for bug reports, feature requests, and beta feedback
- Enhanced documentation including contributing guide, installation instructions, and configuration guide
- Detailed scheduling examples and best practices documentation

### Changed

- Improved command-line argument handling with support for multiple execution modes
- Enhanced logging with scheduler-specific messages and status updates
- Better error handling for invalid cron expressions with user-friendly feedback

## [2.0.0] - TBD

### Added

- Complete rewrite in Go for improved performance and maintainability
- Structured configuration using TOML format
- Comprehensive logging system with support for different log levels
- Advanced error handling with descriptive error messages
- Internationalization (i18n) support
- Multiple export formats (watched movies, watchlist, collections)
- Command-line interface with various options and flags
- Rate limiting for API requests to prevent exceeding Trakt.tv limits
- Retry mechanism for handling transient API failures
- Progress indication during exports
- Enhanced movie matching using TMDb IDs
- Support for advanced filtering (by rating, date range)
- Better handling of rewatched movies
- Cross-platform compatibility (Linux, macOS, Windows, ARM)
- Docker support with multi-arch images

### Changed

- Improved configuration handling with support for environment variables
- Enhanced Trakt.tv API client with better error handling
- More efficient data processing for large movie collections
- Better date handling with proper timezone support
- Improved CSV generation with proper escaping and formatting
- More reliable authentication flow with token refresh

### Removed

- Dependency on external tools (jq, curl)
- Temporary file usage for data processing

## [1.5.0] - 2023-07-15

### Added

- Docker support with multi-arch images
- GitHub Actions workflows for Docker builds
- Option to include collection items in export

### Changed

- Improved error handling and reporting
- Better support for special characters in movie titles
- Enhanced matching algorithm for movies

## [1.4.0] - 2023-05-20

### Added

- Backup functionality for API responses
- Support for exporting TV shows
- Optional logging to file
- Better date handling options

### Changed

- Improved authentication flow
- Enhanced API request handling

## [1.3.0] - 2023-03-10

### Added

- Support for filtering by minimum rating
- Option to include year in movie titles
- Enhanced watchlist export

### Changed

- Improved CSV formatting
- Better error messages

## [1.2.0] - 2023-01-25

### Added

- Export modes: normal, initial, complete
- Support for watched history with dates
- Automatic detection of rewatched movies

### Changed

- Improved Trakt.tv API integration
- Better configuration handling

## [1.1.0] - 2022-11-12

### Added

- TMDB integration for better movie matching
- Support for exporting ratings
- Configuration file for customization

### Changed

- Enhanced authentication mechanism
- Improved export format

## [1.0.0] - 2022-09-01

### Added

- Initial release
- Basic functionality to export Trakt.tv data
- Support for exporting to Letterboxd CSV format
- Simple authentication with Trakt.tv API
- Basic configuration options
