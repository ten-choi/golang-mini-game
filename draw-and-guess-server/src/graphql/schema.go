package graphql

import (
	"context"
	"draw-and-guess-server/src/database"
	"draw-and-guess-server/src/models"
	"draw-and-guess-server/src/valkey"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "User",
	Description: "게임 사용자 정보",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.String,
			Description: "사용자 고유 ID (MongoDB ObjectID)",
		},
		"nickname": &graphql.Field{
			Type:        graphql.String,
			Description: "사용자 닉네임",
		},
		"profileImage": &graphql.Field{
			Type:        graphql.String,
			Description: "프로필 이미지 URL",
		},
		"winningPoint": &graphql.Field{
			Type:        graphql.Int,
			Description: "승리 포인트 (총 획득 점수)",
		},
		"createdAt": &graphql.Field{
			Type:        graphql.String,
			Description: "생성 일시 (ISO 8601 형식)",
		},
		"updatedAt": &graphql.Field{
			Type:        graphql.String,
			Description: "수정 일시 (ISO 8601 형식)",
		},
	},
})

var playerType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Player",
	Description: "게임방 참가자 정보",
	Fields: graphql.Fields{
		"userId": &graphql.Field{
			Type:        graphql.String,
			Description: "플레이어 ID (username과 동일)",
		},
		"username": &graphql.Field{
			Type:        graphql.String,
			Description: "플레이어 이름",
		},
		"score": &graphql.Field{
			Type:        graphql.Int,
			Description: "현재 게임 점수",
		},
		"isReady": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "준비 상태 (게임 시작 전)",
		},
	},
})

var gameRoomType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "GameRoom",
	Description: "그림 맞추기 게임방",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.String,
			Description: "게임방 고유 ID (UUID)",
		},
		"name": &graphql.Field{
			Type:        graphql.String,
			Description: "게임방 이름",
		},
		"hostId": &graphql.Field{
			Type:        graphql.String,
			Description: "방장 사용자 이름",
		},
		"players": &graphql.Field{
			Type:        graphql.NewList(playerType),
			Description: "참가자 목록",
		},
		"maxPlayers": &graphql.Field{
			Type:        graphql.Int,
			Description: "최대 참가자 수 (기본 10명)",
		},
		"currentRound": &graphql.Field{
			Type:        graphql.Int,
			Description: "현재 라운드 번호 (1부터 시작)",
		},
		"totalRounds": &graphql.Field{
			Type:        graphql.Int,
			Description: "총 라운드 수 (기본 3라운드)",
		},
		"status": &graphql.Field{
			Type:        graphql.String,
			Description: "게임 상태 (waiting: 대기 중, playing: 진행 중, finished: 종료)",
		},
		"createdAt": &graphql.Field{
			Type:        graphql.String,
			Description: "생성 일시 (ISO 8601 형식)",
		},
		"updatedAt": &graphql.Field{
			Type:        graphql.String,
			Description: "수정 일시 (ISO 8601 형식)",
		},
	},
})

var quizType = graphql.NewObject(graphql.ObjectConfig{
	Name: "Quiz",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type: graphql.String,
		},
		"type": &graphql.Field{
			Type: graphql.String,
		},
		"category": &graphql.Field{
			Type: graphql.String,
		},
		"difficulty": &graphql.Field{
			Type: graphql.String,
		},
		"question": &graphql.Field{
			Type: graphql.String,
		},
		"answer": &graphql.Field{
			Type: graphql.String,
		},
		"hint": &graphql.Field{
			Type: graphql.String,
		},
		"explanation": &graphql.Field{
			Type: graphql.String,
		},
		"imageUrl": &graphql.Field{
			Type: graphql.String,
		},
		"usageCount": &graphql.Field{
			Type: graphql.Int,
		},
		"isActive": &graphql.Field{
			Type: graphql.Boolean,
		},
		"createdAt": &graphql.Field{
			Type: graphql.String,
		},
		"updatedAt": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var webSocketInfoType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "WebSocketInfo",
	Description: "실시간 통신을 위한 WebSocket 연결 정보",
	Fields: graphql.Fields{
		"url": &graphql.Field{
			Type:        graphql.String,
			Description: "WebSocket 서버 URL (ws://localhost:8080/app/ws)",
		},
		"description": &graphql.Field{
			Type:        graphql.String,
			Description: "WebSocket 사용 방법 설명",
		},
		"channels": &graphql.Field{
			Type:        graphql.NewList(graphql.String),
			Description: "사용 가능한 채널 목록 (game, chat, draw)",
		},
	},
})

var queryType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Query",
	Description: "데이터 조회를 위한 Query 작업",
	Fields: graphql.Fields{
		"user": &graphql.Field{
			Type:        userType,
			Description: "사용자 ID로 특정 사용자 조회",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "조회할 사용자 ID (MongoDB ObjectID Hex 문자열)",
				},
			},
			Resolve: resolveUser,
		},
		"gameRooms": &graphql.Field{
			Type:        graphql.NewList(gameRoomType),
			Description: "현재 활성화된 모든 게임방 목록 조회",
			Resolve:     resolveGameRooms,
		},
		"gameRoom": &graphql.Field{
			Type:        gameRoomType,
			Description: "게임방 ID로 특정 게임방 조회",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "조회할 게임방 ID (UUID)",
				},
			},
			Resolve: resolveGameRoom,
		},
		"webSocketInfo": &graphql.Field{
			Type:        webSocketInfoType,
			Description: "WebSocket 연결 정보 및 사용 가능한 채널 목록",
			Resolve:     resolveWebSocketInfo,
		},
	},
})

var mutationType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Mutation",
	Description: "데이터 생성/수정/삭제를 위한 Mutation 작업",
	Fields: graphql.Fields{
		"createUser": &graphql.Field{
			Type:        userType,
			Description: "새로운 사용자 생성 (닉네임 중복 불가)",
			Args: graphql.FieldConfigArgument{
				"nickname": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "사용자 닉네임 (필수, 중복 불가)",
				},
			},
			Resolve: resolveCreateUser,
		},
		"updateUser": &graphql.Field{
			Type:        userType,
			Description: "사용자 정보 수정 (닉네임, 프로필 이미지)",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "수정할 사용자 ID (필수)",
				},
				"nickname": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "새 닉네임 (선택)",
				},
				"profileImage": &graphql.ArgumentConfig{
					Type:        graphql.String,
					Description: "새 프로필 이미지 URL (선택)",
				},
			},
			Resolve: resolveUpdateUser,
		},
		"createGameRoom": &graphql.Field{
			Type:        gameRoomType,
			Description: "새 게임방 생성 (방장이 자동으로 그림 그리는 사람이 됨)",
			Args: graphql.FieldConfigArgument{
				"hostUsername": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "방장 사용자 이름 (필수)",
				},
				"maxRounds": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "총 라운드 수 (선택, 기본값 3)",
				},
			},
			Resolve: resolveCreateGameRoom,
		},
		"joinGameRoom": &graphql.Field{
			Type:        gameRoomType,
			Description: "게임방 참가 (같은 사용자 중복 참가 불가)",
			Args: graphql.FieldConfigArgument{
				"roomId": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "참가할 게임방 ID (필수)",
				},
				"username": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "참가자 이름 (필수)",
				},
			},
			Resolve: resolveJoinGameRoom,
		},
		"leaveGameRoom": &graphql.Field{
			Type:        gameRoomType,
			Description: "게임방 나가기",
			Args: graphql.FieldConfigArgument{
				"roomId": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "나갈 게임방 ID (필수)",
				},
				"username": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "나가는 참가자 이름 (필수)",
				},
			},
			Resolve: resolveLeaveGameRoom,
		},
		"startGame": &graphql.Field{
			Type:        gameRoomType,
			Description: "게임 시작 (최소 2명 이상 필요, status가 'playing'으로 변경됨)",
			Args: graphql.FieldConfigArgument{
				"roomId": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "시작할 게임방 ID (필수)",
				},
			},
			Resolve: resolveStartGame,
		},
		"deleteGameRoom": &graphql.Field{
			Type:        graphql.Boolean,
			Description: "게임방 삭제 (성공 시 true 반환)",
			Args: graphql.FieldConfigArgument{
				"roomId": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "삭제할 게임방 ID (필수)",
				},
			},
			Resolve: resolveDeleteGameRoom,
		},
	},
})

// Message types for subscriptions
var gameEventType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "GameEvent",
	Description: "게임 이벤트 (플레이어 입장/퇴장, 라운드 시작/종료 등)",
	Fields: graphql.Fields{
		"type": &graphql.Field{
			Type:        graphql.String,
			Description: "이벤트 타입 (player_joined, player_left, round_start, round_end, game_end 등)",
		},
		"roomId": &graphql.Field{
			Type:        graphql.String,
			Description: "게임방 ID",
		},
		"data": &graphql.Field{
			Type:        graphql.String,
			Description: "이벤트 데이터 (JSON 문자열)",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.String,
			Description: "이벤트 발생 시각",
		},
	},
})

var chatMessageType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "ChatMessage",
	Description: "채팅 메시지",
	Fields: graphql.Fields{
		"roomId": &graphql.Field{
			Type:        graphql.String,
			Description: "게임방 ID",
		},
		"userId": &graphql.Field{
			Type:        graphql.String,
			Description: "발신자 ID",
		},
		"username": &graphql.Field{
			Type:        graphql.String,
			Description: "발신자 이름",
		},
		"message": &graphql.Field{
			Type:        graphql.String,
			Description: "채팅 내용",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.String,
			Description: "전송 시각",
		},
	},
})

var drawingDataType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "DrawingData",
	Description: "실시간 그림 그리기 데이터",
	Fields: graphql.Fields{
		"roomId": &graphql.Field{
			Type:        graphql.String,
			Description: "게임방 ID",
		},
		"action": &graphql.Field{
			Type:        graphql.String,
			Description: "그리기 액션 (draw, clear, undo)",
		},
		"data": &graphql.Field{
			Type:        graphql.String,
			Description: "그리기 데이터 (JSON 문자열 - points, color, width 포함)",
		},
		"timestamp": &graphql.Field{
			Type:        graphql.String,
			Description: "그리기 시각",
		},
	},
})

var subscriptionType = graphql.NewObject(graphql.ObjectConfig{
	Name:        "Subscription",
	Description: "실시간 이벤트 구독 (WebSocket 기반)",
	Fields: graphql.Fields{
		"gameEvents": &graphql.Field{
			Type:        gameEventType,
			Description: "게임 이벤트 실시간 구독 (플레이어 입장/퇴장, 라운드 시작/종료, 게임 종료 등)",
			Args: graphql.FieldConfigArgument{
				"roomId": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "구독할 게임방 ID",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source, nil
			},
		},
		"chatMessages": &graphql.Field{
			Type:        chatMessageType,
			Description: "채팅 메시지 실시간 구독",
			Args: graphql.FieldConfigArgument{
				"roomId": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "구독할 게임방 ID",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source, nil
			},
		},
		"drawingUpdates": &graphql.Field{
			Type:        drawingDataType,
			Description: "그림 그리기 데이터 실시간 구독",
			Args: graphql.FieldConfigArgument{
				"roomId": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "구독할 게임방 ID",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source, nil
			},
		},
	},
})

var Schema, _ = graphql.NewSchema(graphql.SchemaConfig{
	Query:        queryType,
	Mutation:     mutationType,
	Subscription: subscriptionType,
})

// Resolver functions

// User Resolvers
func resolveUser(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	collection := database.GetCollection("users")
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	return map[string]interface{}{
		"id":           user.ID.Hex(),
		"nickname":     user.Nickname,
		"profileImage": user.ProfileImage,
		"winningPoint": user.WinningPoint,
		"createdAt":    user.CreatedAt.Format(time.RFC3339),
		"updatedAt":    user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func resolveCreateUser(p graphql.ResolveParams) (interface{}, error) {
	nickname := p.Args["nickname"].(string)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user already exists
	collection := database.GetCollection("users")
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"nickname": nickname}).Decode(&existingUser)
	if err == nil {
		return nil, fmt.Errorf("user with nickname '%s' already exists", nickname)
	}

	now := time.Now()
	user := models.User{
		ID:           primitive.NewObjectID(),
		Nickname:     nickname,
		ProfileImage: "",
		WinningPoint: 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	return map[string]interface{}{
		"id":           user.ID.Hex(),
		"nickname":     user.Nickname,
		"profileImage": user.ProfileImage,
		"winningPoint": user.WinningPoint,
		"createdAt":    user.CreatedAt.Format(time.RFC3339),
		"updatedAt":    user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func resolveUpdateUser(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateFields := bson.M{"updated_at": time.Now()}
	if nickname, ok := p.Args["nickname"].(string); ok {
		updateFields["nickname"] = nickname
	}
	if profileImage, ok := p.Args["profileImage"].(string); ok {
		updateFields["profileImage"] = profileImage
	}

	collection := database.GetCollection("users")
	_, err = collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updateFields})
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, fmt.Errorf("user not found after update")
	}

	return map[string]interface{}{
		"id":           user.ID.Hex(),
		"nickname":     user.Nickname,
		"profileImage": user.ProfileImage,
		"winningPoint": user.WinningPoint,
		"createdAt":    user.CreatedAt.Format(time.RFC3339),
		"updatedAt":    user.UpdatedAt.Format(time.RFC3339),
	}, nil
}

// GameRoom Resolvers
func resolveGameRooms(p graphql.ResolveParams) (interface{}, error) {
	pattern := "game_room:*"
	keys, err := valkey.GetKeys(pattern)
	if err != nil {
		return []interface{}{}, nil
	}

	rooms := []map[string]interface{}{}
	for _, key := range keys {
		var room models.GameRoom
		if err := valkey.GetJSON(key, &room); err == nil {
			rooms = append(rooms, convertGameRoomToMap(&room))
		}
	}

	return rooms, nil
}

func resolveGameRoom(p graphql.ResolveParams) (interface{}, error) {
	id := p.Args["id"].(string)

	var room models.GameRoom
	err := valkey.GetJSON("game_room:"+id, &room)
	if err != nil {
		return nil, fmt.Errorf("room not found")
	}

	return convertGameRoomToMap(&room), nil
}

func resolveCreateGameRoom(p graphql.ResolveParams) (interface{}, error) {
	hostUsername := p.Args["hostUsername"].(string)
	maxRounds := 3
	if val, ok := p.Args["maxRounds"].(int); ok {
		maxRounds = val
	}

	roomUUID := uuid.New().String()
	now := time.Now()

	room := models.GameRoom{
		UUID:         roomUUID,
		IsActive:     true,
		DrawerUser:   hostUsername,
		RoomCreator:  hostUsername,
		Players:      []models.Player{},
		RoundNumber:  1,
		TimeLeft:     60,
		GameStatus:   "waiting",
		UsedWords:    []string{},
		MaxRounds:    maxRounds,
		WinningScore: 3,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err := valkey.SetJSON("game_room:"+roomUUID, &room, 3600)
	if err != nil {
		return nil, fmt.Errorf("failed to create game room: %v", err)
	}

	return convertGameRoomToMap(&room), nil
}

func resolveJoinGameRoom(p graphql.ResolveParams) (interface{}, error) {
	roomID := p.Args["roomId"].(string)
	username := p.Args["username"].(string)

	var room models.GameRoom
	err := valkey.GetJSON("game_room:"+roomID, &room)
	if err != nil {
		return nil, fmt.Errorf("room not found")
	}

	// Check if player already in room
	for _, player := range room.Players {
		if player.Username == username {
			return nil, fmt.Errorf("player already in room")
		}
	}

	// Add player
	room.Players = append(room.Players, models.Player{
		Username: username,
		Score:    0,
		Attempts: 0,
	})
	room.UpdatedAt = time.Now()

	err = valkey.SetJSON("game_room:"+roomID, &room, 3600)
	if err != nil {
		return nil, fmt.Errorf("failed to join room: %v", err)
	}

	return convertGameRoomToMap(&room), nil
}

func resolveLeaveGameRoom(p graphql.ResolveParams) (interface{}, error) {
	roomID := p.Args["roomId"].(string)
	username := p.Args["username"].(string)

	var room models.GameRoom
	err := valkey.GetJSON("game_room:"+roomID, &room)
	if err != nil {
		return nil, fmt.Errorf("room not found")
	}

	// Remove player
	newPlayers := []models.Player{}
	for _, player := range room.Players {
		if player.Username != username {
			newPlayers = append(newPlayers, player)
		}
	}
	room.Players = newPlayers
	room.UpdatedAt = time.Now()

	err = valkey.SetJSON("game_room:"+roomID, &room, 3600)
	if err != nil {
		return nil, fmt.Errorf("failed to leave room: %v", err)
	}

	return convertGameRoomToMap(&room), nil
}

func resolveStartGame(p graphql.ResolveParams) (interface{}, error) {
	roomID := p.Args["roomId"].(string)

	var room models.GameRoom
	err := valkey.GetJSON("game_room:"+roomID, &room)
	if err != nil {
		return nil, fmt.Errorf("room not found")
	}

	if len(room.Players) < 2 {
		return nil, fmt.Errorf("need at least 2 players to start")
	}

	room.GameStatus = "playing"
	room.RoundNumber = 1
	room.TimeLeft = 60
	room.UpdatedAt = time.Now()

	err = valkey.SetJSON("game_room:"+roomID, &room, 3600)
	if err != nil {
		return nil, fmt.Errorf("failed to start game: %v", err)
	}

	return convertGameRoomToMap(&room), nil
}

func resolveDeleteGameRoom(p graphql.ResolveParams) (interface{}, error) {
	roomID := p.Args["roomId"].(string)

	err := valkey.DeleteKey("game_room:" + roomID)
	if err != nil {
		return false, fmt.Errorf("failed to delete room: %v", err)
	}

	return true, nil
}

// WebSocket Resolver
func resolveWebSocketInfo(p graphql.ResolveParams) (interface{}, error) {
	return map[string]interface{}{
		"url":         "ws://localhost:8080/app/ws",
		"description": "WebSocket endpoint for real-time game communication",
		"channels": []string{
			"game/room-{roomId} - Game events (player join/leave, round start/end)",
			"chat/room-{roomId} - Chat messages",
			"draw/room-{roomId} - Drawing data",
		},
	}, nil
}

// Helper function
func convertGameRoomToMap(room *models.GameRoom) map[string]interface{} {
	players := []map[string]interface{}{}
	for _, player := range room.Players {
		players = append(players, map[string]interface{}{
			"userId":   player.Username,
			"username": player.Username,
			"score":    player.Score,
			"isReady":  false,
		})
	}

	return map[string]interface{}{
		"id":           room.UUID,
		"name":         room.UUID,
		"hostId":       room.RoomCreator,
		"players":      players,
		"maxPlayers":   10,
		"currentRound": room.RoundNumber,
		"totalRounds":  room.MaxRounds,
		"status":       room.GameStatus,
		"createdAt":    room.CreatedAt.Format(time.RFC3339),
		"updatedAt":    room.UpdatedAt.Format(time.RFC3339),
	}
}
