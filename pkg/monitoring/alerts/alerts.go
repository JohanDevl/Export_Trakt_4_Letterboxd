package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JohanDevl/Export_Trakt_4_Letterboxd/pkg/monitoring"
	"github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

// AlertManager manages alerts and notifications
type AlertManager struct {
	logger            *logrus.Logger
	config            monitoring.AlertsConfig
	channels          []monitoring.NotificationChannel
	recentAlerts      map[string]time.Time
	rateLimitMutex    sync.RWMutex
	alertHistory      []monitoring.Alert
	alertHistoryMutex sync.RWMutex
}

// NewAlertManager creates a new alert manager
func NewAlertManager(logger *logrus.Logger, config monitoring.AlertsConfig) *AlertManager {
	am := &AlertManager{
		logger:       logger,
		config:       config,
		channels:     make([]monitoring.NotificationChannel, 0),
		recentAlerts: make(map[string]time.Time),
		alertHistory: make([]monitoring.Alert, 0),
	}

	// Initialize notification channels
	if config.WebhookURL != "" {
		webhookChannel := NewWebhookChannel(logger, config.WebhookURL)
		am.AddChannel(webhookChannel)
	}

	if config.EmailEnabled {
		// Email configuration would come from environment or config
		emailChannel := NewEmailChannel(logger, EmailConfig{
			SMTPHost:     "smtp.gmail.com",
			SMTPPort:     587,
			Username:     "", // From environment
			Password:     "", // From environment
			FromAddress:  "noreply@export-trakt.com",
			FromName:     "Export Trakt Monitor",
		})
		am.AddChannel(emailChannel)
	}

	if config.SlackEnabled {
		// Slack configuration would come from environment
		slackChannel := NewSlackChannel(logger, SlackConfig{
			WebhookURL: "", // From environment
			Channel:    "#alerts",
			Username:   "Export Trakt Monitor",
		})
		am.AddChannel(slackChannel)
	}

	return am
}

// AddChannel adds a notification channel
func (am *AlertManager) AddChannel(channel monitoring.NotificationChannel) {
	am.channels = append(am.channels, channel)
	am.logger.WithField("channel", channel.Name()).Info("Notification channel added")
}

// RemoveChannel removes a notification channel
func (am *AlertManager) RemoveChannel(channelName string) {
	for i, channel := range am.channels {
		if channel.Name() == channelName {
			am.channels = append(am.channels[:i], am.channels[i+1:]...)
			am.logger.WithField("channel", channelName).Info("Notification channel removed")
			break
		}
	}
}

// SendAlert sends an alert through all enabled channels
func (am *AlertManager) SendAlert(ctx context.Context, alert monitoring.Alert) error {
	// Check rate limiting
	if am.isRateLimited(alert) {
		am.logger.WithFields(logrus.Fields{
			"alert_id": alert.ID,
			"level":    alert.Level,
			"title":    alert.Title,
		}).Debug("Alert rate limited, skipping")
		return nil
	}

	// Add to alert history
	am.addToHistory(alert)

	// Send through all enabled channels
	var errors []string
	for _, channel := range am.channels {
		if !channel.IsEnabled() {
			continue
		}

		if err := channel.SendAlert(ctx, alert); err != nil {
			am.logger.WithError(err).WithField("channel", channel.Name()).Error("Failed to send alert")
			errors = append(errors, fmt.Sprintf("%s: %v", channel.Name(), err))
		} else {
			am.logger.WithFields(logrus.Fields{
				"channel":  channel.Name(),
				"alert_id": alert.ID,
				"level":    alert.Level,
			}).Info("Alert sent successfully")
		}
	}

	// Update rate limiting
	am.updateRateLimit(alert)

	if len(errors) > 0 {
		return fmt.Errorf("failed to send alert through some channels: %s", strings.Join(errors, "; "))
	}

	return nil
}

// CreateAlert creates a new alert
func (am *AlertManager) CreateAlert(level monitoring.AlertLevel, title, message, source string, metadata map[string]interface{}) monitoring.Alert {
	return monitoring.Alert{
		ID:        fmt.Sprintf("%s_%d", source, time.Now().UnixNano()),
		Level:     level,
		Title:     title,
		Message:   message,
		Timestamp: time.Now(),
		Source:    source,
		Metadata:  metadata,
		Resolved:  false,
	}
}

// SendExportAlert sends an alert for export events
func (am *AlertManager) SendExportAlert(ctx context.Context, exportType, status string, duration time.Duration, err error) {
	var alert monitoring.Alert

	metadata := map[string]interface{}{
		"export_type": exportType,
		"duration":    duration.String(),
	}

	if err != nil {
		metadata["error"] = err.Error()
		alert = am.CreateAlert(
			monitoring.AlertLevelError,
			fmt.Sprintf("Export Failed: %s", exportType),
			fmt.Sprintf("Export of type %s failed after %v: %v", exportType, duration, err),
			"exporter",
			metadata,
		)
	} else {
		level := monitoring.AlertLevelInfo
		if duration > 5*time.Minute {
			level = monitoring.AlertLevelWarning
		}

		alert = am.CreateAlert(
			level,
			fmt.Sprintf("Export Completed: %s", exportType),
			fmt.Sprintf("Export of type %s completed successfully in %v", exportType, duration),
			"exporter",
			metadata,
		)
	}

	am.SendAlert(ctx, alert)
}

// SendHealthAlert sends an alert for health check events
func (am *AlertManager) SendHealthAlert(ctx context.Context, component string, status monitoring.HealthStatus, message string) {
	var level monitoring.AlertLevel
	var title string

	switch status {
	case monitoring.HealthStatusUnhealthy:
		level = monitoring.AlertLevelCritical
		title = fmt.Sprintf("Component Unhealthy: %s", component)
	case monitoring.HealthStatusDegraded:
		level = monitoring.AlertLevelWarning
		title = fmt.Sprintf("Component Degraded: %s", component)
	case monitoring.HealthStatusHealthy:
		level = monitoring.AlertLevelInfo
		title = fmt.Sprintf("Component Recovered: %s", component)
	default:
		level = monitoring.AlertLevelWarning
		title = fmt.Sprintf("Component Status Unknown: %s", component)
	}

	alert := am.CreateAlert(
		level,
		title,
		message,
		"health_checker",
		map[string]interface{}{
			"component": component,
			"status":    string(status),
		},
	)

	am.SendAlert(ctx, alert)
}

// isRateLimited checks if an alert is rate limited
func (am *AlertManager) isRateLimited(alert monitoring.Alert) bool {
	if am.config.RateLimitMinutes <= 0 {
		return false
	}

	am.rateLimitMutex.RLock()
	defer am.rateLimitMutex.RUnlock()

	key := fmt.Sprintf("%s_%s_%s", alert.Source, alert.Level, alert.Title)
	if lastTime, exists := am.recentAlerts[key]; exists {
		return time.Since(lastTime) < time.Duration(am.config.RateLimitMinutes)*time.Minute
	}

	return false
}

// updateRateLimit updates the rate limiting state
func (am *AlertManager) updateRateLimit(alert monitoring.Alert) {
	if am.config.RateLimitMinutes <= 0 {
		return
	}

	am.rateLimitMutex.Lock()
	defer am.rateLimitMutex.Unlock()

	key := fmt.Sprintf("%s_%s_%s", alert.Source, alert.Level, alert.Title)
	am.recentAlerts[key] = time.Now()
}

// addToHistory adds an alert to the history
func (am *AlertManager) addToHistory(alert monitoring.Alert) {
	am.alertHistoryMutex.Lock()
	defer am.alertHistoryMutex.Unlock()

	am.alertHistory = append(am.alertHistory, alert)

	// Keep only last 1000 alerts
	if len(am.alertHistory) > 1000 {
		am.alertHistory = am.alertHistory[len(am.alertHistory)-1000:]
	}
}

// GetAlertHistory returns the alert history
func (am *AlertManager) GetAlertHistory() []monitoring.Alert {
	am.alertHistoryMutex.RLock()
	defer am.alertHistoryMutex.RUnlock()

	// Return a copy to prevent concurrent access issues
	history := make([]monitoring.Alert, len(am.alertHistory))
	copy(history, am.alertHistory)
	return history
}

// WebhookChannel implements webhook notifications
type WebhookChannel struct {
	logger     *logrus.Logger
	webhookURL string
	client     *http.Client
}

// NewWebhookChannel creates a new webhook notification channel
func NewWebhookChannel(logger *logrus.Logger, webhookURL string) *WebhookChannel {
	return &WebhookChannel{
		logger:     logger,
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert sends an alert via webhook
func (wc *WebhookChannel) SendAlert(ctx context.Context, alert monitoring.Alert) error {
	payload, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal alert: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", wc.webhookURL, strings.NewReader(string(payload)))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Export-Trakt-Monitor/1.0")

	resp, err := wc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// Name returns the channel name
func (wc *WebhookChannel) Name() string {
	return "webhook"
}

// IsEnabled returns true if the channel is enabled
func (wc *WebhookChannel) IsEnabled() bool {
	return wc.webhookURL != ""
}

// EmailConfig represents email configuration
type EmailConfig struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	ToAddresses []string
}

// EmailChannel implements email notifications
type EmailChannel struct {
	logger *logrus.Logger
	config EmailConfig
	dialer *gomail.Dialer
}

// NewEmailChannel creates a new email notification channel
func NewEmailChannel(logger *logrus.Logger, config EmailConfig) *EmailChannel {
	dialer := gomail.NewDialer(config.SMTPHost, config.SMTPPort, config.Username, config.Password)

	return &EmailChannel{
		logger: logger,
		config: config,
		dialer: dialer,
	}
}

// SendAlert sends an alert via email
func (ec *EmailChannel) SendAlert(ctx context.Context, alert monitoring.Alert) error {
	if len(ec.config.ToAddresses) == 0 {
		return fmt.Errorf("no recipient email addresses configured")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", ec.config.FromName, ec.config.FromAddress))
	m.SetHeader("To", ec.config.ToAddresses...)
	m.SetHeader("Subject", fmt.Sprintf("[%s] %s", strings.ToUpper(string(alert.Level)), alert.Title))

	body := fmt.Sprintf(`
Alert Details:
- Level: %s
- Title: %s
- Message: %s
- Source: %s
- Timestamp: %s
- Alert ID: %s

Metadata:
%s
`, 
		alert.Level,
		alert.Title,
		alert.Message,
		alert.Source,
		alert.Timestamp.Format("2006-01-02 15:04:05 MST"),
		alert.ID,
		formatMetadata(alert.Metadata),
	)

	m.SetBody("text/plain", body)

	// Create a context with timeout for the email sending
	emailCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Send email in a goroutine to handle context cancellation
	errChan := make(chan error, 1)
	go func() {
		errChan <- ec.dialer.DialAndSend(m)
	}()

	select {
	case <-emailCtx.Done():
		return fmt.Errorf("email sending timed out")
	case err := <-errChan:
		return err
	}
}

// Name returns the channel name
func (ec *EmailChannel) Name() string {
	return "email"
}

// IsEnabled returns true if the channel is enabled
func (ec *EmailChannel) IsEnabled() bool {
	return ec.config.Username != "" && ec.config.Password != "" && len(ec.config.ToAddresses) > 0
}

// SlackConfig represents Slack configuration
type SlackConfig struct {
	WebhookURL string
	Channel    string
	Username   string
}

// SlackChannel implements Slack notifications
type SlackChannel struct {
	logger *logrus.Logger
	config SlackConfig
	client *http.Client
}

// NewSlackChannel creates a new Slack notification channel
func NewSlackChannel(logger *logrus.Logger, config SlackConfig) *SlackChannel {
	return &SlackChannel{
		logger: logger,
		config: config,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendAlert sends an alert via Slack
func (sc *SlackChannel) SendAlert(ctx context.Context, alert monitoring.Alert) error {
	color := ""
	emoji := ""

	switch alert.Level {
	case monitoring.AlertLevelCritical:
		color = "danger"
		emoji = ":red_circle:"
	case monitoring.AlertLevelError:
		color = "danger"
		emoji = ":exclamation:"
	case monitoring.AlertLevelWarning:
		color = "warning"
		emoji = ":warning:"
	case monitoring.AlertLevelInfo:
		color = "good"
		emoji = ":information_source:"
	}

	payload := map[string]interface{}{
		"channel":   sc.config.Channel,
		"username":  sc.config.Username,
		"icon_emoji": ":robot_face:",
		"attachments": []map[string]interface{}{
			{
				"color":    color,
				"title":    fmt.Sprintf("%s %s", emoji, alert.Title),
				"text":     alert.Message,
				"ts":       alert.Timestamp.Unix(),
				"fields": []map[string]interface{}{
					{
						"title": "Level",
						"value": string(alert.Level),
						"short": true,
					},
					{
						"title": "Source",
						"value": alert.Source,
						"short": true,
					},
					{
						"title": "Alert ID",
						"value": alert.ID,
						"short": true,
					},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sc.config.WebhookURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return fmt.Errorf("failed to create Slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Slack returned status %d", resp.StatusCode)
	}

	return nil
}

// Name returns the channel name
func (sc *SlackChannel) Name() string {
	return "slack"
}

// IsEnabled returns true if the channel is enabled
func (sc *SlackChannel) IsEnabled() bool {
	return sc.config.WebhookURL != ""
}

// formatMetadata formats metadata for display
func formatMetadata(metadata map[string]interface{}) string {
	if len(metadata) == 0 {
		return "None"
	}

	var parts []string
	for key, value := range metadata {
		parts = append(parts, fmt.Sprintf("- %s: %v", key, value))
	}

	return strings.Join(parts, "\n")
} 