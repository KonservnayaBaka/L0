package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ReqCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: "http_requests_total", Help: "Number of HTTP requests"},
		[]string{"path", "method", "status"},
	)
	ReqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: "http_request_duration_seconds", Help: "Request latency in seconds", Buckets: prometheus.DefBuckets},
		[]string{"path", "method"},
	)
)

func Init() {
	prometheus.MustRegister(ReqCount, ReqDuration)
}

func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			h.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		c.Next()
	}
}

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		d := time.Since(start).Seconds()
		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		ReqCount.WithLabelValues(path, c.Request.Method, http.StatusText(status)).Inc()
		ReqDuration.WithLabelValues(path, c.Request.Method).Observe(d)
	}
}
