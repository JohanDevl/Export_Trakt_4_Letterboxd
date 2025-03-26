# API Documentation

This document provides an overview of the packages and APIs available in the Go implementation of Export_Trakt_4_Letterboxd.

## Package Structure

The application is organized into the following packages:

```
/cmd
  /export_trakt     - Main application entry point
/pkg
  /api             - Trakt.tv API client
  /config          - Configuration loading and validation
  /export          - Letterboxd export functionality
  /i18n            - Internationalization support
  /logger          - Logging system
```

## Config Package

The `config` package handles loading, parsing, and validating application configuration from TOML files.

### Types

#### `Config`

The main configuration structure that contains all configuration options.

```go
type Config struct {
    Trakt     TraktConfig     `toml:"trakt"`
    Letterboxd LetterboxdConfig `toml:"letterboxd"`
    Export    ExportConfig    `toml:"export"`
    Logging   LoggingConfig   `toml:"logging"`
    I18n      I18nConfig      `toml:"i18n"`
}
```

### Functions

#### `LoadConfig`

```go
func LoadConfig(path string) (*Config, error)
```

Loads and parses a TOML configuration file from the specified path.

#### `Validate`

```go
func (c *Config) Validate() error
```

Validates that the configuration meets all requirements.

## Logger Package

The `logger` package provides a flexible logging system with support for different output targets and log levels.

### Interfaces

#### `Logger`

```go
type Logger interface {
    Info(messageID string, data ...map[string]interface{})
    Infof(messageID string, data map[string]interface{})
    Error(messageID string, data ...map[string]interface{})
    Errorf(messageID string, data map[string]interface{})
    Warn(messageID string, data ...map[string]interface{})
    Warnf(messageID string, data map[string]interface{})
    Debug(messageID string, data ...map[string]interface{})
    Debugf(messageID string, data map[string]interface{})
    SetLogLevel(level string)
    SetLogFile(path string) error
    SetTranslator(t Translator)
}
```

#### `Translator`

```go
type Translator interface {
    Translate(messageID string, templateData map[string]interface{}) string
}
```

### Functions

#### `NewLogger`

```go
func NewLogger() Logger
```

Creates a new logger instance with default settings.

## API Package

The `api` package handles communication with the Trakt.tv API.

### Types

#### `Client`

The main client for interacting with the Trakt.tv API.

```go
type Client struct {
    // private fields
}
```

#### `Movie`

```go
type Movie struct {
    Movie         MovieInfo `json:"movie"`
    LastWatchedAt string    `json:"last_watched_at,omitempty"`
    Rating        float64   `json:"rating,omitempty"`
}
```

### Functions

#### `NewClient`

```go
func NewClient(cfg *config.Config, log logger.Logger) *Client
```

Creates a new Trakt API client.

#### `GetWatchedMovies`

```go
func (c *Client) GetWatchedMovies() ([]Movie, error)
```

Retrieves the list of watched movies from Trakt.tv.

## Export Package

The `export` package provides functionality for exporting data to Letterboxd format.

### Types

#### `LetterboxdExporter`

```go
type LetterboxdExporter struct {
    // private fields
}
```

### Functions

#### `NewLetterboxdExporter`

```go
func NewLetterboxdExporter(cfg *config.Config, log logger.Logger) *LetterboxdExporter
```

Creates a new Letterboxd exporter.

#### `ExportMovies`

```go
func (e *LetterboxdExporter) ExportMovies(movies []api.Movie) error
```

Exports a list of movies to Letterboxd CSV format.

## i18n Package

The `i18n` package handles internationalization and localization.

### Types

#### `Translator`

```go
type Translator struct {
    // private fields
}
```

### Functions

#### `NewTranslator`

```go
func NewTranslator(cfg *config.I18nConfig, log logger.Logger) (*Translator, error)
```

Creates a new translator instance.

#### `Translate`

```go
func (t *Translator) Translate(messageID string, templateData map[string]interface{}) string
```

Returns the translated message for the given message ID.

#### `SetLanguage`

```go
func (t *Translator) SetLanguage(lang string)
```

Changes the current language for the translator.

## Example Usage

### Loading Configuration

```go
cfg, err := config.LoadConfig("config/config.toml")
if err != nil {
    log.Fatalf("Failed to load configuration: %v", err)
}
```

### Setting Up Logging

```go
log := logger.NewLogger()
log.SetLogLevel(cfg.Logging.Level)
if cfg.Logging.File != "" {
    if err := log.SetLogFile(cfg.Logging.File); err != nil {
        log.Errorf("Failed to set log file: %v", err)
    }
}
```

### Using the API Client

```go
client := api.NewClient(cfg, log)
movies, err := client.GetWatchedMovies()
if err != nil {
    log.Errorf("Failed to get watched movies: %v", err)
    return
}
```

### Exporting Data

```go
exporter := export.NewLetterboxdExporter(cfg, log)
if err := exporter.ExportMovies(movies); err != nil {
    log.Errorf("Failed to export movies: %v", err)
    return
}
```

### Internationalization

```go
translator, err := i18n.NewTranslator(&cfg.I18n, log)
if err != nil {
    log.Errorf("Failed to initialize translator: %v", err)
    return
}

// Set translator for logger
log.SetTranslator(translator)

// Log with translation
log.Info("app.starting", map[string]interface{}{"version": "1.0.0"})

// Get translated message
message := translator.Translate("app.welcome", nil)
fmt.Println(message)
```
