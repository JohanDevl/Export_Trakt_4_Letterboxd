package scheduler

import (
	"os"
	"testing"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
)

// MockLogger pour les tests
type MockLogger struct{}

func (m *MockLogger) Debug(messageID string, data ...map[string]interface{}) {}
func (m *MockLogger) Debugf(messageID string, data map[string]interface{})  {}
func (m *MockLogger) Info(messageID string, data ...map[string]interface{})  {}
func (m *MockLogger) Infof(messageID string, data map[string]interface{})   {}
func (m *MockLogger) Warn(messageID string, data ...map[string]interface{})  {}
func (m *MockLogger) Warnf(messageID string, data map[string]interface{})   {}
func (m *MockLogger) Error(messageID string, data ...map[string]interface{}) {}
func (m *MockLogger) Errorf(messageID string, data map[string]interface{})  {}
func (m *MockLogger) SetLogLevel(level string)                        {}
func (m *MockLogger) SetLogFile(filepath string) error                { return nil }
func (m *MockLogger) SetTranslator(translator logger.Translator)            {}

func TestNewScheduler(t *testing.T) {
	// Créer une configuration et un logger mock
	cfg := &config.Config{}
	log := &MockLogger{}

	// Créer un nouveau scheduler
	sched := NewScheduler(cfg, log)

	// Vérifier que le scheduler a été créé correctement
	if sched == nil {
		t.Error("NewScheduler() a retourné nil")
	}
	if sched.config != cfg {
		t.Error("NewScheduler() n'a pas correctement attribué la configuration")
	}
}

func TestScheduler_Start_NoSchedule(t *testing.T) {
	// Créer une configuration et un logger mock
	cfg := &config.Config{}
	log := &MockLogger{}

	// S'assurer que la variable d'environnement EXPORT_SCHEDULE n'est pas définie
	os.Unsetenv("EXPORT_SCHEDULE")

	// Créer un nouveau scheduler
	sched := NewScheduler(cfg, log)

	// Démarrer le scheduler
	err := sched.Start()

	// Vérifier qu'il n'y a pas d'erreur
	if err != nil {
		t.Errorf("Start() a retourné une erreur: %v", err)
	}
}

func TestScheduler_Start_InvalidSchedule(t *testing.T) {
	// Créer une configuration et un logger mock
	cfg := &config.Config{}
	log := &MockLogger{}

	// Définir une planification non valide
	os.Setenv("EXPORT_SCHEDULE", "invalid-schedule")
	defer os.Unsetenv("EXPORT_SCHEDULE")

	// Créer un nouveau scheduler
	sched := NewScheduler(cfg, log)

	// Démarrer le scheduler
	err := sched.Start()

	// Vérifier qu'il y a une erreur
	if err == nil {
		t.Error("Start() n'a pas retourné d'erreur avec une planification non valide")
	}
}

func TestScheduler_Start_ValidSchedule(t *testing.T) {
	// Créer une configuration et un logger mock
	cfg := &config.Config{}
	log := &MockLogger{}

	// Définir une planification valide
	os.Setenv("EXPORT_SCHEDULE", "* * * * *") // Toutes les minutes
	defer os.Unsetenv("EXPORT_SCHEDULE")

	// Créer un nouveau scheduler
	sched := NewScheduler(cfg, log)

	// Démarrer le scheduler dans une goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- sched.Start()
	}()

	// Attendre un peu pour s'assurer que le scheduler démarre
	time.Sleep(100 * time.Millisecond)

	// Arrêter le scheduler
	sched.Stop()

	// Vérifier qu'il n'y a pas d'erreur
	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("Start() a retourné une erreur: %v", err)
		}
	case <-time.After(1 * time.Second):
		// La méthode Start() n'a pas encore retourné après 1 seconde
		// C'est normal si elle bloque indéfiniment
	}
} 