# CLAUDE.md - Gestion Erreurs Structurées

## Module Overview

Ce module fournit un système complet de gestion d'erreurs avec classification, recovery automatique, logging structuré et handlers spécialisés pour différents types d'erreurs.

## Architecture du Module

### 🚨 Types d'Erreurs
```go
type ErrorType int

const (
    ErrorTypeNetwork ErrorType = iota
    ErrorTypeAuth
    ErrorTypeAPI
    ErrorTypeValidation
    ErrorTypeSystem
    ErrorTypeUser
)

type StructuredError struct {
    Type        ErrorType              `json:"type"`
    Code        string                 `json:"code"`
    Message     string                 `json:"message"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Cause       error                  `json:"cause,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
    Retryable   bool                   `json:"retryable"`
    Severity    Severity               `json:"severity"`
}

type Severity int

const (
    SeverityInfo Severity = iota
    SeverityWarning
    SeverityError
    SeverityCritical
)
```

### 🔧 Error Manager
```go
type Manager struct {
    handlers map[ErrorType]ErrorHandler
    recovery RecoveryStrategy
    logger   logger.Logger
    metrics  *ErrorMetrics
}

type ErrorHandler interface {
    Handle(error) (*HandlerResult, error)
    CanHandle(error) bool
    GetPriority() int
}

type HandlerResult struct {
    Recovered bool
    Action    RecoveryAction
    Message   string
    Data      map[string]interface{}
}

type RecoveryAction int

const (
    ActionRetry RecoveryAction = iota
    ActionFallback
    ActionSkip
    ActionAbort
)
```

### 🔄 Handlers Spécialisés

#### Network Error Handler
```go
type NetworkErrorHandler struct{}

func (neh *NetworkErrorHandler) Handle(err error) (*HandlerResult, error) {
    if netErr, ok := err.(net.Error); ok {
        if netErr.Timeout() {
            return &HandlerResult{
                Recovered: false,
                Action:    ActionRetry,
                Message:   "Network timeout, will retry",
            }, nil
        }
        
        if netErr.Temporary() {
            return &HandlerResult{
                Recovered: false,
                Action:    ActionRetry,
                Message:   "Temporary network error",
            }, nil
        }
    }
    
    return &HandlerResult{
        Action: ActionAbort,
        Message: "Permanent network error",
    }, nil
}
```

#### Auth Error Handler
```go
type AuthErrorHandler struct {
    tokenManager *auth.TokenManager
}

func (aeh *AuthErrorHandler) Handle(err error) (*HandlerResult, error) {
    if authErr, ok := err.(*AuthError); ok {
        switch authErr.Code {
        case "token_expired":
            // Tentative de refresh automatique
            if err := aeh.tokenManager.RefreshToken(); err == nil {
                return &HandlerResult{
                    Recovered: true,
                    Action:    ActionRetry,
                    Message:   "Token refreshed successfully",
                }, nil
            }
            fallthrough
            
        case "invalid_token":
            return &HandlerResult{
                Action:  ActionAbort,
                Message: "Re-authentication required",
                Data: map[string]interface{}{
                    "requires_reauth": true,
                },
            }, nil
        }
    }
    
    return nil, fmt.Errorf("cannot handle error: %w", err)
}
```

### 📊 Métriques d'Erreurs
```go
type ErrorMetrics struct {
    totalErrors   int64
    errorsByType  map[ErrorType]int64
    recoveryRate  float64
    lastError     time.Time
}

func (em *ErrorMetrics) RecordError(err *StructuredError) {
    atomic.AddInt64(&em.totalErrors, 1)
    
    if count, exists := em.errorsByType[err.Type]; exists {
        em.errorsByType[err.Type] = count + 1
    } else {
        em.errorsByType[err.Type] = 1
    }
    
    em.lastError = time.Now()
}
```

### 🚀 Usage

#### Gestion Automatique
```go
errorManager := errors.NewManager(cfg, log)

// Enregistrement des handlers
errorManager.RegisterHandler(ErrorTypeNetwork, &NetworkErrorHandler{})
errorManager.RegisterHandler(ErrorTypeAuth, &AuthErrorHandler{tokenManager})

// Traitement d'erreur avec recovery
err := someOperation()
if err != nil {
    result, handlerErr := errorManager.HandleError(err)
    if handlerErr == nil && result.Recovered {
        // Retry automatique
        err = someOperation()
    }
}
```

Ce module assure une gestion robuste des erreurs avec récupération intelligente et métriques détaillées.