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

// GetPlayerStats는 플레이어의 통계를 조회하는 API 핸들러
// GET /player-stats/:id
// 용도: 사용자 ID로 플레이어의 게임 통계 정보를 조회
func GetPlayerStats(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		log.Printf("GetPlayerStats: missing user ID")
		JSONBadRequest(c, "user ID is required")
		return
	}

	// 문자열 ID를 ObjectID로 변환
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Printf("GetPlayerStats: invalid user ID format: %s, error: %v", userID, err)
		JSONBadRequest(c, "invalid user ID format")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// MongoDB에서 사용자 정보 조회
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

// UpsertPlayerStats는 플레이어의 게임 통계를 생성하거나 업데이트하는 API 핸들러
// POST /player-stats
// JSON 데이터: {nickname, correctGuessesThisGame, drawSuccessesThisGame}
// 용도: 게임 종료 후 플레이어의 누적 통계를 갱신 (기존 통계가 있으면 업데이트, 없으면 생성)
func UpsertPlayerStats(c *gin.Context) {
	var request struct {
		Nickname               string `json:"nickname"`
		CorrectGuessesThisGame int    `json:"correctGuessesThisGame"`
		DrawSuccessesThisGame  int    `json:"drawSuccessesThisGame"`
	}

	// JSON 요청 바디 파싱
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Printf("UpsertPlayerStats: invalid request body: %v", err)
		JSONBadRequest(c, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// 닉네임 필수 확인
	if request.Nickname == "" {
		log.Printf("UpsertPlayerStats: missing nickname")
		JSONBadRequest(c, "nickname is required")
		return
	}

	// 통계 값 유효성 검사 (음수 불가)
	if request.CorrectGuessesThisGame < 0 || request.DrawSuccessesThisGame < 0 {
		log.Printf("UpsertPlayerStats: invalid stats values for %s: correctGuesses=%d, drawSuccesses=%d",
			request.Nickname, request.CorrectGuessesThisGame, request.DrawSuccessesThisGame)
		JSONBadRequest(c, "stats values must be non-negative")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := database.GetCollection("player_stats")

	// 기존 플레이어 통계 확인
	var existingStats models.PlayerStats
	err := collection.FindOne(ctx, bson.M{"nickname": request.Nickname}).Decode(&existingStats)

	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Printf("UpsertPlayerStats: database error when checking existing stats for %s: %v", request.Nickname, err)
			JSONInternalError(c, "failed to check existing player stats")
			return
		}

		// 플레이어 통계가 없으면 새로 생성
		newStats := models.PlayerStats{
			Nickname:       request.Nickname,
			CorrectGuesses: request.CorrectGuessesThisGame,
			DrawSuccesses:  request.DrawSuccessesThisGame,
			TotalGames:     1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		// MongoDB에 새 통계 삽입
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

	// 플레이어 통계가 이미 존재하면 업데이트
	newCorrectGuesses := existingStats.CorrectGuesses + request.CorrectGuessesThisGame
	newDrawSuccesses := existingStats.DrawSuccesses + request.DrawSuccessesThisGame
	newTotalGames := existingStats.TotalGames + 1

	// 업데이트 문서 작성
	update := bson.M{
		"$set": bson.M{
			"correctGuesses": newCorrectGuesses,
			"drawSuccesses":  newDrawSuccesses,
			"totalGames":     newTotalGames,
			"updated_at":     time.Now(),
		},
	}

	// MongoDB 업데이트 실행
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
