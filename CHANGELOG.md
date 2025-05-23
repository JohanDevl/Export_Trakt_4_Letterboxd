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

## [2.0.0]

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
