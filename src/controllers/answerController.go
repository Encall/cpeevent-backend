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
		if err := c.BindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
				log.Printf("Error finding vote: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Vote not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true, "data": vote})

		case "form":
			var form models.AForm
			query := bson.M{"postID": request.PostID, "studentID": request.StudentID}
			if err := transactionCollection.FindOne(ctx, query).Decode(&form); err != nil {
				log.Printf("Error finding form: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Form not found"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"success": true, "data": form})
		}
	}
}

func GetSummaryAnswer() gin.HandlerFunc{
    return func(c *gin.Context){
        var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel() // Ensure cancel is called to release resources

        postIDParam := c.Param("postID")
        postID, err := primitive.ObjectIDFromHex(postIDParam)
        if err != nil{
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Post ID"})
        }

        var post models.Post
        if err := postCollection.FindOne(ctx, bson.M{"_id": postID}).Decode(&post); err != nil{
            c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
            return
        }


        switch post.Kind{
        case "vote":
            aggregateQuery := mongo.Pipeline{
                {{"$match", bson.D{{"postID", postID}}}},
                {{"$group", bson.D{
                    {"_id", "$answer"},
                    {"count", bson.D{{"$sum", 1}}},
                }}},
            }
            cursor, err := transactionCollection.Aggregate(ctx, aggregateQuery)

            if err != nil{
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            }
            defer cursor.Close(ctx)

            type Result struct{
                Answer string `bson:"_id" json:"answer"`
                Count int`bson:"count" json:"count"`
            }

            var results []Result

            if err:= cursor.All(ctx, &results); err != nil{
                c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
                return
            }

            totalVotes := 0
            for _, result := range results{
                totalVotes += result.Count 
            }

            response := struct{
                TotalVotes int `json:"totalVotes"`
                Results [] Result `json:"results"`
            }{
                TotalVotes: totalVotes,
                Results: results,
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

            response := map[string]interface{}{
                "postID":  postID,
                "results": results,
            }

            c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
        default:
            c.JSON(http.StatusBadRequest, gin.H{"error": "Unknown post kind"})
            return

        }

       

        

        

        // if err := cursor.All(ctx, &answer); err != nil{
        //     c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
        //     return
        // }

        // c.JSON(http.StatusOK, answer)

        


    }
}