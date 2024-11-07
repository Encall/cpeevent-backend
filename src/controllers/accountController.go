package controllers

import (
	"context"
	_ "fmt"
	"log"
	"net/http"
	"time"

	models "github.com/encall/cpeevent-backend/src/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// var userCollection *mongo.Collection = database.OpenCollection(database.Client, "users")
// var validate = validator.New()

type UpdateAccountInfo struct { //You have to name the struct field according to the JSON attribute
    FirstName   string `json:"firstName"`
    LastName    string `json:"lastName"`
    Year        int    `json:"year"`
    Email       string `json:"email"`
    PhoneNumber string `json:"phoneNumber"`
}

func GetInfo() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID, exists := c.Get("studentid")

		// fmt.Print(userID)
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "User ID not found in context"})
			return
		}
		var info models.User

		// err := userCollection.FindOne(ctx, bson.M{"studentid": userID}).Decode(&foundUser)

		err := userCollection.FindOne(ctx, bson.M{"studentid": userID}).Decode(&info)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, gin.H{"sucess": true, "data": info})

	}
}

func UpdateInfo() gin.HandlerFunc{
	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userID, exists := c.Get("studentid")
		var info UpdateAccountInfo
		defer cancel()
		

		if err := c.BindJSON(&info); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		log.Println("Incoming JSON payload for UpdateInfo:", info)
		err := userCollection.FindOne(ctx, bson.M{"studentid": userID}).Decode(&info)

		//Unfinished
		_ = err
		_ = exists

	}
}