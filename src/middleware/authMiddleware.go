package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/encall/cpeevent-backend/src/database"
	helper "github.com/encall/cpeevent-backend/src/helpers"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")

// Auth validates token and authorizes users
func Authentication(requiredAccessLevel int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("Authorization")
		if clientToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No Authorization header provided"})
			c.Abort()
			return
		}

		// Check if the token starts with "Bearer "
		if len(clientToken) < 7 || clientToken[:7] != "Bearer " {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		// Extract the token part
		clientToken = clientToken[7:]

		claims, msg := helper.ValidateToken(clientToken)
		if msg != "" {
			if strings.Contains(msg, "token is expired") {
				c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			}
			c.Abort()
			return
		}

		c.Set("studentid", claims.StudentID)
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
