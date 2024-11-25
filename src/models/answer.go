package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AQuestion struct {
	QuestionIndex int      `bson:"questionIndex" json:"questionIndex"`
	InputType     string   `bson:"inputType" json:"inputType"`
	Answers       []string `bson:"answers" json:"answers"`
}

type AForm struct {
	ID         primitive.ObjectID `bson:"_id" json:"_id"`
	PostID     primitive.ObjectID `bson:"postID" json:"postID"`
	StudentID  string             `bson:"studentID" json:"studentID"`
	AnswerList []AQuestion        `bson:"answerList" json:"answerList"`
}

type AVote struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	PostID    primitive.ObjectID `bson:"postID" json:"postID"`
	StudentID string             `bson:"studentID" json:"studentID"`
	Answer    string             `bson:"answer" json:"answer"`
}
