package graph

import (
	"context"
	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/service"
	"draw-and-guess-server/internal/valkey"
	"draw-and-guess-server/pkg/dictionary"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// ========================================
// Quiz Queries
// ========================================

// IsValidWord is the resolver for the isValidWord field.
func (r *queryResolver) IsValidWord(ctx context.Context, word string) (bool, error) {
	dict := dictionary.GetInstance()
	return dict.IsValidWord(word), nil
}

// QuizInfo stores quiz answer and difficulty for verification
type QuizInfo struct {
	Answer     interface{}
	Difficulty int // 1-5
}

// Pre-loaded quizzes storage (in-memory) - stores quiz answers for verification
var quizAnswers = make(map[string]map[string]*QuizInfo) // roomID -> quizID -> QuizInfo
var quizAnswersMutex sync.RWMutex

// Pending scores that will be applied when round ends
var pendingScores = make(map[string]map[string]int32) // roomID -> username -> score to add
var pendingScoresMutex sync.RWMutex

// storeQuizAnswer stores a quiz answer and difficulty for later verification
func storeQuizAnswer(roomID string, quizID string, answer interface{}, difficulty int) {
	quizAnswersMutex.Lock()
	defer quizAnswersMutex.Unlock()

	if quizAnswers[roomID] == nil {
		quizAnswers[roomID] = make(map[string]*QuizInfo)
	}
	quizAnswers[roomID][quizID] = &QuizInfo{
		Answer:     answer,
		Difficulty: difficulty,
	}
	log.Printf("[Quiz] Stored answer for room %s, quiz %s, difficulty %d", roomID, quizID, difficulty)
}

// getQuizInfo retrieves stored quiz information (answer and difficulty)
func getQuizInfo(roomID string, quizID string) (*QuizInfo, bool) {
	quizAnswersMutex.RLock()
	defer quizAnswersMutex.RUnlock()

	if roomAnswers, exists := quizAnswers[roomID]; exists {
		info, ok := roomAnswers[quizID]
		return info, ok
	}
	return nil, false
}

// getQuizAnswer retrieves a stored quiz answer (legacy compatibility)
func getQuizAnswer(roomID string, quizID string) (interface{}, bool) {
	info, ok := getQuizInfo(roomID, quizID)
	if !ok {
		return nil, false
	}
	return info.Answer, true
}

// calculateQuizScore calculates score based on difficulty (1-5 -> 50-250 points)
func calculateQuizScore(difficulty int) int32 {
	// Validate level range (1-5)
	if difficulty < 1 {
		difficulty = 1
	} else if difficulty > 5 {
		difficulty = 5
	}

	// Calculate score: difficulty × 50
	score := int32(difficulty * 50)
	log.Printf("[Quiz] Difficulty %d -> %d points", difficulty, score)
	return score
}

// GetQuizInfo (public) retrieves stored quiz information for WebSocket usage
func GetQuizInfo(roomID string, quizID string) (*QuizInfo, bool) {
	return getQuizInfo(roomID, quizID)
}

// CalculateQuizScore (public) calculates score based on difficulty for WebSocket usage
func CalculateQuizScore(difficulty int) int32 {
	return calculateQuizScore(difficulty)
}

// AddScoreToUser adds score to user in game room (thread-safe)
func AddScoreToUser(roomID string, username string, score int32) {
	addScoreToUser(roomID, username, score, true)
}

// AddScoreToUserSilent adds score to pending scores without broadcasting (for quiz answers)
func AddScoreToUserSilent(roomID string, username string, score int32) {
	addScoreToUser(roomID, username, score, false)
}

// addScoreToUser internal function that handles score addition with optional broadcast
func addScoreToUser(roomID string, username string, score int32, broadcast bool) {
	if !broadcast {
		// Store in pending scores (will be applied at ROUND_ENDED)
		pendingScoresMutex.Lock()
		defer pendingScoresMutex.Unlock()

		if pendingScores[roomID] == nil {
			pendingScores[roomID] = make(map[string]int32)
		}
		pendingScores[roomID][username] += score
		log.Printf("[Quiz] Added %d points to pending scores for %s (pending total: %d)", score, username, pendingScores[roomID][username])
		return
	}

	// Immediate score update with broadcast
	roomMutex.Lock()
	defer roomMutex.Unlock()

	room, exists := gameRooms[roomID]
	if !exists {
		log.Printf("[Quiz] Room not found: %s", roomID)
		return
	}

	for _, user := range room.Users {
		if user.Name == username {
			user.Score += score
			log.Printf("[Quiz] Added %d points to %s (total: %d)", score, username, user.Score)
			break
		}
	}

	gameRooms[roomID] = room

	// Broadcast updated room state to all clients
	go publishRoomUpdateToWebSocket(roomID, room)
}

// cleanupQuizAnswers removes stored answers for a room
func cleanupQuizAnswers(roomID string) {
	quizAnswersMutex.Lock()
	defer quizAnswersMutex.Unlock()

	delete(quizAnswers, roomID)
	log.Printf("[Quiz] Cleaned up quiz answers for room %s", roomID)
}

// applyPendingScores applies all pending scores to the room and clears them
func applyPendingScores(roomID string) {
	pendingScoresMutex.Lock()
	defer pendingScoresMutex.Unlock()

	scoresToApply, exists := pendingScores[roomID]
	if !exists || len(scoresToApply) == 0 {
		log.Printf("[Quiz] No pending scores to apply for room %s", roomID)
		return
	}

	// Apply all pending scores atomically
	roomMutex.Lock()
	room, roomExists := gameRooms[roomID]
	if !roomExists {
		roomMutex.Unlock()
		log.Printf("[Quiz] Room not found when applying pending scores: %s", roomID)
		return
	}

	for username, score := range scoresToApply {
		for _, user := range room.Users {
			if user.Name == username {
				user.Score += score
				log.Printf("[Quiz] ✅ Applied pending score: %s +%d points (total: %d)", username, score, user.Score)
				break
			}
		}
	}

	gameRooms[roomID] = room
	roomMutex.Unlock()

	// Clear pending scores for this room
	delete(pendingScores, roomID)
	log.Printf("[Quiz] Cleared pending scores for room %s", roomID)
}

// Pre-loaded quizzes storage (in-memory)
var preloadedQuizzes = make(map[string][]map[string]interface{})
var preloadedQuizzesMutex sync.RWMutex

// startQuizGame starts a quiz game and pre-loads all quizzes
func startQuizGame(ctx context.Context, roomID string, gameType model.GameType, quizService service.QuizService) {
	log.Printf("[Quiz] ==================== START QUIZ GAME ====================")
	log.Printf("[Quiz] Room: %s, GameType: %s", roomID, gameType)

	// Get room to check if it's still in PLAYING status
	room, exists := GetGameRoom(roomID)
	if !exists {
		log.Printf("[Quiz] ERROR: Room %s not found, aborting quiz game", roomID)
		return
	}
	if room.Status != model.GameRoomStatusPlaying {
		log.Printf("[Quiz] ERROR: Room %s status is %s (not PLAYING), aborting quiz game", roomID, room.Status)
		return
	}

	// Pre-load all quizzes for all rounds
	log.Printf("[Quiz] Pre-loading %d quizzes for room %s", room.TotalRounds, roomID)
	quizzes, err := preloadQuizzes(ctx, roomID, gameType, int(room.TotalRounds), quizService)
	if err != nil {
		log.Printf("[Quiz] ERROR: Failed to pre-load quizzes: %v", err)
		return
	}

	log.Printf("[Quiz] Stored %d pre-loaded quizzes in memory for room %s", len(quizzes), roomID)

	log.Printf("[Quiz] Successfully pre-loaded %d quizzes", len(quizzes))
	log.Printf("[Quiz] Room found and playing, sending first quiz...")

	// Send first quiz (round 1)
	sendPreloadedQuiz(roomID, 0) // 0-indexed

	log.Printf("[Quiz] Starting timer for room %s with %d seconds", roomID, room.RoundTimeLimit)
	// Start timer for the quiz
	go startQuizTimer(ctx, roomID, gameType, room.RoundTimeLimit, quizService)
	log.Printf("[Quiz] ==================== QUIZ GAME STARTED ====================")
}

// preloadQuizzes fetches all quizzes needed for the game at once
func preloadQuizzes(ctx context.Context, roomID string, gameType model.GameType, totalRounds int, quizService service.QuizService) ([]map[string]interface{}, error) {
	log.Printf("[Quiz Pre-Loading] Starting to pre-load %d quizzes for room %s (type: %s)", totalRounds, roomID, gameType)
	quizzes := make([]map[string]interface{}, 0, totalRounds)
	var excludedIds []string

	// Get excluded quiz IDs from room
	room, exists := GetGameRoom(roomID)
	if exists && room.UsedQuizIds != nil {
		excludedIds = room.UsedQuizIds
	}

	for i := 0; i < totalRounds; i++ {
		log.Printf("[Quiz Pre-Loading] Fetching quiz %d/%d...", i+1, totalRounds)
		var quizData map[string]interface{}

		if gameType == model.GameTypeOx {
			quiz, err := quizService.GetRandomOXQuiz(ctx, excludedIds)
			if err != nil {
				return nil, fmt.Errorf("failed to get OX quiz for round %d: %w", i+1, err)
			}

			quizID := fmt.Sprintf("%d", quiz.ID)
			excludedIds = append(excludedIds, quizID)

			quizData = map[string]interface{}{
				"id":          quizID,
				"type":        common.GameTypeOX,
				"category":    quiz.Category,
				"difficulty":  quiz.Difficulty,
				"question":    quiz.Question,
				"answer":      quiz.Answer,
				"explanation": quiz.Explanation,
			}
			log.Printf("[Quiz Pre-Loading] 笨・Round %d: OX quiz loaded (ID: %s, Question: %.50s...)", i+1, quizID, quiz.Question)

		} else if gameType == model.GameTypeQa {
			quiz, err := quizService.GetRandomQAQuiz(ctx, excludedIds)
			if err != nil {
				return nil, fmt.Errorf("failed to get QA quiz for round %d: %w", i+1, err)
			}

			quizID := fmt.Sprintf("%d", quiz.ID)
			excludedIds = append(excludedIds, quizID)

			quizData = map[string]interface{}{
				"id":          quizID,
				"type":        common.GameTypeQA,
				"category":    quiz.Category,
				"difficulty":  quiz.Difficulty,
				"question":    quiz.Question,
				"options":     quiz.Options,
				"answer":      quiz.Answer,
				"explanation": quiz.Explanation,
				"imageUrl":    quiz.ImageURL,
			}
			log.Printf("[Quiz Pre-Loading] 笨・Round %d: QA quiz loaded (ID: %s, Options: %d, Question: %.50s...)", i+1, quizID, len(quiz.Options), quiz.Question)
		}

		quizzes = append(quizzes, quizData)
	}

	log.Printf("[Quiz Pre-Loading] 笨・COMPLETED! Successfully pre-loaded %d quizzes for room %s", len(quizzes), roomID)

	// Update room with all used quiz IDs
	if exists {
		room.UsedQuizIds = excludedIds
		SetGameRoom(roomID, room)
	}

	// Store preloaded quizzes with mutex protection
	preloadedQuizzesMutex.Lock()
	preloadedQuizzes[roomID] = quizzes
	preloadedQuizzesMutex.Unlock()

	return quizzes, nil
}

// sendPreloadedQuiz sends a pre-loaded quiz to clients
func sendPreloadedQuiz(roomID string, roundIndex int) {
	log.Printf("[Quiz Delivery] 豆 Sending pre-loaded quiz for room %s, round %d (index: %d)", roomID, roundIndex+1, roundIndex)

	preloadedQuizzesMutex.RLock()
	quizzes, exists := preloadedQuizzes[roomID]
	preloadedQuizzesMutex.RUnlock()

	if !exists {
		log.Printf("[Quiz Delivery] 笶・ERROR: No pre-loaded quizzes found for room %s", roomID)
		return
	}

	if roundIndex >= len(quizzes) {
		log.Printf("[Quiz Delivery] 笶・ERROR: Round index %d out of range (total quizzes: %d) for room %s", roundIndex, len(quizzes), roomID)
		return
	}

	quizData := quizzes[roundIndex]

	log.Printf("[Quiz Delivery] 笨・Quiz found - Type: %s, ID: %s", quizData["type"], quizData["id"])
	log.Printf("[Quiz Delivery] 笨・Question: %.80s...", quizData["question"])

	// Store quiz answer for verification
	quizID := quizData["id"].(string)
	answer := quizData["answer"]
	difficulty := quizData["difficulty"].(int)
	storeQuizAnswer(roomID, quizID, answer, difficulty)
	log.Printf("[Quiz Delivery] 笨・Stored answer for quiz %s (difficulty: %d)", quizID, difficulty)

	// Broadcast quiz to all users via WebSocket (without answer and explanation)
	quizWithoutAnswer := map[string]interface{}{
		"id":         quizData["id"],
		"type":       quizData["type"],
		"category":   quizData["category"],
		"difficulty": quizData["difficulty"],
		"question":   quizData["question"],
	}

	// Add options for QA quiz
	if quizData["type"] == common.GameTypeQA {
		quizWithoutAnswer["options"] = quizData["options"]
		if imageUrl, exists := quizData["imageUrl"]; exists {
			quizWithoutAnswer["imageUrl"] = imageUrl
		}
	}

	message := map[string]interface{}{
		"type": "quiz",
		"data": quizWithoutAnswer,
	}

	log.Printf("[Quiz Delivery] 藤 Broadcasting quiz to all users in room %s", roomID)

	// Publish to Valkey for WebSocket distribution
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("[Quiz] Failed to marshal quiz message: %v", err)
		return
	}

	channelName := fmt.Sprintf("game/%s", roomID)
	err = valkey.Client.Publish(context.Background(), channelName, jsonData).Err()
	if err != nil {
		log.Printf("[Quiz] Failed to publish quiz to Valkey: %v", err)
		return
	}

	log.Printf("[Quiz] Successfully published quiz to channel: %s", channelName)
}

// cleanupPreloadedQuizzes removes pre-loaded quizzes when game ends
func cleanupPreloadedQuizzes(roomID string) {
	preloadedQuizzesMutex.Lock()
	delete(preloadedQuizzes, roomID)
	preloadedQuizzesMutex.Unlock()
	log.Printf("[Quiz] Cleaned up pre-loaded quizzes for room %s", roomID)
}

// startQuizTimer manages the countdown timer for a quiz
func startQuizTimer(ctx context.Context, roomID string, gameType model.GameType, timeLimit int32, quizService service.QuizService) {
	log.Printf("[Quiz] Starting timer for room %s: %d seconds", roomID, timeLimit)

	channelName := fmt.Sprintf("game/%s", roomID)

	// Send initial time
	sendTimerUpdate(channelName, int(timeLimit))

	// Countdown
	for i := int(timeLimit) - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			log.Printf("[Quiz] Timer cancelled for room %s (context done)", roomID)
			return
		case <-time.After(1 * time.Second):
			// Check if room still exists and is playing
			room, exists := GetGameRoom(roomID)
			if !exists || room.Status != model.GameRoomStatusPlaying {
				log.Printf("[Quiz] Timer stopped for room %s (room ended)", roomID)
				return
			}

			sendTimerUpdate(channelName, i)
		}
	}

	log.Printf("[Quiz] Timer finished for room %s", roomID)

	// Apply pending scores before revealing
	applyPendingScores(roomID)

	// Get current quiz info to reveal answer
	preloadedQuizzesMutex.RLock()
	quizzes, quizzesExist := preloadedQuizzes[roomID]
	preloadedQuizzesMutex.RUnlock()

	// Publish ROUND_ENDED with correct answer and scoreboard
	if quizzesExist {
		room, exists := GetGameRoom(roomID)
		if exists && room.CurrentRound > 0 && int(room.CurrentRound) <= len(quizzes) {
			currentQuizIndex := int(room.CurrentRound) - 1
			currentQuiz := quizzes[currentQuizIndex]

			// Prepare scoreboard with all users' current scores
			scoreboard := make([]map[string]interface{}, 0, len(room.Users))
			for _, user := range room.Users {
				scoreboard = append(scoreboard, map[string]interface{}{
					"userId": user.UserID,
					"name":   user.Name,
					"score":  user.Score,
				})
			}

			// Broadcast ROUND_ENDED with correct answer and scoreboard
			channelName := fmt.Sprintf("game/%s", roomID)
			message := map[string]interface{}{
				"type": "ROUND_ENDED",
				"data": map[string]interface{}{
					"round":         room.CurrentRound,
					"quizId":        currentQuiz["id"],
					"correctAnswer": currentQuiz["answer"],
					"explanation":   currentQuiz["explanation"],
					"scoreboard":    scoreboard,
				},
			}

			jsonData, err := json.Marshal(message)
			if err == nil {
				valkey.Client.Publish(context.Background(), channelName, jsonData)
				log.Printf("[Quiz] Published ROUND_ENDED for room %s round %d with answer and scoreboard", roomID, room.CurrentRound)
			}
		}
	}

	// Move to next round or end game
	room, exists := GetGameRoom(roomID)
	if !exists || room.Status != model.GameRoomStatusPlaying {
		cleanupPreloadedQuizzes(roomID)
		return
	}

	// Increment round
	if room.CurrentRound >= room.TotalRounds {
		// Game finished
		log.Printf("[Quiz] Game finished for room %s", roomID)
		room.Status = model.GameRoomStatusFinished
		SetGameRoom(roomID, room)
		publishRoomUpdateToWebSocket(roomID, room)
		cleanupPreloadedQuizzes(roomID)
	} else {
		// Next round
		log.Printf("[Quiz] Moving to next round for room %s (round %d -> %d)", roomID, room.CurrentRound, room.CurrentRound+1)
		room.CurrentRound++
		roundTimeLimit := room.RoundTimeLimit // Save before updating room
		SetGameRoom(roomID, room)
		publishRoomUpdateToWebSocket(roomID, room)

		// Wait before next round to let players see the answer
		time.Sleep(common.RoundDelaySeconds * time.Second)
		log.Printf("[Quiz] Sending next pre-loaded quiz for room %s, round %d", roomID, room.CurrentRound)
		sendPreloadedQuiz(roomID, int(room.CurrentRound)-1) // CurrentRound is 1-indexed
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
			"type":        common.GameTypeOX,
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
			"type":        common.GameTypeQA,
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

	// Broadcast quiz to all users via WebSocket
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
