package config

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

// Config holds all configuration settings
type Config struct {
	Trakt     TraktConfig     `toml:"trakt"`
	Letterboxd LetterboxdConfig `toml:"letterboxd"`
	Export    ExportConfig    `toml:"export"`
	Logging   LoggingConfig   `toml:"logging"`
	I18n      I18nConfig      `toml:"i18n"`
}

// TraktConfig holds Trakt.tv API configuration
type TraktConfig struct {
	ClientID     string `toml:"client_id"`
	ClientSecret string `toml:"client_secret"`
	AccessToken  string `toml:"access_token"`
	APIBaseURL   string `toml:"api_base_url"`
}

// LetterboxdConfig holds Letterboxd export configuration
type LetterboxdConfig struct {
	ExportDir string `toml:"export_dir"`
}

// ExportConfig holds export settings
type ExportConfig struct {
	Format     string `toml:"format"`
	DateFormat string `toml:"date_format"`
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

// LoadConfig reads the config file and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

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