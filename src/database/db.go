package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Dbinstance() *mongo.Client {
	mode := os.Getenv("GIN_MODE")
	if mode == "" || mode == "debug" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	mongoURI := os.Getenv("MONGO_URI")
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client
}

// Client Database instance
var Client *mongo.Client = Dbinstance()

func OpenCollection(client *mongo.Client, collecionName string) *mongo.Collection {
	databaseName := os.Getenv("DATABASE_NAME")
	var collection *mongo.Collection = client.Database(databaseName).Collection(collecionName)
	return collection

}
