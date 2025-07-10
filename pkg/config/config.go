package config

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/security"
)

// Config holds all configuration settings
type Config struct {
	Trakt     TraktConfig     `toml:"trakt"`
	Letterboxd LetterboxdConfig `toml:"letterboxd"`
	Export    ExportConfig    `toml:"export"`
	Logging   LoggingConfig   `toml:"logging"`
	I18n      I18nConfig      `toml:"i18n"`
	Security  security.Config `toml:"security"`
	Auth      AuthConfig      `toml:"auth"`
}

// TraktConfig holds Trakt.tv API configuration
type TraktConfig struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	AccessToken  string `toml:"access_token"`
	APIBaseURL   string `toml:"api_base_url"`
	ExtendedInfo string `toml:"extended_info"`
}

// LetterboxdConfig holds Letterboxd export configuration
type LetterboxdConfig struct {
	ExportDir                string `toml:"export_dir"`
	WatchedFilename          string `toml:"watched_filename"`
	CollectionFilename       string `toml:"collection_filename"`
	ShowsFilename            string `toml:"shows_filename"`
	RatingsFilename          string `toml:"ratings_filename"`
	WatchlistFilename        string `toml:"watchlist_filename"`
	LetterboxdImportFilename string `toml:"letterboxd_import_filename"`
}

// ExportConfig holds export settings
type ExportConfig struct {
	Format     string `toml:"format"`
	DateFormat string `toml:"date_format"`
	Timezone   string `toml:"timezone"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level string `toml:"level"`
	File  string `toml:"file"`
}

// I18nConfig holds internationalization settings
type I18nConfig struct {
	DefaultLanguage string `toml:"default_language"`
	Language       string `toml:"language"`
	LocalesDir    string `toml:"locales_dir"`
}

// AuthConfig holds OAuth authentication settings
type AuthConfig struct {
	RedirectURI    string `toml:"redirect_uri"`
	CallbackPort   int    `toml:"callback_port"`
	UseOAuth       bool   `toml:"use_oauth"`
	AutoRefresh    bool   `toml:"auto_refresh"`
}

// LoadConfig reads the config file and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Set defaults before validation
	config.SetDefaults()

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if err := c.Trakt.Validate(); err != nil {
		return fmt.Errorf("trakt config: %w", err)
	}

	if err := c.Letterboxd.Validate(); err != nil {
		return fmt.Errorf("letterboxd config: %w", err)
	}

	if err := c.Export.Validate(); err != nil {
		return fmt.Errorf("export config: %w", err)
	}

	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config: %w", err)
	}

	if err := c.I18n.Validate(); err != nil {
		return fmt.Errorf("i18n config: %w", err)
	}

	if err := c.Security.Validate(); err != nil {
		return fmt.Errorf("security config: %w", err)
	}

	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("auth config: %w", err)
	}

	return nil
}

// Validate checks if the Trakt configuration is valid
func (c *TraktConfig) Validate() error {
	if c.APIBaseURL == "" {
		return fmt.Errorf("api_base_url is required")
	}
	return nil
}

// Validate checks if the Letterboxd configuration is valid
func (c *LetterboxdConfig) Validate() error {
	if c.ExportDir == "" {
		return fmt.Errorf("export_dir is required")
	}
	return nil
}

// Validate checks if the Export configuration is valid
func (c *ExportConfig) Validate() error {
	if c.Format == "" {
		return fmt.Errorf("format is required")
	}
	if c.DateFormat == "" {
		return fmt.Errorf("date_format is required")
	}
	// If timezone is empty, we'll use UTC as default, so no error needed
	return nil
}

// Validate checks if the Logging configuration is valid
func (c *LoggingConfig) Validate() error {
	if c.Level == "" {
		return fmt.Errorf("level is required")
	}
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}
	return nil
}

// Validate checks if the I18n configuration is valid
func (c *I18nConfig) Validate() error {
	if c.DefaultLanguage == "" {
		return fmt.Errorf("default_language is required")
	}
	if c.Language == "" {
		return fmt.Errorf("language is required")
	}
	if c.LocalesDir == "" {
		return fmt.Errorf("locales_dir is required")
	}
	return nil
}

// Validate checks if the Auth configuration is valid
func (c *AuthConfig) Validate() error {
	// Set defaults if not specified
	if c.RedirectURI == "" {
		c.RedirectURI = "http://localhost:8080/callback"
	}
	if c.CallbackPort == 0 {
		c.CallbackPort = 8080
	}
	
	return nil
}

// SetDefaults sets default values for the configuration
func (c *Config) SetDefaults() {
	// Trakt defaults
	if c.Trakt.APIBaseURL == "" {
		c.Trakt.APIBaseURL = "https://api.trakt.tv"
	}
	if c.Trakt.ExtendedInfo == "" {
		c.Trakt.ExtendedInfo = "full"
	}

	// Letterboxd defaults
	if c.Letterboxd.ExportDir == "" {
		c.Letterboxd.ExportDir = "./exports"
	}
	if c.Letterboxd.WatchedFilename == "" {
		c.Letterboxd.WatchedFilename = "watched.csv"
	}
	if c.Letterboxd.CollectionFilename == "" {
		c.Letterboxd.CollectionFilename = "collection.csv"
	}
	if c.Letterboxd.ShowsFilename == "" {
		c.Letterboxd.ShowsFilename = "shows.csv"
	}
	if c.Letterboxd.RatingsFilename == "" {
		c.Letterboxd.RatingsFilename = "ratings.csv"
	}
	if c.Letterboxd.WatchlistFilename == "" {
		c.Letterboxd.WatchlistFilename = "watchlist.csv"
	}
	if c.Letterboxd.LetterboxdImportFilename == "" {
		c.Letterboxd.LetterboxdImportFilename = "letterboxd_import.csv"
	}

	// Export defaults
	if c.Export.Format == "" {
		c.Export.Format = "csv"
	}
	if c.Export.DateFormat == "" {
		c.Export.DateFormat = "2006-01-02"
	}
	if c.Export.Timezone == "" {
		c.Export.Timezone = "UTC"
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.File == "" {
		c.Logging.File = "./logs/app.log"
	}

	// I18n defaults
	if c.I18n.DefaultLanguage == "" {
		c.I18n.DefaultLanguage = "en"
	}
	if c.I18n.Language == "" {
		c.I18n.Language = "en"
	}
	if c.I18n.LocalesDir == "" {
		c.I18n.LocalesDir = "./locales"
	}

	// Auth defaults
	if c.Auth.RedirectURI == "" {
		c.Auth.RedirectURI = "http://localhost:8080/callback"
	}
	if c.Auth.CallbackPort == 0 {
		c.Auth.CallbackPort = 8080
	}
	// OAuth is enabled by default
	c.Auth.UseOAuth = true
	c.Auth.AutoRefresh = true
} 