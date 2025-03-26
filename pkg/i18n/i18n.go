package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// Translator handles all internationalization operations
type Translator struct {
	bundle    *i18n.Bundle
	config    *config.I18nConfig
	log       logger.Logger
	localizer *i18n.Localizer
}

// NewTranslator creates a new translator instance
func NewTranslator(cfg *config.I18nConfig, log logger.Logger) (*Translator, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	t := &Translator{
		bundle: bundle,
		config: cfg,
		log:    log,
	}

	if err := t.loadTranslations(); err != nil {
		return nil, err
	}

	t.localizer = i18n.NewLocalizer(bundle, cfg.Language, cfg.DefaultLanguage)
	return t, nil
}

// loadTranslations loads all translation files from the locales directory
func (t *Translator) loadTranslations() error {
	t.log.Debug("i18n.loading_translations", map[string]interface{}{
		"dir": t.config.LocalesDir,
	})
	
	entries, err := os.ReadDir(t.config.LocalesDir)
	if err != nil {
		return fmt.Errorf("failed to read locales directory: %w", err)
	}

	t.log.Debug("i18n.found_files", map[string]interface{}{
		"count": len(entries),
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(t.config.LocalesDir, entry.Name())
		if _, err := t.bundle.LoadMessageFile(path); err != nil {
			t.log.Warn("errors.translation_file_load_failed", map[string]interface{}{
				"path":  path,
				"error": err.Error(),
			})
			continue
		}

		t.log.Debug("i18n.translation_file_loaded", map[string]interface{}{
			"path": path,
		})
	}

	return nil
}

// Translate returns the translated message for the given message ID
func (t *Translator) Translate(messageID string, templateData map[string]interface{}) string {
	// Simple protection against recursion
	if messageID == "" {
		return ""
	}

	// Prevent recursion for error messages that might be logged during translation
	if messageID == "errors.translation_failed" || 
	   messageID == "errors.translation_file_load_failed" {
		return messageID
	}

	// Create a message to translate
	msg := i18n.Message{
		ID: messageID,
	}

	// Attempt translation
	translation, err := t.localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &msg,
		TemplateData:   templateData,
	})

	if err != nil {
		// Log a warning about the missing translation
		t.log.Warn("errors.translation_not_found", map[string]interface{}{
			"messageID": messageID,
			"error":     err.Error(),
		})
		
		// If error, return the original ID
		return messageID
	}

	return translation
}

// SetLanguage changes the current language
func (t *Translator) SetLanguage(lang string) {
	t.localizer = i18n.NewLocalizer(t.bundle, lang, t.config.DefaultLanguage)
	t.config.Language = lang
	t.log.Info("i18n.language_changed", map[string]interface{}{
		"language": lang,
	})
} 