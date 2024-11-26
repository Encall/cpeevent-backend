package controllers

import (
	"context"
	"fmt"
	"log"
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

func CreateNewEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var event models.Event

		if err := c.BindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := eventCollection.InsertOne(ctx, event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result, "message": "Event created successfully"})
	}
}

func GetEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		eventID := c.Param("eventID")

		if eventID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
			return
		}

		objectID, err := primitive.ObjectIDFromHex(eventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid eventID format"})
			return
		}

		var event models.Event
		err = eventCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&event)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": event})
	}
}

func UpdateEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Define a struct to represent the request body
		var req models.Event
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.ID == primitive.NilObjectID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
			return
		}

		objectID := req.ID

		update := bson.D{{Key: "$set", Value: req}}

		result, err := eventCollection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result, "message": "Event updated successfully"})
	}
}

func DeleteEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Define a struct to represent the request body
		type DeleteEventRequest struct {
			EventID string `json:"_id"`
		}

		var req DeleteEventRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.EventID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
			return
		}

		objectID, err := primitive.ObjectIDFromHex(req.EventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid eventID format"})
			return
		}

		result, err := eventCollection.DeleteOne(ctx, bson.M{"_id": objectID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting event"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": result, "message": "Event deleted successfully"})
	}
}

func AddPostToPostList(postID primitive.ObjectID, eventID primitive.ObjectID) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// Log the postID and eventID
	log.Printf("Adding postID: %s to eventID: %s", postID.Hex(), eventID.Hex())

	update := bson.D{
		{"$addToSet", bson.D{
			{"postList", postID},
		}},
	}

	// Perform the update operation
	result, err := eventCollection.UpdateOne(ctx, bson.M{"_id": eventID}, update)
	if err != nil {
		log.Printf("Error updating event: %v", err)
		return err
	}

	// Log the result of the update operation
	log.Printf("MatchedCount: %d, ModifiedCount: %d", result.MatchedCount, result.ModifiedCount)

	return nil
}

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

func GetAllRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID := c.Param("eventID")
		if eventID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
			return
		}
		var event models.Event
		objectID, err := primitive.ObjectIDFromHex(eventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid eventID format"})
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		err = eventCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Event not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"data": event.Role})

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
		for _, staff := range event.Staff {
			if staff.StdID == userID {
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

		var staffMember models.StaffMember

		update := bson.D{}
		if joinRequest.Role == "staff" {
			staffMember.StdID = userID.(string)
			staffMember.Role = ""
			update = bson.D{
				{"$addToSet", bson.D{
					{"staff", staffMember},
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
		for _, staff := range event.Staff {
			if staff.StdID == userID {
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

func GetEventMembers() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		eventID := c.Param("eventID")
		if eventID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "eventID is required"})
			return
		}

		objectID, err := primitive.ObjectIDFromHex(eventID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid eventID format"})
			return
		}

		pipeline := mongo.Pipeline{
			{{"$match", bson.D{{"_id", objectID}}}},
			{{"$lookup", bson.D{
				{"from", "users"},
				{"localField", "participants"},
				{"foreignField", "studentID"},
				{"as", "participantDetails"},
			}}},
			{{"$lookup", bson.D{
				{"from", "users"},
				{"localField", "staff.stdID"},
				{"foreignField", "studentID"},
				{"as", "staffDetails"},
			}}},
			{{"$project", bson.D{
				{"eventID", "$_id"},
				{"participants", bson.D{
					{"$map", bson.D{
						{"input", "$participantDetails"},
						{"as", "participant"},
						{"in", bson.D{
							{"stdID", "$$participant.studentID"},
							{"name", bson.D{{"$concat", []interface{}{"$$participant.firstName", " ", "$$participant.lastName"}}}},
							{"phoneNumber", "$$participant.phoneNumber"},
						}},
					}},
				}},
				{"staff", bson.D{
					{"$map", bson.D{
						{"input", "$staffDetails"},
						{"as", "staffMember"},
						{"in", bson.D{
							{"stdID", "$$staffMember.studentID"},
							{"name", bson.D{{"$concat", []interface{}{"$$staffMember.firstName", " ", "$$staffMember.lastName"}}}},
							{"phoneNumber", "$$staffMember.phoneNumber"},
							{"role", bson.D{
								{"$arrayElemAt", []interface{}{"$staff.role", bson.D{{"$indexOfArray", []interface{}{"$staff.stdID", "$$staffMember.studentID"}}}}},
							}},
						}},
					}},
				}},
			}}},
			{{"$addFields", bson.D{
				{"participants", bson.D{
					{"$filter", bson.D{
						{"input", "$participants"},
						{"as", "participant"},
						{"cond", bson.D{{"$ne", bson.A{"$$participant.name", nil}}}},
					}},
				}},
			}}},
		}

		cursor, err := eventCollection.Aggregate(ctx, pipeline)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error fetching event members"})
			return
		}
		defer cursor.Close(ctx)

		var results []bson.M
		if err = cursor.All(ctx, &results); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error decoding event members"})
			return
		}

		if len(results) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
			return
		}

		c.JSON(http.StatusOK, results[0])
	}
}

/*
Expeced Response:
{
eventID: eventid,
participants: [
0: {
stdID: studentid,
name: firstname lastname
phoneNumber: phonenumber},
]
staff: [
0: {
stdID: studentid,
name: firstname lastname
phoneNumber: phonenumber,
role: role}]

}
*/
