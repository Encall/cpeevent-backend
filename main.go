package main

import (
	"log"
	"os"

	"github.com/encall/cpeevent-backend/src/db"
	"github.com/encall/cpeevent-backend/src/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get MongoDB URI from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MONGO_URI is not set in .env file")
	}

	// Connect to MongoDB
	client := db.ConnectMongoDB(mongoURI)
	defer client.Disconnect(nil) // Ensure the MongoDB client is disconnected at shutdown

	// Set up routes
	r := routes.SetupRouter(client)

	// Start the Gin server on port 8080
	r.Run(":8080")
}
