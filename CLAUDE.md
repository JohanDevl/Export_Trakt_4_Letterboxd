# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Export Trakt 4 Letterboxd is a Go-based application that exports movie data from Trakt.tv to Letterboxd-compatible CSV files. The application features enterprise-grade architecture with performance optimization, comprehensive security, monitoring, and internationalization support.

## Core Development Commands

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run coverage script (enforces 70% minimum coverage)
./scripts/coverage.sh

# Run benchmarks
go test -bench=. ./pkg/performance/
```

### Building
```bash
# Development build
go build -o export_trakt ./cmd/export_trakt/

# Production build with optimizations
go build -ldflags "-w -s" -o export_trakt ./cmd/export_trakt/

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o export_trakt_linux ./cmd/export_trakt/
GOOS=windows GOARCH=amd64 go build -o export_trakt_windows.exe ./cmd/export_trakt/
GOOS=darwin GOARCH=amd64 go build -o export_trakt_darwin ./cmd/export_trakt/
```

### Docker Development
```bash
# Build Docker image
docker build -t export-trakt-dev .

# Run with Docker Compose profiles
docker compose --profile dev --profile run-all up --build      # Development build
docker compose --profile run-all up                           # Production image
docker compose --profile schedule-6h up -d                    # Production scheduler
docker compose --profile dev --profile schedule-test up -d --build  # Test scheduler

# Use published images
docker pull johandevl/export-trakt-4-letterboxd:latest         # Latest stable version
docker pull johandevl/export-trakt-4-letterboxd:develop        # Development version
docker pull johandevl/export-trakt-4-letterboxd:PR-123         # Test specific PR
docker pull johandevl/export-trakt-4-letterboxd:v1.2.3         # Specific version
```

### Application Usage
```bash
# OAuth Authentication (required for first use)
./export_trakt auth

# Check token status
./export_trakt token-status

# One-time export (default aggregated mode)
./export_trakt --run --export all --mode complete

# Individual watch history export (all viewing events)
./export_trakt --run --export watched --history-mode individual

# Aggregated export (one entry per movie, original behavior)
./export_trakt --run --export watched --history-mode aggregated

# Scheduled export
./export_trakt --schedule "0 */6 * * *" --export all --mode complete

# Validate configuration
./export_trakt validate

# Security validation
./export_trakt --validate-security
```

## Architecture Overview

### Core Packages Structure
- **cmd/export_trakt/** - Main application entry point with CLI handling
- **pkg/api/** - Trakt.tv API client with optimized HTTP client and retry logic
- **pkg/auth/** - OAuth 2.0 authentication and token management
- **pkg/config/** - Configuration management with TOML support
- **pkg/export/** - CSV export functionality for Letterboxd format
- **pkg/scheduler/** - Cron-based scheduling system
- **pkg/logger/** - Structured logging with i18n support
- **pkg/security/** - Security features including encryption, audit logging, keyring management
- **pkg/performance/** - Performance optimization features (worker pools, LRU cache, streaming)
- **pkg/monitoring/** - Observability with Prometheus metrics and OpenTelemetry tracing
- **pkg/retry/** - Resilient API calls with circuit breaker and exponential backoff
- **pkg/i18n/** - Internationalization support (English, French, German, Spanish)

### Key Architectural Patterns

**Performance Layer**: The application uses a 3-tier performance optimization:
1. **Worker Pool System** (`pkg/performance/pool/`) - Concurrent processing of API requests
2. **LRU Cache** (`pkg/performance/cache/`) - Caches API responses to reduce calls by 70-90%
3. **Streaming Processor** (`pkg/streaming/`) - Memory-efficient processing of large datasets

**Security Layer**: Enterprise-grade security with:
- **OAuth 2.0 Authentication** (`pkg/auth/`) - Modern OAuth 2.0 authentication with automatic token refresh
- **Keyring Integration** (`pkg/security/keyring/`) - Secure credential storage across multiple backends
- **Encryption** (`pkg/security/encryption/`) - AES encryption for sensitive data
- **Audit Logging** (`pkg/security/audit/`) - Comprehensive security event logging
- **Rate Limiting** (`pkg/security/ratelimit.go`) - API rate limiting protection

**Resilience Layer**: Built-in resilience patterns:
- **Circuit Breaker** (`pkg/retry/circuit/`) - Prevents cascading failures
- **Exponential Backoff** (`pkg/retry/backoff/`) - Smart retry logic
- **Checkpoints** (`pkg/resilience/checkpoints/`) - Resume interrupted operations

**Monitoring Layer**: Full observability stack:
- **Prometheus Metrics** (`pkg/monitoring/metrics/`) - Application metrics
- **OpenTelemetry Tracing** (`pkg/monitoring/tracing/`) - Distributed tracing
- **Health Checks** (`pkg/monitoring/health/`) - Application health monitoring
- **Alerting** (`pkg/monitoring/alerts/`) - Alert management

## Export Modes

The application supports two distinct export modes for watched movies:

### Aggregated Mode (Default)
- **Behavior**: One entry per unique movie with the most recent watch date
- **Use Case**: Standard export compatible with original Letterboxd import expectations
- **Configuration**: `history_mode = "aggregated"` or `--history-mode aggregated`
- **Output**: Single CSV row per movie, rewatch flag based on total play count

### Individual History Mode (New)
- **Behavior**: One entry per viewing event with complete watch history
- **Use Case**: Detailed tracking of all viewing dates and proper rewatch sequences
- **Configuration**: `history_mode = "individual"` or `--history-mode individual`
- **Output**: Multiple CSV rows per movie (one per viewing), chronological rewatch tracking
- **Data Source**: Uses `/sync/history/movies` API endpoint with 'watch' and 'scrobble' actions
- **Benefits**: 
  - Complete viewing history preservation
  - Accurate rewatch tracking (first watch = false, subsequent = true)
  - Chronological sorting (most recent first)
  - Same rating applied consistently across all viewings of a movie

**Example Individual Mode Output:**
```csv
Title,Year,WatchedDate,Rating10,imdbID,tmdbID,Rewatch
Cars,2006,2025-07-10,7,tt0317219,920,true
Cars,2006,2024-12-01,7,tt0317219,920,false
```

## Configuration Management

Configuration is handled through TOML files with the following priority:
1. Command-line flags
2. Environment variables
3. `config/config.toml` file
4. `config/config.example.toml` defaults

Key configuration files:
- `config/config.example.toml` - Complete configuration template with security features
- `config/performance.toml` - Performance optimization settings

## Security Considerations

The application implements multiple security layers:
- **OAuth 2.0 Authentication**: Modern authentication flow with PKCE and automatic token refresh
- **Credential Management**: Uses system keyring, environment variables, or encrypted files
- **Multi-Backend Token Storage**: Supports system keyring, file encryption, and environment variables
- **HTTPS Enforcement**: Configurable requirement for HTTPS-only communications
- **Input Validation**: Comprehensive sanitization and validation
- **Audit Logging**: Detailed security event logging with configurable retention
- **File Permissions**: Secure file creation and access controls

## Testing Strategy

The codebase maintains high test coverage (78%+) with:
- **Unit Tests**: All major packages have comprehensive test coverage
- **Integration Tests**: End-to-end workflow testing
- **Benchmark Tests**: Performance testing for critical paths
- **Security Tests**: Security feature validation

Critical test files:
- `pkg/performance/benchmarks_test.go` - Performance benchmarks
- `pkg/security/manager_test.go` - Security feature tests
- `pkg/api/trakt_test.go` - API client tests

## Performance Characteristics

Recent performance optimizations deliver:
- **10x throughput improvement** via worker pool system
- **70-90% API call reduction** through LRU caching
- **80% memory reduction** with streaming processing
- **Sub-second response times** for most operations

### Web Interface Performance Optimizations (2025-07-29)

The Export page has been significantly optimized to handle hundreds of export folders efficiently:

**Key Improvements:**
- **Intelligent Caching**: 5-minute in-memory cache eliminates redundant filesystem scans
- **Lazy Loading**: Prioritizes recent exports (30 days) and loads older ones only if needed  
- **Smart CSV Record Counting**: Uses file size estimation for large files instead of reading entire contents
- **Optimized Scanning**: Limits older export scans to 100 items to prevent excessive latency
- **Efficient Sorting**: Replaced bubble sort with `sort.Slice` for better performance

**Performance Impact:**
- **Page Load Time**: Reduced from ~10s to <1s for typical usage patterns
- **Memory Usage**: Minimal increase due to lightweight caching of metadata only
- **I/O Operations**: Dramatically reduced through intelligent estimation and caching
- **User Experience**: Responsive interface even with hundreds of export folders

**Implementation Details:**
- Cache TTL: 5 minutes (configurable)
- Recent exports window: 30 days (prioritized loading)  
- CSV estimation: ~80 characters per line for large files
- Fallback: Precise counting for files < 1MB

## Internationalization

The application supports multiple languages through the `pkg/i18n` package:
- Message keys are used throughout the codebase
- Translation files are located in `locales/` directory
- Supported languages: English, French, German, Spanish

## Error Handling

Comprehensive error handling system:
- **Error Types** (`pkg/errors/types/`) - Structured error definitions
- **Error Handlers** (`pkg/errors/handlers/`) - Centralized error handling
- **Error Recovery** (`pkg/errors/recovery/`) - Panic recovery mechanisms
- **Validation** (`pkg/errors/validation/`) - Input validation with detailed error messages

## Development Workflow

1. **Make Changes**: Edit code following existing patterns
2. **Run Tests**: Execute `go test ./...` to ensure no regressions
3. **Check Coverage**: Use `./scripts/coverage.sh` to verify coverage meets 70% threshold
4. **Build**: Use `go build -o export_trakt ./cmd/export_trakt/`
5. **Test Integration**: Use Docker Compose profiles for integration testing
6. **Security Check**: Run `./export_trakt --validate-security` for security validation

## CI/CD Pipeline

The project uses GitHub Actions with:
- **Go Tests** (.github/workflows/go-tests.yml) - Testing with 57% minimum coverage
- **Docker Build** (.github/workflows/docker-build.yml) - Multi-platform container builds with intelligent tagging
- **Docker Cleanup** (.github/workflows/docker-cleanup.yml) - Automatic cleanup of obsolete images
- **Auto Tag** (.github/workflows/auto-tag.yml) - Automatic semantic versioning on PR merge
- **Security Scan** (.github/workflows/security-scan.yml) - gosec, trivy, and dependency scanning
- **Release** (.github/workflows/release.yml) - Automated release creation

### Docker Image Management

The project implements an intelligent Docker image management system with semantic versioning, automatic cleanup, and monitoring. All Docker images are published to both Docker Hub (`johandevl/export-trakt-4-letterboxd`) and GitHub Container Registry (`ghcr.io/johandevl/export_trakt_4_letterboxd`).

#### Docker Tag Strategy

**Release Tags (Stable):**
- `latest` - Points to the latest stable release (v*.*.* tags, excluding beta/rc)
- `stable` - Alias for `latest`, more explicit for users
- `v2.0.16` - Specific semantic version tags for production use

**Branch Tags (Development):**
- `main` - Latest code from main branch (production-ready but not tagged)
- `develop` - Latest code from develop branch (development/testing)

**Testing Tags:**
- `PR-123` - Specific pull request builds for testing
- `beta` - Pre-release beta versions (v*.*.*-beta*)
- `rc` - Release candidate versions (v*.*.*-rc*)

#### Automated Workflows

**Auto-Tagging Workflow** (.github/workflows/auto-tag.yml):
- Triggers on PR merge to main branch
- Creates semantic version tags automatically (v2.0.x)
- Uses PAT_TOKEN to trigger downstream Docker builds
- Creates GitHub releases automatically

**CI/CD Pipeline** (.github/workflows/ci-cd.yml):
- Builds multi-platform images (linux/amd64, linux/arm64, linux/arm/v7)
- Applies tags based on trigger type (git tag, branch push, PR)
- Publishes to Docker Hub and GitHub Container Registry
- Runs security scans with Trivy

**Docker Tag Monitor** (.github/workflows/docker-tag-monitor.yml):
- Runs daily at 6 AM UTC to check for missing Docker images
- Automatically triggers CI/CD for missing tags
- Creates GitHub issues for tracking problems
- Ensures every git tag has corresponding Docker image

#### Usage Examples

```bash
# Production (stable releases)
docker pull johandevl/export-trakt-4-letterboxd:latest
docker pull johandevl/export-trakt-4-letterboxd:stable
docker pull johandevl/export-trakt-4-letterboxd:v2.0.16

# Development
docker pull johandevl/export-trakt-4-letterboxd:develop
docker pull johandevl/export-trakt-4-letterboxd:main

# Testing specific PR
docker pull johandevl/export-trakt-4-letterboxd:PR-123

# Pre-releases
docker pull johandevl/export-trakt-4-letterboxd:beta
docker pull johandevl/export-trakt-4-letterboxd:rc
```

#### Key Features

- **Automatic Tag Creation**: Tags are created automatically on PR merge
- **Guaranteed Docker Images**: Monitor ensures every tag gets a Docker image
- **Multi-Registry Publishing**: Images available on Docker Hub and GitHub
- **Security Scanning**: All images are scanned for vulnerabilities
- **Automatic Cleanup**: Obsolete images are cleaned up automatically

For comprehensive Docker documentation including compose profiles and deployment examples, see the [Docker Wiki](https://github.com/JohanDevl/Export_Trakt_4_Letterboxd/wiki/Docker).

## Common Patterns

**Structured Logging**: All log messages use structured format with message keys:
```go
log.Info("export.starting", map[string]interface{}{
    "export_type": exportType,
    "timestamp": time.Now(),
})
```

**Configuration Access**: Use dependency injection for configuration:
```go
func NewClient(cfg *config.Config, log logger.Logger) *Client {
    return &Client{config: cfg, logger: log}
}
```

**Error Handling**: Use structured error types with context:
```go
return fmt.Errorf("failed to export movies: %w", err)
```

**Performance Optimization**: Leverage worker pools for concurrent operations:
```go
pool := workerPool.New(cfg.Performance.WorkerCount)
defer pool.Close()
```

When making changes, always follow existing architectural patterns and maintain the security-first approach. The codebase emphasizes reliability, performance, and security over simplicity.