package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	httpRequests    *prometheus.CounterVec
	httpLatency     *prometheus.HistogramVec
	messagesReceived *prometheus.CounterVec
	messagesSent    *prometheus.CounterVec
	eventsReceived  *prometheus.CounterVec
	errors          *prometheus.CounterVec
	panics          prometheus.Counter
	tokenRefreshes  prometheus.Counter
}

// NewMetrics creates new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		httpRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wechat_http_requests_total",
				Help: "Total HTTP requests",
			},
			[]string{"endpoint", "method", "status"},
		),
		httpLatency: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "wechat_http_latency_seconds",
				Help:    "HTTP request latency",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint", "method"},
		),
		messagesReceived: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wechat_messages_received_total",
				Help: "Total messages received by type",
			},
			[]string{"type"},
		),
		messagesSent: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wechat_messages_sent_total",
				Help: "Total messages sent by type",
			},
			[]string{"type"},
		),
		eventsReceived: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wechat_events_received_total",
				Help: "Total events received by type",
			},
			[]string{"type"},
		),
		errors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "wechat_errors_total",
				Help: "Total errors by type",
			},
			[]string{"type"},
		),
		panics: promauto.NewCounter(prometheus.CounterOpts{
			Name: "wechat_panics_total",
			Help: "Total panics",
		}),
		tokenRefreshes: promauto.NewCounter(prometheus.CounterOpts{
			Name: "wechat_token_refreshes_total",
			Help: "Total token refreshes",
		}),
	}
}

// IncHTTPRequest increments HTTP request counter
func (m *Metrics) IncHTTPRequest(endpoint, method, status string) {
	m.httpRequests.WithLabelValues(endpoint, method, status).Inc()
}

// IncHTTPError increments HTTP error counter
func (m *Metrics) IncHTTPError(endpoint, method string, status int) {
	m.httpRequests.WithLabelValues(endpoint, method, string(rune(status))).Inc()
}

// ObserveLatency records request latency
func (m *Metrics) ObserveLatency(endpoint, method string, duration float64) {
	m.httpLatency.WithLabelValues(endpoint, method).Observe(duration)
}

// IncMessageReceived increments message received counter
func (m *Metrics) IncMessageReceived(msgType string) {
	m.messagesReceived.WithLabelValues(msgType).Inc()
}

// IncMessageSent increments message sent counter
func (m *Metrics) IncMessageSent(msgType string) {
	m.messagesSent.WithLabelValues(msgType).Inc()
}

// IncEventReceived increments event received counter
func (m *Metrics) IncEventReceived(eventType string) {
	m.eventsReceived.WithLabelValues(eventType).Inc()
}

// IncMessageError increments message error counter
func (m *Metrics) IncMessageError(errType string) {
	m.errors.WithLabelValues(errType).Inc()
}

// IncMessagePanic increments panic counter
func (m *Metrics) IncMessagePanic() {
	m.panics.Inc()
}

// IncTokenRefresh increments token refresh counter
func (m *Metrics) IncTokenRefresh() {
	m.tokenRefreshes.Inc()
}
