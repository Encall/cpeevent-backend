package models

import "time"

type User struct {
	StudentID     string    `json:"studentID" validate:"required"`
	FirstName     string    `json:"firstName" validate:"required"`
	LastName      string    `json:"lastName" validate:"required"`
	Year          int       `json:"year" validate:"required"`
	ImgProfile    *string   `json:"imgProfile"` // Optional
	Email         string    `json:"email" validate:"required,email"`
	Password      string    `json:"password" validate:"required,min=6"`
	PhoneNumber   string    `json:"phoneNumber" validate:"required"`
	Username      string    `json:"username" validate:"required"`
	Access        int       `json:"access" validate:"required"`
	Token         *string   `json:"token"`
	Refresh_token *string   `json:"refresh_token" `
	Created_at    time.Time `json:"created_at" `
	Updated_at    time.Time `json:"updated_at" `
}
