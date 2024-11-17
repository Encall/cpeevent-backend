package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	database "github.com/encall/cpeevent-backend/src/database"
	models "github.com/encall/cpeevent-backend/src/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var transactionCollection *mongo.Collection = database.OpenCollection(database.Client, "transactions")

func SubmitAnswer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources

		// Parse the request body to get the postID
		var request struct {
			PostID primitive.ObjectID `json:"postID"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Query the post by its ID
		var post models.Post
		if err := postCollection.FindOne(ctx, bson.M{"_id": request.PostID}).Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Post not found"})
			return
		}

		switch post.Kind {
		case "vote":
			var voteRequest models.AVote
			if err := c.BindJSON(&voteRequest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// insert the vote request into transactions
			_, err := transactionCollection.InsertOne(ctx, voteRequest)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		case "form":
			var formRequest models.AForm
			if err := c.BindJSON(&formRequest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// insert the form request into transactions
			_, err := transactionCollection.InsertOne(ctx, formRequest)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown post kind"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": "answer submitted"})
	}
}

func GetUserAnswer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources
		var request struct {
			StudentID string             `json:"studentID"`
			PostID    primitive.ObjectID `json:"postID"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Query the post by its ID
		log.Print(request.PostID)
		var post models.Post
		if err := postCollection.FindOne(ctx, bson.M{"_id": request.PostID}).Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Post not found"})
			return
		}
		log.Print(post.Kind)
		switch post.Kind {
		case "vote":
			var vote models.AVote
			if err := transactionCollection.FindOne(ctx, bson.M{"postID": request.PostID, "studentID": request.StudentID}).Decode(&vote); err != nil {
				log.Printf("Error finding vote: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Vote not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": vote})

		case "form":
			var form models.AForm
			query := bson.M{"postID": request.PostID, "studentID": request.StudentID}
			if err := transactionCollection.FindOne(ctx, query).Decode(&form); err != nil {
				log.Printf("Error finding form: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Form not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true, "data": form})
		}
	}
}
