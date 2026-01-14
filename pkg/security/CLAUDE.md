# CLAUDE.md - Sécurité, Chiffrement et Audit

## Module Overview

Ce module implémente une couche de sécurité complète avec chiffrement AES-256, gestion de keyring multi-backend, audit logging, validation d'entrées, et protection contre les vulnérabilités courantes.

## Architecture du Module

### 🛡️ Security Manager
```go
type Manager struct {
    config      Config
    encryption  *encryption.AESManager
    keyring     *keyring.Manager
    validator   *validation.Validator
    auditLogger *audit.Logger
    rateLimiter *RateLimiter
}

type Config struct {
    SecurityLevel        string `toml:"security_level"`        // low, medium, high
    EncryptionEnabled    bool   `toml:"encryption_enabled"`
    KeyringBackend      string `toml:"keyring_backend"`       // system, env, file, memory
    RequireHTTPS        bool   `toml:"require_https"`
    AuditLogging        bool   `toml:"audit_logging"`
    RateLimitEnabled    bool   `toml:"rate_limit_enabled"`
}
```

### 🔐 Chiffrement AES-256

#### AES Manager
```go
type AESManager struct {
    key    []byte
    cipher cipher.Block
}

func (a *AESManager) EncryptData(plaintext string) (string, error) {
    // Chiffrement AES-256-GCM avec nonce aléatoire
    gcm, _ := cipher.NewGCM(a.cipher)
    nonce := make([]byte, gcm.NonceSize())
    rand.Read(nonce)
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a *AESManager) DecryptData(ciphertext string) (string, error) {
    // Déchiffrement avec validation d'intégrité
    data, _ := base64.StdEncoding.DecodeString(ciphertext)
    gcm, _ := cipher.NewGCM(a.cipher)
    
    nonceSize := gcm.NonceSize()
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    return string(plaintext), err
}
```

### 🔑 Gestion Keyring Multi-Backend

#### Backends Supportés
```go
type Backend int

const (
    SystemBackend Backend = iota  // Keyring système natif
    EnvBackend                    // Variables d'environnement
    FileBackend                   // Fichier chiffré
    MemoryBackend                 // Mémoire (non persistant)
)

type Manager struct {
    backend Backend
    store   KeyStore
    options []Option
}
```

#### System Backend (Recommandé)
- **macOS** : Keychain
- **Windows** : Credential Manager
- **Linux** : Secret Service (GNOME/KDE)

#### File Backend (Chiffré)
```go
type EncryptedFileStore struct {
    filePath   string
    encryption *AESManager
    data       map[string]string
}

func (e *EncryptedFileStore) Set(key, value string) error {
    encrypted, _ := e.encryption.EncryptData(value)
    e.data[key] = encrypted
    return e.writeToFile()
}
```

### 📝 Audit Logging

#### Audit Logger
```go
type Logger struct {
    config     AuditConfig
    file       *os.File
    encoder    json.Encoder
    rotator    *FileRotator
}

type AuditEvent struct {
    Timestamp   time.Time              `json:"timestamp"`
    EventType   string                 `json:"event_type"`
    UserID      string                 `json:"user_id,omitempty"`
    IPAddress   string                 `json:"ip_address,omitempty"`
    Action      string                 `json:"action"`
    Resource    string                 `json:"resource,omitempty"`
    Result      string                 `json:"result"`
    Details     map[string]interface{} `json:"details,omitempty"`
    Severity    string                 `json:"severity"`
}
```

#### Types d'Événements Audités
- **AUTH** : Authentification, token refresh, logout
- **API** : Appels API, rate limiting, erreurs
- **DATA** : Accès aux données, exports, modifications
- **SYSTEM** : Démarrage/arrêt, configuration, erreurs système
- **SECURITY** : Tentatives d'intrusion, violations de sécurité

### 🔍 Validation et Sanitisation

#### Input Validator
```go
type Validator struct {
    config ValidationConfig
}

func (v *Validator) SanitizeInput(input string) string {
    // Suppression des caractères dangereux
    input = html.EscapeString(input)
    
    // Protection XSS
    input = strings.ReplaceAll(input, "<script", "&lt;script")
    input = strings.ReplaceAll(input, "javascript:", "")
    
    // Nettoyage des caractères de contrôle
    return strings.Map(func(r rune) rune {
        if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
            return -1
        }
        return r
    }, input)
}

func (v *Validator) ValidateFilePath(path string) error {
    // Protection path traversal
    if strings.Contains(path, "..") {
        return fmt.Errorf("path traversal attempt detected")
    }
    
    if strings.HasPrefix(path, "/etc/") || strings.HasPrefix(path, "/sys/") {
        return fmt.Errorf("access to system directory denied")
    }
    
    return nil
}
```

### 🚦 Rate Limiting

#### Rate Limiter
```go
type RateLimiter struct {
    requests map[string]*bucket
    mutex    sync.RWMutex
    config   RateLimitConfig
}

type bucket struct {
    tokens    int
    lastRefill time.Time
    capacity  int
}

func (rl *RateLimiter) Allow(identifier string) bool {
    rl.mutex.Lock()
    defer rl.mutex.Unlock()
    
    bucket := rl.getBucket(identifier)
    
    // Refill tokens basé sur le temps écoulé
    now := time.Now()
    elapsed := now.Sub(bucket.lastRefill)
    tokensToAdd := int(elapsed.Seconds() * float64(rl.config.RefillRate))
    
    bucket.tokens = min(bucket.capacity, bucket.tokens + tokensToAdd)
    bucket.lastRefill = now
    
    if bucket.tokens > 0 {
        bucket.tokens--
        return true
    }
    
    return false
}
```

### 🌐 Protection HTTPS

#### HTTPS Enforcer
```go
type HTTPSConfig struct {
    Required    bool   `toml:"required"`
    MinTLSVersion string `toml:"min_tls_version"`
    CertFile    string `toml:"cert_file"`
    KeyFile     string `toml:"key_file"`
}

func (h *HTTPSEnforcer) ValidateURL(url string) error {
    if h.config.Required && !strings.HasPrefix(url, "https://") {
        return fmt.Errorf("HTTPS required but HTTP URL provided: %s", url)
    }
    return nil
}

func (h *HTTPSEnforcer) ConfigureTLS() *tls.Config {
    return &tls.Config{
        MinVersion: h.getTLSVersion(),
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
        },
    }
}
```

### 📊 Niveaux de Sécurité

#### Configuration par Niveau
```go
func (c *Config) SecurityLevel() string {
    score := 0
    
    if c.EncryptionEnabled { score += 25 }
    if c.AuditLogging { score += 20 }
    if c.RequireHTTPS { score += 20 }
    if c.RateLimitEnabled { score += 15 }
    if c.KeyringBackend == "system" { score += 20 }
    
    switch {
    case score >= 80: return "high"
    case score >= 50: return "medium"
    default: return "low"
    }
}
```

- **LOW** : Sécurité minimale, development only
- **MEDIUM** : Sécurité standard, staging/test
- **HIGH** : Sécurité maximale, production

### 🔧 Configuration

#### Configuration Complète
```toml
[security]
security_level = "high"
encryption_enabled = true
encryption_algorithm = "AES-256-GCM"
keyring_backend = "system"
require_https = true
tls_min_version = "1.2"
audit_logging = true
log_sensitive_data = false
input_sanitization = true
path_traversal_protection = true
rate_limit_enabled = true
rate_limit_requests_per_second = 100

[security.audit]
enabled = true
file_path = "./logs/audit.log"
max_file_size = "100MB"
retention_days = 90
include_sensitive = false
compress_old_logs = true

[security.encryption]
key_rotation_days = 30
key_derivation_iterations = 100000

[security.validation]
max_input_length = 10000
allowed_file_extensions = [".csv", ".json", ".toml"]
blocked_patterns = ["<script", "javascript:", "data:"]
```

### 🛠️ Validation Sécurité

#### Security Validator
```go
func ValidateSecurityConfiguration(cfg *Config, log logger.Logger) int {
    var errors []string
    var warnings []string
    
    // Tests de chiffrement
    if cfg.EncryptionEnabled {
        testData := "test-encryption-data"
        encrypted, err := mgr.EncryptData(testData)
        if err == nil {
            decrypted, err := mgr.DecryptData(encrypted)
            if err != nil || decrypted != testData {
                errors = append(errors, "Encryption test failed")
            }
        }
    }
    
    // Tests de validation
    maliciousInput := "<script>alert('xss')</script>"
    sanitized := mgr.SanitizeInput(maliciousInput)
    if sanitized == maliciousInput {
        warnings = append(warnings, "Input sanitization not working")
    }
    
    // Tests de path traversal
    maliciousPath := "../../../etc/passwd"
    if err := mgr.ValidateFilePath(maliciousPath); err == nil {
        errors = append(errors, "Path traversal protection not working")
    }
    
    return len(errors) // 0 = succès
}
```

### 📚 Usage

#### Initialisation Complète
```go
// Configuration sécurité
securityCfg := security.Config{
    SecurityLevel:     "high",
    EncryptionEnabled: true,
    KeyringBackend:    "system",
    RequireHTTPS:      true,
    AuditLogging:      true,
    RateLimitEnabled:  true,
}

// Création du manager
securityMgr, err := security.NewManager(securityCfg)
if err != nil {
    log.Fatal("Failed to create security manager:", err)
}
defer securityMgr.Close()

// Validation complète
exitCode := security.ValidateSecurityConfiguration(&securityCfg, log)
if exitCode != 0 {
    log.Fatal("Security validation failed")
}
```

#### Usage Pratique
```go
// Chiffrement de données sensibles
encrypted, _ := securityMgr.EncryptData("sensitive-token")
decrypted, _ := securityMgr.DecryptData(encrypted)

// Validation d'entrées
cleaned := securityMgr.SanitizeInput(userInput)
err := securityMgr.ValidateFilePath(filePath)

// Audit logging
securityMgr.LogAuditEvent("AUTH", "token_refresh", "success", details)

// Rate limiting
if !securityMgr.RateLimit(clientIP) {
    return errors.New("rate limit exceeded")
}
```

Ce module fournit une sécurité de niveau entreprise avec chiffrement robuste, audit complet et protection multicouche contre les vulnérabilités courantes.