package main

import (
	"github.com/encall/cpeevent-backend/src/routes"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables from .env file
	r := gin.Default()
	routes.UserRoutes(r)
	// r.Use(middleware.Authentication())

	// r.GET("/", func(ctx *gin.Context) {
	// 	ctx.JSON(http.StatusOK, gin.H{"data": "Hello World"})
	// })

	// Start the Gin server on port 8080
	
	r.Run(":8080")
	
}
