package graph

import (
	"context"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/service"
	"draw-and-guess-server/internal/valkey"
	"draw-and-guess-server/pkg/dictionary"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// ========================================
// Quiz Queries
// ========================================

// RandomOXQuiz is the resolver for the randomOXQuiz field.
func (r *queryResolver) RandomOXQuiz(ctx context.Context, roomID *string) (*model.OXQuiz, error) {
	var excludedIds []string

	// If roomId is provided, get the list of already used quiz IDs
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists && room.UsedQuizIds != nil {
			excludedIds = room.UsedQuizIds
		}
	}

	quiz, err := r.QuizService.GetRandomOXQuiz(ctx, excludedIds)
	if err != nil {
		return nil, err
	}

	// If roomId is provided, add this quiz ID to the used list
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists {
			quizID := fmt.Sprintf("%d", quiz.ID)
			room.UsedQuizIds = append(room.UsedQuizIds, quizID)
			SetGameRoom(*roomID, room)
		}
	}

	return &model.OXQuiz{
		ID:          fmt.Sprintf("%d", quiz.ID),
		Category:    quiz.Category,
		Difficulty:  quiz.Difficulty,
		Question:    quiz.Question,
		Answer:      quiz.Answer,
		Explanation: &quiz.Explanation,
		UsageCount:  int32(quiz.UsageCount),
		IsActive:    quiz.IsActive,
		CreatedAt:   quiz.CreatedAt,
		UpdatedAt:   quiz.UpdatedAt,
	}, nil
}

// RandomQAQuiz is the resolver for the randomQAQuiz field.
func (r *queryResolver) RandomQAQuiz(ctx context.Context, roomID *string) (*model.GeneralQuiz, error) {
	var excludedIds []string

	// If roomId is provided, get the list of already used quiz IDs
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists && room.UsedQuizIds != nil {
			excludedIds = room.UsedQuizIds
		}
	}

	quiz, err := r.QuizService.GetRandomQAQuiz(ctx, excludedIds)
	if err != nil {
		return nil, err
	}

	// If roomId is provided, add this quiz ID to the used list
	if roomID != nil {
		room, exists := GetGameRoom(*roomID)
		if exists {
			quizID := fmt.Sprintf("%d", quiz.ID)
			room.UsedQuizIds = append(room.UsedQuizIds, quizID)
			SetGameRoom(*roomID, room)
		}
	}

	return &model.GeneralQuiz{
		ID:          fmt.Sprintf("%d", quiz.ID),
		Category:    quiz.Category,
		Difficulty:  quiz.Difficulty,
		Question:    quiz.Question,
		Options:     quiz.Options,
		Answer:      int32(quiz.Answer),
		Explanation: &quiz.Explanation,
		ImageURL:    &quiz.ImageURL,
		UsageCount:  int32(quiz.UsageCount),
		IsActive:    quiz.IsActive,
		CreatedAt:   quiz.CreatedAt,
		UpdatedAt:   quiz.UpdatedAt,
	}, nil
}

// IsValidWord is the resolver for the isValidWord field.
func (r *queryResolver) IsValidWord(ctx context.Context, word string) (bool, error) {
	dict := dictionary.GetInstance()
	return dict.IsValidWord(word), nil
}

// startQuizGame starts a quiz game and sends the first quiz
func startQuizGame(ctx context.Context, roomID string, gameType model.GameType, quizService service.QuizService) {
	log.Printf("[Quiz] ==================== START QUIZ GAME ====================")
	log.Printf("[Quiz] Room: %s, GameType: %s", roomID, gameType)

	// Get room to check if it's still in PLAYING status
	room, exists := GetGameRoom(roomID)
	if !exists {
		log.Printf("[Quiz] ERROR: Room %s not found, aborting quiz game", roomID)
		return
	}
	if room.Status != model.GameStatusPlaying {
		log.Printf("[Quiz] ERROR: Room %s status is %s (not PLAYING), aborting quiz game", roomID, room.Status)
		return
	}

	log.Printf("[Quiz] Room found and playing, sending first quiz...")
	// Send first quiz
	sendNextQuiz(ctx, roomID, gameType, quizService)
	
	log.Printf("[Quiz] Starting timer for room %s with %d seconds", roomID, room.RoundTimeLimit)
	// Start timer for the quiz
	go startQuizTimer(ctx, roomID, gameType, room.RoundTimeLimit, quizService)
	log.Printf("[Quiz] ==================== QUIZ GAME STARTED ====================")
}

// startQuizTimer manages the countdown timer for a quiz
func startQuizTimer(ctx context.Context, roomID string, gameType model.GameType, timeLimit int32, quizService service.QuizService) {
	log.Printf("[Quiz] Starting timer for room %s: %d seconds", roomID, timeLimit)
	
	channelName := fmt.Sprintf("game/%s", roomID)
	
	// Send initial time
	sendTimerUpdate(channelName, int(timeLimit))
	
	// Countdown
	for i := int(timeLimit) - 1; i >= 0; i-- {
		time.Sleep(1 * time.Second)
		
		// Check if room still exists and is playing
		room, exists := GetGameRoom(roomID)
		if !exists || room.Status != model.GameStatusPlaying {
			log.Printf("[Quiz] Timer stopped for room %s (room ended)", roomID)
			return
		}
		
		sendTimerUpdate(channelName, i)
	}
	
	log.Printf("[Quiz] Timer finished for room %s", roomID)
	
	// Move to next round or end game
	room, exists := GetGameRoom(roomID)
	if !exists || room.Status != model.GameStatusPlaying {
		return
	}
	
	// Increment round
	if room.CurrentRound >= room.TotalRounds {
		// Game finished
		log.Printf("[Quiz] Game finished for room %s", roomID)
		room.Status = model.GameStatusFinished
		SetGameRoom(roomID, room)
		publishRoomUpdateToWebSocket(roomID, room)
	} else {
		// Next round
		log.Printf("[Quiz] Moving to next round for room %s (round %d -> %d)", roomID, room.CurrentRound, room.CurrentRound+1)
		room.CurrentRound++
		roundTimeLimit := room.RoundTimeLimit // Save before updating room
		SetGameRoom(roomID, room)
		publishRoomUpdateToWebSocket(roomID, room)
		
		// Send next quiz after a short delay
		time.Sleep(2 * time.Second)
		log.Printf("[Quiz] Sending next quiz for room %s, round %d", roomID, room.CurrentRound)
		sendNextQuiz(ctx, roomID, gameType, quizService)
		go startQuizTimer(ctx, roomID, gameType, roundTimeLimit, quizService)
	}
}

// sendTimerUpdate broadcasts timer update to WebSocket clients
func sendTimerUpdate(channelName string, timeLeft int) {
	message := map[string]interface{}{
		"type": "timer",
		"data": map[string]interface{}{
			"timeLeft": timeLeft,
		},
	}
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Quiz] Failed to marshal timer message: %v", err)
		return
	}
	
	err = valkey.Client.Publish(context.Background(), channelName, jsonData).Err()
	if err != nil {
		log.Printf("[Quiz] Failed to publish timer to Valkey: %v", err)
	}
}

// sendNextQuiz fetches and broadcasts the next quiz question
func sendNextQuiz(ctx context.Context, roomID string, gameType model.GameType, quizService service.QuizService) {
	log.Printf("[Quiz] ########## SEND NEXT QUIZ ##########")
	log.Printf("[Quiz] Room: %s, GameType: %s", roomID, gameType)

	var quizData map[string]interface{}

	// Get excluded quiz IDs from room
	var excludedIds []string
	room, exists := GetGameRoom(roomID)
	if exists && room.UsedQuizIds != nil {
		excludedIds = room.UsedQuizIds
		log.Printf("[Quiz] Excluded quiz IDs: %v", excludedIds)
	}

	if gameType == model.GameTypeOx {
		log.Printf("[Quiz] Fetching OX quiz...")
		// Get random OX quiz
		quiz, err := quizService.GetRandomOXQuiz(ctx, excludedIds)
		if err != nil {
			log.Printf("[Quiz] ERROR: Failed to get OX quiz: %v", err)
			return
		}

		// Add quiz ID to used list
		if exists {
			quizID := fmt.Sprintf("%d", quiz.ID)
			room.UsedQuizIds = append(room.UsedQuizIds, quizID)
			SetGameRoom(roomID, room)
		}

		quizData = map[string]interface{}{
			"id":          fmt.Sprintf("%d", quiz.ID),
			"type":        "OX",
			"category":    quiz.Category,
			"difficulty":  quiz.Difficulty,
			"question":    quiz.Question,
			"answer":      quiz.Answer,
			"explanation": quiz.Explanation,
		}
	} else if gameType == model.GameTypeQa {
		log.Printf("[Quiz] Fetching QA quiz...")
		// Get random QA quiz
		quiz, err := quizService.GetRandomQAQuiz(ctx, excludedIds)
		if err != nil {
			log.Printf("[Quiz] ERROR: Failed to get QA quiz: %v", err)
			return
		}

		log.Printf("[Quiz] SUCCESS: Got QA quiz ID=%d, question=%s", quiz.ID, quiz.Question)
		log.Printf("[Quiz] Options count=%d, Options: %v", len(quiz.Options), quiz.Options)

		// Add quiz ID to used list
		if exists {
			quizID := fmt.Sprintf("%d", quiz.ID)
			room.UsedQuizIds = append(room.UsedQuizIds, quizID)
			SetGameRoom(roomID, room)
		}

		quizData = map[string]interface{}{
			"id":          fmt.Sprintf("%d", quiz.ID),
			"type":        "QA",
			"category":    quiz.Category,
			"difficulty":  quiz.Difficulty,
			"question":    quiz.Question,
			"options":     quiz.Options,
			"answer":      quiz.Answer,
			"explanation": quiz.Explanation,
			"imageUrl":    quiz.ImageURL,
		}
		
		log.Printf("[Quiz] Created quiz data with %d options", len(quiz.Options))
	}

	// Broadcast quiz to all players via WebSocket
	message := map[string]interface{}{
		"type": "quiz",
		"data": quizData,
	}

	log.Printf("[Quiz] Broadcasting quiz message: type=%s, has options=%v", quizData["type"], quizData["options"] != nil)

	// Publish to Valkey for WebSocket distribution
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Quiz] Failed to marshal quiz message: %v", err)
		return
	}

	log.Printf("[Quiz] Marshaled quiz JSON length: %d bytes", len(jsonData))

	channelName := fmt.Sprintf("game/%s", roomID)
	err = valkey.Client.Publish(context.Background(), channelName, jsonData).Err()
	if err != nil {
		log.Printf("[Quiz] Failed to publish quiz to Valkey: %v", err)
		return
	}

	log.Printf("[Quiz] Successfully published quiz to channel: %s", channelName)
}
