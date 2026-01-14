# CLAUDE.md - Internationalisation Multilingue

## Module Overview

Ce module fournit un système d'internationalisation complet avec support de 4 langues, traduction contextuelle, pluralisation, formatage localisé et intégration transparente avec le système de logging.

## Architecture du Module

### 🌍 Translator Principal
```go
type Translator struct {
    defaultLanguage string
    currentLanguage string
    messages        map[string]map[string]string
    pluralRules     map[string]PluralRule
    formatters      map[string]Formatter
    mutex           sync.RWMutex
}

type TranslationData struct {
    Language string                 `json:"language"`
    Messages map[string]string      `json:"messages"`
    Plurals  map[string]PluralForms `json:"plurals"`
}

type PluralForms struct {
    Zero  string `json:"zero,omitempty"`
    One   string `json:"one"`
    Two   string `json:"two,omitempty"`
    Few   string `json:"few,omitempty"`
    Many  string `json:"many,omitempty"`
    Other string `json:"other"`
}
```

### 🗣️ Langues Supportées

#### Langues Disponibles
- **🇺🇸 en** : English (par défaut)
- **🇫🇷 fr** : Français
- **🇩🇪 de** : Deutsch
- **🇪🇸 es** : Español

#### Structure des Fichiers
```
locales/
├── en.json              # English (langue de référence)
├── fr.json              # Français  
├── de.json              # Allemand
├── es.json              # Espagnol
└── README.md            # Guide de traduction
```

### 📝 Système de Clés

#### Hiérarchie des Clés
```json
{
  "app": {
    "name": "Export Trakt 4 Letterboxd",
    "description": "Export your Trakt.tv data to Letterboxd format",
    "version": "Version {{version}}"
  },
  "export": {
    "starting": "Starting export of {{type}} data",
    "completed": "Export completed: {{count}} items in {{duration}}",
    "failed": "Export failed: {{error}}",
    "progress": "Processing {{current}} of {{total}} items"
  },
  "auth": {
    "required": "Authentication required",
    "success": "Authentication successful",
    "token_expired": "Token expired, refreshing...",
    "failed": "Authentication failed: {{reason}}"
  },
  "errors": {
    "network": "Network error: {{details}}",
    "api_limit": "API rate limit exceeded. Retry in {{seconds}} seconds",
    "file_write": "Failed to write file: {{filename}}"
  }
}
```

### 🔄 Traduction Contextuelle

#### Méthode Translate
```go
func (t *Translator) Translate(key string, data map[string]interface{}) string {
    t.mutex.RLock()
    defer t.mutex.RUnlock()
    
    // Récupération du message dans la langue courante
    if langMessages, exists := t.messages[t.currentLanguage]; exists {
        if message, exists := langMessages[key]; exists {
            return t.interpolate(message, data)
        }
    }
    
    // Fallback vers la langue par défaut
    if defaultMessages, exists := t.messages[t.defaultLanguage]; exists {
        if message, exists := defaultMessages[key]; exists {
            return t.interpolate(message, data)
        }
    }
    
    // Fallback vers la clé elle-même
    return key
}

func (t *Translator) interpolate(template string, data map[string]interface{}) string {
    if data == nil {
        return template
    }
    
    result := template
    for key, value := range data {
        placeholder := fmt.Sprintf("{{%s}}", key)
        replacement := fmt.Sprintf("%v", value)
        result = strings.ReplaceAll(result, placeholder, replacement)
    }
    
    return result
}
```

### 📊 Pluralisation

#### Règles de Pluralisation
```go
type PluralRule func(n int) PluralForm

type PluralForm int

const (
    PluralZero PluralForm = iota
    PluralOne
    PluralTwo
    PluralFew
    PluralMany
    PluralOther
)

// Règle anglaise/française (simple)
func EnglishPluralRule(n int) PluralForm {
    if n == 0 {
        return PluralZero
    } else if n == 1 {
        return PluralOne
    }
    return PluralOther
}

// Règle allemande 
func GermanPluralRule(n int) PluralForm {
    if n == 1 {
        return PluralOne
    }
    return PluralOther
}

// Règle espagnole
func SpanishPluralRule(n int) PluralForm {
    if n == 1 {
        return PluralOne
    }
    return PluralOther
}
```

#### Usage Pluralisation
```go
func (t *Translator) TranslatePlural(key string, count int, data map[string]interface{}) string {
    pluralRule := t.pluralRules[t.currentLanguage]
    form := pluralRule(count)
    
    pluralKey := fmt.Sprintf("%s.%s", key, form.String())
    
    if data == nil {
        data = make(map[string]interface{})
    }
    data["count"] = count
    
    return t.Translate(pluralKey, data)
}

// Exemples de clés plurielles
{
  "items": {
    "zero": "No items",
    "one": "{{count}} item", 
    "other": "{{count}} items"
  },
  "movies_exported": {
    "zero": "Aucun film exporté",
    "one": "{{count}} film exporté",
    "other": "{{count}} films exportés"
  }
}
```

### 🕒 Formatage Localisé

#### Formatters par Langue
```go
type Formatter interface {
    FormatNumber(n int) string
    FormatFloat(f float64, precision int) string
    FormatDate(t time.Time) string
    FormatDuration(d time.Duration) string
}

type EnglishFormatter struct{}

func (ef EnglishFormatter) FormatNumber(n int) string {
    return fmt.Sprintf("%d", n)
}

func (ef EnglishFormatter) FormatDate(t time.Time) string {
    return t.Format("January 2, 2006")
}

type FrenchFormatter struct{}

func (ff FrenchFormatter) FormatNumber(n int) string {
    // Format français avec espaces pour milliers
    str := fmt.Sprintf("%d", n)
    if len(str) > 3 {
        // Insertion d'espaces pour les milliers
        return insertSpaces(str)
    }
    return str
}

func (ff FrenchFormatter) FormatDate(t time.Time) string {
    months := []string{
        "janvier", "février", "mars", "avril", "mai", "juin",
        "juillet", "août", "septembre", "octobre", "novembre", "décembre",
    }
    return fmt.Sprintf("%d %s %d", t.Day(), months[t.Month()-1], t.Year())
}
```

### 📁 Fichiers de Traduction

#### en.json (Référence)
```json
{
  "app": {
    "name": "Export Trakt 4 Letterboxd",
    "description": "Export your Trakt.tv data to Letterboxd format"
  },
  "export": {
    "starting": "Starting export of {{type}} data",
    "completed": "Export completed: {{count}} items in {{duration}}",
    "movies_exported": {
      "zero": "No movies exported",
      "one": "{{count}} movie exported",
      "other": "{{count}} movies exported"
    }
  },
  "auth": {
    "required": "Authentication required",
    "success": "Authentication successful",
    "token_expires": "Token expires in {{duration}}"
  }
}
```

#### fr.json (Français)
```json
{
  "app": {
    "name": "Export Trakt 4 Letterboxd",
    "description": "Exportez vos données Trakt.tv vers le format Letterboxd"
  },
  "export": {
    "starting": "Démarrage de l'export des données {{type}}",
    "completed": "Export terminé : {{count}} éléments en {{duration}}",
    "movies_exported": {
      "zero": "Aucun film exporté",
      "one": "{{count}} film exporté",
      "other": "{{count}} films exportés"
    }
  },
  "auth": {
    "required": "Authentification requise",
    "success": "Authentification réussie", 
    "token_expires": "Le token expire dans {{duration}}"
  }
}
```

### ⚙️ Configuration

#### Configuration i18n
```toml
[i18n]
default_language = "en"      # Langue par défaut
language = "fr"              # Langue active
locales_dir = "./locales"    # Répertoire des traductions
auto_detect = false          # Détection auto langue système
fallback_enabled = true      # Fallback vers langue par défaut
```

#### Variables d'Environnement
```bash
I18N_LANGUAGE=fr
I18N_LOCALES_DIR=./locales
I18N_AUTO_DETECT=false
```

### 🚀 Utilisation

#### Initialisation
```go
// Configuration
cfg := &config.I18nConfig{
    DefaultLanguage: "en",
    Language:        "fr",
    LocalesDir:      "./locales",
}

// Création translator
translator, err := i18n.NewTranslator(cfg, log)
if err != nil {
    log.Fatal("Failed to initialize translator:", err)
}

// Intégration avec logger
log.SetTranslator(translator)
```

#### Traduction Simple
```go
// Traduction basique
message := translator.Translate("auth.success", nil)
// EN: "Authentication successful"
// FR: "Authentification réussie"

// Traduction avec variables
message := translator.Translate("export.completed", map[string]interface{}{
    "count":    150,
    "duration": "2.3s",
})
// EN: "Export completed: 150 items in 2.3s"
// FR: "Export terminé : 150 éléments en 2.3s"
```

#### Traduction avec Pluralisation
```go
// Pluralisation
message := translator.TranslatePlural("movies_exported", count, nil)

// count = 0 => FR: "Aucun film exporté"
// count = 1 => FR: "1 film exporté"  
// count = 5 => FR: "5 films exportés"
```

#### Intégration avec Logging
```go
// Logs traduits automatiquement
log.Info("export.starting", map[string]interface{}{
    "type": "movies",
})
// EN: "Starting export of movies data"
// FR: "Démarrage de l'export des données movies"

log.Error("errors.network", map[string]interface{}{
    "details": "connection timeout",
})
// EN: "Network error: connection timeout"
// FR: "Erreur réseau : connection timeout"
```

### 🔄 Changement de Langue Dynamique
```go
// Changement de langue à la volée
translator.SetLanguage("de")

// Rechargement des traductions
err := translator.ReloadTranslations()
if err != nil {
    log.Error("Failed to reload translations:", err)
}

// Vérification langue active
currentLang := translator.GetCurrentLanguage()
availableLangs := translator.GetAvailableLanguages()
```

### 📈 Validation et Maintenance

#### Validation des Traductions
```go
func ValidateTranslations(translator *Translator) []ValidationError {
    var errors []ValidationError
    
    referenceKeys := translator.GetKeysForLanguage("en")
    
    for _, lang := range []string{"fr", "de", "es"} {
        langKeys := translator.GetKeysForLanguage(lang)
        
        // Clés manquantes
        for _, key := range referenceKeys {
            if !contains(langKeys, key) {
                errors = append(errors, ValidationError{
                    Language: lang,
                    Key:      key,
                    Type:     "missing_key",
                })
            }
        }
        
        // Clés orphelines
        for _, key := range langKeys {
            if !contains(referenceKeys, key) {
                errors = append(errors, ValidationError{
                    Language: lang,
                    Key:      key,
                    Type:     "orphan_key",
                })
            }
        }
    }
    
    return errors
}
```

Ce module assure une expérience utilisateur cohérente et localisée avec gestion automatique des fallbacks et support complet de la pluralisation selon les règles linguistiques.