package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/encall/cpeevent-backend/src/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetEvents retrieves all events from the MongoDB
func GetEvents(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		collection := client.Database("cpeEVO").Collection("events")
		var results []models.Event

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := collection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
	}
}

// CreateEvent inserts a new event into the MongoDB
func CreateEvent(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		collection := client.Database("cpeEVO").Collection("events")

		var event models.Event
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := collection.InsertOne(ctx, event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Event created successfully!"})
	}
}
