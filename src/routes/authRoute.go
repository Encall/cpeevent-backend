package routes

import (
	controllers "github.com/encall/cpeevent-backend/src/controllers"
	"github.com/gin-gonic/gin"
)

// UserRoutes
func UserRoutes(route *gin.Engine) {
	route.POST("/users/signup", controllers.SignUp())
	route.POST("/users/login", controllers.Login())
	route.GET("/events", controllers.GetEvents())
}