package middleware

import (
	"strconv"
	"time"

	"shortlink/internal/metrics"

	"github.com/gin-gonic/gin"
)

// Metrics 统计 HTTP 请求次数和响应耗时。
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()

		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}

		method := c.Request.Method
		status := strconv.Itoa(c.Writer.Status())

		metrics.HTTPRequestsTotal.
			WithLabelValues(method, route, status).
			Inc()

		metrics.HTTPRequestDuration.
			WithLabelValues(method, route).
			Observe(duration)
	}
}
