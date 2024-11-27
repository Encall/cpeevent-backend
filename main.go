package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/encall/cpeevent-backend/src/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cpeevo_backend_http_request_duration_seconds",
			Help:    "Duration of HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path"},
	)
	requestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cpeevo_backend_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"path", "method"},
	)
	errorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cpeevo_backend_http_errors_total",
			Help: "Total number of HTTP error responses.",
		},
		[]string{"path", "method", "status"},
	)
)

func init() {
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(requestCount)
	prometheus.MustRegister(errorCount)
}

func main() {

	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = "debug"
	}

	// Set Gin mode
	if mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize Gin router with Logger and Recovery middleware
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	origin := os.Getenv("ORIGIN_URL")
	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{origin}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "refresh_token"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	r.Use(cors.New(config))

	r.Use(func(c *gin.Context) {
		if c.FullPath() != "/healthcheck" && c.FullPath() != "/metrics" {
			start := time.Now()
			c.Next()
			duration := time.Since(start).Seconds()
			path := c.FullPath()
			method := c.Request.Method
			status := c.Writer.Status()

			requestDuration.WithLabelValues(path).Observe(duration)
			requestCount.WithLabelValues(path, method).Inc()
			if status >= 400 {
				errorCount.WithLabelValues(path, method, http.StatusText(status)).Inc()
			}
		} else {
			c.Next()
		}
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register all routes with /api prefix
	api := r.Group("/api")
	routes.UserRoutes(api)

	// Health Check endpoint
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
