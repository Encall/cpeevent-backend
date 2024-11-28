package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")
// var validate = validator.New()

type UpdateAccountInfo struct { //You have to name the struct field according to the JSON attribute
	StudentID   string `json:"studentID" bson:"studentID"`
	FirstName   string `json:"firstName" bson:"firstName"`
	LastName    string `json:"lastName" bson:"lastName"`
	Year        int    `json:"year" bson:"year"`
	Email       string `json:"email" bson:"email"`
	PhoneNumber string `json:"phoneNumber" bson:"phoneNumber"`
}

type UpdateUsernameInfo struct {
	Username string `json:"username" bson:"username"`
}

func GetUsername() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID, exists := c.Get("studentid")

		fmt.Println(userID)

		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}

		var username UpdateUsernameInfo

		projection := bson.M{
			"username": 1,
		}

		err := userCollection.FindOne(ctx, bson.M{"studentID": userID}, options.FindOne().SetProjection(projection)).Decode(&username)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{"sucess": true, "data": username})

	}
}

func UpdateUsername() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID, exists := c.Get("studentid")
		defer cancel()
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}
		var username UpdateUsernameInfo

		if err := c.BindJSON(&username); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateInfo := bson.M{
			"username": username.Username,
		}

		result, err := userCollection.UpdateOne(ctx, bson.M{"studentID": userID}, bson.M{"$set": updateInfo})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "result": result, "message": "Account info updated successfully"})
	}
}

func GetInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID, exists := c.Get("studentid")

		// fmt.Print(userID)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}
		var info UpdateAccountInfo

		projection := bson.M{
			"studentID":   1,
			"firstName":   1,
			"lastName":    1,
			"year":        1,
			"email":       1,
			"phoneNumber": 1,
		}

		// err := userCollection.FindOne(ctx, bson.M{"studentid": userID}).Decode(&foundUser)

		err := userCollection.FindOne(ctx, bson.M{"studentID": userID}, options.FindOne().SetProjection(projection)).Decode(&info)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{"sucess": true, "data": info})

	}
}

func UpdateInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID, exists := c.Get("studentid")
		var info UpdateAccountInfo
		defer cancel()

		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}

		if err := c.BindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateInfo := bson.M{
			"firstName":   info.FirstName,
			"lastName":    info.LastName,
			"year":        info.Year,
			"email":       info.Email,
			"phoneNumber": info.PhoneNumber,
		}

		result, err := userCollection.UpdateOne(ctx, bson.M{"studentID": userID}, bson.M{"$set": updateInfo})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "result": result, "message": "Account info updated successfully"})
	}
}
