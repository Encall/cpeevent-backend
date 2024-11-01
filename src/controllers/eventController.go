package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	models "github.com/encall/cpeevent-backend/src/models"

	database "github.com/encall/cpeevent-backend/src/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var eventCollection *mongo.Collection = database.OpenCollection(database.Client, "events")

func GetEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var events []models.Event

		cursor, err := eventCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		if err = cursor.All(ctx, &events); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": events})
	}
}

func SearchEvents() gin.HandlerFunc{
	return func(c *gin.Context){
		name := c.Query("name")
		fmt.Println(name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Missing the name parameter"})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var events []models.Event

		query := bson.M{"eventName": bson.M{"$regex": name, "$options":"i"}}

		cursor, err :=eventCollection.Find(ctx, query)
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		if err = cursor.All(ctx, &events); err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": events})
	}
}

func TestEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("studentid")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"userID": userID})
	}
}