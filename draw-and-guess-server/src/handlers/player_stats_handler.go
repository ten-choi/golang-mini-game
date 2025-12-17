package handlers

import (
	"context"
	"draw-and-guess-server/src/database"
	"draw-and-guess-server/src/models"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPlayerStats(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		log.Printf("GetPlayerStats: missing user ID")
		JSONBadRequest(c, "user ID is required")
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("GetPlayerStats: invalid user ID format: %s, error: %v", userID, err)
		JSONBadRequest(c, "invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	collection := database.GetCollection("users")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("GetPlayerStats: user not found: %s", userID)
			JSONNotFound(c, "user not found")
			return
		}
		log.Printf("GetPlayerStats: database error for user %s: %v", userID, err)
		JSONInternalError(c, "failed to retrieve user information")
		return
	}

	log.Printf("GetPlayerStats: successfully retrieved stats for user: %s", userID)
	JSONSuccess(c, user)
}

func UpsertPlayerStats(c *gin.Context) {
	var request struct {
		Nickname               string `json:"nickname"`
		CorrectGuessesThisGame int    `json:"correctGuessesThisGame"`
		DrawSuccessesThisGame  int    `json:"drawSuccessesThisGame"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("UpsertPlayerStats: invalid request body: %v", err)
		JSONBadRequest(c, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if request.Nickname == "" {
		log.Printf("UpsertPlayerStats: missing nickname")
		JSONBadRequest(c, "nickname is required")
		return
	}

	// Validate stats values
	if request.CorrectGuessesThisGame < 0 || request.DrawSuccessesThisGame < 0 {
		log.Printf("UpsertPlayerStats: invalid stats values for %s: correctGuesses=%d, drawSuccesses=%d",
			request.Nickname, request.CorrectGuessesThisGame, request.DrawSuccessesThisGame)
		JSONBadRequest(c, "stats values must be non-negative")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.GetCollection("player_stats")

	// Check if player stats exist
	var existingStats models.PlayerStats
	err := collection.FindOne(ctx, bson.M{"nickname": request.Nickname}).Decode(&existingStats)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("UpsertPlayerStats: database error when checking existing stats for %s: %v", request.Nickname, err)
			JSONInternalError(c, "failed to check existing player stats")
			return
		}

		// Player stats don't exist, create new
		newStats := models.PlayerStats{
			Nickname:       request.Nickname,
			CorrectGuesses: request.CorrectGuessesThisGame,
			DrawSuccesses:  request.DrawSuccessesThisGame,
			TotalGames:     1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		result, err := collection.InsertOne(ctx, newStats)
		if err != nil {
			log.Printf("UpsertPlayerStats: failed to insert stats for %s: %v", request.Nickname, err)
			JSONInternalError(c, "failed to create player stats")
			return
		}

		log.Printf("UpsertPlayerStats: created new stats for %s (ID: %v)", request.Nickname, result.InsertedID)

		JSONSuccess(c, map[string]interface{}{
			"id":             result.InsertedID,
			"nickname":       request.Nickname,
			"correctGuesses": newStats.CorrectGuesses,
			"drawSuccesses":  newStats.DrawSuccesses,
			"totalGames":     newStats.TotalGames,
		})
		return
	}

	// Player stats exist, update them
	newCorrectGuesses := existingStats.CorrectGuesses + request.CorrectGuessesThisGame
	newDrawSuccesses := existingStats.DrawSuccesses + request.DrawSuccessesThisGame
	newTotalGames := existingStats.TotalGames + 1

	update := bson.M{
		"$set": bson.M{
			"correctGuesses": newCorrectGuesses,
			"drawSuccesses":  newDrawSuccesses,
			"totalGames":     newTotalGames,
			"updated_at":     time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, bson.M{"nickname": request.Nickname}, update)
	if err != nil {
		log.Printf("UpsertPlayerStats: failed to update stats for %s: %v", request.Nickname, err)
		JSONInternalError(c, "failed to update player stats")
		return
	}

	if result.ModifiedCount == 0 {
		log.Printf("UpsertPlayerStats: warning - no documents modified for %s", request.Nickname)
	} else {
		log.Printf("UpsertPlayerStats: updated stats for %s (correctGuesses: %d, drawSuccesses: %d, totalGames: %d)",
			request.Nickname, newCorrectGuesses, newDrawSuccesses, newTotalGames)
	}

	JSONSuccess(c, map[string]interface{}{
		"nickname":       request.Nickname,
		"correctGuesses": newCorrectGuesses,
		"drawSuccesses":  newDrawSuccesses,
		"totalGames":     newTotalGames,
		"modified":       result.ModifiedCount > 0,
	})
}
