# CLAUDE.md - Système de Planification Cron

## Module Overview

Ce module implémente un système de planification robuste basé sur des expressions cron pour automatiser les exports de données. Il gère les tâches programmées, les fuseaux horaires, la surveillance des jobs et l'arrêt gracieux.

## Architecture du Module

### 🕒 Scheduler Principal
```go
type Scheduler struct {
    config *config.Config
    logger logger.Logger
    cron   *cron.Cron
    jobs   map[string]*ScheduledJob
}

type ScheduledJob struct {
    ID       string
    Schedule string
    NextRun  time.Time
    LastRun  time.Time
    Status   JobStatus
    Function func()
}
```

### 📅 Fonctionnalités Principales

#### Expressions Cron Supportées
- **`0 */6 * * *`** : Toutes les 6 heures
- **`0 9 * * 1`** : Chaque lundi à 9h00
- **`30 14 * * *`** : Chaque jour à 14h30
- **`0 0 1 * *`** : Le 1er de chaque mois à minuit

#### Gestion des Fuseaux Horaires
- Support des timezones IANA (Europe/Paris, America/New_York, etc.)
- Conversion automatique UTC ↔ timezone locale
- Gestion de l'heure d'été/hiver

#### Types de Jobs
- **Export unique** : Un seul type d'export (watched, collection, etc.)
- **Export complet** : Tous les types d'exports en séquence
- **Maintenance** : Nettoyage des anciens exports, logs, etc.

### 🛠️ Configuration

#### Configuration TOML
```toml
[scheduler]
enabled = true
timezone = "Europe/Paris"
max_concurrent_jobs = 1
job_timeout = "30m"

[[scheduler.jobs]]
name = "daily_export"
schedule = "0 9 * * *"
export_type = "all"
export_mode = "complete"
enabled = true

[[scheduler.jobs]]
name = "hourly_watched"
schedule = "0 */1 * * *"
export_type = "watched"
export_mode = "normal"
enabled = false
```

### 📊 Monitoring et Logging

#### Métriques de Performance
- Temps d'exécution des jobs
- Taux de succès/échec
- Prochaine exécution programmée
- Statut du scheduler

#### Logs Structurés
```go
log.Info("scheduler.job_triggered", map[string]interface{}{
    "schedule":    schedule,
    "export_type": exportType,
    "timestamp":   time.Now().Format(time.RFC3339),
})
```

### 🔧 Gestion des Erreurs
- Recovery automatique en cas d'échec
- Retry avec backoff exponentiel
- Notifications d'erreur configurables
- Arrêt gracieux avec SIGINT/SIGTERM

### 🚀 Usage

#### Démarrage Simple
```bash
./export_trakt --schedule "0 */6 * * *" --export all --mode complete
```

#### Avec Serveur Web
```bash
./export_trakt server --schedule "0 */6 * * *" --export watched
```

Ce module assure une automatisation fiable des exports avec surveillance complète et gestion robuste des erreurs.