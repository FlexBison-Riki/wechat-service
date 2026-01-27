package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// GinMiddleware returns a Gin middleware for recording HTTP metrics
func GinMiddleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		m.HTTPRequestsInFlight.Inc()

		c.Next()

		m.HTTPRequestsInFlight.Dec()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		m.RecordHTTPRequest(method, endpoint, status, duration)
	}
}

// PrometheusHandler returns the Prometheus metrics HTTP handler
func PrometheusHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}

// RegisterGinMetrics registers Gin metrics collector
func RegisterGinMetrics() {
	// Prometheus Go collector already includes Go runtime metrics
	// No additional registration needed
}
