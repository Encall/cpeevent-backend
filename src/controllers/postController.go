package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	database "github.com/encall/cpeevent-backend/src/database"
	models "github.com/encall/cpeevent-backend/src/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var postCollection *mongo.Collection = database.OpenCollection(database.Client, "posts")

func DeleteAllPosts(eventID primitive.ObjectID) error{
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	var event models.Event

	// This module is to find the event according to the input eventID
	filter := bson.M{"_id": eventID}
	err := eventCollection.FindOne(ctx, filter).Decode(&event)
	if err != nil{
		log.Println("error finding event :", err)
		return err
	}

	//List all the posts which are in event ID
	var postList = event.PostList
	log.Println(postList)

	// DeleteAllAnswers(event.PostList[0])

	for _, postID := range event.PostList{
		if err := DeleteAllAnswers(postID); err != nil{
			log.Println("error deleting for postID: ", postID, err)
			return err
		}
	}

	// This module is to delete all post which are in the postList	
	deleteFilter := bson.M{"_id": bson.M{"$in": postList}}
	_, err = postCollection.DeleteMany(ctx, deleteFilter)
	if err != nil{
		log.Println("Error deleting posts: ", err)
		return err
	}
	return err

}

func UpdatePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var post models.Post
		if err := c.BindJSON(&post); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updatePost := bson.M{
			"assignTo":    post.AssignTo,
			"public":      post.Public,
			"title":       post.Title,
			"description": post.Description,
			"endDate":     post.EndDate,
		}

		switch post.Kind {
		case "post":
			updatePost["markdown"] = post.Markdown
		case "vote":
			updatePost["voteQuestions"] = post.VoteQuestions
		case "form":
			updatePost["formQuestions"] = post.FormQuestions
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post kind"})
			return
		}

		objID, err := primitive.ObjectIDFromHex(post.PostID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postID format"})
		}
		result, err := postCollection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updatePost})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": post})
	}
}

func DeletePost() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		type DeletePost struct {
			EventID string `json:"eventID"`
			PostID  string `json:"postID"`
		}

		var postID DeletePost
		if err := c.BindJSON(&postID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		postObjID, err := primitive.ObjectIDFromHex(postID.PostID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postID format"})
			return
		}

		eventObjID, err := primitive.ObjectIDFromHex(postID.EventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid eventID format"})
			return
		}

		err = DeletePostFromPostList(postObjID, eventObjID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		_, err = postCollection.DeleteOne(ctx, bson.M{"_id": postObjID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		_, err = transactionCollection.DeleteMany(ctx, bson.M{"postID": postObjID})
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting transactions"})
            return
        }

		c.JSON(http.StatusOK, gin.H{"success": true, "data": postID})
	}
}

func NewPost(post models.Post, timeUp bool) interface{} {
	switch post.Kind {
	case "post":
		// Create and return a PPost
		return models.PPost{Post: post, TimeUp: timeUp}
	case "vote":
		// Create and return a PVote with questions
		return models.PVote{Post: post, Questions: post.VoteQuestions, TimeUp: timeUp}
	case "form":
		return models.PForm{Post: post, Questions: post.FormQuestions, TimeUp: timeUp}
	default:
		// Handle unknown post kinds, return nil or an error if needed
		return nil
	}
}

func CreateNewPost() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources

		// Bind the JSON data to a CreatePostRequest struct
		var request models.CreatePostRequest

		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// Log the eventID
		eventID := request.EventID

		// Initialize the ID field if it's not already set
		if request.UpdatedPost.ID.IsZero() {
			request.UpdatedPost.ID = primitive.NewObjectID()
		}

		// Insert the post document
		_, err := postCollection.InsertOne(ctx, request.UpdatedPost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Call AddPostToPostList to add the post ID to the event's post list
		err = AddPostToPostList(request.UpdatedPost.ID, eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": request})
	}
}

func GetPostFromEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources

		userID, exists := c.Get("studentid")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}

		// Get the eventID from the URL parameters
		eventID := c.Param("eventID")
		if eventID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
			return
		}

		// Parse eventID as an ObjectID
		objectID, err := primitive.ObjectIDFromHex(eventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid eventID format"})
			return
		}

		// Query the event by its ID
		var event models.Event
		if err := eventCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&event); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Event not found"})
			return
		}

		isParticipant := false
		isStaff := false
		for _, participant := range event.Participants {
			if participant == userID {
				isParticipant = true
				break
			}
		}
		role := ""
		for _, staff := range event.Staff {
			if staff.StdID == userID {
				isStaff = true
				role = staff.Role
				break
			}
		}

		// Check if the user is a participant or staff in the event
		if !isParticipant && !isStaff {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}})
			return
		}

		// Query the posts collection based on user role
		var posts []models.Post
		var cursor *mongo.Cursor
		if isStaff {
			cursor, err = postCollection.Find(ctx, bson.M{"_id": bson.M{"$in": event.PostList}, "$or": []bson.M{{"assignTo": role}, {"assignTo": "everyone"}, {"public": true}}})
		} else {
			cursor, err = postCollection.Find(ctx, bson.M{"_id": bson.M{"$in": event.PostList}, "public": true})
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving posts"})
			return
		}
		defer cursor.Close(ctx)

		// Decode all the posts from the cursor
		if err = cursor.All(ctx, &posts); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error decoding posts"})
			return
		}

		// Create a slice to hold specific post types
		var specificPosts []interface{}

		// Convert each post to its specific type based on the Kind
		for _, post := range posts {
			specificPost := NewPost(post, false) // Convert to specific type
			if post.EndDate != nil {
				postEndDateLocal := post.EndDate.Time()
				currentTimeLocal := time.Now().Add(time.Hour * 7)

				if postEndDateLocal.Before(currentTimeLocal) {
					specificPost = NewPost(post, true)
				}
			}

			if specificPost == nil {
				continue // Or handle unknown kind if needed
			}
			specificPosts = append(specificPosts, specificPost)
		}

		// Respond with the specific posts data
		c.JSON(http.StatusOK, gin.H{"success": true, "data": specificPosts})
	}
}

func UpdateEventHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var cancel context.CancelFunc
		_, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Define a struct to represent the request body
		type UpdateEventRequest struct {
			EventID string       `json:"eventID"`
			Event   models.Event `json:"event"`
		}

		var req UpdateEventRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.EventID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
			return
		}

		_, err := primitive.ObjectIDFromHex(req.EventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid eventID format"})
			return
		}

		// Use req.Event for the event data
		event := req.Event

		// Continue with the rest of the update logic...
		// For example, update the event in the database
		// db.UpdateEvent(objectID, event)

		c.JSON(http.StatusOK, gin.H{"data": event})
	}
}
func GetPostFromPostId() gin.HandlerFunc {

	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources

		// Get the postID from the URL parameters
		postID := c.Param("postID")
		if postID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "postID is required"})
			return
		}

		// Parse postID as an ObjectID
		objectID, err := primitive.ObjectIDFromHex(postID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid postID format"})
			return
		}

		// Query the post by its ID
		var post models.Post
		if err := postCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Post not found"})
			return
		}

		// Convert the post to its specific type based on the Kind
		specificPost := NewPost(post, false)
		if post.EndDate != nil {
			postEndDateLocal := post.EndDate.Time()
			currentTimeLocal := time.Now().Add(time.Hour * 7)

			if postEndDateLocal.Before(currentTimeLocal) {
				specificPost = NewPost(post, true)
			}
		}

		if specificPost == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown post kind"})
			return
		}

		// Respond with the specific post data
		c.JSON(http.StatusOK, gin.H{"success": true, "data": specificPost})
	}
}
