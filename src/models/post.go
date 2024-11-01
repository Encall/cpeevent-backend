package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Question represents a question in a vote post.
type Question struct {
	Question string   `bson:"question" json:"question"`
	Type     string   `bson:"type" json:"type"`
	Options  []string `bson:"options" json:"options"`
}

// Post represents a general post.
type Post struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	Kind        string             `bson:"kind" json:"kind"`
	AssignTo    string             `bson:"assignTo" json:"assignTo"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	PostDate    primitive.DateTime `bson:"postDate" json:"postDate"`
	EndDate     *primitive.DateTime `bson:"endDate,omitempty" json:"endDate,omitempty"` // Nullable
	Author      string             `bson:"author" json:"author"`
	Mardown     string             `bson:"mardown" json:"mardown"` // Correct the spelling here
	Questions    []Question        `bson:"questions,omitempty" json:"questions,omitempty"` // For vote posts
}

// PPost extends Post for regular posts.
type PPost struct {
	Post
}

// PVote extends Post for vote posts.
type PVote struct {
	Post
	Questions []Question `bson:"questions" json:"questions"` // Include questions for vote posts
}
