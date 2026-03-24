# CLAUDE.md - Trakt.tv API Client

## Module Overview

Ce module implémente un client API avancé pour Trakt.tv avec des optimisations de performance, mise en cache intelligente, gestion de la concurrence, et résilience aux pannes. Il fournit une abstraction unifiée pour toutes les opérations d'API Trakt.tv.

## Architecture du Module

### 🏗️ Structure Unifiée des Clients

#### TraktAPIClient (Interface de Base)
- **Opérations CRUD** : `GetWatchedMovies()`, `GetCollectionMovies()`, `GetWatchedShows()`, etc.
- **Gestion de Configuration** : Accès centralisé à la configuration
- **Cycle de Vie** : Méthodes d'initialisation et nettoyage

#### OptimizedTraktAPIClient (Interface Avancée)
- **Opérations Concurrentes** : `GetWatchedMoviesConcurrent()`, `GetRatingsConcurrent()`
- **Traitement par Lots** : `ProcessBatchRequests()` pour requêtes groupées
- **Métriques de Performance** : `GetCacheStats()`, `GetPerformanceMetrics()`
- **Gestion du Cache** : `ClearCache()` pour contrôle du cache

#### UnifiedTraktClient (Interface Complète)
- **Santé et Métriques** : `HealthCheck()`, `GetRequestCount()`, `GetAverageResponseTime()`
- **Exécution d'Opérations** : `ExecuteOperation()`, `ExecuteBatchOperations()`
- **Capacités Dynamiques** : `GetClientType()`, `GetCapabilities()`, `Reconfigure()`

### 📊 Types de Données API

#### Structures de Films
```go
type MovieInfo struct {
    Title    string   `json:"title"`
    Year     int      `json:"year"`
    IDs      MovieIDs `json:"ids"`
    Tagline  string   `json:"tagline,omitempty"`
    Overview string   `json:"overview,omitempty"`
    // ... métadonnées complètes
}

type Movie struct {
    Movie         MovieInfo `json:"movie"`
    LastWatchedAt string    `json:"last_watched_at"`
    Plays         int       `json:"plays,omitempty"`
}
```

#### Structures de Séries
```go
type ShowInfo struct {
    Title      string   `json:"title"`
    Year       int      `json:"year"`
    IDs        ShowIDs  `json:"ids"`
    FirstAired string   `json:"first_aired,omitempty"`
    Network    string   `json:"network,omitempty"`
    // ... métadonnées étendues
}

type WatchedShow struct {
    Show    ShowInfo    `json:"show"`
    Seasons []Season    `json:"seasons"`
    LastWatchedAt string `json:"last_watched_at"`
}
```

#### Structures d'Évaluation et Historique
```go
type Rating struct {
    Movie   MovieInfo `json:"movie"`
    Rating  int       `json:"rating"`
    RatedAt string    `json:"rated_at"`
}

type HistoryItem struct {
    Movie     MovieInfo `json:"movie"`
    WatchedAt string    `json:"watched_at"`
    Action    string    `json:"action"`
}
```

### ⚡ Optimisations de Performance

#### Client Optimisé (OptimizedClient)
```go
type OptimizedClient struct {
    config      *config.Config
    logger      logger.Logger
    httpClient  *http.Client
    cache       *cache.APIResponseCache   // Cache LRU avec TTL
    metrics     *metrics.PerformanceMetrics
    workerPool  *pool.WorkerPool         // Pool de workers concurrent
    rateLimiter chan struct{}            // Rate limiting intégré
    transport   *http.Transport          // Pool de connexions HTTP
}
```

#### Configuration Optimisée
```go
type OptimizedClientConfig struct {
    Config           *config.Config
    Logger           logger.Logger
    CacheConfig      cache.CacheConfig    // TTL: 24h, Capacity: 1000
    WorkerPoolSize   int                  // Défaut: 10 workers
    RateLimitPerSec  int                  // Défaut: 100 req/s
    ConnectionPool   int                  // Défaut: 20 connexions
    RequestTimeout   time.Duration        // Défaut: 30s
}
```

### 🔄 Patterns d'Exécution

#### Opérations API Génériques
```go
type APIOperation interface {
    Execute(ctx context.Context, client TraktAPIClient) (interface{}, error)
    GetOperationName() string
    GetMaxRetries() int
    IsRetryable(error) bool
}

type APIOperationResult struct {
    Data      interface{}
    Error     error
    Duration  time.Duration
    Attempts  int
    Operation string
}
```

#### Politiques de Retry
```go
type RetryPolicy struct {
    MaxAttempts   int           // Tentatives max
    InitialDelay  time.Duration // Délai initial
    MaxDelay      time.Duration // Délai maximum
    BackoffFactor float64       // Facteur d'augmentation
    RetryOn       []error       // Erreurs spécifiques à retry
}
```

#### Opérations par Lots
```go
type BatchOperation struct {
    Operations     []APIOperation
    MaxConcurrency int
    FailFast       bool // Arrêt au premier échec
}
```

### 🚀 Fonctionnalités Avancées

#### 1. Cache LRU Intelligent
- **TTL Configurable** : Durée de vie de 24h par défaut
- **Capacité Adaptative** : 1000 entrées par défaut
- **Invalidation Automatique** : Basée sur l'âge et l'usage
- **Metrics de Cache** : Hit ratio, miss count, evictions

#### 2. Pool de Workers Concurrent
- **Workers Configurables** : 10 workers par défaut
- **Buffer Circulaire** : Queue de tâches avec taille adaptative
- **Gestion d'Erreurs** : Retry automatique et circuit breaker
- **Monitoring** : Métriques de throughput et latence

#### 3. Rate Limiting Adaptatif
- **Limite Configurable** : 100 requêtes/seconde par défaut
- **Refill Automatique** : Rechargement constant du bucket
- **Backoff Intelligent** : Adaptation automatique en cas de limitation

#### 4. Pool de Connexions HTTP
- **Connexions Persistantes** : Réutilisation des connexions TCP
- **Idle Timeout** : 90 secondes de timeout d'inactivité
- **Connection Pooling** : 20 connexions max par défaut
- **Compression** : Support gzip/deflate automatique

### 📈 Métriques et Observabilité

#### Métriques de Performance
```go
type ClientMetrics interface {
    GetRequestCount() int64
    GetErrorCount() int64
    GetAverageResponseTime() time.Duration
    GetCacheHitRatio() float64
    GetCircuitBreakerStatus() string
}
```

#### Health Checks
```go
type ClientHealth interface {
    HealthCheck(ctx context.Context) error
    GetLastError() error
    GetUptime() time.Duration
    IsHealthy() bool
}
```

### 🔧 Factory Pattern

#### ClientFactory pour Création Dynamique
```go
type ClientFactory interface {
    CreateBasicClient(cfg *config.Config) (TraktAPIClient, error)
    CreateOptimizedClient(cfg OptimizedClientConfig) (OptimizedTraktAPIClient, error)
    CreateClientWithCapabilities(cfg ClientCapabilitiesConfig) (TraktAPIClient, error)
}

type ClientCapabilitiesConfig struct {
    BaseConfig        *config.Config
    EnableCaching     bool
    EnableMetrics     bool
    EnableRetry       bool
    EnableRateLimit   bool
    EnableConcurrency bool
    WorkerPoolSize    int
    CacheConfig       *cache.CacheConfig
}
```

### 🛡️ Résilience et Gestion d'Erreurs

#### Client Conscient des Erreurs (ErrorAwareClient)
- **Classification d'Erreurs** : Erreurs temporaires vs permanentes
- **Retry Intelligent** : Backoff exponentiel avec jitter
- **Circuit Breaker** : Protection contre les pannes en cascade
- **Fallback** : Stratégies de dégradation gracieuse

#### Adaptateurs pour Compatibilité
- **Legacy Support** : Compatibilité avec anciennes versions API
- **Format Adaptation** : Conversion automatique des formats de données
- **Version Management** : Support multi-versions simultané

### 📋 Endpoints API Supportés

#### Films
- **`/sync/watched/movies`** : Films regardés (mode agrégé)
- **`/sync/history/movies`** : Historique complet (mode individuel)
- **`/sync/collection/movies`** : Collection personnelle
- **`/sync/ratings/movies`** : Notes attribuées aux films
- **`/sync/watchlist/movies`** : Liste de films à voir

#### Séries TV
- **`/sync/watched/shows`** : Séries regardées avec épisodes
- **`/sync/ratings/shows`** : Notes des séries
- **`/sync/ratings/episodes`** : Notes des épisodes
- **`/sync/collection/shows`** : Collection de séries

#### Configuration Dynamique
- **Extended Info** : Support `full`, `metadata`, `letterboxd`
- **Pagination** : Gestion automatique des pages multiples
- **Filtres** : Date ranges, statuts, genres, etc.

### 🎯 Modes d'Export

#### Mode Agrégé (Performance)
```go
func (c *OptimizedClient) GetWatchedMovies() ([]Movie, error) {
    // Une entrée par film avec date de dernière écoute
    // Optimisé pour rapidité et compatibilité Letterboxd
}
```

#### Mode Historique Individuel (Complet)
```go
func (c *OptimizedClient) GetMovieHistory() ([]HistoryItem, error) {
    // Historique complet de tous les visionnages
    // Support des re-visionnages chronologiques
}
```

### 🚧 Exemples d'Usage

#### Client Basic
```go
cfg := &config.Config{...}
log := logger.NewLogger()
client := api.NewClient(cfg, log)

movies, err := client.GetWatchedMovies()
if err != nil {
    log.Error("Failed to get movies", err)
}
```

#### Client Optimisé
```go
optimizedCfg := api.OptimizedClientConfig{
    Config:          cfg,
    Logger:          log,
    WorkerPoolSize:  15,
    RateLimitPerSec: 50,
    CacheConfig: cache.CacheConfig{
        Capacity: 2000,
        TTL:      12 * time.Hour,
    },
}

client := api.NewOptimizedClient(optimizedCfg)
defer client.Close()

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

movies, err := client.GetWatchedMoviesConcurrent(ctx)
stats := client.GetCacheStats()
metrics := client.GetPerformanceMetrics()
```

#### Opérations par Lots
```go
operations := []api.APIOperation{
    &GetWatchedMoviesOperation{},
    &GetCollectionMoviesOperation{},
    &GetRatingsOperation{},
}

batch := api.BatchOperation{
    Operations:     operations,
    MaxConcurrency: 3,
    FailFast:      false,
}

results := client.ExecuteBatchOperations(ctx, batch)
```

### ⚙️ Configuration et Tuning

#### Optimisation par Type d'Usage
```toml
[api.performance]
# Configuration pour gros volumes
worker_pool_size = 20
rate_limit_per_sec = 200
connection_pool = 50
cache_capacity = 5000
cache_ttl = "48h"

[api.cache]
# Stratégies de cache
enable_cache = true
cache_strategy = "lru"
cache_compression = true
cache_encryption = false
```

#### Monitoring et Alerting
- **Latence** : Temps de réponse par endpoint
- **Throughput** : Requêtes/seconde traitées
- **Error Rate** : Pourcentage d'erreurs par type
- **Cache Performance** : Hit ratio et évictions
- **Resource Usage** : CPU, mémoire, connexions

### 🔍 Debugging et Diagnostics

#### Logs Structurés
```go
client.Logger.Info("api.request.starting", map[string]interface{}{
    "endpoint": "/sync/watched/movies",
    "cache_enabled": true,
    "worker_id": workerID,
})
```

#### Métriques Détaillées
- **Temps de Réponse** : Par endpoint et percentiles
- **Statut Cache** : Hit/miss ratio par endpoint
- **Pool de Workers** : Utilisation et queue length
- **Rate Limiting** : Rejets et délais d'attente

### 🔄 Évolution et Extensions

Le module API est conçu pour être extensible :
- **Nouveaux Endpoints** : Ajout facile de nouvelles méthodes
- **Stratégies de Cache** : Support de différents backends
- **Formats de Données** : Adaptation pour nouveaux formats
- **Clients Spécialisés** : Création de clients dédiés par usage

Cette architecture modulaire permet une évolution progressive tout en maintenant la compatibilité et les performances optimales.