# Mini Game Server

실시간 멀티플레이어 게임 서버 (GraphQL + WebSocket)

## 서버 스펙

- **Language**: Go 1.24+
- **GraphQL**: gqlgen (schema-first)
- **Database**: MongoDB 7.0
- **Cache/PubSub**: Valkey 7.2 (Redis 호환)
- **WebSocket**: gorilla/websocket
- **Framework**: Gin

### 지원 게임 모드
- **WORDCHAIN**: 일본어 끝말잇기 (174,709 단어)
- **OX**: O/X 퀴즈 (참/거짓)
- **QA**: 4지선다 퀴즈

### API 엔드포인트
- `POST /graphql` - GraphQL Query/Mutation
- `GET /graphql` - GraphQL Playground
- `GET /health` - 헬스 체크

## 서버 실행

### 1. 의존성 설치
```bash
go mod download
```

### 2. 데이터베이스 배포 (Kubernetes)
```bash
cd deployments/k8s
.\deploy-all.ps1
```

### 3. 환경 변수 설정
```env
MONGO_URI=mongodb://admin:password@localhost:30017
MONGO_DB=draw_and_guess_db
VALKEY_ADDR=localhost:30379
SERVER_PORT=8080
GIN_MODE=debug
```

### 4. 서버 실행
```bash
go run ./cmd/server/main.go
```

서버가 `http://localhost:8080`에서 실행됩니다.

## 프로젝트 구조

```
.
├── cmd/
│   ├── server/                    # 메인 서버
│   ├── add_more_quizzes/          # 퀴즈 데이터 추가 도구
│   ├── check_db/                  # DB 확인 도구
│   ├── list_collections/          # 컬렉션 조회
│   └── migrate_*/                 # 데이터 마이그레이션
├── internal/
│   ├── common/                    # 공통 상수, 에러, 로거
│   ├── config/                    # 환경 설정
│   ├── database/                  # MongoDB 연결
│   ├── graph/                     # GraphQL 리졸버 및 스키마
│   ├── handlers/                  # HTTP 핸들러
│   ├── middleware/                # HTTP 미들웨어
│   ├── models/                    # 도메인 모델
│   ├── repository/                # 데이터 액세스
│   ├── routes/                    # 라우터
│   ├── service/                   # 비즈니스 로직
│   ├── valkey/                    # Valkey/Redis 클라이언트
│   └── websocket/                 # WebSocket 핸들러
├── pkg/
│   ├── dictionary/                # 일본어 사전 (174,709개)
│   └── utils/                     # 유틸리티
├── data/
│   └── japanese_words_jmdict.txt  # 단어 데이터
└── deployments/
    └── k8s/                       # Kubernetes 매니페스트
```

## 데이터베이스 스키마

### users
```javascript
{
  _id: ObjectID,
  hange_id: String (unique),      // 로그인 ID
  name: String (unique),          // 사용자명
  avatar_url: String,
  level: Number,
  credit: Number,
  guild_id: ObjectID,
  created_at: Date,
  updated_at: Date
}
```

### ox_quizzes / qa_quizzes
```javascript
{
  _id: ObjectID,
  category: String,
  difficulty: Number (1-5),
  question: String,
  answer: Boolean | String,
  options: Array<String>,         // QA만
  explanation: String,
  is_active: Boolean,
  created_at: Date
}
```

### user_stats
```javascript
{
  _id: ObjectID,
  user_id: String (unique),
  total_games: Number,
  total_wins: Number,
  total_score: Number,
  wordchain_games: Number,
  ox_games: Number,
  qa_games: Number,
  created_at: Date
}
```

## WebSocket 메시지

### Client → Server
```javascript
// 채널 구독
{"type": "subscribe", "channel": "lobby"}
{"type": "subscribe", "channel": "game/{roomID}"}

// 사용자 ID 설정
{"type": "set_user_id", "userId": "user-id"}

// 끝말잇기 정답 제출
{"type": "wordchain_submit", "data": {"userId": "user-id", "word": "단어", "lastWord": "이전단어"}}

// 로비 채팅
{"type": "lobby_chat", "data": {"userId": "user-id", "username": "user", "message": "메시지"}}
```

### Server → Client
```javascript
// 끝말잇기 단어 제시 (라운드 시작)
{"type": "wordchain_prompt", "prompt": null, "lastWord": "단어", "currentRound": 1}

// 정답 결과
{"type": "word_result", "userName": "user", "word": "단어", "correct": true, "reason": "정답!"}

// 게임방 업데이트
{"type": "room_update", "data": {GameRoom}}

// 라운드 종료
{"type": "round_end", "data": {"round": 1, "reason": "시간 초과"}}

// 게임 종료
{"type": "game_end", "data": {"totalRounds": 5, "users": [...], "message": "게임 종료"}}
```

## 참고 문서

- [schema.graphqls](internal/graph/schema.graphqls) - GraphQL 스키마
- [API_REFERENCE.md](docs/API_REFERENCE.md) - API 상세 문서
- [deployments/k8s/README.md](deployments/k8s/README.md) - Kubernetes 배포 가이드
