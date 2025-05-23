package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/robfig/cron/v3"
)

// Scheduler manages the scheduling of export jobs
type Scheduler struct {
	config *config.Config
	log    logger.Logger
	cron   *cron.Cron
}

// NewScheduler creates a new scheduler
func NewScheduler(cfg *config.Config, log logger.Logger) *Scheduler {
	return &Scheduler{
		config: cfg,
		log:    log,
		cron:   cron.New(),
	}
}

// Start initializes the scheduler from environment variables
func (s *Scheduler) Start() error {
	// Get schedule from environment variable
	schedule := os.Getenv("EXPORT_SCHEDULE")
	if schedule == "" {
		s.log.Info("scheduler.no_schedule_defined", map[string]interface{}{
			"message": "No EXPORT_SCHEDULE environment variable defined. Scheduler will not run.",
		})
		return nil
	}

	// Get export mode and type from environment variables or use defaults
	exportMode := os.Getenv("EXPORT_MODE")
	if exportMode == "" {
		exportMode = "complete" // Default to complete mode
	}

	exportType := os.Getenv("EXPORT_TYPE")
	if exportType == "" {
		exportType = "all" // Default to export all
	}

	s.log.Info("scheduler.starting", map[string]interface{}{
		"schedule":    schedule,
		"export_mode": exportMode,
		"export_type": exportType,
	})

	// Add the job to the cron scheduler
	_, err := s.cron.AddFunc(schedule, func() {
		s.runExport(exportMode, exportType)
	})
	if err != nil {
		s.log.Error("scheduler.invalid_schedule", map[string]interface{}{
			"schedule": schedule,
			"error":    err.Error(),
			"details":  "Format should be standard cron format: minute hour day-of-month month day-of-week",
		})
		return fmt.Errorf("invalid schedule format: %w", err)
	}

	// Start the cron scheduler
	s.cron.Start()
	
	entries := s.cron.Entries()
	if len(entries) > 0 {
		s.log.Info("scheduler.started", map[string]interface{}{
			"next_run": entries[0].Next.Format(time.RFC3339),
		})
	} else {
		s.log.Warn("scheduler.no_entries", map[string]interface{}{
			"message": "Scheduler started but no entries were added",
		})
	}

	// Set up a signal handler to gracefully shut down the scheduler
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signalChan
		s.Stop()
	}()

	return nil
}

// Stop gracefully stops the scheduler
func (s *Scheduler) Stop() {
	if s.cron != nil {
		s.log.Info("scheduler.stopping", nil)
		ctx := s.cron.Stop()
		<-ctx.Done()
		s.log.Info("scheduler.stopped", nil)
	}
}

// runExport executes the export command with the specified mode and type
func (s *Scheduler) runExport(mode, exportType string) {
	s.log.Info("scheduler.running_export", map[string]interface{}{
		"mode": mode,
		"type": exportType,
	})

	// Create command to run export
	cmd := exec.Command(os.Args[0], "export", "--mode", mode, "--export", exportType)
	
	// Get output
	output, err := cmd.CombinedOutput()
	if err != nil {
		s.log.Error("scheduler.export_failed", map[string]interface{}{
			"error":  err.Error(),
			"output": strings.TrimSpace(string(output)),
		})
		return
	}

	s.log.Info("scheduler.export_completed", map[string]interface{}{
		"output": strings.TrimSpace(string(output)),
	})
} 