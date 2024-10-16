package controllers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	models "github.com/encall/cpeevent-backend/src/models"

	database "github.com/encall/cpeevent-backend/src/database"

	helper "github.com/encall/cpeevent-backend/src/helpers"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")
var validate = validator.New()

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// HashPassword is used to encrypt the password
func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

// VerifyPassword checks
func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = "Login or Password is incorrect"
		check = false
	}

	return check, msg
}

// sign up user
func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		defer cancel()

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Println("Incoming JSON payload for SignUp:", user)

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "data": nil, "message": validationErr.Error()})
			return
		}

		count, err := userCollection.CountDocuments(ctx, bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"studentID": user.StudentID},
			},
		})

		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError,
				gin.H{"success": false, "data": nil, "message": "error occured while checking for the email"})

			return
		}

		password := HashPassword(user.Password)
		user.Password = password

		// Default user access when signing up
		user.Access = 1

		if count > 0 {
			c.JSON(http.StatusConflict,
				gin.H{"success": false, "data": nil, "message": "email or studentid ready exists"})

			return
		}

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		token, refreshToken, _ := helper.GenerateAllTokens(user.Email, user.Access)
		user.Token = &token
		user.Refresh_token = &refreshToken

		result, insertErr := userCollection.InsertOne(ctx, user)
		log.Println("insertErr:", insertErr)
		log.Println("result:", result)

		if insertErr != nil {
			msg := "User item was not created"
			c.JSON(http.StatusInternalServerError,
				gin.H{"success": false, "data": nil, "message": msg})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": result.InsertedID, "message": "user signup success"})
	}
}

// Login user
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var loginRequest LoginRequest
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Println("Incoming JSON payload for Login:", loginRequest)

		var foundUser models.User
		err := userCollection.FindOne(ctx, bson.M{"email": loginRequest.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Login or Password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(loginRequest.Password, foundUser.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		if foundUser.Email == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		token, refreshToken, _ := helper.GenerateAllTokens(foundUser.Email, foundUser.Access)
		helper.UpdateAllTokens(token, refreshToken, foundUser.Email)

		err = userCollection.FindOne(ctx, bson.M{"email": foundUser.Email}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"user":          foundUser.Email,
			"access":        foundUser.Access,
			"token":         foundUser.Token,
			"refresh_token": foundUser.Refresh_token},
			"message": "return successfully"})
	}
}

func Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		accessToken := c.GetHeader("Authorization")
		log.Println("accessToken:", accessToken)
		if accessToken == "" || !strings.HasPrefix(accessToken, "Bearer ") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No access token provided"})
			return
		}
		token := strings.TrimPrefix(accessToken, "Bearer ")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No access token provided"})
			return
		}

		claims, msg := helper.ValidateToken(token)
		if msg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		updateObj := bson.D{
			{"token", ""},
			{"refresh_token", ""},
		}

		_, err := userCollection.UpdateOne(
			ctx,
			bson.M{"email": claims.Email},
			bson.D{
				{"$set", updateObj},
			},
		)
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging out"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}

func RefreshToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		accessToken := c.Request.Header.Get("token")
		if accessToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No access token provided"})
			return
		}

		accessClaims, msg := helper.ValidateToken(accessToken)
		if msg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		refreshToken := c.Request.Header.Get("refresh_token")
		if refreshToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No refresh token provided"})
			return
		}

		_, msg = helper.ValidateToken(refreshToken)
		if msg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return
		}

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"email": accessClaims.Email, "refresh_token": refreshToken}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
			return
		}

		newToken, newRefreshToken, err := helper.GenerateAllTokens(accessClaims.Email, accessClaims.Access)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		updateObj := bson.D{
			{"token", newToken},
			{"refresh_token", newRefreshToken},
		}

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		_, err = userCollection.UpdateOne(
			ctx,
			bson.M{"email": accessClaims.Email},
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating tokens"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{
			"token":         newToken,
			"refresh_token": newRefreshToken},
			"message": "return successfully"})
	}
}
