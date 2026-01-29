package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
)

// GinMiddleware returns Gin middleware for metrics
func GinMiddleware(m *Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		m.IncHTTPRequest(path, method, string(rune(status)))
		m.ObserveLatency(path, method, duration)
	}
}
