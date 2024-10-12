package models

import "time"

// Event represents an event in the system
type Event struct {
	EventName        string    `json:"eventName" binding:"required"`
	EventDescription string    `json:"eventDescription"`
	NParticipant     int       `json:"nParticipant"`
	Participants     []string  `json:"participants"`
	NStaff           int       `json:"nStaff"`
	StartDate        time.Time `json:"startDate"`
	EndDate          time.Time `json:"endDate"`
	President        string    `json:"president"`
	Kind             string    `json:"kind"`
	Role             []string  `json:"role"`
	Icon             *string   `json:"icon"`
	Poster           *string   `json:"poster"`
}
type User struct {
	StudentID   string `json:"studentID"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Year        string `json:"year"`
	ImgProfile  string `json:"imgProfile"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
	Username    string `json:"username"`
	Access      string `json:"access"`
}
