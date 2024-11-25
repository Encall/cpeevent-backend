package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents an event in the system
type StaffMember struct {
	StdID string `json:"stdID" bson:"stdID"`
	Role  string `json:"role" bson:"role"`
}

// Define the Event struct with updated types
type Event struct {
	ID               primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	EventName        string               `json:"eventName" bson:"eventName" binding:"required"`
	EventDescription string               `json:"eventDescription" bson:"eventDescription"`
	NParticipant     *int                 `json:"nParticipant" bson:"nParticipant"`
	Participants     []string             `json:"participants" bson:"participants"`
	NStaff           *int                 `json:"nStaff" bson:"nStaff"`
	Staff            []StaffMember        `json:"staff" bson:"staff"`
	StartDate        time.Time            `json:"startDate" bson:"startDate"`
	EndDate          time.Time            `json:"endDate" bson:"endDate"`
	President        *string              `json:"president" bson:"president"`
	Kind             string               `json:"kind" bson:"kind"`
	Role             []string             `json:"role" bson:"role"`
	Icon             *string              `json:"icon" bson:"icon"`
	Poster           *string              `json:"poster" bson:"poster"`
	PostList         []primitive.ObjectID `json:"postList" bson:"postList"`
}
