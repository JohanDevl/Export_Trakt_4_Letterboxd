# CLAUDE.md - Gestion de Configuration

## Module Overview

Ce module gère la configuration centralisée de l'application via des fichiers TOML avec validation, définition de valeurs par défaut, et structure hiérarchique. Il fournit une interface unifiée pour tous les paramètres de configuration avec support des variables d'environnement et validation robuste.

## Architecture du Module

### 🏗️ Structure de Configuration Principale

#### Config (Structure Racine)
```go
type Config struct {
    Trakt      TraktConfig     `toml:"trakt"`      // Configuration API Trakt.tv
    Letterboxd LetterboxdConfig `toml:"letterboxd"` // Configuration export Letterboxd
    Export     ExportConfig    `toml:"export"`     // Paramètres d'export
    Logging    LoggingConfig   `toml:"logging"`    // Configuration logging
    I18n       I18nConfig      `toml:"i18n"`       // Internationalisation
    Security   security.Config `toml:"security"`   // Paramètres de sécurité
    Auth       AuthConfig      `toml:"auth"`       // Configuration OAuth
}
```

### 📋 Sections de Configuration

#### 1. Configuration Trakt.tv
```go
type TraktConfig struct {
    ClientID     string `toml:"client_id"`     // ID client OAuth Trakt.tv
    ClientSecret string `toml:"client_secret"` // Secret client OAuth
    AccessToken  string `toml:"access_token"`  // Token d'accès (legacy)
    APIBaseURL   string `toml:"api_base_url"`  // URL de base API
    ExtendedInfo string `toml:"extended_info"` // Niveau d'infos étendues
}
```

**Valeurs par défaut :**
- `APIBaseURL` : `"https://api.trakt.tv"`
- `ExtendedInfo` : `"full"`

**Exemple TOML :**
```toml
[trakt]
client_id = "your_trakt_client_id"
client_secret = "your_trakt_client_secret"
api_base_url = "https://api.trakt.tv"
extended_info = "full"  # full, metadata, letterboxd
```

#### 2. Configuration Export Letterboxd
```go
type LetterboxdConfig struct {
    ExportDir                string `toml:"export_dir"`                 // Répertoire d'export
    WatchedFilename          string `toml:"watched_filename"`           // Nom fichier films regardés
    CollectionFilename       string `toml:"collection_filename"`        // Nom fichier collection
    ShowsFilename            string `toml:"shows_filename"`             // Nom fichier séries
    RatingsFilename          string `toml:"ratings_filename"`           // Nom fichier notes
    WatchlistFilename        string `toml:"watchlist_filename"`         // Nom fichier watchlist
    LetterboxdImportFilename string `toml:"letterboxd_import_filename"` // Nom fichier import Letterboxd
}
```

**Valeurs par défaut :**
- `ExportDir` : `"./exports"`
- `WatchedFilename` : `"watched.csv"`
- `CollectionFilename` : `"collection.csv"`
- `ShowsFilename` : `"shows.csv"`
- `RatingsFilename` : `"ratings.csv"`
- `WatchlistFilename` : `"watchlist.csv"`
- `LetterboxdImportFilename` : `"letterboxd-import.csv"`

**Exemple TOML :**
```toml
[letterboxd]
export_dir = "./exports"
watched_filename = "watched.csv"
collection_filename = "collection.csv"
shows_filename = "shows.csv"
ratings_filename = "ratings.csv"
watchlist_filename = "watchlist.csv"
```

#### 3. Configuration Export
```go
type ExportConfig struct {
    Format      string `toml:"format"`       // Format d'export (csv)
    DateFormat  string `toml:"date_format"`  // Format des dates
    Timezone    string `toml:"timezone"`     // Fuseau horaire
    HistoryMode string `toml:"history_mode"` // Mode historique (aggregated/individual)
}
```

**Valeurs par défaut :**
- `Format` : `"csv"`
- `DateFormat` : `"2006-01-02"`
- `Timezone` : `"UTC"`
- `HistoryMode` : `"aggregated"`

**Modes d'historique :**
- **`aggregated`** : Une entrée par film (mode par défaut, compatible Letterboxd)
- **`individual`** : Une entrée par visionnage (historique complet)

**Exemple TOML :**
```toml
[export]
format = "csv"
date_format = "2006-01-02"
timezone = "Europe/Paris"
history_mode = "individual"  # aggregated ou individual
```

#### 4. Configuration Logging
```go
type LoggingConfig struct {
    Level string `toml:"level"` // Niveau de log (debug, info, warn, error)
    File  string `toml:"file"`  // Fichier de log (optionnel)
}
```

**Niveaux de log supportés :**
- `debug` : Logs détaillés pour développement
- `info` : Informations générales (par défaut)
- `warn` : Avertissements
- `error` : Erreurs uniquement

**Exemple TOML :**
```toml
[logging]
level = "info"
file = "./logs/export.log"
```

#### 5. Configuration Internationalisation
```go
type I18nConfig struct {
    DefaultLanguage string `toml:"default_language"` // Langue par défaut
    Language        string `toml:"language"`         // Langue active
    LocalesDir      string `toml:"locales_dir"`      // Répertoire des traductions
}
```

**Langues supportées :**
- `en` : Anglais (par défaut)
- `fr` : Français
- `de` : Allemand
- `es` : Espagnol

**Exemple TOML :**
```toml
[i18n]
default_language = "en"
language = "fr"
locales_dir = "./locales"
```

#### 6. Configuration Authentification
```go
type AuthConfig struct {
    RedirectURI  string `toml:"redirect_uri"`   // URI de redirection OAuth
    CallbackPort int    `toml:"callback_port"`  // Port serveur callback
    UseOAuth     bool   `toml:"use_oauth"`      // Utiliser OAuth 2.0
    AutoRefresh  bool   `toml:"auto_refresh"`   // Rafraîchissement auto des tokens
}
```

**Valeurs par défaut :**
- `RedirectURI` : `"http://localhost:8080/callback"`
- `CallbackPort` : `8080`
- `UseOAuth` : `true`
- `AutoRefresh` : `true`

**Exemple TOML :**
```toml
[auth]
redirect_uri = "http://localhost:8080/callback"
callback_port = 8080
use_oauth = true
auto_refresh = true
```

### 🔧 Chargement et Validation

#### Processus de Chargement
```go
func LoadConfig(path string) (*Config, error) {
    var config Config
    
    // 1. Décoder le fichier TOML
    if _, err := toml.DecodeFile(path, &config); err != nil {
        return nil, fmt.Errorf("failed to decode config file: %w", err)
    }
    
    // 2. Appliquer les valeurs par défaut
    config.SetDefaults()
    
    // 3. Valider la configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("invalid configuration: %w", err)
    }
    
    return &config, nil
}
```

#### Système de Validation Hiérarchique
```go
func (c *Config) Validate() error {
    // Validation de chaque section
    validationErrors := []error{
        c.Trakt.Validate(),
        c.Letterboxd.Validate(),
        c.Export.Validate(),
        c.Logging.Validate(),
        c.I18n.Validate(),
        c.Security.Validate(),
        c.Auth.Validate(),
    }
    
    // Consolidation des erreurs
    for _, err := range validationErrors {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

#### Validation Spécialisée par Section
```go
// Validation configuration export
func (c *ExportConfig) Validate() error {
    if c.Format == "" {
        return fmt.Errorf("format is required")
    }
    
    // Validation mode historique
    if c.HistoryMode != "" {
        validModes := map[string]bool{
            "aggregated": true,
            "individual": true,
        }
        if !validModes[c.HistoryMode] {
            return fmt.Errorf("invalid history_mode: %s", c.HistoryMode)
        }
    }
    
    return nil
}

// Validation configuration logging
func (c *LoggingConfig) Validate() error {
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
```

### 📁 Hiérarchie de Configuration

#### Ordre de Priorité
1. **Arguments CLI** (flags) - priorité la plus haute
2. **Variables d'environnement**
3. **Fichier de configuration** (`config/config.toml`)
4. **Valeurs par défaut** - priorité la plus basse

#### Fichiers de Configuration
```
config/
├── config.toml          # Configuration principale
├── config.example.toml  # Template avec toutes les options
├── performance.toml     # Configuration performance (optionnel)
└── credentials.enc      # Credentials chiffrés (si backend file)
```

### 🌐 Support Variables d'Environnement

#### Mapping Automatique
```bash
# Variables d'environnement supportées
TRAKT_CLIENT_ID="your_client_id"
TRAKT_CLIENT_SECRET="your_secret"
TRAKT_API_BASE_URL="https://api.trakt.tv"

EXPORT_DIR="./exports"
EXPORT_HISTORY_MODE="individual"
EXPORT_TIMEZONE="Europe/Paris"

LOG_LEVEL="debug"
LOG_FILE="./logs/export.log"

AUTH_REDIRECT_URI="http://localhost:8080/callback"
AUTH_CALLBACK_PORT="8080"
AUTH_USE_OAUTH="true"

I18N_LANGUAGE="fr"
I18N_LOCALES_DIR="./locales"
```

### 🔒 Configuration de Sécurité (Intégrée)

#### Section Sécurité
```toml
[security]
# Niveau de sécurité global
security_level = "high"  # low, medium, high

# Chiffrement des données sensibles
encryption_enabled = true
encryption_algorithm = "AES-256-GCM"

# Backend de stockage credentials
keyring_backend = "system"  # system, env, file, memory

# Application HTTPS
require_https = true
tls_min_version = "1.2"

# Audit et logging
audit_logging = true
log_sensitive_data = false

# Validation des entrées
input_sanitization = true
path_traversal_protection = true

# Rate limiting
rate_limit_enabled = true
rate_limit_requests_per_second = 100

# Configuration audit
[security.audit]
enabled = true
file_path = "./logs/audit.log"
max_file_size = "100MB"
retention_days = 90
include_sensitive = false
compress_old_logs = true
```

### 📊 Configuration Performance (Avancée)

#### Fichier performance.toml
```toml
[cache]
enabled = true
capacity = 1000
ttl = "24h"
compression = true

[worker_pool]
size = 10
buffer_size = 20
max_queue_size = 1000

[api]
timeout = "30s"
max_retries = 3
backoff_factor = 2.0
connection_pool_size = 20

[streaming]
enabled = true
buffer_size = 8192
chunk_size = 1000
```

### 🛠️ Méthodes Utilitaires

#### Valeurs par Défaut
```go
func (c *Config) SetDefaults() {
    // Trakt
    if c.Trakt.APIBaseURL == "" {
        c.Trakt.APIBaseURL = "https://api.trakt.tv"
    }
    if c.Trakt.ExtendedInfo == "" {
        c.Trakt.ExtendedInfo = "full"
    }
    
    // Export
    if c.Export.Format == "" {
        c.Export.Format = "csv"
    }
    if c.Export.DateFormat == "" {
        c.Export.DateFormat = "2006-01-02"
    }
    if c.Export.Timezone == "" {
        c.Export.Timezone = "UTC"
    }
    if c.Export.HistoryMode == "" {
        c.Export.HistoryMode = "aggregated"
    }
    
    // Logging
    if c.Logging.Level == "" {
        c.Logging.Level = "info"
    }
    
    // I18n
    if c.I18n.DefaultLanguage == "" {
        c.I18n.DefaultLanguage = "en"
    }
    if c.I18n.Language == "" {
        c.I18n.Language = c.I18n.DefaultLanguage
    }
    if c.I18n.LocalesDir == "" {
        c.I18n.LocalesDir = "./locales"
    }
    
    // Auth
    if c.Auth.RedirectURI == "" {
        c.Auth.RedirectURI = "http://localhost:8080/callback"
    }
    if c.Auth.CallbackPort == 0 {
        c.Auth.CallbackPort = 8080
    }
    c.Auth.UseOAuth = true
    c.Auth.AutoRefresh = true
    
    // Letterboxd
    if c.Letterboxd.ExportDir == "" {
        c.Letterboxd.ExportDir = "./exports"
    }
    // ... autres valeurs par défaut
}
```

### 📚 Exemples d'Usage

#### Configuration Basique
```toml
# config/config.toml
[trakt]
client_id = "your_client_id"
client_secret = "your_client_secret"

[letterboxd]
export_dir = "./exports"

[export]
history_mode = "individual"
timezone = "Europe/Paris"

[logging]
level = "info"

[i18n]
language = "fr"

[auth]
use_oauth = true
```

#### Configuration Avancée avec Sécurité
```toml
[trakt]
client_id = "${TRAKT_CLIENT_ID}"
client_secret = "${TRAKT_CLIENT_SECRET}"
api_base_url = "https://api.trakt.tv"

[security]
security_level = "high"
encryption_enabled = true
keyring_backend = "system"
require_https = true
audit_logging = true

[security.audit]
enabled = true
file_path = "./logs/audit.log"
retention_days = 90

[export]
format = "csv"
history_mode = "individual"
timezone = "Europe/Paris"

[logging]
level = "info"
file = "./logs/export.log"
```

#### Chargement en Code
```go
// Chargement simple
cfg, err := config.LoadConfig("config/config.toml")
if err != nil {
    log.Fatal("Failed to load config:", err)
}

// Accès aux valeurs
clientID := cfg.Trakt.ClientID
exportDir := cfg.Letterboxd.ExportDir
historyMode := cfg.Export.HistoryMode
logLevel := cfg.Logging.Level

// Validation personnalisée
if cfg.Export.HistoryMode == "individual" {
    log.Info("Using individual history mode for complete watch tracking")
}
```

### 🔧 Extension et Personnalisation

#### Ajout de Nouvelles Sections
1. Définir la structure dans `config.go`
2. Ajouter la validation correspondante
3. Implémenter les valeurs par défaut
4. Ajouter les tests unitaires

#### Support Nouveaux Formats
Le module est extensible pour supporter d'autres formats de configuration (JSON, YAML) via des adaptateurs.

Ce module fournit une fondation robuste et flexible pour la gestion de toute la configuration de l'application, avec validation automatique, sécurité intégrée, et facilité d'extension.