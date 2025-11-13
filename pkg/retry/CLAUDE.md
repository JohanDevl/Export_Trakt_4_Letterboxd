# CLAUDE.md - Résilience API avec Circuit Breaker

## Module Overview

Ce module implémente un système de résilience avancé avec circuit breaker, exponential backoff, retry policies et fallback strategies pour garantir la robustesse des appels API même en cas de pannes ou de dégradations.

## Architecture du Module

### ⚡ Circuit Breaker
```go
type CircuitBreaker struct {
    state           State
    failureCount    int
    successCount    int
    threshold       int
    timeout         time.Duration
    lastFailureTime time.Time
    mutex          sync.RWMutex
}

type State int

const (
    StateClosed   State = iota  // Trafic normal
    StateOpen                   // Circuit ouvert, rejets
    StateHalfOpen              // Test de récupération
)

func (cb *CircuitBreaker) Call(fn func() (interface{}, error)) (interface{}, error) {
    cb.mutex.Lock()
    defer cb.mutex.Unlock()
    
    switch cb.state {
    case StateOpen:
        if time.Since(cb.lastFailureTime) > cb.timeout {
            cb.state = StateHalfOpen
            cb.successCount = 0
        } else {
            return nil, ErrCircuitBreakerOpen
        }
    case StateHalfOpen:
        // Limite les tentatives en mode test
        if cb.successCount >= 3 {
            cb.state = StateClosed
            cb.failureCount = 0
        }
    }
    
    result, err := fn()
    
    if err != nil {
        cb.onFailure()
    } else {
        cb.onSuccess()
    }
    
    return result, err
}
```

### 📈 Exponential Backoff
```go
type ExponentialBackoff struct {
    initialDelay  time.Duration
    maxDelay      time.Duration
    multiplier    float64
    jitter        bool
    maxRetries    int
}

func (eb *ExponentialBackoff) NextDelay(attempt int) time.Duration {
    if attempt <= 0 {
        return eb.initialDelay
    }
    
    delay := time.Duration(float64(eb.initialDelay) * math.Pow(eb.multiplier, float64(attempt-1)))
    
    if delay > eb.maxDelay {
        delay = eb.maxDelay
    }
    
    if eb.jitter {
        // Ajoute +/-10% de jitter pour éviter thundering herd
        jitterAmount := float64(delay) * 0.1
        jitterOffset := (rand.Float64() - 0.5) * 2 * jitterAmount
        delay += time.Duration(jitterOffset)
    }
    
    return delay
}
```

### 🔄 Retry Client
```go
type RetryClient struct {
    httpClient      *http.Client
    circuitBreaker  *CircuitBreaker
    backoffStrategy BackoffStrategy
    retryPolicy     RetryPolicy
    metrics         *RetryMetrics
}

type RetryPolicy struct {
    MaxRetries      int
    RetryableErrors []error
    RetryableStatus []int
    RetryCondition  func(error) bool
}

func (rc *RetryClient) DoWithRetry(req *http.Request) (*http.Response, error) {
    var lastErr error
    
    for attempt := 0; attempt <= rc.retryPolicy.MaxRetries; attempt++ {
        if attempt > 0 {
            delay := rc.backoffStrategy.NextDelay(attempt)
            rc.metrics.RecordRetryDelay(delay)
            time.Sleep(delay)
        }
        
        resp, err := rc.circuitBreaker.Call(func() (interface{}, error) {
            return rc.httpClient.Do(req)
        })
        
        if err == nil {
            rc.metrics.RecordSuccess(attempt)
            return resp.(*http.Response), nil
        }
        
        if !rc.shouldRetry(err) {
            rc.metrics.RecordFailure(attempt, err)
            return nil, err
        }
        
        lastErr = err
        rc.metrics.RecordRetryAttempt(attempt)
    }
    
    rc.metrics.RecordMaxRetriesExceeded()
    return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (rc *RetryClient) shouldRetry(err error) bool {
    // Vérification des erreurs retryables
    for _, retryableErr := range rc.retryPolicy.RetryableErrors {
        if errors.Is(err, retryableErr) {
            return true
        }
    }
    
    // Vérification custom condition
    if rc.retryPolicy.RetryCondition != nil {
        return rc.retryPolicy.RetryCondition(err)
    }
    
    return false
}
```

### 📊 Métriques de Retry
```go
type RetryMetrics struct {
    totalAttempts     int64
    successfulRetries int64
    failedRetries     int64
    circuitBreakTrips int64
    averageDelay      time.Duration
    maxRetryReached   int64
}

func (rm *RetryMetrics) RecordRetryAttempt(attempt int) {
    atomic.AddInt64(&rm.totalAttempts, 1)
}

func (rm *RetryMetrics) RecordSuccess(attempts int) {
    if attempts > 0 {
        atomic.AddInt64(&rm.successfulRetries, 1)
    }
}

func (rm *RetryMetrics) GetStats() RetryStats {
    return RetryStats{
        TotalAttempts:     atomic.LoadInt64(&rm.totalAttempts),
        SuccessfulRetries: atomic.LoadInt64(&rm.successfulRetries),
        FailedRetries:     atomic.LoadInt64(&rm.failedRetries),
        SuccessRate:       rm.calculateSuccessRate(),
        CircuitBreakTrips: atomic.LoadInt64(&rm.circuitBreakTrips),
    }
}
```

### 🛡️ Fallback Strategies
```go
type FallbackStrategy interface {
    Execute(originalError error) (interface{}, error)
    CanFallback(error) bool
}

type CacheFallbackStrategy struct {
    cache cache.Cache
    ttl   time.Duration
}

func (cfs *CacheFallbackStrategy) Execute(originalError error) (interface{}, error) {
    // Tentative de récupération depuis le cache
    cachedData, found := cfs.cache.Get("fallback_data")
    if found {
        return cachedData, nil
    }
    
    return nil, fmt.Errorf("no fallback data available: %w", originalError)
}

type DefaultValueFallbackStrategy struct {
    defaultValue interface{}
}

func (dvfs *DefaultValueFallbackStrategy) Execute(originalError error) (interface{}, error) {
    return dvfs.defaultValue, nil
}
```

### ⚙️ Configuration

#### Retry Configuration
```toml
[retry]
enabled = true
max_retries = 3
initial_delay = "1s"
max_delay = "30s"
multiplier = 2.0
jitter = true

[retry.circuit_breaker]
enabled = true
failure_threshold = 5
success_threshold = 3
timeout = "60s"

[retry.policy]
retryable_status_codes = [429, 500, 502, 503, 504]
retryable_errors = ["connection_timeout", "dns_resolution"]

[retry.fallback]
enabled = true
strategy = "cache"  # cache, default_value, degraded
cache_ttl = "1h"
```

### 🚦 Types d'Erreurs Gérées

#### Erreurs Retryables
- **429 Too Many Requests** : Rate limiting
- **500, 502, 503, 504** : Erreurs serveur temporaires
- **Connection Timeout** : Timeout réseau
- **DNS Resolution** : Problèmes DNS temporaires
- **Connection Reset** : Connexion interrompue

#### Erreurs Non-Retryables
- **400 Bad Request** : Requête malformée
- **401 Unauthorized** : Problème d'authentification
- **403 Forbidden** : Accès interdit
- **404 Not Found** : Ressource inexistante

### 📈 Patterns de Résilience

#### Retry avec Circuit Breaker
```go
func NewResilientAPIClient(cfg Config) *ResilientClient {
    circuitBreaker := &CircuitBreaker{
        threshold: cfg.CircuitBreaker.FailureThreshold,
        timeout:   cfg.CircuitBreaker.Timeout,
    }
    
    backoffStrategy := &ExponentialBackoff{
        initialDelay: cfg.Retry.InitialDelay,
        maxDelay:     cfg.Retry.MaxDelay,
        multiplier:   cfg.Retry.Multiplier,
        jitter:       cfg.Retry.Jitter,
    }
    
    retryPolicy := RetryPolicy{
        MaxRetries:      cfg.Retry.MaxRetries,
        RetryableStatus: cfg.Retry.RetryableStatusCodes,
        RetryCondition: func(err error) bool {
            return isTemporaryError(err)
        },
    }
    
    return &ResilientClient{
        retryClient: &RetryClient{
            circuitBreaker:  circuitBreaker,
            backoffStrategy: backoffStrategy,
            retryPolicy:     retryPolicy,
        },
        fallbackStrategy: NewCacheFallbackStrategy(cache),
    }
}
```

### 🚀 Usage Pratique

#### Client API Résilient
```go
// Configuration résilience
cfg := retry.Config{
    MaxRetries:   3,
    InitialDelay: time.Second,
    MaxDelay:     30 * time.Second,
    Multiplier:   2.0,
    CircuitBreaker: retry.CircuitBreakerConfig{
        FailureThreshold: 5,
        Timeout:         60 * time.Second,
    },
}

// Création client résilient
client := retry.NewResilientAPIClient(cfg)

// Appel avec résilience automatique
movies, err := client.GetWatchedMovies()
if err != nil {
    // Gestion d'erreur après tous les retries
    log.Error("Failed after all retries", err)
}

// Statistiques de résilience
stats := client.GetRetryStats()
log.Info("Retry stats", map[string]interface{}{
    "success_rate": stats.SuccessRate,
    "total_attempts": stats.TotalAttempts,
    "circuit_breaks": stats.CircuitBreakTrips,
})
```

#### Patterns d'Usage Avancés
```go
// Retry avec condition custom
retryPolicy := RetryPolicy{
    MaxRetries: 5,
    RetryCondition: func(err error) bool {
        // Retry uniquement sur erreurs temporaires spécifiques
        if apiErr, ok := err.(*APIError); ok {
            return apiErr.IsTemporary()
        }
        return false
    },
}

// Fallback multi-niveaux
fallbackChain := []FallbackStrategy{
    NewCacheFallbackStrategy(cache),
    NewDefaultValueFallbackStrategy(defaultMovies),
    NewDegradedServiceFallbackStrategy(),
}

// Circuit breaker avec monitoring
circuitBreaker := NewCircuitBreakerWithMetrics(cfg, prometheus.DefaultRegisterer)
```

### 📊 Monitoring et Alertes

#### Métriques Exposées
- **retry_attempts_total** : Nombre total de tentatives
- **retry_success_rate** : Taux de succès après retry
- **circuit_breaker_state** : État du circuit breaker
- **fallback_activations** : Activations des fallbacks
- **retry_delay_seconds** : Distribution des délais

#### Alertes Recommandées
- Circuit breaker ouvert > 5 minutes
- Taux de retry > 20% sur 15 minutes  
- Fallback activé > 10 fois en 5 minutes
- Délai moyen de retry > 10 secondes

Ce module assure une résilience robuste face aux pannes temporaires et permanentes, avec récupération automatique et dégradation gracieuse des services.