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
```

### Application Usage
```bash
# One-time export
./export_trakt --run --export all --mode complete

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
- **Keyring Integration** (`pkg/security/keyring/`) - Secure credential storage
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
- **Credential Management**: Uses system keyring, environment variables, or encrypted files
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
- **Docker Build** (.github/workflows/docker-build.yml) - Multi-platform container builds
- **Security Scan** (.github/workflows/security-scan.yml) - gosec, trivy, and dependency scanning
- **Release** (.github/workflows/release.yml) - Automated release creation

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