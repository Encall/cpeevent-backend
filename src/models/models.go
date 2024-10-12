package models

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
