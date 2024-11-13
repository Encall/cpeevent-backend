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
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func SearchEvents() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		fmt.Println(name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"Error": "Missing the name parameter"})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var events []models.Event

		query := bson.M{"eventName": bson.M{"$regex": name, "$options": "i"}}

		cursor, err := eventCollection.Find(ctx, query)
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

func JoinEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID, exists := c.Get("studentid")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}

		type JoinRequest struct {
			EventID string `json:"eventID" binding:"required"`
			Role    string `json:"role" binding:"required"`
		}

		var joinRequest JoinRequest
		if err := c.BindJSON(&joinRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var event models.Event
		eventID, err := primitive.ObjectIDFromHex(joinRequest.EventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}

		err = eventCollection.FindOne(ctx, bson.M{"_id": eventID}).Decode(&event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Event not found"})
			return
		}

		// Check if user is already a staff member or participant
		isStaff := false
		isParticipant := false
		for _, staffID := range event.Staff {
			if staffID == userID {
				isStaff = true
				break
			}
		}
		for _, participantID := range event.Participants {
			if participantID == userID {
				isParticipant = true
				break
			}
		}

		if (joinRequest.Role == "staff" && isParticipant) || (joinRequest.Role == "participant" && isStaff) {
			c.JSON(http.StatusConflict, gin.H{"error": "User cannot be both staff and participant"})
			return
		}

		var StaffMember models.StaffMember

		update := bson.D{}
		if joinRequest.Role == "staff" {
			StaffMember.StdID = userID.(string)
			StaffMember.Role = ""
			update = bson.D{
				{"$addToSet", bson.D{
					{"staff", StaffMember},
				}},
			}
		} else if joinRequest.Role == "participant" {
			update = bson.D{
				{"$addToSet", bson.D{
					{"participants", userID},
				}},
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
			return
		}

		result, err := eventCollection.UpdateOne(ctx, bson.M{"_id": eventID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error joining event"})
			return
		}

		if result.ModifiedCount == 0 {
			c.JSON(http.StatusConflict, gin.H{"message": "User already in event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result, "message": "Joined event successfully"})
	}
}

func LeaveEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		userID, exists := c.Get("studentid")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}

		type LeaveRequest struct {
			EventID string `json:"eventID" binding:"required"`
		}

		var leaveRequest LeaveRequest
		if err := c.BindJSON(&leaveRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var event models.Event
		eventID, err := primitive.ObjectIDFromHex(leaveRequest.EventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
			return
		}

		err = eventCollection.FindOne(ctx, bson.M{"_id": eventID}).Decode(&event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Event not found"})
			return
		}

		// Check if user is a staff member or participant
		isStaff := false
		isParticipant := false
		for _, staffID := range event.Staff {
			if staffID.StdID == userID {
				isStaff = true
				break
			}
		}
		for _, participantID := range event.Participants {
			if participantID == userID {
				isParticipant = true
				break
			}
		}

		if !isStaff && !isParticipant {
			c.JSON(http.StatusConflict, gin.H{"error": "User is not part of the event"})
			return
		}

		update := bson.D{}
		if isStaff {
			update = bson.D{
				{"$pull", bson.D{
					{"staff", bson.D{{"stdID", userID}}},
				}},
			}
		} else if isParticipant {
			update = bson.D{
				{"$pull", bson.D{
					{"participants", userID},
				}},
			}
		}

		result, err := eventCollection.UpdateOne(ctx, bson.M{"_id": eventID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error leaving event"})
			return
		}

		if result.ModifiedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "User not in event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result, "message": "Left event successfully"})
	}
}
