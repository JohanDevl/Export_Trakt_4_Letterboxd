package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/config"
	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/logger"
	"github.com/robfig/cron/v3"
)

// getConfiguredTimezone returns the configured timezone or UTC as fallback
func getConfiguredTimezone(cfg *config.Config, log logger.Logger) *time.Location {
	// Try environment variable first (Docker TZ)
	if tz := os.Getenv("TZ"); tz != "" {
		if loc, err := time.LoadLocation(tz); err == nil {
			log.Info("scheduler.using_env_timezone", map[string]interface{}{
				"timezone": tz,
			})
			return loc
		}
		log.Warn("scheduler.invalid_env_timezone", map[string]interface{}{
			"timezone": tz,
		})
	}

	// Try config timezone
	if cfg.Export.Timezone != "" {
		if loc, err := time.LoadLocation(cfg.Export.Timezone); err == nil {
			log.Info("scheduler.using_config_timezone", map[string]interface{}{
				"timezone": cfg.Export.Timezone,
			})
			return loc
		}
		log.Warn("scheduler.invalid_config_timezone", map[string]interface{}{
			"timezone": cfg.Export.Timezone,
		})
	}

	// Fallback to UTC
	log.Info("scheduler.using_default_timezone", map[string]interface{}{
		"timezone": "UTC",
	})
	return time.UTC
}

// runWithSchedule sets up a cron scheduler and runs the export according to the schedule
func runWithSchedule(cfg *config.Config, log logger.Logger, schedule, exportType, exportMode string) {
	log.Info("scheduler.initializing", map[string]interface{}{
		"schedule":    schedule,
		"export_type": exportType,
		"export_mode": exportMode,
	})

	// Check for verbose logging environment variable
	if os.Getenv("EXPORT_VERBOSE") == "true" {
		log.SetLogLevel("debug")
		log.Info("scheduler.verbose_mode_enabled", nil)
	}

	// Override log level if LOG_LEVEL environment variable is set
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		log.SetLogLevel(logLevel)
		log.Info("scheduler.log_level_set", map[string]interface{}{
			"level": logLevel,
		})
	}

	// Get configured timezone for display
	configuredTZ := getConfiguredTimezone(cfg, log)

	// Validate cron expression
	cronParser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	_, err := cronParser.Parse(schedule)
	if err != nil {
		log.Error("errors.invalid_cron_schedule", map[string]interface{}{
			"schedule": schedule,
			"error":    err.Error(),
		})
		fmt.Printf("Invalid cron schedule format: %s\nError: %s\n", schedule, err.Error())
		fmt.Println("Example formats:")
		fmt.Println("  '0 */6 * * *'   - Every 6 hours")
		fmt.Println("  '0 9 * * 1'     - Every Monday at 9:00 AM")
		fmt.Println("  '30 14 * * *'   - Every day at 2:30 PM")
		os.Exit(1)
	}

	log.Info("scheduler.cron_validation_successful", map[string]interface{}{
		"schedule": schedule,
	})

	// Create a new cron scheduler
	c := cron.New()

	// Add the export job to the scheduler
	entryID, err := c.AddFunc(schedule, func() {
		log.Info("scheduler.job_triggered", map[string]interface{}{
			"schedule":    schedule,
			"export_type": exportType,
			"export_mode": exportMode,
			"timestamp":   time.Now().Format(time.RFC3339),
		})

		// Run the export with additional logging
		log.Info("scheduler.starting_export_execution", map[string]interface{}{
			"export_type": exportType,
			"export_mode": exportMode,
		})

		startTime := time.Now()
		runExportOnce(cfg, log, exportType, exportMode, "")
		duration := time.Since(startTime)

		// Get next run time for display
		entries := c.Entries()
		var nextRunDisplay string
		if len(entries) > 0 {
			nextRun := entries[0].Next.In(configuredTZ)
			nextRunDisplay = nextRun.Format("2006-01-02 15:04:05 MST")
		}

		log.Info("scheduler.export_execution_completed", map[string]interface{}{
			"export_type": exportType,
			"export_mode": exportMode,
			"duration":    duration.String(),
			"next_run":    nextRunDisplay,
		})

		// Display visual completion message with next run
		fmt.Printf("\n✅ === EXPORT COMPLETED ===\n")
		fmt.Printf("⏱️  Duration: %s\n", duration.String())
		fmt.Printf("▶️  Next run: %s\n", nextRunDisplay)
		fmt.Printf("============================\n\n")
	})

	if err != nil {
		log.Error("errors.scheduler_add_failed", map[string]interface{}{
			"schedule": schedule,
			"error":    err.Error(),
		})
		fmt.Printf("Failed to add scheduled job: %s\n", err.Error())
		os.Exit(1)
	}

	log.Info("scheduler.job_added_successfully", map[string]interface{}{
		"entry_id": entryID,
		"schedule": schedule,
	})

	// Start the cron scheduler
	c.Start()
	log.Info("scheduler.cron_started", nil)

	// Get the next run time and display in configured timezone
	entries := c.Entries()
	if len(entries) > 0 {
		nextRun := entries[0].Next
		nextRunInTZ := nextRun.In(configuredTZ)

		log.Info("scheduler.started", map[string]interface{}{
			"schedule":        schedule,
			"entry_id":        entryID,
			"next_run":        nextRun.Format(time.RFC3339),
			"next_run_local":  nextRunInTZ.Format("2006-01-02 15:04:05 MST"),
			"timezone":        configuredTZ.String(),
		})
		fmt.Printf("\n🎯 === EXPORT SCHEDULER STARTED ===\n")
		fmt.Printf("⏰ Schedule: %s\n", schedule)
		fmt.Printf("📺 Export Type: %s\n", exportType)
		fmt.Printf("🔧 Export Mode: %s\n", exportMode)
		fmt.Printf("🌍 Timezone: %s\n", configuredTZ.String())
		fmt.Printf("▶️  Next run: %s\n", nextRunInTZ.Format("2006-01-02 15:04:05 MST"))
		fmt.Printf("=====================================\n\n")

		// Log upcoming executions for the next hour in configured timezone
		now := time.Now()
		oneHourLater := now.Add(time.Hour)
		log.Info("scheduler.upcoming_executions_preview", map[string]interface{}{
			"next_hour_from": now.Format(time.RFC3339),
			"next_hour_to":   oneHourLater.Format(time.RFC3339),
			"timezone":       configuredTZ.String(),
		})

		count := 0
		if len(entries) > 0 {
			entry := entries[0]
			nextExec := entry.Next
			for nextExec.Before(oneHourLater) && count < 10 {
				nextExecInTZ := nextExec.In(configuredTZ)
				log.Info("scheduler.upcoming_execution", map[string]interface{}{
					"execution_time": nextExecInTZ.Format("2006-01-02 15:04:05 MST"),
					"in_minutes":     int(time.Until(nextExec).Minutes()),
					"timezone":       configuredTZ.String(),
				})
				// Calculate next execution after this one
				schedule, _ := cronParser.Parse(schedule)
				nextExec = schedule.Next(nextExec)
				count++
			}
		}
	}

	// Keep the program running until interrupted
	log.Info("scheduler.waiting", map[string]interface{}{
		"message": "Scheduler is running. Press Ctrl+C to stop.",
		"pid":     os.Getpid(),
	})
	fmt.Println("Scheduler is running. Press Ctrl+C to stop...")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("scheduler.shutdown_signal_received", map[string]interface{}{
			"signal": sig.String(),
		})
		fmt.Printf("\nReceived signal %s, shutting down gracefully...\n", sig)
		c.Stop()
		log.Info("scheduler.shutdown_complete", nil)
		os.Exit(0)
	}()

	// Block forever (or until SIGINT/SIGTERM)
	select {}
}
