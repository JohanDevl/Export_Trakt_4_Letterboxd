# Export Trakt 4 Letterboxd

Go application that exports Trakt.tv movie data to Letterboxd-compatible CSV files.

## Commands

### Build & Test
- `go build -o export_trakt ./cmd/export_trakt/` - Dev build
- `go test ./...` - All tests
- `./scripts/coverage.sh` - Coverage check (enforces 70% minimum)
- `go test -bench=. ./pkg/performance/` - Benchmarks

### Docker
- `docker compose --profile dev --profile run-all up --build` - Dev build
- `docker compose --profile run-all up` - Production
- `docker compose --profile schedule-6h up -d` - Scheduled export (6h)

### CLI
- `./export_trakt auth` - OAuth setup (required first use)
- `./export_trakt --run --export all --mode complete` - Full export
- `./export_trakt --run --export watched --history-mode individual` - Individual history
- `./export_trakt --run --export watched --history-mode aggregated` - Aggregated (default)
- `./export_trakt --schedule "0 */6 * * *" --export all --mode complete` - Scheduled

## Important Files
- `config/config.example.toml` - Configuration template (config priority: CLI flags > env vars > config.toml)
- `config/performance.toml` - Performance tuning settings
- `locales/` - i18n translations (EN, FR, DE, ES)

## Export Modes
- **Aggregated** (default): One CSV row per unique movie, most recent watch date
- **Individual**: One CSV row per viewing event, chronological rewatch tracking (first watch = no rewatch, subsequent = rewatch)

## Rules
- **ALWAYS** use structured logging with i18n message keys: `log.Info("key.name", map[string]interface{}{...})`
- **ALWAYS** use dependency injection: `NewClient(cfg *config.Config, log logger.Logger)`
- **ALWAYS** wrap errors with context: `fmt.Errorf("failed to ...: %w", err)`
- CI enforces 57% minimum coverage; local script enforces 70%
- Docker images: Docker Hub (`johandevl/export-trakt-4-letterboxd`) + GHCR (`ghcr.io/johandevl/export_trakt_4_letterboxd`)
