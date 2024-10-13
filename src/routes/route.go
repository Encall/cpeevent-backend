package routes

import (
	"net/http"

	controllers "github.com/encall/cpeevent-backend/src/controllers"
	"github.com/encall/cpeevent-backend/src/middleware"
	"github.com/gin-gonic/gin"
)

// UserRoutes
func UserRoutes(route *gin.RouterGroup) {
	route.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	})

	route.POST("/user/signup", controllers.SignUp())
	route.POST("/user/login", controllers.Login())
	route.GET("/events", controllers.GetEvents())
	route.GET("/searchEvents", controllers.SearchEvents()) //usage: /searchEvents?name=XXXXXX

	// Group routes that require authentication
	protected := route.Group("/")
	protected.Use(middleware.Authentication(1)) // Example: Access level 1 required
	{
		protected.GET("/protected-route", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "This is a protected route"})
		})
	}

	protected.Use(middleware.Authentication(2)) // Example: Access level 2 required
	{
		protected.GET("/protected-route2", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "This is a protected route with level 2 access"})
		})
	}
}
