# 🎮 미니 게임 서버 - API 가이드

> 멀티플레이어 게임 서버의 핵심 API 문서입니다.

---

## 📡 접속 정보

| 항목 | URL |
|------|-----|
| 🌐 **Base URL** | `http://localhost:8080` |
| 🔌 **GraphQL** | `/graphql` |
| 💬 **Lobby WebSocket** | `ws://localhost:8080/ws/lobby` |
| 🎯 **Room WebSocket** | `ws://localhost:8080/ws/rooms/:roomId` |
| ❤️ **Health Check** | `GET /health` |

---

## 📚 목차

### 1️⃣ GraphQL API
- [👤 사용자 관리](#-1-사용자-관리)
- [🎲 게임방 관리](#-2-게임방-관리)
- [🎮 게임 플레이](#-3-게임-플레이)

### 2️⃣ WebSocket API
- [💬 로비 WebSocket](#-로비-websocket)
- [🎯 게임방 WebSocket](#-게임방-websocket)

---

# 1️⃣ GraphQL API

---

## 👤 1. 사용자 관리

---

### 1.1 사용자 생성

**Mutation: createUser**

> 새 사용자 계정을 생성합니다.

**Request**:
```json
{
  "query": "mutation($input: CreateUserInput!) { createUser(input: $input) { id hangeId name avatarUrl level credit createdAt } }",
  "variables": {
    "input": {
      "hangeId": "user123",
      "name": "플레이어1",
      "avatarUrl": "https://example.com/avatar.jpg"
    }
  }
}
```

**Response**:
```json
{
  "data": {
    "createUser": {
      "id": "507f1f77bcf86cd799439011",
      "hangeId": "user123",
      "name": "플레이어1",
      "avatarUrl": "https://example.com/avatar.jpg",
      "level": 0,
      "credit": 0,
      "createdAt": "2026-01-21T10:00:00Z"
    }
  }
}
```

---

---

### 1.2 사용자 조회 (HangeId로)

**Query: userByHangeId**

> 로그인용 HangeId로 사용자를 조회합니다.

**Request**:
```json
{
  "query": "query($hangeId: String!) { userByHangeId(hangeId: $hangeId) { id hangeId name avatarUrl level credit } }",
  "variables": {
    "name": "user123"
  }
}
```

**Response**:
```json
{
  "data": {
    "userByHangeId": {
      "id": "507f1f77bcf86cd799439011",
      "hangeId": "user123",
      "name": "플레이어1",
      "avatarUrl": "https://example.com/avatar.jpg",
      "level": 5,
      "credit": 1000
    }
  }
}
```

---

---

## 🎲 2. 게임방 관리

---

### 2.1 게임방 생성

**Mutation: createGameRoom**

> 새로운 게임방을 생성합니다. 생성자가 자동으로 방장이 됩니다.

**Request**:
```json
{
  "query": "mutation($input: CreateGameRoomInput!) { createGameRoom(input: $input) { id name gameType status maxUsers totalRounds roundTimeLimit hostUserId isPrivate users { userId name score isReady } } }",
  "variables": {
    "input": {
      "name": "즐거운 끝말잇기",
      "gameType": "WORDCHAIN",
      "maxUsers": 4,
      "totalRounds": 5,
      "roundTimeLimit": 30,
      "hostUserId": "507f1f77bcf86cd799439011",
      "isPrivate": false
    }
  }
}
```

**Response**:
```json
{
  "data": {
    "createGameRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "즐거운 끝말잇기",
      "gameType": "WORDCHAIN",
      "status": "WAITING",
      "maxUsers": 4,
      "totalRounds": 5,
      "roundTimeLimit": 30,
      "hostUserId": "507f1f77bcf86cd799439011",
      "isPrivate": false,
      "users": [
        {
          "userId": "507f1f77bcf86cd799439011",
          "name": "플레이어1",
          "score": 0,
          "isReady": true
        }
      ]
    }
  }
}
```

---

---

### 2.2 게임방 목록 조회

**Query: gameRooms**

> 게임방 목록을 조회합니다. 다양한 필터 옵션 지원.

**Request**:
```json
{
  "query": "query($gameType: GameType, $status: GameStatus, $hasSpace: Boolean, $limit: Int) { gameRooms(gameType: $gameType, status: $status, hasSpace: $hasSpace, limit: $limit) { id name gameType status maxUsers currentRound totalRounds users { name score } } }",
  "variables": {
    "gameType": "WORDCHAIN",
    "status": "WAITING",
    "hasSpace": true,
    "limit": 20
  }
}
```

**Response**:
```json
{
  "data": {
    "gameRooms": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "즐거운 끝말잇기",
        "gameType": "WORDCHAIN",
        "status": "WAITING",
        "maxUsers": 4,
        "currentRound": 0,
        "totalRounds": 5,
        "users": [
          {
            "name": "플레이어1",
            "score": 0
          }
        ]
      }
    ]
  }
}
```

---

---

### 2.3 게임방 입장

**Mutation: joinGameRoom**

> 게임방에 입장합니다. 비공개방은 비밀번호가 필요합니다.

**Request**:
```json
{
  "query": "mutation($roomId: ID!, $userId: ID!, $password: String) { joinGameRoom(roomId: $roomId, userId: $userId, password: $password) { id name users { userId name score isReady } } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "507f1f77bcf86cd799439012"
  }
}
```

**Response**:
```json
{
  "data": {
    "joinGameRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "즐거운 끝말잇기",
      "users": [
        {
          "userId": "507f1f77bcf86cd799439011",
          "name": "플레이어1",
          "score": 0,
          "isReady": true
        },
        {
          "userId": "507f1f77bcf86cd799439012",
          "name": "플레이어2",
          "score": 0,
          "isReady": false
        }
      ]
    }
  }
}
```

---

---

### 2.4 게임방 퇴장

**Mutation: leaveGameRoom**

> 게임방에서 퇴장합니다.

**Request**:
```json
{
  "query": "mutation($roomId: ID!, $userId: ID!) { leaveGameRoom(roomId: $roomId, userId: $userId) { id users { name } hostUserId } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "507f1f77bcf86cd799439012"
  }
}
```

**Response**:
```json
{
  "data": {
    "leaveGameRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "users": [
        {
          "name": "플레이어1"
        }
      ],
      "hostUserId": "507f1f77bcf86cd799439011"
    }
  }
}
```

---

---

### 2.5 게임방 삭제

**Mutation: deleteGameRoom**

> 게임방을 삭제합니다. 방장만 가능합니다.

**Request**:
```json
{
  "query": "mutation($roomId: ID!) { deleteGameRoom(roomId: $roomId) }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**:
```json
{
  "data": {
    "deleteGameRoom": true
  }
}
```

---

---

## 🎮 3. 게임 플레이

---

### 3.1 게임 시작

**Mutation: startGame**

> 게임을 시작합니다. 방장만 실행 가능합니다.

**Request**:
```json
{
  "query": "mutation($roomId: ID!) { startGame(roomId: $roomId) { id status currentRound } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**:
```json
{
  "data": {
    "startGame": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "PLAYING",
      "currentRound": 1
    }
  }
}
```

---

---

### 3.2 OX 퀴즈 조회

**Query: randomOXQuiz**

> 랜덤 OX 퀴즈를 조회합니다.

**Request**:
```json
{
  "query": "query($roomId: ID) { randomOXQuiz(roomId: $roomId) { id category difficulty question } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**:
```json
{
  "data": {
    "randomOXQuiz": {
      "id": "123456789",
      "category": "과학",
      "difficulty": "medium",
      "question": "지구는 태양 주위를 돈다."
    }
  }
}
```

---

---

### 3.3 QA 퀴즈 조회

**Query: randomQAQuiz**

> 랜덤 4지선다 퀴즈를 조회합니다.

**Request**:
```json
{
  "query": "query($roomId: ID) { randomQAQuiz(roomId: $roomId) { id category difficulty question options } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**:
```json
{
  "data": {
    "randomQAQuiz": {
      "id": "987654321",
      "category": "지리",
      "difficulty": "easy",
      "question": "대한민국의 수도는?",
      "options": ["서울", "부산", "대구", "인천"]
    }
  }
}
```

---

---

### 3.4 끝말잇기 시작 단어 조회

**Query: randomWordchainPrompt**

> 끝말잇기 게임용 랜덤 시작 단어를 조회합니다.

**Request**:
```json
{
  "query": "query { randomWordchainPrompt { word hint } }"
}
```

**Response**:
```json
{
  "data": {
    "randomWordchainPrompt": {
      "word": "りんご",
      "hint": "사과"
    }
  }
}
```

---

---

### 3.5 단어 유효성 검증

**Query: isValidWord**

> 끝말잇기 단어의 유효성을 검증합니다.

**Request**:
```json
{
  "query": "query($word: String!) { isValidWord(word: $word) }",
  "variables": {
    "word": "りんご"
  }
}
```

**Response**:
```json
{
  "data": {
    "isValidWord": true
  }
}
```

---

---

### 3.6 답안 제출 (OX 퀴즈)

**Mutation: submitAnswer**

> OX 퀴즈 답안을 제출합니다.

**Request**:
```json
{
  "query": "mutation($roomId: ID!, $userId: ID!, $answer: String!) { submitAnswer(roomId: $roomId, userId: $userId, answer: $answer) { success isCorrect earnedScore totalScore explanation } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "507f1f77bcf86cd799439011",
    "answer": "true"
  }
}
```

**Response**:
```json
{
  "data": {
    "submitAnswer": {
      "success": true,
      "isCorrect": true,
      "earnedScore": 100,
      "totalScore": 100,
      "explanation": "정답입니다! 지구는 태양 주위를 공전합니다."
    }
  }
}
```

---

---

### 3.7 답안 제출 (QA 퀴즈)

**Mutation: submitAnswer**

> 4지선다 퀴즈 답안을 제출합니다.

**Request**:
```json
{
  "query": "mutation($roomId: ID!, $userId: ID!, $answer: String!) { submitAnswer(roomId: $roomId, userId: $userId, answer: $answer) { success isCorrect earnedScore totalScore explanation } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "507f1f77bcf86cd799439011",
    "answer": "0"
  }
}
```

**Response**:
```json
{
  "data": {
    "submitAnswer": {
      "success": true,
      "isCorrect": true,
      "earnedScore": 100,
      "totalScore": 200,
      "explanation": "정답입니다!"
    }
  }
}
```

---

---

### 3.8 게임 종료

**Mutation: endGame**

> 게임을 종료합니다. 방장만 실행 가능합니다.

**Request**:
```json
{
  "query": "mutation($roomId: ID!) { endGame(roomId: $roomId) { id status users { name score } } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**:
```json
{
  "data": {
    "endGame": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "FINISHED",
      "users": [
        {
          "name": "플레이어1",
          "score": 500
        },
        {
          "name": "플레이어2",
          "score": 400
        }
      ]
    }
  }
}
```

---

---

# 2️⃣ WebSocket API

---

## 💬 로비 WebSocket

> **연결**: `ws://localhost:8080/ws/lobby`

---

### 1. 사용자 식별

**WebSocket Event: identify**

> WebSocket 연결 후 사용자 정보를 서버에 알립니다.

**Client → Server**:
```json
{
  "type": "identify",
  "data": {
    "UserName": "플레이어1",
    "roomId": "lobby"
  }
}
```

**Response**: 없음 (연결 확립)

---

---

### 2. 로비 채널 구독

**WebSocket Event: subscribe**

> 로비 채널을 구독하여 로비 메시지를 받습니다.

**Client → Server**:
```json
{
  "type": "subscribe",
  "channel": "lobby"
}
```

**Response**: 없음

---

---

### 3. 로비 채팅

**WebSocket Event: lobby_chat**

> 로비에서 채팅 메시지를 전송합니다.

**Client → Server**:
```json
{
  "type": "lobby_chat",
  "data": {
    "userId": "507f1f77bcf86cd799439011",
    "userName": "플레이어1",
    "message": "같이 게임 하실 분?"
  }
}
```

**Server → All Clients**:
```json
{
  "type": "message",
  "channel": "lobby",
  "data": "{\"type\":\"LOBBY_CHAT\",\"payload\":{\"userId\":\"507f1f77bcf86cd799439011\",\"userName\":\"플레이어1\",\"message\":\"같이 게임 하실 분?\"}}"
}
```

---

---

## 🎯 게임방 WebSocket

> **연결**: `ws://localhost:8080/ws/rooms/:roomId`

---

### 1. 게임방 채널 자동 구독

> 게임방 WebSocket 연결 시 자동으로 게임방 채널이 구독됩니다.

**Response**: 없음 (자동 처리)

---

---

### 2. 게임방 채팅

**WebSocket Event: chat**

> 게임방에서 채팅 메시지를 전송합니다.

**Client → Server**:
```json
{
  "type": "chat",
  "data": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "507f1f77bcf86cd799439011",
    "userName": "플레이어1",
    "message": "안녕하세요!"
  }
}
```

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"CHAT_MESSAGE\",\"payload\":{\"userId\":\"507f1f77bcf86cd799439011\",\"userName\":\"플레이어1\",\"message\":\"안녕하세요!\",\"timestamp\":\"2026-01-21T10:40:00Z\"}}"
}
```

---

---

### 3. 게임 상태 업데이트

**WebSocket Event: game_state**

> 게임 상태가 변경될 때 서버에서 전송됩니다.

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"GAME_STATE\",\"payload\":{\"status\":\"PLAYING\",\"currentRound\":2,\"totalRounds\":5,\"currentTurn\":\"플레이어1\",\"timeRemaining\":30}}"
}
```

---

---

### 4. 라운드 시작

**WebSocket Event: round_start**

> 새로운 라운드가 시작될 때 서버에서 전송됩니다.

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"ROUND_START\",\"payload\":{\"roundNumber\":3,\"question\":\"대한민국의 수도는?\",\"options\":[\"서울\",\"부산\",\"대구\",\"인천\"],\"timeLimit\":30}}"
}
```

---

---

### 5. 라운드 종료

**WebSocket Event: round_end**

> 라운드가 종료될 때 서버에서 전송됩니다.

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"ROUND_END\",\"payload\":{\"roundNumber\":3,\"correctAnswer\":\"서울\",\"results\":[{\"userName\":\"플레이어1\",\"isCorrect\":true,\"earnedScore\":100,\"totalScore\":300},{\"userName\":\"플레이어2\",\"isCorrect\":false,\"earnedScore\":0,\"totalScore\":200}]}}"
}
```

---

---

### 6. 정답 알림

**WebSocket Event: correct_answer**

> 누군가 정답을 맞췄을 때 서버에서 전송됩니다.

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"CORRECT_ANSWER\",\"payload\":{\"userName\":\"플레이어1\",\"answer\":\"서울\",\"earnedScore\":100,\"totalScore\":300}}"
}
```

---

---

### 7. 게임 종료 알림

**WebSocket Event: game_end**

> 게임이 종료될 때 서버에서 전송됩니다.

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"GAME_END\",\"payload\":{\"winner\":\"플레이어1\",\"finalScores\":[{\"userName\":\"플레이어1\",\"score\":500,\"rank\":1},{\"userName\":\"플레이어2\",\"score\":400,\"rank\":2}],\"reason\":\"모든 라운드 완료\"}}"
}
```

---

---

### 8. 방 삭제 알림

**WebSocket Event: room_deleted**

> 게임방이 삭제될 때 서버에서 전송됩니다.

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"ROOM_DELETED\",\"payload\":{\"roomId\":\"550e8400-e29b-41d4-a716-446655440000\",\"reason\":\"방장이 방을 삭제했습니다.\"}}"
}
```

---

---

### 9. 끝말잇기 단어 제출

**WebSocket Event: wordchain_submit**

> 끝말잇기 게임에서 단어를 제출합니다.

**Client → Server**:
```json
{
  "type": "wordchain_submit",
  "data": {
    "username": "플레이어1",
    "word": "ごりら",
    "lastWord": "りんご"
  }
}
```

**Server → All Room Clients (성공)**:
```json
{
  "type": "message",
  "data": "{\"type\":\"WORDCHAIN_RESULT\",\"payload\":{\"success\":true,\"userName\":\"플레이어1\",\"word\":\"ごりら\",\"earnedScore\":100,\"totalScore\":100}}"
}
```

**Server → All Room Clients (실패)**:
```json
{
  "type": "message",
  "data": "{\"type\":\"WORDCHAIN_RESULT\",\"payload\":{\"success\":false,\"userName\":\"플레이어1\",\"word\":\"ごりら\",\"reason\":\"이미 사용된 단어입니다. 다른 단어를 입력해주세요.\",\"earnedScore\":0,\"totalScore\":0}}"
}
```

---

---

### 10. 그림 그리기 (향후 기능)

**WebSocket Event: drawing**

> 그림 맞추기 게임에서 그림 데이터를 전송합니다.

**Client → Server**:
```json
{
  "type": "drawing",
  "data": {
    "x": 150,
    "y": 200,
    "color": "#FF0000",
    "brushSize": 5,
    "action": "draw"
  }
}
```

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"DRAWING\",\"payload\":{\"x\":150,\"y\":200,\"color\":\"#FF0000\",\"brushSize\":5,\"action\":\"draw\",\"userId\":\"507f1f77bcf86cd799439011\"}}"
}
```

---

---

### 11. 게임 액션

**WebSocket Event: game_action**

> 게임 내 액션(준비 등)을 전송합니다.

**Client → Server**:
```json
{
  "type": "game_action",
  "data": {
    "action": "ready",
    "userId": "507f1f77bcf86cd799439011"
  }
}
```

**Server → All Room Clients**:
```json
{
  "type": "message",
  "data": "{\"type\":\"GAME_ACTION\",\"payload\":{\"action\":\"ready\",\"userId\":\"507f1f77bcf86cd799439011\",\"userName\":\"플레이어1\"}}"
}
```

---

---

## 오류 처리

### GraphQL 오류

```json
{
  "data": null,
  "errors": [
    {
      "message": "room not found",
      "path": ["gameRoom"],
      "extensions": {
        "code": "NOT_FOUND"
      }
    }
  ]
}
```

---

## HTTP 상태 코드

| 코드 | 의미 | 설명 |
|------|------|------|
| ✅ **200** | Success | 요청 성공 |
| ⚠️ **400** | Bad Request | 잘못된 요청 |
| 🔒 **401** | Unauthorized | 인증 실패 |
| ❌ **404** | Not Found | 리소스 없음 |
| 💥 **500** | Server Error | 서버 오류 |

---

## 열거형 타입

### GameType (게임 타입)

| 값 | 설명 |
|------|------|
| `OX` | ⭕ OX 퀴즈 (참/거짓) |
| `QA` | 📝 4지선다 퀴즈 |
| `WORDCHAIN` | 🔤 끝말잇기 (일본어) |
| `DRAWING` | 🎨 그림 맞추기 |

### GameStatus (게임 상태)

| 값 | 설명 |
|------|------|
| `WAITING` | ⏳ 대기 중 (플레이어 모집) |
| `PLAYING` | 🎮 게임 진행 중 |
| `FINISHED` | 🏁 게임 종료 |

---

## 개발 도구

| 도구 | URL |
|------|-----|
| 🎮 **GraphQL Playground** | http://localhost:8080/graphql |
| ❤️ **Health Check** | GET http://localhost:8080/health |

---

> **💡 Tip**: 노션에서 코드 블록은 `/code` 명령어로, 토글은 `/toggle` 명령어로 생성할 수 있습니다!
