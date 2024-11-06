package routes

import (
	"net/http"

	controllers "github.com/encall/cpeevent-backend/src/controllers"
	"github.com/encall/cpeevent-backend/src/middleware"
	"github.com/gin-gonic/gin"
)

// UserRoutes
func UserRoutes(route *gin.RouterGroup) {
	v1 := route.Group("/v1")

	v1.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello World"})
	})

	v1.GET("/events", controllers.GetEvents())
	v1.GET("/searchEvents", controllers.SearchEvents()) //usage: /searchEvents?name=XXXXXX
	// v1.GET("/event/:eventID/posts", controllers.GetPostFromEvent())

	// Group routes for user related operations
	userRoute := v1.Group("/user")
	userRoute.POST("/signup", controllers.SignUp())
	userRoute.POST("/login", controllers.Login())
	userRoute.POST("/logout", controllers.Logout())
	userRoute.POST("/refresh", controllers.RefreshToken())


	// Group routes that require authentication
	protected := v1.Group("/")
	protected.Use(middleware.Authentication(1)) // Example: Access level 1 required
	{
		protected.GET("/protected-route", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "This is a protected route"})
		})
		protected.GET("/testevent", controllers.TestEvents())
		protected.GET("/event/:eventID/posts", controllers.GetPostFromEvent())

		profile := protected.Group("/account")
		profile.GET("/info", controllers.GetInfo())
	}

	protected.Use(middleware.Authentication(2)) // Example: Access level 2 required
	{
		protected.GET("/protected-route2", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "This is a protected route with level 2 access"})
		})
	}
}
