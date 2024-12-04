// main_test.go
package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/encall/cpeevent-backend/src/controllers"
	"github.com/encall/cpeevent-backend/src/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupRouter() *gin.Engine {
	// Initialize Gin router with middleware
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "refresh_token"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	r.Use(cors.New(config))

	r.Use(func(c *gin.Context) {
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
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Register all routes with /api prefix
	api := r.Group("/api")
	routes.UserRoutes(api)

	// Health Check endpoint
	r.GET("/healthcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	return r
}

func TestMetrics(t *testing.T) {
	router := setupRouter()

	// Send a GET request to /healthcheck
	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Send a request to a non-existent route to generate an error
	req, _ = http.NewRequest("GET", "/api/nonexistent", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	// Fetch metrics
	req, _ = http.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /metrics, got %d", w.Code)
	}

	body, _ := io.ReadAll(w.Body)
	metrics := string(body)

	// Check if metrics are present
	if !strings.Contains(metrics, "http_requests_total") {
		t.Errorf("Metric http_requests_total not found")
	}
	if !strings.Contains(metrics, "http_errors_total") {
		t.Errorf("Metric http_errors_total not found")
	}
	if !strings.Contains(metrics, "http_request_duration_seconds") {
		t.Errorf("Metric http_request_duration_seconds not found")
	}
}

func TestHealthCheck(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /healthcheck, got %d", w.Code)
	}

	expectedBody := `{"status":"healthy"}`
	if strings.TrimSpace(w.Body.String()) != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}

func TestInvalidMethod(t *testing.T) {
	router := setupRouter()

	req, _ := http.NewRequest("POST", "/healthcheck", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusOK {
		t.Errorf("Expected status 405 or 200 for invalid method on /healthcheck, got %d", w.Code)
	}
}

func TestAPIEndpoint(t *testing.T) {
	router := setupRouter()

	// Assuming there's a /api/users endpoint
	req, _ := http.NewRequest("GET", "/api/v1/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Adjust the expected status code based on actual implementation
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for /api/users, got %d", w.Code)
	}
}

func TestGetPostFromEvent(t *testing.T) {
	// Setup the Gin router
	setupRouter()

	// Create a test context and response recorder
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// Set the necessary parameters and headers
	ctx.Params = gin.Params{
		{Key: "eventID", Value: "6748b6dcbfc7262e96ab9d59"}, // Example eventID
	}
	ctx.Set("studentid", "exampleStudentID")

	// Call the function
	controllers.GetPostFromEvent()(ctx)

	// Check the response status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check the response body
	expectedBody := `{"data":[],"success":true}` // Adjust based on actual implementation
	if w.Body.String() != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, w.Body.String())
	}
}
