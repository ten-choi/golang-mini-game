package handlers

import (
	"context"
	"draw-and-guess-server/database"
	"draw-and-guess-server/models"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		JSONBadRequest(c, "user ID is required")
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		JSONBadRequest(c, "invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	collection := database.GetCollection("users")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		log.Printf("User %s not found: %v", userID, err)
		JSONNotFound(c, "user not found")
		return
	}

	JSONSuccess(c, user)
}

func CreateUser(c *gin.Context) {
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")
	birthDate := c.PostForm("birth_date")

	if nicknameLen := len(nickname); nicknameLen < 3 || nicknameLen > 20 {
		JSONBadRequest(c, "nickname must be between 3 and 20 characters")
		return
	}

	if password == "" {
		JSONBadRequest(c, "password is missing")
		return
	}

	parsedBirthDate, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		JSONBadRequest(c, "invalid birth_date format, expected YYYY-MM-DD")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.GetCollection("users")

	// Check if nickname already exists
	count, err := collection.CountDocuments(ctx, bson.M{"nickname": nickname})
	if err != nil {
		log.Printf("Failed to check nickname existence: %v", err)
		JSONInternalError(c, "database error")
		return
	}
	if count > 0 {
		JSONConflict(c, "nickname already exists")
		return
	}

	user := models.User{
		Nickname:     nickname,
		Password:     password,
		BirthDate:    parsedBirthDate,
		WinningPoint: 0,
		ProfileImage: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to create user %s: %v", nickname, err)
		JSONInternalError(c, "failed to create user")
		return
	}

	log.Printf("Created user: %s", nickname)

	JSONSuccess(c, map[string]interface{}{
		"id":       result.InsertedID,
		"nickname": nickname,
	})
}

func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		JSONBadRequest(c, "user ID is required")
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		JSONBadRequest(c, "invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.GetCollection("users")

	// Build update document
	update := bson.M{
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	// Update fields if provided
	if password := c.PostForm("password"); password != "" {
		update["$set"].(bson.M)["password"] = password
	}
	if birthDate := c.PostForm("birth_date"); birthDate != "" {
		parsedBirthDate, err := time.Parse("2006-01-02", birthDate)
		if err != nil {
			JSONBadRequest(c, "invalid birth_date format, expected YYYY-MM-DD")
			return
		}
		update["$set"].(bson.M)["birth_date"] = parsedBirthDate
	}
	if profileImage := c.PostForm("profile_image"); profileImage != "" {
		update["$set"].(bson.M)["profile_image"] = profileImage
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": objectID}, update)
	if err != nil {
		log.Printf("Failed to update user %s: %v", userID, err)
		JSONInternalError(c, "failed to update user")
		return
	}

	if result.MatchedCount == 0 {
		JSONNotFound(c, "user not found")
		return
	}

	log.Printf("Updated user: %s", userID)

	JSONSuccess(c, map[string]interface{}{
		"id":             userID,
		"modified_count": result.ModifiedCount,
	})
}
