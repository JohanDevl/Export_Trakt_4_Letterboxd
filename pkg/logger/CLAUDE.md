# CLAUDE.md - Logging Structuré avec i18n

## Module Overview

Ce module fournit un système de logging structuré avec support d'internationalisation, niveaux configurables, rotation automatique des logs et intégration avec les métriques de performance.

## Architecture du Module

### 📝 Logger Principal
```go
type Logger interface {
    Debug(key string, data map[string]interface{})
    Info(key string, data map[string]interface{})
    Warn(key string, data map[string]interface{})
    Error(key string, data map[string]interface{})
    SetLogLevel(level string)
    SetLogFile(filepath string) error
    SetTranslator(translator *i18n.Translator)
}

type StructuredLogger struct {
    level      LogLevel
    output     io.Writer
    file       *os.File
    translator *i18n.Translator
    mutex      sync.RWMutex
}
```

### 🌍 Intégration i18n

#### Messages Traduits
```go
// Utilisation avec clés de traduction
log.Info("export.starting", map[string]interface{}{
    "export_type": "watched",
    "count": 150,
})

// Rendu selon la langue configurée :
// EN: "Export starting: watched (150 items)"
// FR: "Export démarré : watched (150 éléments)"
// DE: "Export gestartet: watched (150 Elemente)"
```

#### Clés de Messages Structurées
- **`export.*`** : Messages d'export
- **`auth.*`** : Messages d'authentification
- **`api.*`** : Messages API
- **`errors.*`** : Messages d'erreur
- **`scheduler.*`** : Messages de planification

### 📊 Niveaux de Log

#### Hiérarchie des Niveaux
1. **DEBUG** : Informations de débogage détaillées
2. **INFO** : Informations générales (par défaut)
3. **WARN** : Avertissements non critiques
4. **ERROR** : Erreurs critiques

#### Configuration Dynamique
```go
// Via configuration
log.SetLogLevel("debug")

// Via variable d'environnement
LOG_LEVEL=debug ./export_trakt
```

### 📁 Gestion des Fichiers

#### Rotation Automatique
- Rotation par taille (100MB par défaut)
- Rétention configurable (30 jours)
- Compression des anciens logs
- Nettoyage automatique

#### Structure des Logs
```
logs/
├── export.log           # Log actuel
├── export.log.1         # Rotation précédente
├── export.log.2.gz      # Archive compressée
├── audit.log            # Logs de sécurité
└── README.md
```

### 🏷️ Format Structuré

#### Format JSON
```json
{
  "timestamp": "2025-07-11T15:43:22Z",
  "level": "INFO",
  "message": "Export completed successfully",
  "context": {
    "export_type": "watched",
    "duration": "2.3s",
    "records": 150,
    "file": "watched.csv"
  },
  "source": "pkg/export/letterboxd.go:245"
}
```

#### Format Console (Développement)
```
2025-07-11 15:43:22 INFO  Export completed successfully export_type=watched duration=2.3s records=150
```

### 🔧 Configuration

#### Configuration TOML
```toml
[logging]
level = "info"                    # debug, info, warn, error
file = "./logs/export.log"        # Fichier de log (optionnel)
format = "json"                   # json, text
max_size = "100MB"               # Taille max avant rotation
max_age = 30                     # Jours de rétention
max_backups = 10                 # Nombre de backups
compress = true                  # Compression des anciens logs
```

### 🚀 Usage

#### Logging Simple
```go
log := logger.NewLogger()
log.Info("application.starting", nil)
log.Error("database.connection_failed", map[string]interface{}{
    "host": "localhost",
    "error": err.Error(),
})
```

#### Avec Traduction
```go
translator, _ := i18n.NewTranslator(&cfg.I18n, log)
log.SetTranslator(translator)

log.Info("export.success", map[string]interface{}{
    "count": 150,
    "duration": "2.3s",
})
// FR: "Export réussi : 150 éléments en 2.3s"
```

### 📈 Intégration Métriques

#### Corrélation avec Monitoring
- Logs liés aux métriques Prometheus
- Tracing distribué avec OpenTelemetry
- Alertes basées sur les patterns de logs
- Dashboard des erreurs en temps réel

Ce module assure un logging professionnel avec traçabilité complète et support multilingue pour une expérience utilisateur optimale.