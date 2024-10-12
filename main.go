package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"github.com/gin-gonic/gin"
)

var db = make(map[string]string)

func connectToMongo() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27023")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal("Could not connect to MongoDB:", err)
	}

	fmt.Println("Connected to MongoDB!")
	return client
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// GET /events (retrieve all events)
	r.GET("/events", func(c *gin.Context) {
		client := connectToMongo()
		collection := client.Database("cpeEVO").Collection("events")

		var results []struct {
			EventName       string    `json:"eventName"`
			EventDescription string   `json:"eventDescription"`
			NParticipant    int       `json:"nParticipant"`
			Participants    []string  `json:"participants"`
			NStaff          int       `json:"nStaff"`
			StartDate       time.Time `json:"startDate"`
			EndDate         time.Time `json:"endDate"`
			President       string    `json:"president"`
			Kind            string    `json:"kind"`
			Role            []string  `json:"role"`
			Icon            *string   `json:"icon"`
			Poster          *string   `json:"poster"`
		}

		cursor, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cursor.Close(context.Background())

		if err = cursor.All(context.Background(), &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
	})

	// POST /events (insert new event)
	r.POST("/events", func(c *gin.Context) {
		client := connectToMongo()
		collection := client.Database("cpeEVO").Collection("events")

		var event struct {
			EventName       string    `json:"eventName" binding:"required"`
			EventDescription string   `json:"eventDescription"`
			NParticipant    int       `json:"nParticipant"`
			Participants    []string  `json:"participants"`
			NStaff          int       `json:"nStaff"`
			StartDate       time.Time `json:"startDate"`
			EndDate         time.Time `json:"endDate"`
			President       string    `json:"president"`
			Kind            string    `json:"kind"`
			Role            []string  `json:"role"`
			Icon            *string   `json:"icon"`
			Poster          *string   `json:"poster"`
		}

		// Bind the incoming JSON to the event struct
		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Insert the event into MongoDB
		_, err := collection.InsertOne(context.Background(), event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Event created successfully!"})
	})

	return r
}

func main() {
	r := setupRouter()
	r.Run(":8080")
}
