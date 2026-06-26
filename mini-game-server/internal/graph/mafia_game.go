package graph

import (
	"context"
	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/valkey"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

type MafiaRole string

type MafiaPhase string

const (
	MafiaRoleMafia   MafiaRole = "MAFIA"
	MafiaRoleCitizen MafiaRole = "CITIZEN"
	MafiaRolePolice  MafiaRole = "POLICE"
	MafiaRoleDoctor  MafiaRole = "DOCTOR"
)

const (
	MafiaPhaseDayChat MafiaPhase = "DAY_CHAT"
	MafiaPhaseDayVote MafiaPhase = "DAY_VOTE"
	MafiaPhaseNight   MafiaPhase = "NIGHT"
)

type mafiaState struct {
	RoomID        string
	Phase         MafiaPhase
	Roles         map[string]MafiaRole
	Alive         map[string]bool
	DayVotes      map[string]string
	NightVotes    map[string]string
	DoctorSave    *string
	PoliceCheck   *string
	DayCount      int
	LastDayKilled *string
	LastNightKill *string
	mu            sync.Mutex
}

var (
	mafiaStates   = make(map[string]*mafiaState)
	mafiaStatesMu sync.RWMutex
)

func startMafiaGame(roomID string) {
	room, exists := GetGameRoom(roomID)
	if !exists {
		log.Printf("[Mafia] Room not found: %s", roomID)
		return
	}

	if room.Status != model.GameRoomStatus(common.RoomStatusPlaying) {
		log.Printf("[Mafia] Room %s not in PLAYING state", roomID)
		return
	}

	if len(room.Users) != 6 {
		log.Printf("[Mafia] Room %s requires exactly 6 players, got %d", roomID, len(room.Users))
		return
	}

	state := &mafiaState{
		RoomID:     roomID,
		Phase:      MafiaPhaseDayChat,
		Roles:      map[string]MafiaRole{},
		Alive:      map[string]bool{},
		DayVotes:   map[string]string{},
		NightVotes: map[string]string{},
		DayCount:   1,
	}

	userIDs := make([]string, 0, len(room.Users))
	for _, user := range room.Users {
		userIDs = append(userIDs, user.UserID)
		state.Alive[user.UserID] = true
	}

	roles := []MafiaRole{MafiaRoleMafia, MafiaRoleMafia, MafiaRoleCitizen, MafiaRoleCitizen, MafiaRolePolice, MafiaRoleDoctor}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(roles), func(i, j int) { roles[i], roles[j] = roles[j], roles[i] })

	for i, userID := range userIDs {
		state.Roles[userID] = roles[i]
	}

	setMafiaState(roomID, state)

	mafiaIDs := make([]string, 0, 2)
	for userID, role := range state.Roles {
		if role == MafiaRoleMafia {
			mafiaIDs = append(mafiaIDs, userID)
		}
	}

	for _, userID := range userIDs {
		role := state.Roles[userID]
		teammates := []string{}
		if role == MafiaRoleMafia {
			for _, mafiaID := range mafiaIDs {
				if mafiaID != userID {
					teammates = append(teammates, mafiaID)
				}
			}
		}
		publishMafiaToUser(userID, map[string]interface{}{
			"type":      "MAFIA_ROLE",
			"roomId":    roomID,
			"role":      role,
			"teammates": teammates,
		})
	}

	room.CurrentRound = 1
	SetGameRoom(roomID, room)
	publishRoomUpdateToWebSocket(roomID, room)

	publishMafiaToRoom(roomID, map[string]interface{}{
		"type":   "MAFIA_GAME_STARTED",
		"roomId": roomID,
		"alive":  aliveList(state.Alive),
		"day":    state.DayCount,
	})

	go runMafiaLoop(roomID)
}

func runMafiaLoop(roomID string) {
	for {
		state, ok := getMafiaState(roomID)
		if !ok {
			return
		}

		if !runMafiaPhase(roomID, MafiaPhaseDayChat, 60) {
			return
		}

		if !runMafiaPhase(roomID, MafiaPhaseDayVote, 20) {
			return
		}

		resolveDayVote(roomID)
		if checkMafiaWin(roomID) {
			return
		}

		if !runMafiaPhase(roomID, MafiaPhaseNight, 20) {
			return
		}

		resolveNight(roomID)
		if checkMafiaWin(roomID) {
			return
		}

		state, ok = getMafiaState(roomID)
		if !ok {
			return
		}
		state.mu.Lock()
		state.DayCount++
		state.mu.Unlock()

		room, exists := GetGameRoom(roomID)
		if exists {
			room.CurrentRound = int32(state.DayCount)
			SetGameRoom(roomID, room)
			publishRoomUpdateToWebSocket(roomID, room)
		}
	}
}

func runMafiaPhase(roomID string, phase MafiaPhase, duration int) bool {
	state, ok := getMafiaState(roomID)
	if !ok {
		return false
	}

	state.mu.Lock()
	state.Phase = phase
	if phase == MafiaPhaseDayVote {
		state.DayVotes = map[string]string{}
		state.LastDayKilled = nil
	} else if phase == MafiaPhaseNight {
		state.NightVotes = map[string]string{}
		state.DoctorSave = nil
		state.PoliceCheck = nil
		state.LastNightKill = nil
	}
	dayCount := state.DayCount
	state.mu.Unlock()

	publishMafiaToRoom(roomID, map[string]interface{}{
		"type":     "MAFIA_PHASE",
		"roomId":   roomID,
		"phase":    phase,
		"duration": duration,
		"day":      dayCount,
	})

	for timeLeft := duration; timeLeft > 0; timeLeft-- {
		if !isMafiaPhase(roomID, phase) {
			return false
		}
		publishMafiaToRoom(roomID, map[string]interface{}{
			"type":     "MAFIA_TIMER",
			"roomId":   roomID,
			"phase":    phase,
			"timeLeft": timeLeft,
			"day":      dayCount,
		})
		time.Sleep(1 * time.Second)
	}

	return true
}

func SubmitMafiaDayVote(roomID, userID, targetID string) error {
	state, ok := getMafiaState(roomID)
	if !ok {
		return fmt.Errorf("mafia state not found")
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.Phase != MafiaPhaseDayVote {
		return fmt.Errorf("not in day vote phase")
	}
	if !state.Alive[userID] {
		return fmt.Errorf("voter not alive")
	}
	if !state.Alive[targetID] {
		return fmt.Errorf("target not alive")
	}

	state.DayVotes[userID] = targetID

	publishMafiaToRoom(roomID, map[string]interface{}{
		"type":     "MAFIA_DAY_VOTE_CAST",
		"roomId":   roomID,
		"voterId":  userID,
		"targetId": targetID,
	})

	return nil
}

func SubmitMafiaNightVote(roomID, userID, targetID string) error {
	state, ok := getMafiaState(roomID)
	if !ok {
		return fmt.Errorf("mafia state not found")
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.Phase != MafiaPhaseNight {
		return fmt.Errorf("not in night phase")
	}
	if !state.Alive[userID] {
		return fmt.Errorf("voter not alive")
	}
	if state.Roles[userID] != MafiaRoleMafia {
		return fmt.Errorf("only mafia can vote at night")
	}
	if !state.Alive[targetID] {
		return fmt.Errorf("target not alive")
	}
	if state.Roles[targetID] == MafiaRoleMafia {
		return fmt.Errorf("cannot target mafia")
	}

	state.NightVotes[userID] = targetID
	return nil
}

func SubmitMafiaDoctorSave(roomID, userID, targetID string) error {
	state, ok := getMafiaState(roomID)
	if !ok {
		return fmt.Errorf("mafia state not found")
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.Phase != MafiaPhaseNight {
		return fmt.Errorf("not in night phase")
	}
	if !state.Alive[userID] {
		return fmt.Errorf("doctor not alive")
	}
	if state.Roles[userID] != MafiaRoleDoctor {
		return fmt.Errorf("only doctor can save")
	}
	if !state.Alive[targetID] {
		return fmt.Errorf("target not alive")
	}

	state.DoctorSave = &targetID
	publishMafiaToUser(userID, map[string]interface{}{
		"type":     "MAFIA_DOCTOR_SELECTED",
		"roomId":   roomID,
		"targetId": targetID,
	})

	return nil
}

func SubmitMafiaPoliceCheck(roomID, userID, targetID string) error {
	state, ok := getMafiaState(roomID)
	if !ok {
		return fmt.Errorf("mafia state not found")
	}

	state.mu.Lock()
	defer state.mu.Unlock()

	if state.Phase != MafiaPhaseNight {
		return fmt.Errorf("not in night phase")
	}
	if !state.Alive[userID] {
		return fmt.Errorf("police not alive")
	}
	if state.Roles[userID] != MafiaRolePolice {
		return fmt.Errorf("only police can check")
	}
	if !state.Alive[targetID] {
		return fmt.Errorf("target not alive")
	}

	state.PoliceCheck = &targetID
	publishMafiaToUser(userID, map[string]interface{}{
		"type":     "MAFIA_POLICE_SELECTED",
		"roomId":   roomID,
		"targetId": targetID,
	})

	return nil
}

func HandleMafiaUserLeft(roomID, userID string) {
	state, ok := getMafiaState(roomID)
	if !ok {
		return
	}

	state.mu.Lock()
	if !state.Alive[userID] {
		state.mu.Unlock()
		return
	}
	state.Alive[userID] = false
	aliveNow := aliveList(state.Alive)
	state.mu.Unlock()

	publishMafiaToRoom(roomID, map[string]interface{}{
		"type":     "MAFIA_PLAYER_LEFT",
		"roomId":   roomID,
		"userId":   userID,
		"aliveNow": aliveNow,
	})

	checkMafiaWin(roomID)
}

func cleanupMafiaState(roomID string) {
	mafiaStatesMu.Lock()
	delete(mafiaStates, roomID)
	mafiaStatesMu.Unlock()
}

func resolveDayVote(roomID string) {
	state, ok := getMafiaState(roomID)
	if !ok {
		return
	}

	state.mu.Lock()

	aliveCount := countAlive(state.Alive)
	selected, maxVotes, tie := tallyVotes(state.DayVotes, state.Alive)

	var executed *string
	if maxVotes > aliveCount/2 && !tie {
		if selected != "" {
			state.Alive[selected] = false
			executed = &selected
			state.LastDayKilled = executed
		}
	}
	aliveNow := aliveList(state.Alive)
	state.mu.Unlock()

	payload := map[string]interface{}{
		"type":       "MAFIA_DAY_VOTE_RESULT",
		"roomId":     roomID,
		"executedId": executed,
		"aliveNow":   aliveNow,
	}
	publishMafiaToRoom(roomID, payload)
}

func resolveNight(roomID string) {
	state, ok := getMafiaState(roomID)
	if !ok {
		return
	}

	state.mu.Lock()

	mafiaAlive := mafiaAliveCount(state.Roles, state.Alive)
	target, maxVotes, tie := tallyVotes(state.NightVotes, state.Alive)

	var killed *string
	saved := false
	if maxVotes > mafiaAlive/2 && !tie && target != "" {
		if state.DoctorSave != nil && *state.DoctorSave == target {
			saved = true
		} else {
			state.Alive[target] = false
			killed = &target
			state.LastNightKill = killed
		}
	}

	policeCheck := state.PoliceCheck
	policeID := ""
	checkedRole := MafiaRole("")
	if policeCheck != nil {
		policeID = findRoleOwner(state.Roles, MafiaRolePolice, state.Alive)
		if policeID != "" {
			checkedRole = state.Roles[*policeCheck]
		}
	}
	aliveNow := aliveList(state.Alive)
	state.mu.Unlock()

	if policeCheck != nil && policeID != "" {
		publishMafiaToUser(policeID, map[string]interface{}{
			"type":     "MAFIA_POLICE_RESULT",
			"roomId":   roomID,
			"targetId": *policeCheck,
			"role":     checkedRole,
		})
	}

	publishMafiaToRoom(roomID, map[string]interface{}{
		"type":     "MAFIA_NIGHT_RESULT",
		"roomId":   roomID,
		"killedId": killed,
		"saved":    saved,
		"aliveNow": aliveNow,
	})
}

func checkMafiaWin(roomID string) bool {
	state, ok := getMafiaState(roomID)
	if !ok {
		return true
	}

	state.mu.Lock()
	mafiaCount := mafiaAliveCount(state.Roles, state.Alive)
	aliveCount := countAlive(state.Alive)
	state.mu.Unlock()

	others := aliveCount - mafiaCount
	if mafiaCount == 0 {
		endMafiaGame(roomID, "CITIZEN")
		return true
	}

	if mafiaCount > others {
		endMafiaGame(roomID, "MAFIA")
		return true
	}

	return false
}

func endMafiaGame(roomID, winner string) {
	room, exists := GetGameRoom(roomID)
	if !exists {
		cleanupMafiaState(roomID)
		return
	}

	room.Status = model.GameRoomStatus(common.RoomStatusFinished)
	SetGameRoom(roomID, room)
	publishRoomUpdateToWebSocket(roomID, room)

	publishMafiaToRoom(roomID, map[string]interface{}{
		"type":    "MAFIA_GAME_ENDED",
		"roomId":  roomID,
		"winner":  winner,
		"alive":   aliveListFromRoom(roomID),
		"message": "게임이 종료되었습니다.",
	})

	cleanupMafiaState(roomID)

	go func() {
		time.Sleep(common.RoomResetDelaySeconds * time.Second)
		roomMutex.Lock()
		defer roomMutex.Unlock()

		room, exists := gameRooms[roomID]
		if !exists {
			return
		}

		room.Status = model.GameRoomStatus(common.RoomStatusWaiting)
		room.CurrentRound = 0

		for i := range room.Users {
			room.Users[i].IsReady = false
			room.Users[i].Score = 0
		}

		gameRooms[roomID] = room
		go publishRoomUpdateToWebSocket(roomID, room)
		go publishLobbyUpdate()
	}()
}

func publishMafiaToRoom(roomID string, payload map[string]interface{}) {
	channel := common.ChannelGamePrefix + roomID
	publishPayload(channel, payload)
}

func publishMafiaToUser(userID string, payload map[string]interface{}) {
	channel := common.ChannelUserPrefix + userID
	publishPayload(channel, payload)
}

func publishPayload(channel string, payload map[string]interface{}) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[Mafia] Failed to marshal payload: %v", err)
		return
	}

	if err := valkey.PublishMessage(context.Background(), channel, string(jsonData)); err != nil {
		log.Printf("[Mafia] Failed to publish payload: %v", err)
	}
}

func getMafiaState(roomID string) (*mafiaState, bool) {
	mafiaStatesMu.RLock()
	state, ok := mafiaStates[roomID]
	mafiaStatesMu.RUnlock()
	return state, ok
}

func setMafiaState(roomID string, state *mafiaState) {
	mafiaStatesMu.Lock()
	mafiaStates[roomID] = state
	mafiaStatesMu.Unlock()
}

func isMafiaPhase(roomID string, phase MafiaPhase) bool {
	state, ok := getMafiaState(roomID)
	if !ok {
		return false
	}

	state.mu.Lock()
	defer state.mu.Unlock()
	return state.Phase == phase
}

func countAlive(alive map[string]bool) int {
	count := 0
	for _, isAlive := range alive {
		if isAlive {
			count++
		}
	}
	return count
}

func mafiaAliveCount(roles map[string]MafiaRole, alive map[string]bool) int {
	count := 0
	for userID, role := range roles {
		if role == MafiaRoleMafia && alive[userID] {
			count++
		}
	}
	return count
}

func aliveList(alive map[string]bool) []string {
	result := []string{}
	for userID, isAlive := range alive {
		if isAlive {
			result = append(result, userID)
		}
	}
	return result
}

func aliveListFromRoom(roomID string) []string {
	state, ok := getMafiaState(roomID)
	if !ok {
		return []string{}
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	return aliveList(state.Alive)
}

func tallyVotes(votes map[string]string, alive map[string]bool) (string, int, bool) {
	counts := map[string]int{}
	for voter, target := range votes {
		if !alive[voter] {
			continue
		}
		if !alive[target] {
			continue
		}
		counts[target]++
	}

	maxVotes := 0
	selected := ""
	tie := false
	for target, count := range counts {
		if count > maxVotes {
			maxVotes = count
			selected = target
			tie = false
		} else if count == maxVotes && count > 0 {
			tie = true
		}
	}

	return selected, maxVotes, tie
}

func findRoleOwner(roles map[string]MafiaRole, role MafiaRole, alive map[string]bool) string {
	for userID, userRole := range roles {
		if userRole == role && alive[userID] {
			return userID
		}
	}
	return ""
}
