package handlers

import (
	"context"
	"draw-and-guess-server/database"
	"draw-and-guess-server/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		respondError(c, http.StatusBadRequest, "user ID is required")
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	collection := database.GetCollection("users")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		log.Printf("User %s not found: %v", userID, err)
		respondError(c, http.StatusNotFound, "user not found")
		return
	}

	respondSuccess(c, "User retrieved successfully", map[string]interface{}{
		"user": user,
	})
}

func CreateUser(c *gin.Context) {
	nickname := c.PostForm("nickname")
	password := c.PostForm("password")
	birthDate := c.PostForm("birth_date")

	if nicknameLen := len(nickname); nicknameLen < 3 || nicknameLen > 20 {
		respondError(c, http.StatusBadRequest, "nickname must be between 3 and 20 characters")
		return
	}

	if password == "" {
		respondError(c, http.StatusBadRequest, "password is missing")
		return
	}

	parsedBirthDate, err := time.Parse("2006-01-02", birthDate)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid birth_date format, expected YYYY-MM-DD")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.GetCollection("users")

	// Check if nickname already exists
	count, err := collection.CountDocuments(ctx, bson.M{"nickname": nickname})
	if err != nil {
		log.Printf("Failed to check nickname existence: %v", err)
		respondError(c, http.StatusInternalServerError, "database error")
		return
	}
	if count > 0 {
		respondError(c, http.StatusConflict, "nickname already exists")
		return
	}

	user := models.User{
		Nickname:     nickname,
		PassWrod:     password,
		BirthDate:    parsedBirthDate,
		WinningPoint: 0,
		ProfileImage: "",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		log.Printf("Failed to create user %s: %v", nickname, err)
		respondError(c, http.StatusInternalServerError, "failed to create user")
		return
	}

	log.Printf("Created user: %s", nickname)

	respondSuccess(c, "User created successfully", map[string]interface{}{
		"id":       result.InsertedID,
		"nickname": nickname,
	})
}

func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		respondError(c, http.StatusBadRequest, "user ID is required")
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		respondError(c, http.StatusBadRequest, "invalid user ID format")
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
			respondError(c, http.StatusBadRequest, "invalid birth_date format, expected YYYY-MM-DD")
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
		respondError(c, http.StatusInternalServerError, "failed to update user")
		return
	}

	if result.MatchedCount == 0 {
		respondError(c, http.StatusNotFound, "user not found")
		return
	}

	log.Printf("Updated user: %s", userID)

	respondSuccess(c, "User updated successfully", map[string]interface{}{
		"modified_count": result.ModifiedCount,
	})
}
