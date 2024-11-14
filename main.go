package main

import (
	"os"
	"time"

	"github.com/encall/cpeevent-backend/src/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

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
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	origin := os.Getenv("ORIGIN_URL")
	// CORS configuration
	config := cors.Config{
		AllowOrigins:     []string{origin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "refresh_token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(config))

	// Register all routes with /api prefix
	api := r.Group("/api")
	routes.UserRoutes(api)

	// Start the Gin server on port 8080
	r.Run(":8080")

}
