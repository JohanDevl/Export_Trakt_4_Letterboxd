# Changelog

All notable changes to the Export Trakt for Letterboxd project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **🎨 Modern GitHub Templates**: Complete YAML-based issue forms with structured validation
  - 🐛 Comprehensive bug report template with environment detection
  - ✨ Feature request template with priority levels and implementation tracking
  - 📚 Documentation issue template for targeted improvements
  - ❓ Question & support template with categorized help requests
- **📋 Enhanced Pull Request Template**: Professional template with comprehensive sections
  - Testing, security, performance, and deployment considerations
  - Code quality checklist and documentation requirements
  - Professional review guidelines and contributor confirmation
- **🤖 Professional Bot Configuration**: Modern community automation with helpful guidance
- **📖 Complete README Overhaul**: Enterprise-grade documentation with modern structure
  - Professional layout with emojis and clear visual hierarchy
  - Comprehensive Docker Compose usage examples for production and development
  - Detailed troubleshooting guide and development setup instructions
  - Enhanced configuration documentation and internationalization details
- **🔧 GitHub Actions Workflow Integration**: Proper badge references for build status
- **🌍 Multi-language Support**: Enhanced internationalization documentation
- **🐳 Production-Ready Docker Examples**: Complete Docker Compose profiles for various use cases
- **📊 Comprehensive Export Documentation**: Detailed export modes, types, and usage examples

### Changed

- **Improved GitHub Community Management**: Structured issue templates replace basic markdown forms
- **Enhanced Documentation Standards**: Professional documentation matching enterprise open source projects
- **Better User Experience**: Streamlined issue creation with guided forms and helpful links
- **Modernized Project Appearance**: Professional badges, layout, and visual hierarchy
- **Updated Bot Messages**: Helpful, actionable guidance for new contributors

### Fixed

- **Badge References**: Corrected GitHub Actions workflow badge URLs
- **Documentation Links**: Fixed broken references and outdated information
- **Template Structure**: Removed legacy markdown templates and reorganized structure

## [2.0.0] - 2025-05-23

### Added

- **🚀 New Execution Modes**: `--run` flag for immediate one-time execution and `--schedule` flag for cron-based scheduling
- **⏰ Comprehensive Cron Scheduler**: Built-in scheduler with detailed logging and status reporting
  - Cron schedule validation with helpful error messages and examples
  - Support for immediate execution mode for testing and CI/CD integration
  - Enhanced command-line interface with new scheduling options
- **🧪 Comprehensive Test Suite**: Unit and integration tests with high coverage
  - Package-specific test coverage reporting
  - Integration tests for API interactions
  - Mock objects for reliable testing
- **🌍 Internationalization (i18n)**: Full multilingual support
  - English and French translations (with German and Spanish support)
  - Configurable language selection
  - Localized error messages and user interface
- **🔄 GitHub Actions CI/CD**: Automated testing and deployment pipeline
  - Automated release workflow for cross-platform binary generation
  - Multi-platform Docker image builds (amd64, arm64, armv7)
  - Comprehensive testing on multiple platforms
- **📚 Enhanced Documentation**: Complete documentation overhaul
  - Contributing guide with development setup instructions
  - Configuration guide with detailed examples
  - Installation instructions for multiple platforms
  - Troubleshooting guide with common solutions

### Changed

- **Complete Rewrite in Go**: Improved performance and maintainability over original shell scripts
- **Structured Configuration**: TOML-based configuration with comprehensive validation
- **Enhanced Command-Line Interface**: Improved argument handling with support for multiple execution modes
- **Better Error Handling**: Descriptive error messages with actionable guidance
- **Improved Logging System**: Structured logging with configurable levels and file output
- **Enhanced API Client**: Better error handling for Trakt.tv API interactions

### Removed

- **Legacy Dependencies**: No longer requires external tools (jq, curl)
- **Temporary File Usage**: Improved data processing without temporary files
- **Shell Script Implementation**: Replaced with robust Go implementation

## [1.5.0] - 2023-07-15

### Added

- **🐳 Docker Support**: Multi-architecture Docker images (amd64, arm64, armv7)
- **🔄 GitHub Actions Workflows**: Automated Docker builds and testing
- **📦 Collection Export**: Option to include collection items in export
- **🎯 Enhanced Filtering**: Better support for filtering by various criteria

### Changed

- **Improved Error Handling**: Better error reporting and recovery
- **Enhanced Character Support**: Better handling of special characters in movie titles
- **Refined Matching Algorithm**: Improved movie matching accuracy

### Fixed

- **Character Encoding Issues**: Resolved problems with international characters
- **API Rate Limiting**: Better handling of Trakt.tv API rate limits

## [1.4.0] - 2023-05-20

### Added

- **💾 Backup Functionality**: Automatic backup of API responses for recovery
- **📺 TV Show Support**: Basic support for exporting TV show data
- **📝 File Logging**: Optional logging to file with rotation
- **📅 Enhanced Date Handling**: Better date parsing and formatting options

### Changed

- **Improved Authentication**: Enhanced OAuth flow with better token management
- **Better API Handling**: More robust API request processing

### Fixed

- **Date Format Issues**: Resolved problems with various date formats
- **Authentication Tokens**: Better handling of expired tokens

## [1.3.0] - 2023-03-10

### Added

- **⭐ Rating Filters**: Support for filtering by minimum rating
- **📅 Year in Titles**: Option to include release year in movie titles
- **📋 Enhanced Watchlist**: Improved watchlist export functionality

### Changed

- **CSV Formatting**: Improved CSV output formatting for better Letterboxd compatibility
- **Error Messages**: More descriptive and actionable error messages

### Fixed

- **CSV Escaping**: Proper escaping of special characters in CSV output
- **Unicode Support**: Better handling of Unicode characters in titles

## [1.2.0] - 2023-01-25

### Added

- **🔄 Export Modes**: Multiple export modes (normal, initial, complete)
- **📈 Watched History**: Support for exporting watched history with dates
- **🔍 Rewatch Detection**: Automatic detection and handling of rewatched movies

### Changed

- **Enhanced Trakt.tv Integration**: Improved API integration with better error handling
- **Configuration Management**: Better configuration file handling

### Fixed

- **Duplicate Entries**: Resolved issues with duplicate movie entries
- **Date Accuracy**: Improved accuracy of watch dates

## [1.1.0] - 2022-11-12

### Added

- **🎬 TMDB Integration**: Integration with The Movie Database for better matching
- **⭐ Rating Export**: Support for exporting user ratings
- **⚙️ Configuration File**: TOML-based configuration for customization

### Changed

- **Authentication Mechanism**: Enhanced OAuth authentication with Trakt.tv
- **Export Format**: Improved export format for better Letterboxd compatibility

### Fixed

- **Movie Matching**: Better movie matching between Trakt.tv and Letterboxd
- **API Reliability**: Improved reliability of API calls

## [1.0.0] - 2022-09-01

### Added

- **🎬 Initial Release**: Basic functionality to export Trakt.tv movie data
- **📊 Letterboxd Format**: Support for exporting to Letterboxd-compatible CSV format
- **🔐 Trakt.tv Authentication**: Simple authentication with Trakt.tv API
- **⚙️ Basic Configuration**: Essential configuration options for basic usage

---

## Migration Guide

### From v1.x to v2.0

1. **Configuration Update**: Convert your configuration from environment variables to `config.toml`
2. **Command Changes**: Update your commands to use the new CLI interface
3. **Docker Updates**: Pull the latest Docker images which now support multi-architecture
4. **Schedule Format**: Update cron schedules to use the new `--schedule` flag format

### Updating to Latest

```bash
# Docker users
docker pull johandevl/export-trakt-4-letterboxd:latest

# Source builds
git pull origin main
go build -o export_trakt ./cmd/export_trakt/
```

For detailed migration instructions, see our [Migration Guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Migration-Guide).
