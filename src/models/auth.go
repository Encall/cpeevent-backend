package models

import "time"

type User struct {
	StudentID     string    `json:"studentID" bson:"studentID" validate:"required"`
	FirstName     string    `json:"firstName" bson:"firstName" validate:"required"`
	LastName      string    `json:"lastName" bson:"lastName" validate:"required"`
	Year          int       `json:"year" bson:"year" validate:"required"`
	ImgProfile    *string   `json:"imgProfile" bson:"imgProfile"` // Optional
	Email         string    `json:"email" bson:"email" validate:"required,email"`
	Password      string    `json:"password" bson:"password" validate:"required,min=6"`
	PhoneNumber   string    `json:"phoneNumber" bson:"phoneNumber" validate:"required"`
	Username      string    `json:"username" bson:"username" validate:"required"`
	Access        int       `json:"access" bson:"access"`
	Token         *string   `json:"token" bson:"token"`
	Refresh_token *string   `json:"refresh_token" bson:"refresh_token"`
	Created_at    time.Time `json:"created_at" bson:"created_at"`
	Updated_at    time.Time `json:"updated_at" bson:"updated_at"`
}
