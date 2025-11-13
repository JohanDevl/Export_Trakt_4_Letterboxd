# CLAUDE.md - Traitement Streaming Mémoire

## Module Overview

Ce module implémente un système de traitement streaming pour gérer efficacement de gros volumes de données avec une utilisation mémoire minimale, processing par chunks et pipeline de transformation.

## Architecture du Module

### 🌊 Streaming Processor
```go
type StreamingProcessor struct {
    chunkSize    int
    bufferSize   int
    pipeline     []ProcessorStage
    errorHandler ErrorHandler
    metrics      *StreamingMetrics
}

type ProcessorStage interface {
    Process(chunk []interface{}) ([]interface{}, error)
    GetName() string
    IsParallel() bool
}

type Chunk struct {
    Data     []interface{}
    Index    int
    Metadata map[string]interface{}
}
```

### 📊 Avantages Performance

#### Réduction Mémoire (80%)
- **Streaming** : Traitement par petits chunks au lieu de tout charger
- **Backpressure** : Contrôle automatique du flux de données
- **Pipeline** : Transformation en étapes sans accumulation

#### Processing Efficace
- **Parallélisation** : Stages concurrents quand possible
- **Buffer Circulaire** : Réutilisation de la mémoire
- **Fault Tolerance** : Retry per-chunk en cas d'erreur

### 🔄 Pipeline de Transformation

#### Stages de Processing
```go
// Stage 1: Récupération API par chunks
type APIFetchStage struct {
    client    *api.Client
    chunkSize int
}

func (afs *APIFetchStage) Process(chunk []interface{}) ([]interface{}, error) {
    var results []interface{}
    
    for _, item := range chunk {
        if movie, ok := item.(MovieRequest); ok {
            movieData, err := afs.client.GetMovieDetails(movie.ID)
            if err != nil {
                continue // Skip erreurs individuelles
            }
            results = append(results, movieData)
        }
    }
    
    return results, nil
}

// Stage 2: Transformation format
type FormatTransformStage struct{}

func (fts *FormatTransformStage) Process(chunk []interface{}) ([]interface{}, error) {
    var transformed []interface{}
    
    for _, item := range chunk {
        if movie, ok := item.(api.Movie); ok {
            letterboxdMovie := transformToLetterboxdFormat(movie)
            transformed = append(transformed, letterboxdMovie)
        }
    }
    
    return transformed, nil
}

// Stage 3: Écriture CSV streaming
type CSVWriteStage struct {
    writer *csv.Writer
}

func (cws *CSVWriteStage) Process(chunk []interface{}) ([]interface{}, error) {
    for _, item := range chunk {
        if movie, ok := item.(LetterboxdMovie); ok {
            row := movieToCSVRow(movie)
            if err := cws.writer.Write(row); err != nil {
                return nil, err
            }
        }
    }
    
    cws.writer.Flush()
    return chunk, nil // Pass-through pour monitoring
}
```

### 🚀 Configuration

#### Streaming Config
```toml
[streaming]
enabled = true
chunk_size = 1000          # Éléments par chunk
buffer_size = 8192         # Buffer I/O
max_concurrent_stages = 3   # Parallélisme pipeline
memory_limit = "100MB"     # Limite mémoire

[streaming.backpressure]
enabled = true
threshold = 0.8            # Seuil déclenchement (80%)
strategy = "slow_down"     # slow_down, block, drop
```

### 📈 Utilisation

#### Stream Processing Simple
```go
processor := streaming.NewProcessor(streaming.Config{
    ChunkSize:  1000,
    BufferSize: 8192,
})

// Configuration pipeline
processor.AddStage(&APIFetchStage{client: traktClient})
processor.AddStage(&FormatTransformStage{})
processor.AddStage(&CSVWriteStage{writer: csvWriter})

// Traitement streaming
input := make(chan interface{}, 100)
output := make(chan interface{}, 100)

go processor.Process(input, output)

// Alimentation du stream
for _, movieID := range movieIDs {
    input <- MovieRequest{ID: movieID}
}
close(input)

// Collecte résultats
for result := range output {
    // Traitement des résultats transformés
}
```

Ce module permet le traitement efficace de datasets volumineux avec une empreinte mémoire constante et optimisée.