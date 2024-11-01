package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Event represents an event in the system
type StaffMember struct {
	StdID string `json:"stdID"`
	Role  string `json:"role"`
}

// Define the Event struct with updated types
type Event struct {
	ID               primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	EventName        string        `json:"eventName" binding:"required"`
	EventDescription string        `json:"eventDescription"`
	NParticipant     int           `json:"nParticipant"`
	Participants     []string      `json:"participants"`
	NStaff           int           `json:"nStaff"`
	Staff            []StaffMember `json:"staff"` // Use StaffMember struct here
	StartDate        time.Time     `json:"startDate"`
	EndDate          time.Time     `json:"endDate"`
	President        string        `json:"president"`
	Kind             string        `json:"kind"`
	Role             []string      `json:"role"`
	Icon             *string       `json:"icon"`
	Poster           *string       `json:"poster"`
	PostList         []primitive.ObjectID `json:"postList"`
}
