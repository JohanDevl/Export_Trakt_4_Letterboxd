package config

import (
	"github.com/BurntSushi/toml"
)

// Config holds all configuration settings
type Config struct {
	Trakt     TraktConfig     `toml:"trakt"`
	Letterboxd LetterboxdConfig `toml:"letterboxd"`
	Export    ExportConfig    `toml:"export"`
	Logging   LoggingConfig   `toml:"logging"`
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

// LoadConfig reads the config file and returns a Config struct
func LoadConfig(path string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
} 