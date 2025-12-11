
Draw and Guess Game - Backend Server

## 필요한 환경

- Go 1.24 이상
- Valkey (Redis 호환)

## 주요 기능

- 멀티플레이어 그림 그리기 게임
- WebSocket 기반 실시간 통신
- Valkey를 사용한 게임 상태 관리
- Swagger API 문서 제공

## 세팅

### 1. 의존성 설치

```powershell
go mod download
```

### 2. Valkey 실행

Docker를 사용하여 Valkey를 실행하거나 로컬에 설치:

```powershell
docker run -d --name valkey -p 6379:6379 valkey/valkey:7.2-alpine
```

### 3. 환경 변수 설정 (.env)

```env
MONGO_URI=mongodb://localhost:27017
DATABASE_NAME=draw_and_guess_db
VALKEY_ADDR=localhost:6379
SERVER_PORT=8080
```

### 4. 서버 시작

```powershell
go run main.go
```

서버는 `http://localhost:8080`에서 실행됩니다.

### 5. Swagger UI 접속

API 문서: `http://localhost:8080/swagger/index.html`

## API 엔드포인트

### Game Rooms
- `GET /app/game/rooms` - 방 목록 조회
  - Query: `id` (optional) - 특정 방 조회
- `POST /app/game/room` - 방 생성
  - Form: `ldap_user` (required)
- `POST /app/game/room/:id/join` - 방 참가
  - Form: `ldap_user` (required)
- `POST /app/game/room/:id/leave` - 방 나가기
  - Form: `ldap_user` (required)
- `POST /app/game/room/:id/start` - 게임 시작
- `POST /app/game/room/:id/answer` - 정답 제출
  - Form: `ldap_user`, `answer` (required)
- `POST /app/game/room/:id/chat` - 채팅 메시지 전송
  - Form: `ldap_user`, `message` (required)
- `PATCH /app/game/room/:id` - 그림 데이터 업데이트
  - Body: drawing data (JSON)
- `DELETE /app/game/room/:id` - 방 삭제

### WebSocket
- `GET /app/ws` - WebSocket 연결 (실시간 통신)

## 데이터 구조

### GameRoom (Valkey에 저장)

```json
{
  "uuid": "string",
  "is_active": true,
  "drawer_user": "string",
  "players": [
    {
      "username": "string",
      "score": 0,
      "attempts": 0
    }
  ],
  "current_word": "string",
  "round_number": 1,
  "time_left": 60,
  "game_status": "waiting|playing|finished",
  "used_words": ["string"],
  "max_rounds": 3,
  "winning_score": 3,
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

## WebSocket

- 엔드포인트: `ws://localhost:8080/app/ws`
- 메시지 형식: `{ "type": "subscribe|unsubscribe|message", "channel": "chat/<roomId>", "data": {...} }`
- 채널 예시
  - `draw/{roomId}`: 그림 스트로크 데이터
  - `chat/{roomId}`: 채팅/정답 피드백
  - `game/{roomId}`: 타이머 및 게임 상태
- 클라이언트는 `subscribe` 메시지를 보내 채널을 구독하고, 서버는 Valkey Pub/Sub을 통해 수신한 내용을 실시간으로 중계합니다.
