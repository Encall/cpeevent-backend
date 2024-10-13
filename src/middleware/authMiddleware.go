package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/encall/cpeevent-backend/src/database"
	helper "github.com/encall/cpeevent-backend/src/helpers"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")

// Auth validates token and authorizes users
func Authentication(requiredAccessLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("access", claims.Access)

		// Check if the user has the required access level
		if claims.Access < requiredAccessLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient access level"})
			c.Abort()
			return
		}

		c.Next()
	}
}
