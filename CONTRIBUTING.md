# Contributing to Export Trakt for Letterboxd

Thank you for your interest in contributing to Export Trakt for Letterboxd! This document provides comprehensive guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
  - [Development Setup](#development-setup)
  - [Project Structure](#project-structure)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Enhancements](#suggesting-enhancements)
  - [Improving Documentation](#improving-documentation)
  - [Pull Requests](#pull-requests)
- [Development Guidelines](#development-guidelines)
  - [Coding Standards](#coding-standards)
  - [Testing](#testing)
  - [Documentation](#documentation)
  - [Internationalization](#internationalization)
- [Release Process](#release-process)
- [Community](#community)
- [License](#license)

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) to foster an open and welcoming environment for all contributors.

## Getting Started

### Development Setup

1. **Fork the repository**

   - Click the "Fork" button on the GitHub repository page
   - Clone your fork locally

2. **Clone your fork**

   ```bash
   git clone https://github.com/YOUR-USERNAME/Export_Trakt_4_Letterboxd.git
   cd Export_Trakt_4_Letterboxd
   ```

3. **Set up the development environment**

   ```bash
   # Install Go 1.22+ (if not already installed)

   # macOS (using Homebrew):
   brew install go

   # Ubuntu/Debian:
   sudo apt-get update
   sudo apt-get install golang-go

   # Windows:
   # Download from https://golang.org/dl/

   # Verify Go installation
   go version  # Should show Go 1.22 or higher

   # Install dependencies
   go mod download
   go mod tidy
   ```

4. **Set up pre-commit hooks (optional but recommended)**

   ```bash
   # Install pre-commit (if available)
   brew install pre-commit  # macOS
   # or
   pip install pre-commit  # Python users

   # Install hooks
   pre-commit install
   ```

5. **Create a branch for your work**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/issue-number-description
   ```

### Project Structure

```
Export_Trakt_4_Letterboxd/
├── cmd/                     # 🎯 Command-line applications
│   └── export_trakt/        # Main application entry point
├── pkg/                     # 📦 Reusable packages
│   ├── api/                 # 🌐 API client for Trakt.tv
│   ├── config/              # ⚙️ Configuration handling
│   ├── export/              # 📊 Export functionality
│   ├── i18n/                # 🌍 Internationalization
│   ├── logger/              # 📝 Logging facilities
│   └── scheduler/           # ⏰ Cron scheduling
├── internal/                # 🔒 Private application code
│   ├── models/              # 🗂️ Data models
│   └── utils/               # 🛠️ Private utilities
├── tests/                   # 🧪 Test suites
│   ├── integration/         # Integration tests
│   └── mocks/               # Mock objects for testing
├── locales/                 # 🗣️ Translation files
├── config/                  # 📋 Configuration examples
├── scripts/                 # 🚀 Build and utility scripts
├── .github/                 # 🏗️ GitHub workflows and templates
│   ├── workflows/           # CI/CD pipelines
│   └── ISSUE_TEMPLATE/      # Modern issue forms
└── docs/                    # 📖 Additional documentation
```

## How to Contribute

### Reporting Bugs

We use **structured issue forms** to ensure we get all the information needed to help you effectively.

1. **Search existing issues** first in the [Issues](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues)
2. If no existing issue matches, **create a new bug report**:
   - Use the **🐛 Bug Report** template
   - Fill out **all required sections** completely
   - Include your **version, platform, and configuration**
   - Provide **detailed steps to reproduce**
   - Add **relevant logs** (set log level to `debug` for more detail)

### Suggesting Enhancements

We welcome feature suggestions and improvements!

1. **Check existing feature requests** in [Issues](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues)
2. If your idea is new, **create a feature request**:
   - Use the **✨ Feature Request** template
   - Explain the **problem** you're trying to solve
   - Describe your **proposed solution**
   - Include **use cases and examples**
   - Indicate if you're willing to **help implement** it

### Improving Documentation

Documentation improvements are always welcome!

1. **Check for existing documentation issues**
2. **Create a documentation issue**:
   - Use the **📚 Documentation Issue** template
   - Specify the **type of issue** (typo, missing info, unclear instructions)
   - Indicate the **location** of the problem
   - Provide **specific suggestions** for improvement

### Pull Requests

We use a **comprehensive pull request template** to ensure quality contributions.

#### Before Submitting

1. **Update your fork** to the latest main branch
2. **Create a new branch** for your changes
3. **Make your changes** following our [Development Guidelines](#development-guidelines)
4. **Add or update tests** as needed
5. **Ensure all tests pass** locally
6. **Update documentation** as required
7. **Test your changes** thoroughly

#### Submitting Your PR

1. **Fill out the PR template completely**:

   - Describe **what changes** you're introducing
   - Explain **why** these changes are needed
   - Link to **related issues**
   - Complete all **relevant checklists**

2. **Required sections include**:
   - Type of change (bug fix, feature, docs, etc.)
   - Testing information
   - Documentation updates
   - Security considerations

#### PR Review Process

1. **Automated checks** will run (tests, builds, linting)
2. **Maintainer review** will be scheduled
3. **Address feedback** promptly and professionally
4. **Update your PR** as needed
5. **Celebrate** when your PR gets merged! 🎉

## Development Guidelines

### Coding Standards

- **Follow Go best practices** and idioms
- **Use `gofmt` and `goimports`** for consistent formatting
- **Use meaningful names** for variables, functions, and types
- **Keep functions focused** on a single responsibility
- **Write clear comments** for complex logic
- **Document exported functions** and types with Go doc comments

#### Example Code Style

```go
// ExportMovies exports user's movie data from Trakt.tv to CSV format.
// It returns the number of movies exported and any error encountered.
func ExportMovies(ctx context.Context, client *api.Client, config *Config) (int, error) {
    if client == nil {
        return 0, errors.New("client cannot be nil")
    }

    // Get user's watched movies
    movies, err := client.GetWatchedMovies(ctx)
    if err != nil {
        return 0, fmt.Errorf("failed to fetch watched movies: %w", err)
    }

    return len(movies), nil
}
```

### Testing

We maintain **high test coverage** across the codebase.

#### Writing Tests

- **Write unit tests** for new functionality
- **Use table-driven tests** for multiple scenarios
- **Mock external dependencies** (API calls, file system)
- **Test error conditions** and edge cases
- **Include integration tests** for critical workflows

#### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test -v ./pkg/api/

# Run tests with race detection
go test -race ./...

# Generate coverage report
./scripts/coverage.sh
```

#### Test Examples

```go
func TestExportMovies(t *testing.T) {
    tests := []struct {
        name     string
        client   *api.Client
        want     int
        wantErr  bool
    }{
        {
            name:    "nil client",
            client:  nil,
            want:    0,
            wantErr: true,
        },
        // Add more test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ExportMovies(context.Background(), tt.client, nil)
            if (err != nil) != tt.wantErr {
                t.Errorf("ExportMovies() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ExportMovies() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Documentation

- **Update README.md** for new features or changed workflows
- **Update code documentation** for public APIs
- **Add configuration examples** for new options
- **Update CLI help text** for new commands or flags
- **Consider Wiki updates** for complex features

### Internationalization

We support multiple languages. When adding user-facing text:

1. **Add to translation files** in `locales/`
2. **Use translation keys** instead of hardcoded strings
3. **Test with different languages** if possible
4. **Update translation documentation**

```go
// Good: Use translation key
logger.Info(i18n.Get("export.started"))

// Bad: Hardcoded string
logger.Info("Export started")
```

## Release Process

The project follows [Semantic Versioning](https://semver.org/):

- **Major versions** (v3.0.0): Breaking changes
- **Minor versions** (v2.1.0): New features, backward compatible
- **Patch versions** (v2.0.1): Bug fixes, backward compatible

### Release Workflow

1. **Version bump** in relevant files
2. **Update CHANGELOG.md** with new version
3. **Create release PR** with all changes
4. **Automated testing** runs on all platforms
5. **Manual review** and approval
6. **Merge and tag** creates automated release
7. **Docker images** are built and published
8. **GitHub release** with binaries is created

## Community

### Getting Help

- **📖 Documentation**: Check our [Wiki](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki)
- **❓ Questions**: Use the **Question & Support** issue template
- **💬 Discussions**: Join [GitHub Discussions](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/discussions)
- **🐳 Docker**: See [Docker Hub](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)

### Communication Guidelines

- **Be respectful** and professional
- **Follow the Code of Conduct**
- **Provide context** when asking questions
- **Search before posting** to avoid duplicates
- **Use templates** for structured communication

### Recognition

Contributors are recognized in:

- **GitHub Contributors** section
- **Release notes** for significant contributions
- **Special thanks** in documentation
- **Community highlights** for ongoing support

## License

By contributing, you agree that your contributions will be licensed under the project's [MIT License](LICENSE).

### Copyright

- **Original work** by Thierry Beugnet (u2pitchjami)
- **Current maintainer**: JohanDevl
- **Contributors** retain copyright on their contributions
- **Project license** covers the combined work

---

## Quick Start for Contributors

Ready to contribute? Here's the fast track:

1. **🍴 Fork the repo** and clone locally
2. **🔧 Set up Go 1.22+** and install dependencies
3. **🌿 Create a feature branch** from main
4. **💻 Make your changes** following our guidelines
5. **🧪 Add tests** and ensure they pass
6. **📝 Update docs** as needed
7. **📤 Submit a PR** using our template
8. **🎉 Celebrate** your contribution!

**Thank you for helping make Export Trakt 4 Letterboxd better for everyone!** 🚀
