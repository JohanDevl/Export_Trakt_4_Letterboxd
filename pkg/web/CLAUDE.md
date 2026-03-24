# CLAUDE.md - Interface Web et Serveur HTTP

## Module Overview

Ce module fournit une interface web complète avec serveur HTTP, endpoints REST, pagination server-side, gestion des téléchargements, callbacks OAuth et dashboard temps réel pour monitoring et contrôle de l'application.

## Architecture du Module

### 🌐 Serveur Web Principal
```go
type Server struct {
    config       *config.Config
    logger       logger.Logger
    tokenManager *auth.TokenManager
    router       *http.ServeMux
    httpServer   *http.Server
    exportCache  *ExportCache
}

type ExportCache struct {
    exports    []ExportItem
    lastScan   time.Time
    ttl        time.Duration
    mutex      sync.RWMutex
}
```

### 🔗 Endpoints Principaux

#### Dashboard et Navigation
- **`GET /`** : Dashboard principal avec statut et métriques
- **`GET /exports`** : Page d'exports avec pagination
- **`GET /status`** : Statut des services en temps réel
- **`GET /health`** : Health check détaillé

#### Authentification OAuth
- **`GET /auth-url`** : Génération d'URL d'authentification
- **`GET /callback`** : Callback OAuth après autorisation
- **`POST /auth/refresh`** : Rafraîchissement manuel des tokens

#### Exports et Téléchargements  
- **`POST /export`** : Déclenchement d'export via API
- **`GET /download/{file}`** : Téléchargement direct de fichiers
- **`GET /api/exports`** : API JSON des exports avec pagination

### 📄 Pagination Server-Side

#### Système de Pagination Intelligent
```go
type PaginationConfig struct {
    DefaultPageSize int
    MaxPageSize     int
    EnableLazyLoad  bool
    CacheTTL        time.Duration
}

type PaginatedResponse struct {
    Data       []interface{} `json:"data"`
    Page       int           `json:"page"`
    PageSize   int           `json:"page_size"`
    Total      int           `json:"total"`
    TotalPages int           `json:"total_pages"`
    HasNext    bool          `json:"has_next"`
    HasPrev    bool          `json:"has_prev"`
}

func (s *Server) getPaginatedExports(w http.ResponseWriter, r *http.Request) {
    page := getIntParam(r, "page", 1)
    pageSize := getIntParam(r, "page_size", 20)
    
    // Limitation de la taille de page
    if pageSize > s.config.Web.MaxPageSize {
        pageSize = s.config.Web.MaxPageSize
    }
    
    // Récupération avec cache
    allExports, err := s.getCachedExports()
    if err != nil {
        http.Error(w, "Failed to get exports", http.StatusInternalServerError)
        return
    }
    
    // Calcul pagination
    total := len(allExports)
    totalPages := (total + pageSize - 1) / pageSize
    start := (page - 1) * pageSize
    end := start + pageSize
    
    if start > total {
        start = total
    }
    if end > total {
        end = total
    }
    
    response := PaginatedResponse{
        Data:       allExports[start:end],
        Page:       page,
        PageSize:   pageSize,
        Total:      total,
        TotalPages: totalPages,
        HasNext:    page < totalPages,
        HasPrev:    page > 1,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### 📁 Optimisations Scan d'Exports

#### Cache Intelligent avec Lazy Loading
```go
func (s *Server) getCachedExports() ([]ExportItem, error) {
    s.exportCache.mutex.RLock()
    
    // Vérification cache valide
    if time.Since(s.exportCache.lastScan) < s.exportCache.ttl && 
       len(s.exportCache.exports) > 0 {
        exports := s.exportCache.exports
        s.exportCache.mutex.RUnlock()
        return exports, nil
    }
    s.exportCache.mutex.RUnlock()
    
    // Scan nécessaire
    return s.refreshExportCache()
}

func (s *Server) refreshExportCache() ([]ExportItem, error) {
    s.exportCache.mutex.Lock()
    defer s.exportCache.mutex.Unlock()
    
    // Scan prioritaire des exports récents (30 jours)
    recentExports := s.scanRecentExports(30 * 24 * time.Hour)
    
    // Scan paresseux des exports anciens si nécessaire
    var olderExports []ExportItem
    if len(recentExports) < 50 { // Seuil pour scan étendu
        olderExports = s.scanOlderExports(100) // Limite à 100 anciens
    }
    
    allExports := append(recentExports, olderExports...)
    
    // Tri par date (plus récent en premier)
    sort.Slice(allExports, func(i, j int) bool {
        return allExports[i].Date.After(allExports[j].Date)
    })
    
    // Mise à jour cache
    s.exportCache.exports = allExports
    s.exportCache.lastScan = time.Now()
    
    return allExports, nil
}
```

#### Estimation Intelligente CSV
```go
func estimateCSVRecords(filePath string) int {
    info, err := os.Stat(filePath)
    if err != nil {
        return 0
    }
    
    fileSize := info.Size()
    
    // Pour fichiers < 1MB, comptage précis
    if fileSize < 1024*1024 {
        return countCSVRecordsExact(filePath)
    }
    
    // Pour gros fichiers, estimation basée sur la taille
    // Estimation : ~80 caractères par ligne moyenne
    estimatedLines := int(fileSize / 80)
    if estimatedLines > 0 {
        estimatedLines-- // Soustraire header
    }
    
    return estimatedLines
}
```

### 🎨 Interface Utilisateur

#### Dashboard HTML
```html
<!DOCTYPE html>
<html>
<head>
    <title>Export Trakt 4 Letterboxd</title>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        .dashboard { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .status-card { background: #f5f5f5; padding: 20px; margin: 10px 0; border-radius: 8px; }
        .export-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .pagination { display: flex; justify-content: center; margin: 20px 0; }
        .btn { padding: 10px 20px; margin: 5px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="dashboard">
        <h1>🎬 Export Trakt 4 Letterboxd</h1>
        
        {{if .AuthRequired}}
        <div class="status-card">
            <h3>🔐 Authentication Required</h3>
            <p>Please authenticate with Trakt.tv to start using the export features.</p>
            <a href="/auth-url" class="btn">Authenticate Now</a>
        </div>
        {{else}}
        <div class="status-card">
            <h3>✅ Ready to Export</h3>
            <p>Token valid until: {{.TokenExpiry}}</p>
        </div>
        {{end}}
        
        <div class="export-actions">
            <h3>🚀 Quick Export</h3>
            <a href="/export?type=watched" class="btn">Export Watched Movies</a>
            <a href="/export?type=all" class="btn">Export All Data</a>
        </div>
        
        <div class="recent-exports">
            <h3>📁 Recent Exports</h3>
            <div id="exports-container">
                <!-- Chargement dynamique via JavaScript -->
            </div>
            <div class="pagination" id="pagination">
                <!-- Pagination dynamique -->
            </div>
        </div>
    </div>
    
    <script src="/static/js/dashboard.js"></script>
</body>
</html>
```

#### JavaScript Frontend
```javascript
class ExportDashboard {
    constructor() {
        this.currentPage = 1;
        this.pageSize = 10;
        this.loadExports();
    }
    
    async loadExports(page = 1) {
        try {
            const response = await fetch(`/api/exports?page=${page}&page_size=${this.pageSize}`);
            const data = await response.json();
            
            this.renderExports(data.data);
            this.renderPagination(data);
            
        } catch (error) {
            console.error('Failed to load exports:', error);
            this.showError('Failed to load export data');
        }
    }
    
    renderExports(exports) {
        const container = document.getElementById('exports-container');
        
        if (exports.length === 0) {
            container.innerHTML = `
                <div class="no-exports">
                    <p>📭 No exports found</p>
                    <p>Create your first export using the buttons above!</p>
                </div>
            `;
            return;
        }
        
        const exportsHTML = exports.map(exp => `
            <div class="export-item">
                <h4>${exp.type_icon} ${exp.type_name}</h4>
                <div class="export-meta">
                    <span>📅 ${exp.date}</span>
                    <span>⏱️ ${exp.duration}</span>
                    <span>💾 ${exp.file_size}</span>
                    <span>📊 ${exp.record_count} records</span>
                </div>
                <div class="export-actions">
                    ${exp.files.map(file => 
                        `<a href="/download/${exp.dir_name}/${file}" class="btn btn-sm">📥 ${file}</a>`
                    ).join('')}
                </div>
            </div>
        `).join('');
        
        container.innerHTML = exportsHTML;
    }
    
    renderPagination(data) {
        const container = document.getElementById('pagination');
        
        let paginationHTML = '';
        
        if (data.has_prev) {
            paginationHTML += `<button onclick="dashboard.loadExports(${data.page - 1})" class="btn">← Previous</button>`;
        }
        
        paginationHTML += `<span class="page-info">Page ${data.page} of ${data.total_pages}</span>`;
        
        if (data.has_next) {
            paginationHTML += `<button onclick="dashboard.loadExports(${data.page + 1})" class="btn">Next →</button>`;
        }
        
        container.innerHTML = paginationHTML;
    }
}

// Initialisation
const dashboard = new ExportDashboard();

// Auto-refresh toutes les 30 secondes
setInterval(() => {
    dashboard.loadExports(dashboard.currentPage);
}, 30000);
```

### 🔧 Configuration Web

#### Configuration Serveur
```toml
[web]
enabled = true
port = 8080
host = "0.0.0.0"
read_timeout = "30s"
write_timeout = "30s"
max_page_size = 50
default_page_size = 20
cache_ttl = "5m"

[web.static]
enabled = true
path = "/static/"
directory = "./web/static"

[web.exports]
scan_recent_days = 30
max_older_exports = 100
estimate_threshold = "1MB"
```

### 🚀 Fonctionnalités Avancées

#### WebSocket pour Updates Temps Réel
```go
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
    conn, err := websocket.Upgrade(w, r, nil)
    if err != nil {
        s.logger.Error("websocket_upgrade_failed", map[string]interface{}{
            "error": err.Error(),
        })
        return
    }
    defer conn.Close()
    
    // Envoi du statut initial
    status := s.getSystemStatus()
    conn.WriteJSON(status)
    
    // Updates périodiques
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            status := s.getSystemStatus()
            if err := conn.WriteJSON(status); err != nil {
                return
            }
        }
    }
}
```

#### Déclenchement d'Export via API
```go
func (s *Server) handleExportAPI(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var req ExportRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    // Déclenchement export en arrière-plan
    go func() {
        s.logger.Info("export.api_triggered", map[string]interface{}{
            "type": req.Type,
            "mode": req.Mode,
        })
        
        // Logique d'export...
        err := s.performExport(req.Type, req.Mode)
        if err != nil {
            s.logger.Error("export.api_failed", map[string]interface{}{
                "error": err.Error(),
            })
        }
    }()
    
    response := map[string]interface{}{
        "status":  "started",
        "message": "Export started in background",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### 📊 Métriques Web

#### Endpoints de Monitoring
- **`/metrics`** : Métriques Prometheus
- **`/debug/pprof/`** : Profiling Go
- **`/api/stats`** : Statistiques JSON

#### Performance Web
- **Lazy Loading** : Chargement différé des exports anciens
- **Client-Side Caching** : Cache JavaScript pour navigation fluide  
- **Server-Side Pagination** : Traitement optimisé côté serveur
- **Response Compression** : Gzip automatique
- **Static Assets** : Serving optimisé des ressources statiques

Ce module fournit une interface web moderne et performante avec pagination intelligente, téléchargements optimisés et monitoring temps réel pour une expérience utilisateur exceptionnelle.