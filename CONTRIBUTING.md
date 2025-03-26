# Contributing to Export Trakt for Letterboxd

Thank you for your interest in contributing to Export Trakt for Letterboxd! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
  - [Development Setup](#development-setup)
  - [Project Structure](#project-structure)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Pull Requests](#pull-requests)
- [Development Guidelines](#development-guidelines)
  - [Coding Standards](#coding-standards)
  - [Testing](#testing)
  - [Documentation](#documentation)
- [Release Process](#release-process)
- [License](#license)

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) to foster an open and welcoming environment.

## Getting Started

### Development Setup

1. **Fork the repository**

2. **Clone your fork**

   ```bash
   git clone https://github.com/YOUR-USERNAME/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

3. **Set up the development environment**

   ```bash
   # Install Go (if not already installed)
   # macOS (using Homebrew):
   brew install go

   # Ubuntu/Debian:
   sudo apt-get update
   sudo apt-get install golang

   # Windows:
   # Download from https://golang.org/dl/

   # Install dependencies
   go mod download
   ```

4. **Create a branch for your work**
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Project Structure

```
Export_Trakt_4_Letterboxd/
├── cmd/                 # Command-line applications
│   └── export_trakt/    # Main application entry point
├── pkg/                 # Reusable packages
│   ├── api/             # API client for Trakt.tv
│   ├── config/          # Configuration handling
│   ├── export/          # Export functionality
│   ├── i18n/            # Internationalization
│   └── logger/          # Logging facilities
├── tests/               # Test suites
│   ├── integration/     # Integration tests
│   └── mocks/           # Mock objects for testing
├── docs/                # Documentation
├── locales/             # Translation files
└── .github/             # GitHub specific files
```

## How to Contribute

### Reporting Bugs

1. Check if the bug has already been reported in the [Issues](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues)
2. If not, create a new issue using the Bug Report template
3. Provide detailed steps to reproduce the bug
4. Include relevant information about your environment

### Suggesting Enhancements

1. Check if the enhancement has already been suggested in the [Issues](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues)
2. If not, create a new issue using the Feature Request template
3. Describe the enhancement in detail and why it would be valuable

### Pull Requests

1. Update your fork to the latest main branch
2. Create a new branch for your changes
3. Make your changes following the [Development Guidelines](#development-guidelines)
4. Add or update tests as needed
5. Ensure all tests pass
6. Update documentation as required
7. Submit your pull request with a clear description of the changes

## Development Guidelines

### Coding Standards

- Follow Go best practices and style guidelines (use `gofmt` or `goimports`)
- Use meaningful variable and function names
- Keep functions small and focused on a single responsibility
- Write clear comments for complex logic
- Document exported functions and types

### Testing

- Write unit tests for new functionality
- Ensure existing tests pass with your changes
- Use integration tests for API interactions
- Aim for high test coverage, especially for critical code paths

### Documentation

- Update code documentation for public APIs
- Update README.md when adding new features
- Document configuration options
- Consider adding examples for complex features

## Release Process

The project follows [Semantic Versioning](https://semver.org/). For more details on the release process, see the [Release Plan](docs/RELEASE_PLAN.md).

## License

By contributing, you agree that your contributions will be licensed under the project's license. See the [LICENSE](LICENSE) file for details.
