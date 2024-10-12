package routes

import (
	"net/http"

	"github.com/encall/cpeevent-backend/src/controllers"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRouter configures the Gin router and routes
func SetupRouter(client *mongo.Client) *gin.Engine {
	r := gin.Default()

	// Ping test route
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Event routes
	r.GET("/events", controllers.GetEvents(client))
	r.POST("/events", controllers.CreateEvent(client))

	return r
}
