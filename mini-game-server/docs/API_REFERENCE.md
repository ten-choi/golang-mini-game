# API Reference

**Base URL**: `http://localhost:8080`
- GraphQL: `/graphql`
- WebSocket (Lobby): `ws://localhost:8080/ws/lobby`
- WebSocket (Room): `ws://localhost:8080/ws/rooms/:id`

---

## 1. 사용자 관리 (User Management)

### 1.1 사용자 생성

**GraphQL Mutation**
```graphql
mutation {
  createUser(input: {
    hangeId: "user123"
    name: "플레이어1"
    avatarUrl: "https://example.com/avatar.jpg"
  }) {
    id
    hangeId
    name
    avatarUrl
    level
    credit
    createdAt
  }
}
```

**Request Body**
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

**Response**
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
      "createdAt": "2026-01-19T10:30:00Z"
    }
  }
}
```

### 1.2 사용자 조회 (HangeId로)

**GraphQL Query**
```json
{
  "query": "query($name: String!) { userByHangeId(name: $name) { id hangeId name level credit } }",
  "variables": {
    "name": "user123"
  }
}
```

**Response**
```json
{
  "data": {
    "userByHangeId": {
      "id": "507f1f77bcf86cd799439011",
      "hangeId": "user123",
      "name": "플레이어1",
      "level": 5,
      "credit": 1000
    }
  }
}
```

### 1.3 사용자 조회 (이름으로)

**GraphQL Query**
```json
{
  "query": "query($name: String!) { userByName(name: $name) { id name level credit } }",
  "variables": {
    "name": "플레이어1"
  }
}
```

**Response**
```json
{
  "data": {
    "userByName": {
      "id": "507f1f77bcf86cd799439011",
      "name": "플레이어1",
      "level": 5,
      "credit": 1000
    }
  }
}
```

### 1.4 사용자 정보 수정

**GraphQL Mutation**
```json
{
  "query": "mutation($name: String!, $input: UpdateUserInput!) { updateUser(name: $name, input: $input) { id name avatarUrl level credit updatedAt } }",
  "variables": {
    "name": "플레이어1",
    "input": {
      "avatarUrl": "https://example.com/new-avatar.jpg",
      "level": 10,
      "credit": 5000
    }
  }
}
```

**Response**
```json
{
  "data": {
    "updateUser": {
      "id": "507f1f77bcf86cd799439011",
      "name": "플레이어1",
      "avatarUrl": "https://example.com/new-avatar.jpg",
      "level": 10,
      "credit": 5000,
      "updatedAt": "2026-01-19T11:00:00Z"
    }
  }
}
```

### 1.5 사용자 삭제

**GraphQL Mutation**
```json
{
  "query": "mutation($name: String!) { deleteUser(name: $name) }",
  "variables": {
    "name": "플레이어1"
  }
}
```

**Response**
```json
{
  "data": {
    "deleteUser": true
  }
}
```

### 1.6 사용자 목록 조회 (검색/페이징)

**GraphQL Query**
```json
{
  "query": "query($limit: Int, $offset: Int, $search: String) { users(limit: $limit, offset: $offset, search: $search) { name level credit } }",
  "variables": {
    "limit": 10,
    "offset": 0,
    "search": "플레이어"
  }
}
```

**Response**
```json
{
  "data": {
    "users": [
      { "name": "플레이어1", "level": 5, "credit": 1000 },
      { "name": "플레이어2", "level": 3, "credit": 500 }
    ]
  }
}
```

### 1.7 사용자 통계 조회

**GraphQL Query**
```json
{
  "query": "query($username: String!, $gameType: String) { userStats(username: $username, gameType: $gameType) { gameType totalGames totalWins totalScore } }",
  "variables": {
    "username": "플레이어1",
    "gameType": "WORDCHAIN"
  }
}
```

**Response**
```json
{
  "data": {
    "userStats": [
      {
        "gameType": "WORDCHAIN",
        "totalGames": 25,
        "totalWins": 8,
        "totalScore": 12500
      }
    ]
  }
}
```

### 1.8 리더보드 조회

**GraphQL Query**
```json
{
  "query": "query($gameType: String!, $limit: Int) { leaderboard(gameType: $gameType, limit: $limit) { username totalScore totalWins } }",
  "variables": {
    "gameType": "OX",
    "limit": 10
  }
}
```

**Response**
```json
{
  "data": {
    "leaderboard": [
      { "username": "프로게이머", "totalScore": 50000, "totalWins": 150 },
      { "username": "천재", "totalScore": 45000, "totalWins": 130 }
    ]
  }
}
```
### 1.9 게임 설정 조회

**GraphQL Query**
```json
{
  "query": "query { gameConfig { maxUsers roundDuration roundsPerGame } }"
}
```

**Response**
```json
{
  "data": {
    "gameConfig": {
      "maxUsers": 10,
      "roundDuration": 30,
      "roundsPerGame": 5
    }
  }
}
```
---

## 2. 게임방 관리 (Game Room Management)

### 2.1 게임방 생성

**GraphQL Mutation**
```json
{
  "query": "mutation($input: CreateGameRoomInput!) { createGameRoom(input: $input) { id name gameType status users { name score isReady } maxUsers hostUserId } }",
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

**Response**
```json
{
  "data": {
    "createGameRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "즐거운 끝말잇기",
      "gameType": "WORDCHAIN",
      "status": "WAITING",
      "users": [
        { "name": "플레이어1", "score": 0, "isReady": true }
      ],
      "maxUsers": 4,
      "hostUserId": "507f1f77bcf86cd799439011"
    }
  }
}
```

### 2.2 게임방 목록 조회

**GraphQL Query**
```json
{
  "query": "query($gameType: GameType, $status: GameStatus, $hasSpace: Boolean, $limit: Int) { gameRooms(gameType: $gameType, status: $status, hasSpace: $hasSpace, limit: $limit) { id name gameType status users { name } maxUsers } }",
  "variables": {
    "gameType": "WORDCHAIN",
    "status": "WAITING",
    "hasSpace": true,
    "limit": 20
  }
}
```

**Response**
```json
{
  "data": {
    "gameRooms": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "name": "즐거운 끝말잇기",
        "gameType": "WORDCHAIN",
        "status": "WAITING",
        "users": [{ "name": "플레이어1" }],
        "maxUsers": 4
      }
    ]
  }
}
```

}
```

### 2.3 게임방 조회 (단일)

**GraphQL Query**
```json
{
  "query": "query($id: ID!) { gameRoom(id: $id) { id name gameType status users { name score isReady } maxUsers currentRound totalRounds } }",
  "variables": {
    "id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**
```json
{
  "data": {
    "gameRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "즐거운 끝말잇기",
      "gameType": "WORDCHAIN",
      "status": "WAITING",
      "users": [
        { "name": "플레이어1", "score": 0, "isReady": true },
        { "name": "플레이어2", "score": 0, "isReady": false }
      ],
      "maxUsers": 4,
      "currentRound": 0,
      "totalRounds": 5
    }
  }
}
```

### 2.4 내 현재 게임방 조회

**GraphQL Query**
```json
{
  "query": "query($username: String!) { myCurrentRoom(username: $username) { id name gameType status currentRound users { name score } } }",
  "variables": {
    "username": "플레이어1"
  }
}
```

**Response**
```json
{
  "data": {
    "myCurrentRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "즐거운 끝말잇기",
      "gameType": "WORDCHAIN",
      "status": "PLAYING",
      "currentRound": 2,
      "users": [
        { "name": "플레이어1", "score": 200 },
        { "name": "플레이어2", "score": 100 }
      ]
    }
  }
}
```

### 2.5 게임방 입장

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!, $username: String!, $password: String) { joinGameRoom(roomId: $roomId, username: $username, password: $password) { id users { name score isReady } } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "username": "플레이어2"
  }
}
```

**Response**
```json
{
  "data": {
    "joinGameRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "users": [
        { "name": "플레이어1", "score": 0, "isReady": true },
        { "name": "플레이어2", "score": 0, "isReady": false }
      ]
    }
  }
}
```

}
```

### 2.6 게임방 퇴장

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!, $username: String!) { leaveGameRoom(roomId: $roomId, username: $username) { id users { name } } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "username": "플레이어2"
  }
}
```

}
```

### 2.7 게임방 설정 수정 (방장 전용)

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!, $input: UpdateGameRoomInput!) { updateGameRoom(roomId: $roomId, input: $input) { id name maxUsers totalRounds roundTimeLimit isPrivate } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "input": {
      "name": "새로운 방 이름",
      "maxUsers": 6,
      "totalRounds": 7,
      "roundTimeLimit": 45
    }
  }
}
```

**Response**
```json
{
  "data": {
    "updateGameRoom": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "새로운 방 이름",
      "maxUsers": 6,
      "totalRounds": 7,
      "roundTimeLimit": 45,
      "isPrivate": false
    }
  }
}
```

### 2.8 게임방 삭제 (방장 전용)

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!) { deleteGameRoom(roomId: $roomId) }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**
```json
{
  "data": {
    "deleteGameRoom": true
  }
}
```

### 2.9 방장 권한 이전

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!, $newHostUsername: String!) { transferHost(roomId: $roomId, newHostUsername: $newHostUsername) { id hostUserId users { name } } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "newHostUsername": "플레이어2"
  }
}
```

**Response**
```json
{
  "data": {
    "transferHost": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "hostUserId": "507f1f77bcf86cd799439022",
      "users": [
        { "name": "플레이어1" },
        { "name": "플레이어2" }
      ]
    }
  }
}
```

### 2.10 준비 상태 변경

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!, $username: String!, $ready: Boolean!) { setReady(roomId: $roomId, username: $username, ready: $ready) { id users { name isReady } } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "username": "플레이어2",
    "ready": true
  }
}
```

}
```

### 2.11 게임방 실시간 업데이트 (Subscription)

**GraphQL Subscription**
```graphql
subscription {
  roomUpdated(roomId: "550e8400-e29b-41d4-a716-446655440000") {
    id
    status
    currentRound
    users { name score isReady }
  }
}
```

**WebSocket Message (서버 → 클라이언트)**
```json
{
  "type": "next",
  "payload": {
    "data": {
      "roomUpdated": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "status": "PLAYING",
        "currentRound": 1,
        "users": [
          { "name": "플레이어1", "score": 0, "isReady": false },
          { "name": "플레이어2", "score": 0, "isReady": false }
        ]
      }
    }
  }
}
```

---

## 3. 게임 플레이 (Game Play)

### 3.1 게임 시작 (방장 전용)

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!) { startGame(roomId: $roomId) { id status currentRound } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**
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

}
```

### 3.2 다음 라운드 시작

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!) { startRound(roomId: $roomId) { id status currentRound } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**
```json
{
  "data": {
    "startRound": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "PLAYING",
      "currentRound": 2
    }
  }
}
```

### 3.3 끝말잇기 시작 단어 조회

**GraphQL Query**
```json
{
  "query": "query { randomWordchainPrompt { word hint } }"
}
```

**Response**
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

### 3.4 끝말잇기 단어 검증

**GraphQL Query**
```json
{
  "query": "query($word: String!, $previousWord: String!) { isValidWord(word: $word, previousWord: $previousWord) { valid reason } }",
  "variables": {
    "word": "ごりら",
    "previousWord": "りんご"
  }
}
```

**Response**
```json
{
  "data": {
    "isValidWord": {
      "valid": true,
      "reason": ""
    }
  }
}
```

### 3.5 OX 퀴즈 조회

**GraphQL Query**
```json
{
  "query": "query($roomId: ID) { randomOXQuiz(roomId: $roomId) { id category difficulty question } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**
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

### 3.6 QA 퀴즈 조회

**GraphQL Query**
```json
{
  "query": "query($roomId: ID) { randomQAQuiz(roomId: $roomId) { id category question options } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**
```json
{
  "data": {
    "randomQAQuiz": {
      "id": "987654321",
      "category": "지리",
      "question": "대한민국의 수도는?",
      "options": ["서울", "부산", "대구", "인천"]
    }
  }
}
```

### 3.7 답안 제출

**GraphQL Mutation (OX 퀴즈)**
```json
{
  "query": "mutation($roomId: ID!, $username: String!, $answer: String!) { submitAnswer(roomId: $roomId, username: $username, answer: $answer) { success isCorrect earnedScore totalScore explanation } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "username": "플레이어1",
    "answer": "true"
  }
}
```

**Response**
```json
{
  "data": {
    "submitAnswer": {
      "success": true,
      "isCorrect": true,
      "earnedScore": 100,
      "totalScore": 100,
      "explanation": "맞습니다! 지구는 태양 주위를 공전합니다."
    }
  }
}
```

**GraphQL Mutation (QA 퀴즈)**
```json
{
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "username": "플레이어1",
    "answer": "0"
  }
}
```

### 3.8 끝말잇기 단어 제출 (WebSocket)

**클라이언트 → 서버**
```json
{
  "type": "wordchain_submit",
  "data": {
    "UserName": "플레이어1",
    "word": "さくら",
    "lastWord": "りんご"
  }
}
```

**서버 → 클라이언트 (성공)**
```json
{
  "type": "message",
  "data": "{\"type\":\"WORDCHAIN_RESULT\",\"payload\":{\"success\":true,\"userName\":\"플레이어1\",\"word\":\"さくら\",\"earnedScore\":100,\"totalScore\":100}}"
}
```

**서버 → 클라이언트 (실패)**
```json
{
  "type": "message",
  "data": "{\"type\":\"WORDCHAIN_RESULT\",\"payload\":{\"success\":false,\"userName\":\"플레이어1\",\"word\":\"さくら\",\"reason\":\"invalid_connection\",\"earnedScore\":0,\"totalScore\":0}}"
}
```

### 3.9 게임 이벤트 구독 (Subscription)

**GraphQL Subscription**
```graphql
subscription {
  gameEvent(roomId: "550e8400-e29b-41d4-a716-446655440000") {
    type
    roomId
    username
    data
    timestamp
  }
}
```

**이벤트 예시**
```json
{
  "type": "next",
  "payload": {
    "data": {
      "gameEvent": {
        "type": "ANSWER_SUBMITTED",
        "roomId": "550e8400-e29b-41d4-a716-446655440000",
        "username": "플레이어2",
        "data": "{\"hasAnswered\":true}",
        "timestamp": "2026-01-19T10:35:00Z"
      }
    }
  }
}
```

### 3.10 게임 종료

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!) { endGame(roomId: $roomId) { id status users { name score } } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Response**
```json
{
  "data": {
    "endGame": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "status": "FINISHED",
      "users": [
        { "name": "플레이어1", "score": 400 },
        { "name": "플레이어2", "score": 300 }
      ]
    }
  }
}
```

---

## 4. 채팅 (Chat)

### 4.1 게임방 채팅 (GraphQL)

**Mutation**
```json
{
  "query": "mutation($roomId: ID!, $username: String!, $message: String!) { sendChat(roomId: $roomId, username: $username, message: $message) { id message timestamp } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "username": "플레이어1",
    "message": "안녕하세요!"
  }
}
```

**Subscription**
```graphql
subscription {
  chatMessage(roomId: "550e8400-e29b-41d4-a716-446655440000") {
    username
    message
    timestamp
  }
}
```

### 4.2 게임방 채팅 (WebSocket)

**클라이언트 → 서버**
```json
{
  "type": "chat",
  "data": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "userId": "507f1f77bcf86cd799439011",
    "userName": "플레이어1",
    "message": "GG!"
  }
}
```

**서버 → 클라이언트**
```json
{
  "type": "message",
  "data": "{\"type\":\"CHAT_MESSAGE\",\"payload\":{\"userId\":\"507f1f77bcf86cd799439011\",\"userName\":\"플레이어1\",\"message\":\"GG!\",\"timestamp\":\"2026-01-19T10:40:00Z\"}}"
}
```

### 4.3 로비 채팅 (WebSocket)

**클라이언트 → 서버**
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

**서버 → 클라이언트**
```json
{
  "type": "message",
  "channel": "lobby",
  "data": "{\"type\":\"LOBBY_CHAT\",\"payload\":{\"userId\":\"507f1f77bcf86cd799439011\",\"userName\":\"플레이어1\",\"message\":\"같이 게임 하실 분?\"}}"
}
```

---

## 5. 초대 (Invitation)

### 5.1 사용자 초대

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!, $inviteeUsername: String!) { inviteUser(roomId: $roomId, inviteeUsername: $inviteeUsername) { id status expiresAt } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "inviteeUsername": "플레이어3"
  }
}
```

**Response**
```json
{
  "data": {
    "inviteUser": {
      "id": "inv-123456",
      "status": "PENDING",
      "expiresAt": "2026-01-20T10:30:00Z"
    }
  }
}
```

### 5.2 여러 사용자 초대

**GraphQL Mutation**
```json
{
  "query": "mutation($roomId: ID!, $inviteeUsernames: [String!]!) { inviteUsers(roomId: $roomId, inviteeUsernames: $inviteeUsernames) { id status expiresAt } }",
  "variables": {
    "roomId": "550e8400-e29b-41d4-a716-446655440000",
    "inviteeUsernames": ["플레이어3", "플레이어4", "플레이어5"]
  }
}
```

**Response**
```json
{
  "data": {
    "inviteUsers": [
      {
        "id": "inv-123456",
        "status": "PENDING",
        "expiresAt": "2026-01-20T10:30:00Z"
      },
      {
        "id": "inv-123457",
        "status": "PENDING",
        "expiresAt": "2026-01-20T10:30:00Z"
      },
      {
        "id": "inv-123458",
        "status": "PENDING",
        "expiresAt": "2026-01-20T10:30:00Z"
      }
    ]
  }
}
```

### 5.3 내 초대 목록 조회

**GraphQL Query**
```json
{
  "query": "query($username: String!, $status: InvitationStatus) { myInvitations(username: $username, status: $status) { id roomName inviterName status expiresAt } }",
  "variables": {
    "username": "플레이어3",
    "status": "PENDING"
  }
}
```

**Response**
```json
{
  "data": {
    "myInvitations": [
      {
        "id": "inv-123456",
        "roomName": "즐거운 끝말잇기",
        "inviterName": "플레이어1",
        "status": "PENDING",
        "expiresAt": "2026-01-20T10:30:00Z"
      }
    ]
  }
}
```

### 5.4 초대 상세 조회

**GraphQL Query**
```json
{
  "query": "query($invitationId: ID!) { invitation(id: $invitationId) { id roomId roomName inviterName inviteeName status expiresAt } }",
  "variables": {
    "invitationId": "inv-123456"
  }
}
```

**Response**
```json
{
  "data": {
    "invitation": {
      "id": "inv-123456",
      "roomId": "550e8400-e29b-41d4-a716-446655440000",
      "roomName": "즐거운 끝말잇기",
      "inviterName": "플레이어1",
      "inviteeName": "플레이어3",
      "status": "PENDING",
      "expiresAt": "2026-01-20T10:30:00Z"
    }
  }
}
```

### 5.5 초대 수락

**GraphQL Mutation**
```json
{
  "query": "mutation($invitationId: ID!) { acceptInvite(invitationId: $invitationId) { id name users { name } } }",
  "variables": {
    "invitationId": "inv-123456"
  }
}
```

**Response**
```json
{
  "data": {
    "acceptInvite": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "즐거운 끝말잇기",
      "users": [
        { "name": "플레이어1" },
        { "name": "플레이어3" }
      ]
    }
  }
}
```

### 5.6 초대 거절

**GraphQL Mutation**
```json
{
  "query": "mutation($invitationId: ID!) { rejectInvite(invitationId: $invitationId) }",
  "variables": {
    "invitationId": "inv-123456"
  }
}
```

**Response**
```json
{
  "data": {
    "rejectInvite": true
  }
}
```

---

## 6. WebSocket 연결 및 메시지

### 6.1 로비 WebSocket 연결

**연결**: `ws://localhost:8080/ws/lobby`

#### 사용자 식별
**클라이언트 → 서버**
```json
{
  "type": "identify",
  "data": {
    "UserName": "플레이어1",
    "roomId": "lobby"
  }
}
```

#### 채널 구독
**클라이언트 → 서버**
```json
{
  "type": "subscribe",
  "channel": "lobby"
}
```



#### 채팅 히스토리
**서버 → 클라이언트**
```json
{
  "type": "LOBBY_CHAT_HISTORY",
  "messages": [
    {
      "userId": "user-1",
      "userName": "플레이어1",
      "message": "안녕하세요!",
      "timestamp": "2026-01-19T10:30:00Z"
    }
  ]
}
```

### 6.2 게임방 WebSocket 연결

**연결**: `ws://localhost:8080/ws/rooms/:roomId`

#### 게임 액션
**클라이언트 → 서버**
```json
{
  "type": "game_action",
  "data": {
    "action": "ready",
    "userId": "507f1f77bcf86cd799439011"
  }
}
```

**서버 → 클라이언트**
```json
{
  "type": "message",
  "data": "{\"type\":\"GAME_ACTION\",\"payload\":{\"action\":\"ready\",\"userId\":\"507f1f77bcf86cd799439011\",\"userName\":\"플레이어1\"}}"
}
```

#### 그림 그리기 (향후 기능)
**클라이언트 → 서버**
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

**서버 → 클라이언트**
```json
{
  "type": "message",
  "data": "{\"type\":\"DRAW_EVENT\",\"payload\":{\"x\":150,\"y\":200,\"color\":\"#FF0000\",\"brushSize\":5,\"action\":\"draw\",\"userId\":\"507f1f77bcf86cd799439011\"}}"
}
```

---

## 7. 오류 처리

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

### HTTP 상태 코드

| 코드 | 설명 |
|------|------|
| 200 | 성공 |
| 400 | 잘못된 요청 |
| 401 | 인증 실패 |
| 404 | 리소스 없음 |
| 500 | 서버 오류 |

---

## 8. 열거형 타입

### GameType
- `OX` - OX 퀴즈
- `QA` - 4지선다 퀴즈
- `WORDCHAIN` - 끝말잇기
- `DRAWING` - 그림 맞추기

### GameStatus
- `WAITING` - 대기 중
- `PLAYING` - 게임 진행 중
- `FINISHED` - 게임 종료

### InviteStatus
- `PENDING` - 대기 중
- `ACCEPTED` - 수락됨
- `REJECTED` - 거절됨
- `EXPIRED` - 만료됨

---

## 부록

**개발 도구**: http://localhost:8080/graphql (GraphQL Playground)
**Health Check**: GET http://localhost:8080/health
