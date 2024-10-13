package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/encall/cpeevent-backend/src/database"
	helper "github.com/encall/cpeevent-backend/src/helpers"
	"github.com/encall/cpeevent-backend/src/models"
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

		// Fetch user details from the database
		var user models.User
		if err := userCollection.FindOne(context.Background(), bson.M{"email": claims.Email}).Decode(&user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Check if the user has the required access level
		if user.Access < requiredAccessLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient access level"})
			c.Abort()
			return
		}

		c.Next()
	}
}
