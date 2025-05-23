# Export Trakt 4 Letterboxd

[![GitHub release](https://img.shields.io/github/v/release/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/releases)
[![GitHub stars](https://img.shields.io/github/stars/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/stargazers)
[![GitHub issues](https://img.shields.io/github/issues/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/issues)
[![GitHub license](https://img.shields.io/github/license/JohanDevl/Export_Trakt_4_Letterboxd)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/blob/main/LICENSE)
[![Go Build](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/go-build.yml/badge.svg)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/go-build.yml)
[![Docker Build](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-build.yml/badge.svg)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/actions/workflows/docker-build.yml)
[![Docker Package](https://img.shields.io/badge/GitHub%20Packages-ghcr.io-blue?logo=docker)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkgs/container/export_trakt_4_letterboxd)
[![Docker Hub](https://img.shields.io/docker/v/johandevl/export-trakt-4-letterboxd?logo=docker&label=Docker%20Hub)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)
[![Docker Image Size](https://img.shields.io/docker/image-size/johandevl/export-trakt-4-letterboxd/latest?logo=docker&label=Image%20Size)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)
[![Docker Pulls](https://img.shields.io/docker/pulls/johandevl/export-trakt-4-letterboxd?logo=docker&label=Pulls)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd)
[![Platforms](https://img.shields.io/badge/platforms-amd64%20|%20arm64%20|%20armv7-lightgrey?logo=docker)](https://hub.docker.com/r/johandevl/export-trakt-4-letterboxd/tags)
[![Code Coverage](https://img.shields.io/badge/coverage-78%25-brightgreen)](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/blob/main/coverage.html)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org/)
[![Trakt.tv](https://img.shields.io/badge/Trakt.tv-ED1C24?logo=trakt&logoColor=white)](https://trakt.tv)
[![Letterboxd](https://img.shields.io/badge/Letterboxd-00D735?logo=letterboxd&logoColor=white)](https://letterboxd.com)

> **🎬 Seamlessly export your Trakt.tv movie data to Letterboxd-compatible CSV files**

A robust, modern Go application that enables you to migrate your Trakt.tv movie ratings, watchlist, and viewing history to Letterboxd with ease. Built with enterprise-grade reliability and featuring comprehensive internationalization support.

## ✨ Key Features

- **🎯 Complete Data Export**: Export ratings, watchlist, watch history, and collections
- **📊 Letterboxd Optimized**: Native support for Letterboxd's import format
- **🔄 Automatic Scheduling**: Set up cron-based automated exports
- **🌍 Internationalization**: Full i18n support (English, French, German, Spanish)
- **🐳 Docker Ready**: Multi-platform Docker images (amd64, arm64, armv7)
- **📈 High Performance**: Built with Go 1.22+ for optimal speed and reliability
- **🔒 Security First**: Token-based authentication with secure credential handling
- **📝 Comprehensive Logging**: Detailed logging with configurable levels
- **🧪 Well Tested**: 78%+ test coverage across all core packages
- **⚙️ Highly Configurable**: Extensive configuration options via TOML

## 🚀 Quick Start

### Option 1: Docker (Recommended)

**Simple one-time export:**

```bash
# Pull and run directly from Docker Hub
docker run --rm -it \
  -v $(pwd)/config:/app/config \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/exports:/app/exports \
  johandevl/export-trakt-4-letterboxd:latest
```

**Using Docker Compose:**

```bash
# Clone the repository
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd

# Interactive setup (first time)
docker compose --profile setup up

# Run export
docker compose --profile run-all up

# For scheduled exports (every 6 hours)
docker compose --profile schedule-6h up -d
```

### Option 2: Local Installation

**Prerequisites:**

- Go 1.22 or higher
- Git

**Installation:**

```bash
# Clone and build
git clone https://github.com/JohanDevl/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd
go build -o export_trakt ./cmd/export_trakt/

# Copy and configure
cp config/config.example.toml config/config.toml
# Edit config/config.toml with your Trakt.tv credentials

# Run
./export_trakt --run --export all --mode complete
```

## ⚙️ Configuration

### 1. Trakt.tv API Setup

1. Go to [Trakt.tv API Applications](https://trakt.tv/oauth/applications)
2. Create a new application
3. Note your `Client ID` and `Client Secret`
4. Generate an access token

### 2. Configuration File

Copy the example configuration and customize:

```bash
cp config/config.example.toml config/config.toml
```

**Essential settings:**

```toml
[trakt]
client_id = "YOUR_CLIENT_ID"
client_secret = "YOUR_CLIENT_SECRET"
access_token = "YOUR_ACCESS_TOKEN"
extended_info = "letterboxd"  # For optimal Letterboxd compatibility

[export]
format = "csv"
date_format = "2006-01-02"

[logging]
level = "info"
file = "logs/export.log"
```

## 🎯 Usage Examples

### Command Line Interface

```bash
# Complete export of all data
./export_trakt --run --export all --mode complete

# Export only watched movies
./export_trakt --run --export watched --mode normal

# Validate configuration
./export_trakt validate

# Schedule automated exports (every 6 hours)
./export_trakt --schedule "0 */6 * * *" --export all --mode complete
```

### Docker Compose Profiles

```bash
# PRODUCTION WORKFLOWS

# Daily automated exports at 2:30 AM
docker compose --profile schedule-daily up -d

# Every 6 hours (recommended for active users)
docker compose --profile schedule-6h up -d

# One-time complete export
docker compose --profile run-all up

# DEVELOPMENT/TESTING

# Test with local build (every 2 minutes)
docker compose --profile dev --profile schedule-test up -d --build

# Run watched movies only (testing)
docker compose --profile dev --profile run-watched up --build

# Interactive setup
docker compose --profile dev --profile setup up --build
```

### Environment Variables

```bash
# Custom scheduling
SCHEDULE="0 4 * * *" docker compose --profile schedule-custom up -d

# Different export types
EXPORT_TYPE="watched" EXPORT_MODE="normal" docker compose --profile schedule-custom up -d

# Timezone configuration
TZ="America/New_York" docker compose --profile schedule-daily up -d
```

## 📁 Project Structure

```
Export_Trakt_4_Letterboxd/
├── cmd/export_trakt/           # 🎯 Main application entry point
├── pkg/                        # 📦 Core packages
│   ├── api/                    # 🌐 Trakt.tv API client
│   ├── config/                 # ⚙️ Configuration management
│   ├── export/                 # 📊 Export functionality
│   ├── i18n/                   # 🌍 Internationalization
│   ├── logger/                 # 📝 Logging system
│   └── scheduler/              # ⏰ Cron scheduler
├── internal/                   # 🔒 Private application code
│   ├── models/                 # 🗂️ Data models
│   └── utils/                  # 🛠️ Private utilities
├── locales/                    # 🗣️ Translation files
├── config/                     # 📋 Configuration files
├── scripts/                    # 🚀 Build and utility scripts
├── .github/workflows/          # 🤖 CI/CD workflows
└── docker/                     # 🐳 Docker configurations
```

## 🔄 Export Modes

| Mode       | Description                      | Use Case                           |
| ---------- | -------------------------------- | ---------------------------------- |
| `normal`   | Basic export with essential data | Quick exports, testing             |
| `complete` | Full export with all metadata    | Production use, complete migration |
| `initial`  | First-time setup export          | Initial Letterboxd migration       |

## 📊 Export Types

| Type         | Content                        | Letterboxd Import     |
| ------------ | ------------------------------ | --------------------- |
| `watched`    | Rated movies and viewing dates | ✅ Ratings & History  |
| `watchlist`  | Movies in your watchlist       | ✅ Watchlist          |
| `collection` | Collected movies               | ✅ Custom Lists       |
| `shows`      | TV show data                   | ⚠️ Limited support    |
| `all`        | Everything above               | ✅ Complete migration |

## 🌍 Internationalization

Supported languages:

- 🇺🇸 English (`en`) - Default
- 🇫🇷 French (`fr`)
- 🇩🇪 German (`de`)
- 🇪🇸 Spanish (`es`)

Configure in `config.toml`:

```toml
[i18n]
language = "fr"  # Change to your preferred language
```

## 📅 Scheduling

### Cron Schedule Examples

```bash
# Every 5 minutes (testing)
*/5 * * * *

# Every hour
0 * * * *

# Daily at 4 AM
0 4 * * *

# Weekly on Sunday at 4 AM
0 4 * * 0

# Monthly on the 1st at 4 AM
0 4 1 * *
```

### Production Recommendations

- **Active users**: Every 6 hours (`0 */6 * * *`)
- **Regular users**: Daily (`0 4 * * *`)
- **Occasional users**: Weekly (`0 4 * * 0`)

## 🧪 Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run coverage script
./scripts/coverage.sh
```

### Test Coverage

| Package              | Coverage |
| -------------------- | -------- |
| API Client           | 73.3%    |
| Config Management    | 85.4%    |
| Export Functionality | 78.3%    |
| Internationalization | 81.6%    |
| Logging System       | 97.7%    |
| **Overall**          | **78%+** |

## 🐳 Docker

### Available Images

- **GitHub Container Registry**: `ghcr.io/johandevl/export_trakt_4_letterboxd`
- **Docker Hub**: `johandevl/export-trakt-4-letterboxd`

### Supported Platforms

- `linux/amd64` (x86_64)
- `linux/arm64` (ARM 64-bit)
- `linux/arm/v7` (ARM 32-bit)

### Tags

- `latest` - Latest stable release
- `v2.x.x` - Specific version tags
- `main` - Latest development build

## 🛠️ Development

### Building from Source

```bash
# Development build
go build -o export_trakt ./cmd/export_trakt/

# Production build with optimization
go build -ldflags "-w -s" -o export_trakt ./cmd/export_trakt/

# Cross-compilation
GOOS=linux GOARCH=amd64 go build -o export_trakt-linux ./cmd/export_trakt/
```

### Docker Development

```bash
# Build local image
docker build -t export-trakt-dev .

# Run development container
docker compose --profile dev --profile run-all up --build
```

## 🔧 Troubleshooting

### Common Issues

**Authentication Errors**

```bash
# Check your credentials
./export_trakt validate

# Test API connection
curl -H "Authorization: Bearer YOUR_TOKEN" https://api.trakt.tv/users/me
```

**Docker Issues**

```bash
# Check logs
docker compose logs -f

# Reset everything
docker compose down -v
docker system prune -f
```

**Export Problems**

```bash
# Enable debug logging
# In config.toml: level = "debug"

# Check export directory permissions
ls -la exports/

# Verify Trakt.tv profile is public
```

### Log Analysis

```bash
# View recent logs
tail -f logs/export.log

# Search for errors
grep "ERROR" logs/export.log

# Monitor in real-time
docker compose --profile schedule-6h logs -f | grep ERROR
```

## 📖 Documentation

Complete documentation available in our [Wiki](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki):

- [📥 Installation Guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Installation)
- [⌨️ CLI Reference](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/CLI-Reference)
- [📊 Export Features](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Export-Features)
- [🔑 Trakt API Guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Trakt-API-Guide)
- [🌍 Internationalization](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Internationalization)
- [🔄 Migration Guide](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Migration-Guide)
- [🧪 Testing](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Testing)
- [🤖 CI/CD](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/CI-CD)

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/Export_Trakt_4_Letterboxd.git
cd Export_Trakt_4_Letterboxd

# Create feature branch
git checkout -b feature/amazing-feature

# Make changes and test
go test ./...
docker compose --profile dev --profile run-all up --build

# Commit and push
git commit -m "feat: add amazing feature"
git push origin feature/amazing-feature
```

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgements

This project is based on the original work by [u2pitchjami](https://github.com/u2pitchjami/Export_Trakt_4_Letterboxd). Special thanks to the original author for creating the foundation that made this project possible.

**Original repository**: https://github.com/u2pitchjami/Export_Trakt_4_Letterboxd

## 👨‍💻 Author

**JohanDevl**

- 🐦 Twitter: [@0xUta](https://twitter.com/0xUta)
- 🐙 GitHub: [@JohanDevl](https://github.com/JohanDevl)
- 💼 LinkedIn: [@johan-devlaminck](https://linkedin.com/in/johan-devlaminck)

## 🌟 Support

If you find this project helpful, please consider:

- ⭐ Starring the repository
- 🐛 Reporting bugs or requesting features
- 🤝 Contributing code or documentation
- 💬 Sharing with the community

---

<div align="center">

**Made with ❤️ for the movie community**

[Trakt.tv](https://trakt.tv) • [Letterboxd](https://letterboxd.com) • [Documentation](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki)

</div>
