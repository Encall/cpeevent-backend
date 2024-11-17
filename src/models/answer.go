package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AQuestion struct {
	questionIndex int      `bson:"questionIndex" json:"questionIndex"`
	answers       []string `bson:"answers" json:"answers"`
}

type AnswerHeader struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	postID    primitive.ObjectID `bson:"postID" json:"postID"`
	studentID string             `bson:"studentID" json:"studentID"`
}

// PPost extends Post for regular posts.
type AForm struct {
	AnswerHeader
	inputType  string      `bson:"inputType" json:"inputType"`
	answerList []AQuestion `bson:"answerList" json:"answerList"`
}

// PVote extends Post for vote posts.
type AVote struct {
	AnswerHeader
	answer string `bson:"answer" json:"answer"`
}
