package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestsInFlight prometheus.Gauge

	// Message metrics
	MessagesReceived *prometheus.CounterVec
	MessagesSent     *prometheus.CounterVec
	MessagesError    *prometheus.CounterVec

	// Event metrics
	EventsReceived *prometheus.CounterVec
	EventsError    *prometheus.CounterVec

	// User metrics
	UsersSubscribed   prometheus.Counter
	UsersUnsubscribed prometheus.Counter

	// Database metrics
	DBQueriesTotal    *prometheus.CounterVec
	DBQueryDuration   *prometheus.HistogramVec
	DBConnectionsActive prometheus.Gauge
	DBConnectionsIdle  prometheus.Gauge

	// Cache metrics
	CacheHits   prometheus.Counter
	CacheMisses prometheus.Counter
	CacheErrors prometheus.Counter

	// System metrics
	GoRoutines prometheus.Gauge
	GoMemAlloc prometheus.Gauge
}

// NewMetrics creates and registers all metrics
func NewMetrics(namespace string) *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "http_requests_total",
				Help:      "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "http_request_duration_seconds",
				Help:      "HTTP request duration in seconds",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPRequestsInFlight: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "http_requests_in_flight",
				Help:      "Current number of HTTP requests being processed",
			},
		),

		// Message metrics
		MessagesReceived: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_received_total",
				Help:      "Total number of messages received from users",
			},
			[]string{"type"},
		),
		MessagesSent: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_sent_total",
				Help:      "Total number of messages sent to users",
			},
			[]string{"type"},
		),
		MessagesError: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "messages_error_total",
				Help:      "Total number of message processing errors",
			},
			[]string{"type", "error"},
		),

		// Event metrics
		EventsReceived: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "events_received_total",
				Help:      "Total number of events received",
			},
			[]string{"type"},
		),
		EventsError: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "events_error_total",
				Help:      "Total number of event processing errors",
			},
			[]string{"type", "error"},
		),

		// User metrics
		UsersSubscribed: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_subscribed_total",
				Help:      "Total number of user subscriptions",
			},
		),
		UsersUnsubscribed: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "users_unsubscribed_total",
				Help:      "Total number of user unsubscriptions",
			},
		),

		// Database metrics
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "db_queries_total",
				Help:      "Total number of database queries",
			},
			[]string{"operation", "table"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "db_query_duration_seconds",
				Help:      "Database query duration in seconds",
				Buckets:   []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
			},
			[]string{"operation", "table"},
		),
		DBConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_active",
				Help:      "Number of active database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "db_connections_idle",
				Help:      "Number of idle database connections",
			},
		),

		// Cache metrics
		CacheHits: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
		),
		CacheMisses: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
		),
		CacheErrors: promauto.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_errors_total",
				Help:      "Total number of cache errors",
			},
		),

		// System metrics
		GoRoutines: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "go_routines",
				Help:      "Number of running goroutines",
			},
		),
		GoMemAlloc: promauto.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "go_mem_alloc_bytes",
				Help:      "Bytes allocated and still in use",
			},
		),
	}
}

// RecordHTTPRequest records an HTTP request
func (m *Metrics) RecordHTTPRequest(method, endpoint, status string, duration float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordMessageReceived records a received message
func (m *Metrics) RecordMessageReceived(msgType string) {
	m.MessagesReceived.WithLabelValues(msgType).Inc()
}

// RecordMessageSent records a sent message
func (m *Metrics) RecordMessageSent(msgType string) {
	m.MessagesSent.WithLabelValues(msgType).Inc()
}

// RecordMessageError records a message processing error
func (m *Metrics) RecordMessageError(msgType, error string) {
	m.MessagesError.WithLabelValues(msgType, error).Inc()
}

// RecordEventReceived records a received event
func (m *Metrics) RecordEventReceived(eventType string) {
	m.EventsReceived.WithLabelValues(eventType).Inc()
}

// RecordEventError records an event processing error
func (m *Metrics) RecordEventError(eventType, error string) {
	m.EventsError.WithLabelValues(eventType, error).Inc()
}

// RecordDBSuccess records a successful database query
func (m *Metrics) RecordDBSuccess(operation, table string, duration float64) {
	m.DBQueriesTotal.WithLabelValues(operation, table).Inc()
	m.DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordDBError records a database error
func (m *Metrics) RecordDBError(operation, table string) {
	m.DBQueriesTotal.WithLabelValues(operation, table).Inc()
}

// UpdateGoMetrics updates Go runtime metrics
func (m *Metrics) UpdateGoMetrics() {
	// Go runtime metrics are automatically collected by prometheus client_golang
	// No manual collection needed
}
