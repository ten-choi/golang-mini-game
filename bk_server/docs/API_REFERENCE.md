# API Reference

서버의 모든 API 엔드포인트와 WebSocket 채널을 기능별로 정리한 문서입니다.

## 목차
- [기본 정보](#기본-정보)
- [User (사용자)](#user-사용자)
- [Game Room (게임방)](#game-room-게임방)
- [Quiz (퀴즈)](#quiz-퀴즈)
- [Invitation (초대)](#invitation-초대)
- [User Stats (플레이어 통계)](#user-stats-플레이어-통계)
- [Game Flow (게임 진행)](#game-flow-게임-진행)
- [Chat (채팅)](#chat-채팅)
- [WebSocket (실시간 통신)](#websocket-실시간-통신)

---

## 기본 정보

### Base URL
```
HTTP/GraphQL: http://localhost:8080/graphql
WebSocket: ws://localhost:8080/ws
```

### 인증
현재 인증은 구현되지 않음 (추후 JWT 추가 예정)

---

## User (사용자)

사용자 계정 생성, 조회, 수정, 삭제 기능

### GraphQL Queries

#### `userByHangeId` - HangeId로 사용자 조회 (로그인)

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `name` | String | ✅ | 조회할 사용자의 HangeId (로그인용 식별자) |

**Response:** `User` 객체 또는 `null`
| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | ID | MongoDB ObjectID |
| `hangeId` | String | 로그인용 식별자 |
| `name` | String | 고유 사용자명 (3-20자) |
| `avatarUrl` | String | 프로필 이미지 URL |
| `level` | Int | 플레이어 레벨 |
| `credit` | Int | 보유 크레딧 (게임 재화) |
| `createdAt` | Time | 계정 생성 시각 (RFC3339) |
| `updatedAt` | Time | 마지막 업데이트 시각 (RFC3339) |

---

#### `userByName` - 사용자명으로 조회

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `name` | String | ✅ | 조회할 사용자명 |

**Response:** `User` 객체 (필드 구조는 위와 동일)

---

#### `users` - 사용자 목록 조회 (페이징, 검색)

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `limit` | Int | ❌ | 반환할 최대 개수 (기본값: 50, 최대: 100) |
| `offset` | Int | ❌ | 건너뛸 개수 (페이징용, 기본값: 0) |
| `search` | String | ❌ | 사용자명 검색어 (부분 일치, 대소문자 무시) |
| `minLevel` | Int | ❌ | 최소 레벨 필터 |

**Response:** `[User!]!` 배열 (각 User 필드는 위와 동일)

---

### GraphQL Mutations

#### `createUser` - 새 사용자 생성

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `input.hangeId` | String | ✅ | 로그인용 식별자 (3-20자, 영문/숫자/언더스코어) |
| `input.name` | String | ❌ | 고유 사용자명 (미제공 시 hangeId 사용) |
| `input.avatarUrl` | String | ❌ | 프로필 이미지 URL |

**Response:** `User!` 객체
| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | ID | 생성된 사용자 ID |
| `hangeId` | String | 로그인용 식별자 |
| `name` | String | 사용자명 |
| `avatarUrl` | String | 프로필 이미지 URL |
| `level` | Int | 초기 레벨 (0) |
| `credit` | Int | 초기 크레딧 (0) |
| `createdAt` | Time | 생성 시각 |
| `updatedAt` | Time | 업데이트 시각 |

---

#### `updateUser` - 사용자 정보 수정

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `name` | String | ✅ | 수정할 사용자명 (변경 불가, 식별용) |
| `input.avatarUrl` | String | ❌ | 새 프로필 이미지 URL |
| `input.level` | Int | ❌ | 새 레벨 |
| `input.credit` | Int | ❌ | 새 크레딧 |

**Response:** `User!` 객체 (수정된 정보 포함)

---

#### `deleteUser` - 사용자 삭제

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `name` | String | ✅ | 삭제할 사용자명 |

**Response:** `Boolean!`
- `true`: 삭제 성공
- `false`: 삭제 실패

### 예시

```graphql
# 사용자 생성
mutation {
  createUser(input: {
    hangeId: "user123"
    name: "플레이어123"
    avatarUrl: "https://example.com/avatar.jpg"
  }) {
    id
    name
    level
    credit
  }
}

# 사용자 조회
query {
  userByName(name: "플레이어123") {
    id
    name
    level
    credit
    createdAt
  }
}

# 사용자 목록 (검색)
query {
  users(search: "플레이어", limit: 10) {
    name
    level
  }
}
```

---

## Game Room (게임방)

게임방 생성, 입장, 퇴장, 관리 기능

### GraphQL Queries

#### `gameRoom` - 특정 게임방 조회

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `id` | ID | ✅ | 조회할 게임방 UUID |

**Response:** `GameRoom` 객체 또는 `null`
| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | ID | 게임방 UUID |
| `name` | String | 게임방 이름 |
| `gameType` | GameType | 게임 타입 (OX, QA, WORDCHAIN, DRAWING) |
| `status` | GameStatus | 게임 상태 (WAITING, PLAYING, FINISHED) |
| `currentRound` | Int | 현재 라운드 (0부터 시작) |
| `totalRounds` | Int | 총 라운드 수 |
| `roundTimeLimit` | Int | 라운드 제한 시간 (초) |
| `users` | [User!]! | 플레이어 목록 |
| `users[].name` | String | 플레이어 사용자명 |
| `users[].score` | Int | 플레이어 점수 |
| `users[].isReady` | Boolean | 준비 상태 |
| `maxUsers` | Int | 최대 플레이어 수 (2-8) |
| `hostUsername` | String | 방장 사용자명 |
| `isPrivate` | Boolean | 비공개 방 여부 |
| `password` | String | 비밀번호 (비공개 방인 경우) |
| `createdAt` | Time | 생성 시각 |
| `wordchainLastWord` | String | 끝말잇기 마지막 단어 (WORDCHAIN 게임) |
| `currentTurnUsername` | String | 현재 턴 플레이어 (WORDCHAIN 게임) |

---

#### `gameRooms` - 게임방 목록 조회 (필터링)

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `gameType` | GameType | ❌ | 게임 타입 필터 (OX, QA, WORDCHAIN, DRAWING) |
| `status` | GameStatus | ❌ | 게임 상태 필터 (WAITING, PLAYING) |
| `includePrivate` | Boolean | ❌ | 비공개 방 포함 여부 (기본값: false) |
| `hasSpace` | Boolean | ❌ | 입장 가능한 방만 (기본값: true) |
| `limit` | Int | ❌ | 반환 최대 개수 (기본값: 50, 최대: 100) |
| `sortBy` | String | ❌ | 정렬 기준 (CREATED_DESC, CREATED_ASC, PLAYERS_DESC) |

**Response:** `[GameRoom!]!` 배열 (각 GameRoom 필드는 위와 동일)

---

#### `myCurrentRoom` - 사용자의 현재 게임방

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `username` | String | ✅ | 조회할 사용자명 |

**Response:** `GameRoom` 객체 또는 `null` (게임방에 없는 경우)

---

### GraphQL Mutations

#### `createGameRoom` - 게임방 생성

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `input.name` | String | ✅ | 게임방 이름 (1-50자) |
| `input.gameType` | GameType | ✅ | 게임 타입 (OX, QA, WORDCHAIN, DRAWING) |
| `input.totalRounds` | Int | ✅ | 총 라운드 수 (1-20) |
| `input.maxUsers` | Int | ✅ | 최대 플레이어 수 (2-8) |
| `input.hostUsername` | String | ✅ | 방장 사용자명 (생성자) |
| `input.isPrivate` | Boolean | ❌ | 비공개 방 여부 (기본값: false) |
| `input.password` | String | ❌ | 비밀번호 (비공개 방인 경우) |
| `input.roundTimeLimit` | Int | ❌ | 라운드 제한 시간 (5-300초, 기본값: 30) |

**Response:** `GameRoom!` 객체 (생성된 방 정보, 방장이 첫 플레이어로 자동 추가됨)

---

#### `updateGameRoom` - 게임방 설정 변경 (방장만)

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 수정할 게임방 ID |
| `input.name` | String | ❌ | 새 게임방 이름 |
| `input.maxUsers` | Int | ❌ | 새 최대 플레이어 수 (현재 인원 이상) |
| `input.totalRounds` | Int | ❌ | 새 총 라운드 수 |
| `input.roundTimeLimit` | Int | ❌ | 새 라운드 제한 시간 |

**Response:** `GameRoom!` 객체 (수정된 방 정보)

---

#### `joinGameRoom` - 게임방 입장

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 입장할 게임방 ID |
| `username` | String | ✅ | 입장하는 사용자명 |
| `password` | String | ❌ | 비밀번호 (비공개 방인 경우 필수) |

**Response:** `GameRoom!` 객체 (업데이트된 플레이어 목록 포함)

**에러:**
- 방이 가득 참 (maxUsers 도달)
- 잘못된 비밀번호
- 게임 진행 중 (status: PLAYING)

---

#### `leaveGameRoom` - 게임방 퇴장

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 퇴장할 게임방 ID |
| `username` | String | ✅ | 퇴장하는 사용자명 |

**Response:** `GameRoom!` 객체 (업데이트된 방 정보)

**자동 처리:**
- 방장 퇴장 시 다음 플레이어에게 자동 양도
- 마지막 플레이어 퇴장 시 방 자동 삭제

---

#### `deleteGameRoom` - 게임방 삭제 (방장만)

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 삭제할 게임방 ID |

**Response:** `Boolean!`
- `true`: 삭제 성공
- `false`: 삭제 실패

---

#### `transferHost` - 방장 권한 양도

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 게임방 ID |
| `newHostUsername` | String | ✅ | 새 방장의 사용자명 |

**Response:** `GameRoom!` 객체 (hostUsername 변경됨)

---

#### `setReady` - 준비 상태 변경

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 게임방 ID |
| `username` | String | ✅ | 플레이어 사용자명 |
| `ready` | Boolean | ✅ | 준비 상태 (true: 준비 완료, false: 취소) |

**Response:** `GameRoom!` 객체 (업데이트된 플레이어 준비 상태)

### WebSocket

| 엔드포인트 | 설명 | 연결 방법 |
|-----------|------|-----------|
| `/ws/rooms/:id` | 게임방 전용 WebSocket | `ws://localhost:8080/ws/rooms/{roomId}` |

### 예시

```graphql
# 게임방 생성
mutation {
  createGameRoom(input: {
    name: "초보자 환영"
    gameType: WORDCHAIN
    totalRounds: 5
    maxUsers: 4
    hostUsername: "플레이어123"
    isPrivate: false
  }) {
    id
    name
    status
  }
}

# 게임방 목록 조회
query {
  gameRooms(gameType: WORDCHAIN, status: WAITING, hasSpace: true) {
    id
    name
    users { name score }
    maxUsers
  }
}

# 게임방 입장
mutation {
  joinGameRoom(
    roomId: "550e8400-e29b-41d4-a716-446655440000"
    username: "플레이어456"
  ) {
    id
    users { name }
  }
}
```

---

## Quiz (퀴즈)

OX 퀴즈, 4지선다 퀴즈, 끝말잇기 단어 관련 기능

### GraphQL Queries

#### `randomOXQuiz` - 랜덤 OX 퀴즈 조회

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ❌ | 게임방 ID (제공 시 이미 사용된 퀴즈 제외) |

**Response:** `OXQuiz` 객체 또는 `null`
| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | ID | 퀴즈 ID |
| `category` | String | 카테고리 (예: "일반상식", "과학") |
| `difficulty` | Int | 난이도 (1-5) |
| `question` | String | 문제 텍스트 |
| `answer` | Boolean | 정답 (true/false) ⚠️ 클라이언트에 노출 주의 |
| `imageUrl` | String | 문제 이미지 URL (선택사항) |
| `explanation` | String | 정답 해설 |

---

#### `randomQAQuiz` - 랜덤 4지선다 퀴즈 조회

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ❌ | 게임방 ID (제공 시 이미 사용된 퀴즈 제외) |

**Response:** `GeneralQuiz` 객체 또는 `null`
| 필드 | 타입 | 설명 |
|------|------|------|
| `id` | ID | 퀴즈 ID |
| `category` | String | 카테고리 |
| `difficulty` | Int | 난이도 (1-5) |
| `question` | String | 문제 텍스트 |
| `options` | [String!]! | 선택지 배열 (4개) |
| `answer` | Int | 정답 인덱스 (0-3) ⚠️ 클라이언트에 노출 주의 |
| `imageUrl` | String | 문제 이미지 URL (선택사항) |
| `explanation` | String | 정답 해설 |

---

#### `randomWordchainPrompt` - 끝말잇기 시작 단어

**Request:** 없음

**Response:** `WordchainPrompt!` 객체
| 필드 | 타입 | 설명 |
|------|------|------|
| `word` | String | 일본어 시작 단어 (히라가나/카타카나) |
| `hint` | String | 단어 뜻/힌트 |

---

#### `isValidWord` - 단어 유효성 검증

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `word` | String | ✅ | 검증할 일본어 단어 (히라가나/카타카나) |

**Response:** `Boolean!`
- `true`: 유효한 단어 (사전 등록됨)
- `false`: 무효한 단어

**검증 규칙:**
- 2자 이상
- 히라가나 또는 카타카나만
- ん으로 끝나지 않음
- 사전에 등록된 단어

### 예시

```graphql
# OX 퀴즈 조회
query {
  randomOXQuiz(roomId: "room-123") {
    id
    category
    difficulty
    question
  }
}

# 끝말잇기 시작 단어
query {
  randomWordchainPrompt {
    word
    hint
  }
}

# 단어 검증
query {
  isValidWord(word: "りんご")
}
```

---

## Invitation (초대)

게임방 초대 및 수락/거절 기능

### GraphQL Queries

| 쿼리 | 설명 | 파라미터 | 반환 |
|------|------|---------|------|
| `myInvitations` | 받은 초대 목록 | `userId: ID!`, `status: InviteStatus?` | `[Invitation!]!` |
| `invitation` | 특정 초대 조회 | `id: ID!` | `Invitation` |

### GraphQL Mutations

| 뮤테이션 | 설명 | 파라미터 | 반환 |
|---------|------|---------|------|
| `inviteUser` | 사용자 초대 | `roomId: ID!`, `inviteeUsername: String!` | `Invitation!` |
| `inviteUsers` | 여러 사용자 초대 | `roomId: ID!`, `inviteeUsernames: [String!]!` | `[Invitation!]!` |
| `acceptInvite` | 초대 수락 | `invitationId: ID!` | `GameRoom!` |
| `rejectInvite` | 초대 거절 | `invitationId: ID!` | `Boolean!` |

### 예시

```graphql
# 사용자 초대
mutation {
  inviteUser(
    roomId: "room-123"
    inviteeUsername: "친구1"
  ) {
    id
    status
    expiresAt
  }
}

# 초대 수락
mutation {
  acceptInvite(invitationId: "inv-456") {
    id
    name
    users { name }
  }
}
```

---

## User Stats (플레이어 통계)

플레이어 게임 통계 및 리더보드

### GraphQL Queries

| 쿼리 | 설명 | 파라미터 | 반환 |
|------|------|---------|------|
| `userStats` | 플레이어 통계 조회 | `username: String!`, `gameType: String?` | `[UserStats!]!` |
| `leaderboard` | 리더보드 상위 랭커 | `gameType: String!`, `limit: Int?` | `[UserStats!]!` |
| `gameConfig` | 게임 설정 정보 | - | `GameConfig!` |

### 예시

```graphql
# 플레이어 통계
query {
  userStats(username: "플레이어123", gameType: "OX") {
    gameType
    totalGames
    totalWins
    totalScore
  }
}

# 리더보드
query {
  leaderboard(gameType: "WORDCHAIN", limit: 10) {
    username
    totalScore
    totalWins
  }
}

# 게임 설정
query {
  gameConfig {
    maxUsers
    roundDuration
    roundsPerGame
  }
}
```

---

## Game Flow (게임 진행)

게임 시작, 라운드 진행, 답변 제출

### GraphQL Mutations

#### `startGame` - 게임 시작 (방장만)

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 게임방 ID |

**Response:** `GameRoom!` 객체
| 필드 | 설명 |
|------|------|
| `status` | PLAYING으로 변경됨 |
| `currentRound` | 1로 설정됨 |
| `users[].isReady` | 모두 false로 초기화 |

**제약조건:**
- 방장만 실행 가능
- 최소 2명 이상 필요
- 모든 플레이어(방장 제외) 준비 완료 필요

---

#### `startRound` - 다음 라운드 시작

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 게임방 ID |

**Response:** `GameRoom!` 객체
| 필드 | 설명 |
|------|------|
| `currentRound` | 1 증가 |
| 자동 처리 | 새 퀴즈 제시, 이전 답변 초기화 |

**주의:** 일반적으로 서버가 자동 호출, 클라이언트 직접 호출 불필요

---

#### `submitAnswer` - 답변 제출

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 게임방 ID |
| `username` | String | ✅ | 답변 제출자 사용자명 |
| `answer` | String | ✅ | 답변 값 (게임 타입에 따라 다름) |

**답변 형식:**
- **OX 게임:** `"true"` 또는 `"false"`
- **QA 게임:** `"0"`, `"1"`, `"2"`, `"3"` (선택지 인덱스)
- **WORDCHAIN 게임:** 일본어 단어 (예: `"りんご"`)

**Response:** `AnswerResult!` 객체
| 필드 | 타입 | 설명 |
|------|------|------|
| `success` | Boolean | 제출 성공 여부 |
| `isCorrect` | Boolean | 정답 여부 |
| `earnedScore` | Int | 획득 점수 (정답 시 100점) |
| `totalScore` | Int | 플레이어 총점 |
| `explanation` | String | 정답 해설 (OX/QA 게임) |

---

#### `endGame` - 게임 종료 (방장만)

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 게임방 ID |

**Response:** `GameRoom!` 객체
| 필드 | 설명 |
|------|------|
| `status` | FINISHED로 변경 후 곧 WAITING으로 전환 |
| `users[].score` | 최종 점수 포함 |

---

### GraphQL Subscriptions

#### `roomUpdated` - 게임방 상태 변경

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 구독할 게임방 ID |

**Response:** `GameRoom!` 객체 (변경 시마다)

**이벤트 발생 시점:**
- 플레이어 입장/퇴장
- 게임 시작/종료
- 라운드 변경
- 점수 업데이트

---

#### `gameEvent` - 게임 이벤트

**Request:**
| 필드 | 타입 | 필수 | 설명 |
|------|------|------|------|
| `roomId` | ID | ✅ | 구독할 게임방 ID |

**Response:** `GameEvent!` 객체
| 필드 | 타입 | 설명 |
|------|------|------|
| `type` | String | 이벤트 타입 (user_joined, round_started 등) |
| `roomId` | ID | 게임방 ID |
| `username` | String | 관련 플레이어 사용자명 |
| `data` | JSON | 이벤트 추가 데이터 |
| `timestamp` | Time | 이벤트 발생 시각 |

**이벤트 타입:**
- `user_joined`: 플레이어 입장
- `user_left`: 플레이어 퇴장
- `game_started`: 게임 시작
- `round_started`: 라운드 시작
- `round_ended`: 라운드 종료
- `game_ended`: 게임 종료

### 예시

```graphql
# 게임 시작
mutation {
  startGame(roomId: "room-123") {
    id
    status
    currentRound
  }
}

# 답변 제출 (OX)
mutation {
  submitAnswer(
    roomId: "room-123"
    username: "플레이어123"
    answer: "true"
  ) {
    success
    isCorrect
    earnedScore
    totalScore
  }
}

# 게임방 구독
subscription {
  roomUpdated(roomId: "room-123") {
    id
    status
    currentRound
    users { name score }
  }
}
```

---

## Chat (채팅)

게임방 및 로비 채팅 기능

### GraphQL Mutations

| 뮤테이션 | 설명 | 파라미터 | 반환 |
|---------|------|---------|------|
| `sendChat` | 채팅 메시지 전송 | `roomId: ID!`, `username: String!`, `message: String!` | `ChatMessage!` |

### GraphQL Subscriptions

| 구독 | 설명 | 파라미터 | 이벤트 |
|------|------|---------|--------|
| `chatMessage` | 채팅 메시지 구독 | `roomId: ID!` | `ChatMessage!` |

### WebSocket (로비 채팅)

| 메시지 타입 | 설명 | 데이터 |
|------------|------|--------|
| `lobby_chat` | 로비 채팅 전송 | `{ userId, userName, message }` |
| `LOBBY_CHAT` | 로비 채팅 수신 | `{ userId, userName, message }` |
| `LOBBY_CHAT_HISTORY` | 채팅 히스토리 | `{ messages: [...] }` |

### 예시

```graphql
# 채팅 전송
mutation {
  sendChat(
    roomId: "room-123"
    username: "플레이어123"
    message: "안녕하세요!"
  ) {
    id
    message
    timestamp
  }
}

# 채팅 구독
subscription {
  chatMessage(roomId: "room-123") {
    username
    message
    timestamp
  }
}
```

**WebSocket 로비 채팅:**
```javascript
// 연결
const ws = new WebSocket('ws://localhost:8080/ws/lobby');

// 채팅 전송
ws.send(JSON.stringify({
  type: "lobby_chat",
  data: {
    userId: "user-123",
    userName: "플레이어123",
    message: "안녕하세요!"
  }
}));

// 채팅 수신
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === "message") {
    const payload = JSON.parse(data.data);
    if (payload.type === "LOBBY_CHAT") {
      console.log(payload.payload.message);
    }
  }
};
```

---

## WebSocket (실시간 통신)

모든 실시간 기능을 위한 WebSocket 연결

### 엔드포인트

| 엔드포인트 | 설명 | 사용 시점 |
|-----------|------|-----------|
| `/ws/lobby` | 로비 전용 WebSocket | 로비 화면, 게임방 목록, 로비 채팅 |
| `/ws/rooms/:id` | 게임방 전용 WebSocket | 게임방 입장 후, 게임 플레이 |

### 로비 WebSocket (`/ws/lobby`)

**연결:**
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/lobby');
```

**지원 메시지 타입:**

| 타입 | 방향 | 설명 | 데이터 |
|------|------|------|--------|
| `subscribe` | 클라이언트 → 서버 | 채널 구독 | `{ type: "subscribe", channel: "lobby" }` |
| `unsubscribe` | 클라이언트 → 서버 | 구독 취소 | `{ type: "unsubscribe", channel: "lobby" }` |
| `lobby_chat` | 클라이언트 → 서버 | 로비 채팅 전송 | `{ type: "lobby_chat", data: {...} }` |
| `identify` | 클라이언트 → 서버 | 사용자 식별 | `{ type: "identify", data: { UserName, roomId } }` |
| `message` | 서버 → 클라이언트 | 일반 메시지 수신 | `{ type: "message", channel, data }` |
| `LOBBY_CHAT_HISTORY` | 서버 → 클라이언트 | 채팅 히스토리 | `{ type: "LOBBY_CHAT_HISTORY", messages: [...] }` |

**예시:**

```javascript
// 연결
const ws = new WebSocket('ws://localhost:8080/ws/lobby');

ws.onopen = () => {
  // 사용자 식별
  ws.send(JSON.stringify({
    type: "identify",
    data: {
      UserName: "플레이어123",
      roomId: "lobby"
    }
  }));
  
  // 로비 채널 구독 (자동으로 되지만 명시적으로도 가능)
  ws.send(JSON.stringify({
    type: "subscribe",
    channel: "lobby"
  }));
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  
  // 로비 업데이트 (게임방 목록)
  if (message.channel === "lobby" && message.type === "message") {
    const payload = JSON.parse(message.data);
    if (payload.type === "lobby_update") {
      console.log("게임방 목록:", payload.data);
    }
    if (payload.type === "LOBBY_CHAT") {
      console.log("채팅:", payload.payload.message);
    }
  }
  
  // 채팅 히스토리
  if (message.type === "LOBBY_CHAT_HISTORY") {
    console.log("히스토리:", message.messages);
  }
};

// 로비 채팅 전송
ws.send(JSON.stringify({
  type: "lobby_chat",
  data: {
    userId: "user-123",
    userName: "플레이어123",
    message: "안녕하세요!"
  }
}));
```

### 게임방 WebSocket (`/ws/rooms/:id`)

**연결:**
```javascript
const roomId = "550e8400-e29b-41d4-a716-446655440000";
const ws = new WebSocket(`ws://localhost:8080/ws/rooms/${roomId}`);
```

**지원 메시지 타입:**

| 타입 | 방향 | 설명 | 데이터 |
|------|------|------|--------|
| `chat` | 양방향 | 게임방 채팅 | `{ type: "chat", data: { roomId, userId, userName, message } }` |
| `drawing` | 클라이언트 → 서버 | 그림 그리기 데이터 | `{ type: "drawing", data: {...} }` |
| `game_action` | 클라이언트 → 서버 | 게임 액션 | `{ type: "game_action", data: {...} }` |
| `wordchain_submit` | 클라이언트 → 서버 | 끝말잇기 단어 제출 | `{ type: "wordchain_submit", data: { UserName, word, lastWord } }` |
| `DRAW_EVENT` | 서버 → 클라이언트 | 그림 이벤트 | 브로드캐스트 |
| `GAME_ACTION` | 서버 → 클라이언트 | 게임 액션 | 브로드캐스트 |
| `CHAT_MESSAGE` | 서버 → 클라이언트 | 채팅 메시지 | 브로드캐스트 |

**예시:**

```javascript
const roomId = "550e8400-e29b-41d4-a716-446655440000";
const ws = new WebSocket(`ws://localhost:8080/ws/rooms/${roomId}`);

ws.onopen = () => {
  console.log(`방 ${roomId}에 연결됨`);
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  const payload = JSON.parse(message.data);
  
  // 채팅 메시지
  if (payload.type === "CHAT_MESSAGE") {
    console.log(`${payload.payload.userName}: ${payload.payload.message}`);
  }
  
  // 게임 이벤트
  if (payload.type === "GAME_ACTION") {
    console.log("게임 액션:", payload.payload);
  }
  
  // 그림 그리기
  if (payload.type === "DRAW_EVENT") {
    console.log("그림 데이터:", payload.payload);
  }
};

// 채팅 전송
ws.send(JSON.stringify({
  type: "chat",
  data: {
    roomId: roomId,
    userId: "user-123",
    userName: "플레이어123",
    message: "GG!"
  }
}));

// 끝말잇기 단어 제출
ws.send(JSON.stringify({
  type: "wordchain_submit",
  data: {
    UserName: "플레이어123",
    word: "りんご",
    lastWord: "ごりら"
  }
}));

// 게임 액션
ws.send(JSON.stringify({
  type: "game_action",
  data: {
    action: "ready",
    userId: "user-123"
  }
}));
```

### Valkey Pub/Sub 채널

백엔드 내부적으로 Valkey Pub/Sub을 사용하여 여러 서버 인스턴스 간 메시지 전달

| 채널 | 설명 | 메시지 타입 |
|------|------|-------------|
| `lobby` | 로비 업데이트, 로비 채팅 | `lobby_update`, `LOBBY_CHAT` |
| `game/{roomId}` | 게임방 이벤트 | `CHAT_MESSAGE`, `DRAW_EVENT`, `GAME_ACTION` |

---

## 에러 코드

### HTTP 상태 코드

| 코드 | 설명 |
|------|------|
| 200 | 성공 |
| 400 | 잘못된 요청 (파라미터 오류) |
| 401 | 인증 필요 |
| 403 | 권한 없음 |
| 404 | 리소스 없음 |
| 500 | 서버 오류 |

### GraphQL 에러

GraphQL은 항상 HTTP 200을 반환하며, 에러는 `errors` 배열에 포함됩니다.

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

## 데이터 타입

### GameType (게임 타입)
- `OX` - OX 퀴즈
- `QA` - 4지선다 퀴즈
- `WORDCHAIN` - 끝말잇기
- `DRAWING` - 그림 맞추기

### GameStatus (게임 상태)
- `WAITING` - 대기 중
- `PLAYING` - 게임 진행 중
- `FINISHED` - 게임 종료

### InviteStatus (초대 상태)
- `PENDING` - 대기 중
- `ACCEPTED` - 수락됨
- `REJECTED` - 거절됨
- `EXPIRED` - 만료됨

---

## 개발 도구

### Apollo Sandbox
```
http://localhost:8080/graphql
```
브라우저에서 접속하여 GraphQL Playground 사용 가능

### Health Check
```
GET http://localhost:8080/health
```
서버 상태 확인

---

## 버전 정보

- **API Version**: 1.0.0
- **Server**: Golang 1.21+
- **GraphQL**: gqlgen
- **WebSocket**: gorilla/websocket
- **Database**: MongoDB
- **Cache**: Valkey (Redis fork)

---

## 추가 정보

- [PRD.md](../PRD.md) - 프로젝트 요구사항
- [README.md](../README.md) - 프로젝트 개요
- [SCHEMA_IMPROVEMENTS.md](./SCHEMA_IMPROVEMENTS.md) - 스키마 개선 사항
- [FUTURE_FEATURES.md](./FUTURE_FEATURES.md) - 향후 기능
