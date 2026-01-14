# CLAUDE.md - Gestion Résilience Checkpoints

## Module Overview

Ce module implémente un système de checkpoints et de récupération pour assurer la continuité des opérations longues en cas d'interruption, avec sauvegarde automatique de l'état et reprise intelligente.

## Architecture du Module

### 💾 Checkpoint Manager
```go
type Manager struct {
    checkpointDir string
    interval      time.Duration
    maxCheckpoints int
    operations    map[string]*Operation
    mutex         sync.RWMutex
}

type Checkpoint struct {
    ID          string                 `json:"id"`
    OperationType string               `json:"operation_type"`
    Progress    float64                `json:"progress"`
    State       map[string]interface{} `json:"state"`
    Timestamp   time.Time              `json:"timestamp"`
    LastItem    interface{}            `json:"last_item,omitempty"`
}

type Operation struct {
    ID           string
    Type         string
    TotalItems   int
    ProcessedItems int
    State        map[string]interface{}
    StartTime    time.Time
    LastCheckpoint time.Time
}
```

### 🔄 Fonctionnalités Principales

#### Sauvegarde Automatique
- **Checkpoints périodiques** : Sauvegarde automatique toutes les N opérations
- **État complet** : Capture de l'état complet de l'opération
- **Métadonnées** : Progression, timing, contexte
- **Compression** : Optimisation de l'espace disque

#### Récupération Intelligente
- **Détection d'interruption** : Identification des opérations incomplètes
- **Reprise exacte** : Continuation depuis le dernier checkpoint
- **Validation d'état** : Vérification de la cohérence
- **Cleanup automatique** : Nettoyage des checkpoints obsolètes

### 🚀 Usage

#### Export avec Checkpoints
```go
manager := resilience.NewCheckpointManager("./checkpoints")

// Démarrage opération avec checkpoint
op := manager.StartOperation("export_movies", totalMovies)

for i, movie := range movies {
    // Traitement du film
    processMovie(movie)
    
    // Checkpoint tous les 100 films
    if i%100 == 0 {
        op.SaveCheckpoint(map[string]interface{}{
            "last_movie_id": movie.ID,
            "processed_count": i,
        })
    }
}

op.Complete()
```

#### Reprise après Interruption
```go
// Vérification d'opérations interrompues au démarrage
interrupted := manager.GetInterruptedOperations()

for _, op := range interrupted {
    log.Info("Resuming interrupted operation", op.ID)
    
    // Reprise depuis le checkpoint
    lastState := op.GetLastCheckpoint()
    lastProcessed := lastState["processed_count"].(int)
    
    // Continuation du traitement
    for i := lastProcessed; i < op.TotalItems; i++ {
        // Reprendre le traitement...
    }
}
```

Ce module assure la continuité des opérations longues avec récupération automatique en cas d'interruption système.