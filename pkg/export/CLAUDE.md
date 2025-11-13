# CLAUDE.md - Export CSV Letterboxd

## Module Overview

Ce module gère l'exportation des données Trakt.tv vers le format CSV compatible Letterboxd. Il prend en charge les exports de films regardés, collections, séries TV, notes et watchlists avec formatage adapté à l'import Letterboxd et gestion intelligente des répertoires d'export.

## Architecture du Module

### 🏗️ Composant Principal

#### LetterboxdExporter
```go
type LetterboxdExporter struct {
    config *config.Config
    log    logger.Logger
}
```

Responsabilités :
- **Export CSV Letterboxd** : Conversion des données Trakt vers format compatible Letterboxd
- **Gestion des Répertoires** : Création automatique de dossiers d'export horodatés
- **Formats de Dates** : Application des fuseaux horaires configurés
- **Validation des Données** : Nettoyage et formatage des champs obligatoires

### 📁 Gestion des Répertoires d'Export

#### Stratégie de Nommage Intelligente
```go
func (e *LetterboxdExporter) getExportDir() (string, error) {
    // Détection des environnements de test
    isTestDir := strings.Contains(e.config.Letterboxd.ExportDir, "test") ||
                 strings.Contains(e.config.Letterboxd.ExportDir, "tmp") ||
                 strings.Contains(e.config.Letterboxd.ExportDir, "/var/folders/")
    
    if isTestDir {
        // Utilisation directe pour les tests
        return e.config.Letterboxd.ExportDir, nil
    }
    
    // Création de sous-répertoire horodaté pour production
    now := e.getTimeInConfigTimezone()
    dirName := fmt.Sprintf("export_%s_%s", 
        now.Format("2006-01-02"),
        now.Format("15-04"))
    
    exportDir := filepath.Join(e.config.Letterboxd.ExportDir, dirName)
    return exportDir, os.MkdirAll(exportDir, 0755)
}
```

#### Exemples de Structure d'Export
```
exports/
├── export_2025-07-11_15-43/
│   ├── watched.csv
│   ├── collection.csv
│   ├── shows.csv
│   ├── ratings.csv
│   └── watchlist.csv
├── export_2025-07-12_09-24/
│   └── watched.csv
└── README.md
```

### 🕒 Gestion des Fuseaux Horaires

#### Support Multi-Timezone
```go
func (e *LetterboxdExporter) getTimeInConfigTimezone() time.Time {
    now := time.Now().UTC()
    
    if e.config.Export.Timezone == "" {
        return now // Fallback UTC
    }
    
    loc, err := time.LoadLocation(e.config.Export.Timezone)
    if err != nil {
        e.log.Warn("timezone_load_failed", map[string]interface{}{
            "timezone": e.config.Export.Timezone,
            "error": err.Error(),
        })
        return now
    }
    
    return now.In(loc)
}
```

#### Fuseaux Horaires Supportés
- **UTC** (par défaut)
- **Europe/Paris**
- **America/New_York**
- **Asia/Tokyo**
- Tous les fuseaux IANA standard

### 📊 Types d'Export Supportés

#### 1. Films Regardés (Mode Agrégé)
```go
func (e *LetterboxdExporter) ExportMovies(movies []api.Movie, client *api.Client) error
```

**Format CSV Letterboxd :**
```csv
Title,Year,WatchedDate,Rating10,imdbID,tmdbID,Rewatch
Cars,2006,2025-07-10,7,tt0317219,920,false
The Matrix,1999,2025-07-09,9,tt0133093,603,false
```

**Champs :**
- `Title` : Titre du film
- `Year` : Année de sortie
- `WatchedDate` : Date du dernier visionnage (format YYYY-MM-DD)
- `Rating10` : Note sur 10 (conversion depuis Trakt /10)
- `imdbID` : Identifiant IMDb
- `tmdbID` : Identifiant TMDB
- `Rewatch` : Indicateur de re-visionnage

#### 2. Historique Individuel (Mode Complet)
```go
func (e *LetterboxdExporter) ExportMovieHistory(history []api.HistoryItem, client *api.Client) error
```

**Caractéristiques :**
- **Une entrée par visionnage** : Historique complet de tous les events
- **Chronologie Précise** : Ordre chronologique des visionnages
- **Re-visionnages Trackés** : Première écoute = false, suivantes = true
- **Ratings Cohérents** : Même note appliquée à tous les visionnages d'un film

**Exemple pour un film revu :**
```csv
Title,Year,WatchedDate,Rating10,imdbID,tmdbID,Rewatch
Cars,2006,2025-07-10,7,tt0317219,920,true
Cars,2006,2024-12-01,7,tt0317219,920,false
```

#### 3. Collection Personnelle
```go
func (e *LetterboxdExporter) ExportCollectionMovies(movies []api.CollectionMovie) error
```

**Format :**
```csv
Title,Year,CollectedDate,imdbID,tmdbID
Blade Runner,1982,2025-01-15,tt0083658,78
```

#### 4. Séries TV
```go
func (e *LetterboxdExporter) ExportShows(shows []api.WatchedShow) error
```

**Format Adapté :**
```csv
ShowTitle,Year,LastWatchedDate,Seasons,Episodes,imdbID,tmdbID
Breaking Bad,2008,2025-07-10,5,62,tt0903747,1396
```

#### 5. Notes/Ratings
```go
func (e *LetterboxdExporter) ExportRatings(ratings []api.Rating) error
```

**Format :**
```csv
Title,Year,Rating10,RatedDate,imdbID,tmdbID
Inception,2010,9,2025-07-08,tt1375666,27205
```

#### 6. Watchlist
```go
func (e *LetterboxdExporter) ExportWatchlist(watchlist []api.WatchlistMovie) error
```

**Format :**
```csv
Title,Year,AddedDate,imdbID,tmdbID
Dune,2021,2025-07-05,tt1160419,438631
```

### 🔧 Logique de Formatage

#### Conversion des Données Trakt → Letterboxd
```go
func formatMovieForLetterboxd(movie api.Movie, client *api.Client) LetterboxdMovie {
    return LetterboxdMovie{
        Title:       cleanTitle(movie.Movie.Title),
        Year:        movie.Movie.Year,
        WatchedDate: formatDate(movie.LastWatchedAt),
        Rating10:    convertRating(movie.Rating),
        ImdbID:      movie.Movie.IDs.IMDB,
        TmdbID:      strconv.Itoa(movie.Movie.IDs.TMDB),
        Rewatch:     movie.Plays > 1,
    }
}
```

#### Nettoyage des Titres
```go
func cleanTitle(title string) string {
    // Suppression des caractères problématiques pour CSV
    title = strings.ReplaceAll(title, `"`, `""`)
    title = strings.TrimSpace(title)
    
    // Gestion des caractères spéciaux
    title = strings.ReplaceAll(title, "\n", " ")
    title = strings.ReplaceAll(title, "\r", " ")
    
    return title
}
```

#### Conversion des Dates
```go
func formatDate(dateStr string) string {
    // Parse la date Trakt (ISO 8601)
    parsedTime, err := time.Parse(time.RFC3339, dateStr)
    if err != nil {
        // Tentative avec format alternatif
        parsedTime, err = time.Parse("2006-01-02T15:04:05Z", dateStr)
        if err != nil {
            return "" // Date invalide
        }
    }
    
    // Format Letterboxd (YYYY-MM-DD)
    return parsedTime.Format("2006-01-02")
}
```

#### Conversion des Notes
```go
func convertRating(traktRating float64) string {
    if traktRating == 0 {
        return "" // Pas de note
    }
    
    // Trakt utilise une échelle de 1-10
    // Letterboxd utilise aussi 1-10 mais avec demi-points
    rating := int(traktRating)
    if rating < 1 {
        rating = 1
    }
    if rating > 10 {
        rating = 10
    }
    
    return strconv.Itoa(rating)
}
```

### 📝 Écriture CSV Optimisée

#### Writer CSV avec Headers
```go
func (e *LetterboxdExporter) writeCSV(filename string, headers []string, data [][]string) error {
    file, err := os.Create(filename)
    if err != nil {
        return fmt.Errorf("failed to create file: %w", err)
    }
    defer file.Close()
    
    writer := csv.NewWriter(file)
    defer writer.Flush()
    
    // Écriture des en-têtes
    if err := writer.Write(headers); err != nil {
        return fmt.Errorf("failed to write headers: %w", err)
    }
    
    // Écriture des données
    for _, row := range data {
        if err := writer.Write(row); err != nil {
            return fmt.Errorf("failed to write row: %w", err)
        }
    }
    
    return nil
}
```

#### Gestion des Caractères Spéciaux
```go
func escapeCSVField(field string) string {
    // Échappement des guillemets
    if strings.Contains(field, `"`) {
        field = strings.ReplaceAll(field, `"`, `""`)
    }
    
    // Encapsulation si nécessaire
    if strings.ContainsAny(field, ",\n\r\"") {
        field = fmt.Sprintf(`"%s"`, field)
    }
    
    return field
}
```

### 🔍 Validation et Nettoyage

#### Validation des Champs Obligatoires
```go
func validateMovie(movie api.Movie) error {
    if movie.Movie.Title == "" {
        return fmt.Errorf("movie title is required")
    }
    
    if movie.Movie.Year == 0 {
        return fmt.Errorf("movie year is required")
    }
    
    if movie.LastWatchedAt == "" {
        return fmt.Errorf("watched date is required")
    }
    
    return nil
}
```

#### Déduplication pour Mode Agrégé
```go
func deduplicateMovies(movies []api.Movie) []api.Movie {
    seen := make(map[string]*api.Movie)
    
    for _, movie := range movies {
        key := fmt.Sprintf("%s-%d", movie.Movie.Title, movie.Movie.Year)
        
        existing, exists := seen[key]
        if !exists || movie.LastWatchedAt > existing.LastWatchedAt {
            seen[key] = &movie
        }
    }
    
    result := make([]api.Movie, 0, len(seen))
    for _, movie := range seen {
        result = append(result, *movie)
    }
    
    return result
}
```

### 📊 Tri et Organisation

#### Tri Chronologique
```go
func sortMoviesByWatchedDate(movies []api.Movie) {
    sort.Slice(movies, func(i, j int) bool {
        dateI, _ := time.Parse(time.RFC3339, movies[i].LastWatchedAt)
        dateJ, _ := time.Parse(time.RFC3339, movies[j].LastWatchedAt)
        return dateI.After(dateJ) // Plus récent en premier
    })
}
```

#### Tri Alphabétique
```go
func sortMoviesByTitle(movies []api.Movie) {
    sort.Slice(movies, func(i, j int) bool {
        return strings.ToLower(movies[i].Movie.Title) < strings.ToLower(movies[j].Movie.Title)
    })
}
```

### 🛡️ Gestion d'Erreurs

#### Récupération sur Erreurs de Données
```go
func (e *LetterboxdExporter) exportWithErrorRecovery(movies []api.Movie) error {
    var validMovies []api.Movie
    var errorCount int
    
    for _, movie := range movies {
        if err := validateMovie(movie); err != nil {
            e.log.Warn("export.invalid_movie_skipped", map[string]interface{}{
                "title": movie.Movie.Title,
                "error": err.Error(),
            })
            errorCount++
            continue
        }
        validMovies = append(validMovies, movie)
    }
    
    if errorCount > 0 {
        e.log.Info("export.movies_skipped", map[string]interface{}{
            "skipped": errorCount,
            "total": len(movies),
            "exported": len(validMovies),
        })
    }
    
    return e.exportValidMovies(validMovies)
}
```

### 📚 Exemples d'Usage

#### Export Simple
```go
// Configuration
cfg := &config.Config{
    Letterboxd: config.LetterboxdConfig{
        ExportDir: "./exports",
        WatchedFilename: "watched.csv",
    },
    Export: config.ExportConfig{
        DateFormat: "2006-01-02",
        Timezone: "Europe/Paris",
    },
}

// Initialisation
exporter := export.NewLetterboxdExporter(cfg, log)

// Export des films regardés
movies, _ := traktClient.GetWatchedMovies()
err := exporter.ExportMovies(movies, traktClient)
```

#### Export Complet Multi-Types
```go
// Export de tous les types de données
exporter := export.NewLetterboxdExporter(cfg, log)

// Films regardés
movies, _ := traktClient.GetWatchedMovies()
exporter.ExportMovies(movies, traktClient)

// Collection
collection, _ := traktClient.GetCollectionMovies()
exporter.ExportCollectionMovies(collection)

// Notes
ratings, _ := traktClient.GetRatings()
exporter.ExportRatings(ratings)

// Watchlist
watchlist, _ := traktClient.GetWatchlist()
exporter.ExportWatchlist(watchlist)

// Séries
shows, _ := traktClient.GetWatchedShows()
exporter.ExportShows(shows)
```

#### Export Historique Individuel
```go
// Mode historique complet (tous les visionnages)
history, _ := traktClient.GetMovieHistory()
err := exporter.ExportMovieHistory(history, traktClient)

// Résultat : un fichier CSV avec une ligne par visionnage
// Permet le tracking précis des re-visionnages sur Letterboxd
```

### ⚙️ Configuration et Personnalisation

#### Configuration TOML
```toml
[letterboxd]
export_dir = "./exports"
watched_filename = "watched.csv"
collection_filename = "collection.csv"
shows_filename = "shows.csv"
ratings_filename = "ratings.csv"
watchlist_filename = "watchlist.csv"

[export]
format = "csv"
date_format = "2006-01-02"
timezone = "Europe/Paris"
history_mode = "individual"  # ou "aggregated"
```

#### Noms de Fichiers Dynamiques
- **Timestamp** : Intégration automatique de l'horodatage
- **Type Detection** : Détection des environnements de test
- **Fallback** : Noms par défaut si non configurés

### 🚀 Optimisations et Performance

#### Traitement par Lots
- **Buffering** : Écriture par chunks pour gros volumes
- **Memory Efficient** : Traitement streaming pour éviter l'OOM
- **Parallel Processing** : Traitement concurrent des différents types

#### Caching
- **Metadata Cache** : Cache des informations de films pour éviter les doublons
- **Rating Cache** : Cache des notes pour corrélation multi-exports

Ce module fournit une conversion robuste et optimisée des données Trakt.tv vers le format CSV compatible Letterboxd, avec gestion intelligente des cas d'erreur et support complet des différents modes d'historique.