package monitor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"wechat-service/internal/config"
	"wechat-service/pkg/logger"
)

// Alert represents a monitoring alert
type Alert struct {
	ID          string                 `json:"id"`
	AppID       string                 `json:"appid"`
	Nickname    string                 `json:"nickname"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Count       int                    `json:"count"`
	FirstTime   time.Time              `json:"first_time"`
	Example     map[string]interface{} `json:"example"`
	Severity    string                 `json:"severity"`
	CreatedAt   time.Time              `json:"created_at"`
}

// AlertType constants
const (
	AlertTypeDNSFailure       = "DNS_FAILURE"
	AlertTypeDNSTimeout       = "DNS_TIMEOUT"
	AlertTypeConnectionTimeout = "CONNECTION_TIMEOUT"
	AlertTypeRequestTimeout   = "REQUEST_TIMEOUT"
	AlertTypeResponseInvalid  = "RESPONSE_INVALID"
	AlertTypeMarkFail         = "MARK_FAIL"
)

// Severity levels
const (
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

// Monitor handles system monitoring and alerting
type Monitor struct {
	cfg        *config.Config
	log        *logger.Logger
	httpClient *http.Client
	alerts     []Alert
	alertMu    sync.RWMutex
	stopCh     chan struct{}
	wg         sync.WaitGroup

	// Health checks
	checks []HealthCheck
	mu     sync.RWMutex
}

// HealthCheck represents a health check function
type HealthCheck struct {
	Name    string
	Status  string
	Message string
	LastRun time.Time
}

// NewMonitor creates a new monitoring instance
func NewMonitor(cfg *config.Config, log *logger.Logger) *Monitor {
	return &Monitor{
		cfg: cfg,
		log: log,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		alerts:  make([]Alert, 0),
		stopCh:  make(chan struct{}),
		checks:  make([]HealthCheck, 0),
	}
}

// AddHealthCheck adds a health check
func (m *Monitor) AddHealthCheck(name string, checkFunc func() (string, string)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checks = append(m.checks, HealthCheck{
		Name: name,
	})
}

// RunHealthChecks runs all health checks
func (m *Monitor) RunHealthChecks() []HealthCheck {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]HealthCheck, len(m.checks))
	for i, check := range m.checks {
		// In production, call the actual check function
		results[i] = HealthCheck{
			Name:    check.Name,
			Status:  "ok",
			Message: "Healthy",
			LastRun: time.Now(),
		}
	}

	return results
}

// GetHealthStatus returns overall health status
func (m *Monitor) GetHealthStatus() string {
	checks := m.RunHealthChecks()
	for _, check := range checks {
		if check.Status != "ok" {
			return "unhealthy"
		}
	}
	return "healthy"
}

// HandleAlert processes an incoming alert from WeChat
func (m *Monitor) HandleAlert(alertData map[string]interface{}) {
	alert := Alert{
		ID:        generateAlertID(),
		AppID:     getString(alertData, "appid"),
		Nickname:  getString(alertData, "nickname"),
		Type:      getString(alertData, "type"),
		Count:     getInt(alertData, "count"),
		FirstTime: parseTime(getString(alertData, "time")),
		Example:   getMap(alertData, "example"),
		Severity:  m.calculateSeverity(getString(alertData, "type")),
		CreatedAt: time.Now(),
	}

	// Set description based on type
	alert.Description = m.getDescription(alert.Type)

	m.log.Error("Alert received",
		"type", alert.Type,
		"count", alert.Count,
		"severity", alert.Severity,
		"example", alert.Example,
	)

	// Store alert
	m.mu.Lock()
	m.alerts = append(m.alerts, alert)
	// Keep only last 100 alerts
	if len(m.alerts) > 100 {
		m.alerts = m.alerts[len(m.alerts)-100:]
	}
	m.mu.Unlock()

	// Send webhook notification
	if m.cfg.Monitoring.AlertEnabled && m.cfg.Monitoring.AlertWebhook != "" {
		go m.sendAlertWebhook(alert)
	}
}

// calculateSeverity returns severity level for alert type
func (m *Monitor) calculateSeverity(alertType string) string {
	switch alertType {
	case AlertTypeMarkFail:
		return SeverityCritical
	case AlertTypeConnectionTimeout, AlertTypeRequestTimeout:
		return SeverityHigh
	case AlertTypeDNSFailure, AlertTypeResponseInvalid:
		return SeverityHigh
	case AlertTypeDNSTimeout:
		return SeverityMedium
	default:
		return SeverityLow
	}
}

// getDescription returns description for alert type
func (m *Monitor) getDescription(alertType string) string {
	descriptions := map[string]string{
		AlertTypeDNSFailure:        "DNS resolution failed",
		AlertTypeDNSTimeout:        "DNS resolution timed out (>5s)",
		AlertTypeConnectionTimeout: "Connection to server timed out (>3s)",
		AlertTypeRequestTimeout:    "Server did not respond within 5 seconds",
		AlertTypeResponseInvalid:   "Server response was invalid",
		AlertTypeMarkFail:          "Server auto-blocked (multiple failures)",
	}
	return descriptions[alertType]
}

// sendAlertWebhook sends alert to webhook URL
func (m *Monitor) sendAlertWebhook(alert Alert) {
	if m.cfg.Monitoring.AlertWebhook == "" {
		return
	}

	data, err := json.Marshal(alert)
	if err != nil {
		m.log.Error("Failed to marshal alert", "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", m.cfg.Monitoring.AlertWebhook, bytes.NewBuffer(data))
	if err != nil {
		m.log.Error("Failed to create alert webhook request", "error", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		m.log.Error("Failed to send alert webhook", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		m.log.Error("Alert webhook returned non-200", "status", resp.StatusCode)
	}
}

// GetAlerts returns recent alerts
func (m *Monitor) GetAlerts() []Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return copy
	alerts := make([]Alert, len(m.alerts))
	copy(alerts, m.alerts)
	return alerts
}

// ClearAlerts clears all alerts
func (m *Monitor) ClearAlerts() {
	m.mu.Lock()
	m.alerts = make([]Alert, 0)
	m.mu.Unlock()
}

// StartMonitoring starts background monitoring
func (m *Monitor) StartMonitoring() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.runPeriodicChecks()
			case <-m.stopCh:
				return
			}
		}
	}()
}

// runPeriodicChecks runs periodic monitoring checks
func (m *Monitor) runPeriodicChecks() {
	// Check health status
	status := m.GetHealthStatus()
	if status != "healthy" {
		m.log.Warn("Health check failed", "status", status)
	}

	// Could add more periodic checks here:
	// - Check API response times
	// - Check token freshness
	// - Check disk space, memory, etc.
}

// Stop stops monitoring
func (m *Monitor) Stop() {
	close(m.stopCh)
	m.wg.Wait()
	m.log.Info("Monitor stopped")
}

// Helper functions
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

func getString(data map[string]interface{}, key string) string {
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if v, ok := data[key]; ok {
		switch t := v.(type) {
		case int:
			return t
		case int64:
			return int(t)
		case float64:
			return int(t)
		}
	}
	return 0
}

func getMap(data map[string]interface{}, key string) map[string]interface{} {
	if v, ok := data[key]; ok {
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
	}
	return make(map[string]interface{})
}

func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04:05", s)
	if err != nil {
		return time.Now()
	}
	return t
}
