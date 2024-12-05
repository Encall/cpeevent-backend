package controllers

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

	database "github.com/encall/cpeevent-backend/src/database"
	models "github.com/encall/cpeevent-backend/src/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var transactionCollection *mongo.Collection = database.OpenCollection(database.Client, "transactions")

type QuestionForm struct {
	QuestionIndex int    `json:"questionIndex"`
	Question      string `json:"question"`
}

func SubmitAnswer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources

		// Log the request body
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Re-bind the request body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Parse the request body to get the postID
		var request struct {
			PostID primitive.ObjectID `json:"postID"`
		}
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Query the post by its ID
		var post models.Post
		if err := postCollection.FindOne(ctx, bson.M{"_id": request.PostID}).Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Post not found"})
			return
		}

		switch post.Kind {
		case "vote":
			var voteRequest models.AVote

			// Re-bind the request body again for voteRequest
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			if err := c.BindJSON(&voteRequest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			voteRequest.ID = primitive.NewObjectID()

			// Insert the vote request into transactions
			_, err = transactionCollection.InsertOne(ctx, voteRequest)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		case "form":
			var formRequest models.AForm

			// Re-bind the request body again for formRequest
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

			if err := c.BindJSON(&formRequest); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			formRequest.ID = primitive.NewObjectID()

			// Insert the form request into transactions
			_, err = transactionCollection.InsertOne(ctx, formRequest)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown post kind"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "data": "answer submitted"})
	}
}

func GetUserAnswer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources
		var request struct {
			StudentID string             `json:"studentID"`
			PostID    primitive.ObjectID `json:"postID"`
		}

		request.StudentID = c.Param("studentID")
		var postID = c.Param("postID")
		println(postID)
		println(request.StudentID)
		var err error
		request.PostID, err = primitive.ObjectIDFromHex(postID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid postID format"})
			return
		}

		// Query the post by its ID
		log.Print(request.PostID)
		var post models.Post
		if err := postCollection.FindOne(ctx, bson.M{"_id": request.PostID}).Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Post not found"})
			return
		}
		log.Print(post.Kind)
		switch post.Kind {
		case "vote":
			var vote models.AVote
			if err := transactionCollection.FindOne(ctx, bson.M{"postID": request.PostID, "studentID": request.StudentID}).Decode(&vote); err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
					return
				}
				log.Printf("Error finding vote: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Vote not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": vote})

		case "form":
			var form models.AForm
			query := bson.M{"postID": request.PostID, "studentID": request.StudentID}
			if err := transactionCollection.FindOne(ctx, query).Decode(&form); err != nil {
				if err == mongo.ErrNoDocuments {
					c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
					return
				}
				log.Printf("Error finding form: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Form not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true, "data": form})
		}
	}
}

// get Answer option from passing postID
func GetAnswerOptionInVote(postID primitive.ObjectID) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var post models.Post
	if err := postCollection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post); err != nil {
		return nil, err
	}

	if post.Kind != "vote" {
		return nil, nil
	}

	var votes []models.AVote
	query := bson.M{"postID": postID}
	cursor, err := transactionCollection.Find(ctx, query)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &votes); err != nil {
		return nil, err
	}

	optionSet := make(map[string]struct{})
	for _, vote := range votes {
		optionSet[string(vote.Answer)] = struct{}{}
	}

	var options []string
	for option := range optionSet {
		options = append(options, option)
	}
	// sort the options
	sort.Strings(options)
	// log.Print(options)

	return options, nil
}

func GetSummaryAnswer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources

		postIDParam := c.Param("postID")
		postID, err := primitive.ObjectIDFromHex(postIDParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Post ID"})
		}

		var post models.Post
		if err := postCollection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		switch post.Kind {
		case "vote":
			options, err := GetAnswerOptionInVote(postID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			aggregateQuery := mongo.Pipeline{
				{{"$match", bson.D{{Key: "postID", Value: postID}}}},
				{{"$group", bson.D{
					{Key: "_id", Value: "$answer"},
					{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
				}}},
			}
			cursor, err := transactionCollection.Aggregate(ctx, aggregateQuery)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			defer cursor.Close(ctx)

			type Result struct {
				Answer string `bson:"_id" json:"answer"`
				Count  int    `bson:"count"`
			}

			var results []Result

			if err := cursor.All(ctx, &results); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			optionCountMap := make(map[string]int)
			for _, option := range options {
				optionCountMap[option] = 0
			}

			totalVotes := 0
			for _, result := range results {
				if _, exists := optionCountMap[result.Answer]; exists {
					optionCountMap[result.Answer] = result.Count
					totalVotes += result.Count
				}
			}

			mappedResults := make([]struct {
				Option string `json:"option"`
				Count  int    `json:"count"`
			}, 0, len(options))

			for _, option := range options {
				mappedResults = append(mappedResults, struct {
					Option string `json:"option"`
					Count  int    `json:"count"`
				}{
					Option: option,
					Count:  optionCountMap[option],
				})
			}

			response := struct {
				TotalVotes int `json:"totalVotes"`
				Results    []struct {
					Option string `json:"option"`
					Count  int    `json:"count"`
				} `json:"results"`
			}{
				TotalVotes: totalVotes,
				Results:    mappedResults,
			}

			c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
		case "form":
			var answers []models.AForm
			query := bson.M{"postID": postID}
			cursor, err := transactionCollection.Find(ctx, query)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			defer cursor.Close(ctx)

			if err := cursor.All(ctx, &answers); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// Transform the data to the desired structure
			resultMap := make(map[int]map[string][]map[string]interface{})
			for _, answer := range answers {
				for _, question := range answer.AnswerList {
					if _, ok := resultMap[question.QuestionIndex]; !ok {
						resultMap[question.QuestionIndex] = make(map[string][]map[string]interface{})
					}
					if _, ok := resultMap[question.QuestionIndex][question.InputType]; !ok {
						resultMap[question.QuestionIndex][question.InputType] = []map[string]interface{}{}
					}
					resultMap[question.QuestionIndex][question.InputType] = append(resultMap[question.QuestionIndex][question.InputType], map[string]interface{}{
						"studentID": answer.StudentID,
						"answer":    question.Answers,
					})
				}
			}

			// Convert the resultMap to the desired JSON structure
			var results []map[string]interface{}
			for questionIndex, typeMap := range resultMap {
				for inputType, studentAnswers := range typeMap {
					results = append(results, map[string]interface{}{
						"questionIndex": questionIndex,
						"type":          inputType,
						"answers":       studentAnswers,
					})
				}
			}

			sort.Slice(results, func(i, j int) bool {
				return results[i]["questionIndex"].(int) < results[j]["questionIndex"].(int)
			})

			var question []QuestionForm
			for i, fq := range post.FormQuestions {
				question = append(question, QuestionForm{
					QuestionIndex: i,
					Question:      fq.Question,
				})
			}

			response := map[string]interface{}{
				"postID":       postID,
				"formQuestion": question,
				"results":      results,
			}

			c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown post kind"})
			return

		}
	}
}

func DeleteAllAnswers(postID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	log.Println(postID)

	filter := bson.M{"postID": postID}

	_, err := transactionCollection.DeleteMany(ctx, filter)
	if err != nil {
		log.Println("Error deleting transaction:", err)
		return err
	}
	return nil
}
