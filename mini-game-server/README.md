# Draw & Guess Game - Backend Server

> GraphQL + WebSocket 기반의 실시간 멀티플레이어 퀴즈 게임 서버

## 🎯 개요

3가지 게임 모드를 지원하는 실시간 멀티플레이어 퀴즈 플랫폼:
- **WORDCHAIN**: 일본어 끝말잇기 (174,709 단어)
- **OX**: O/X 퀴즈 (참/거짓)
- **QA**: 4지선다 퀴즈

## 📦 기술 스택

- **Language**: Go 1.24+
- **API**: gqlgen (GraphQL schema-first)
- **Database**: MongoDB 7.0
- **Cache**: Valkey 7.2 (Redis 호환)
- **WebSocket**: gorilla/websocket
- **Framework**: Gin

## 🚀 빠른 시작

### 1. 의존성 설치
```bash
go mod download
```

### 2. 데이터베이스 배포 (Kubernetes)
```bash
cd deployments/k8s
.\deploy-all.ps1
```

### 3. 환경 변수 설정 (.env)
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

## 📚 API 엔드포인트

| 엔드포인트 | 설명 |
|-----------|------|
| `POST /graphql` | GraphQL Query/Mutation |
| `GET /graphql` | GraphQL Playground (개발용) |
 
| `GET /health` | 헬스 체크 |

## 📖 GraphQL API 예제

### 사용자 생성
```graphql
mutation {
  createUser(input: {
    UserName: "user1"
    avatarUrl: "https://example.com/avatar.png"
  }) {
    id
    UserName
    level
    credit
  }
}
```

### 게임방 생성 및 참가
```graphql
# 게임방 생성
mutation {
  createGameRoom(input: {
    name: "My Quiz Room"
    gameType: QA
    maxUsers: 8
    totalRounds: 5
    hostUserName: "user1"
  }) {
    id
    name
    status
  }
}

# 게임방 참가
mutation {
  joinGameRoom(roomId: "room-id", UserName: "user2") {
    id
    users {
      UserName
      score
      isReady
    }
  }
}
```

### 게임 플레이
```graphql
# 게임 시작 (방장만)
mutation {
  startGame(roomId: "room-id") {
    id
    status
    currentRound
  }
}

# 답변 제출
mutation {
  submitAnswer(
    roomId: "room-id"
    UserName: "user1"
    answer: "1"  # OX: "true"/"false", QA: "0"-"3"
  )
}
```

### 실시간 구독
```graphql
# 게임방 업데이트
subscription {
  gameRoomUpdated(roomId: "room-id") {
    id
    status
    currentRound
    users { UserName score }
  }
}

# 채팅 메시지
subscription {
  chatMessage(roomId: "room-id") {
    UserName
    message
    timestamp
  }
}
```

자세한 API 문서는 [schema.graphqls](internal/graph/schema.graphqls)를 참고하세요.

## 🎮 게임 흐름

1. **방 생성**: 방장이 게임방 생성 (게임 타입 선택)
2. **입장**: 플레이어들이 입장 및 준비 완료
3. **시작**: 방장이 게임 시작 (최소 2명)
4. **플레이**: 각 라운드마다 퀴즈 출제 → 답변 제출
5. **점수**: 정답 시 100점, 실시간 점수 업데이트
6. **종료**: 모든 라운드 완료 후 최종 순위 발표

## 📂 프로젝트 구조

```
.
├── cmd/
│   ├── server/              # 메인 서버 애플리케이션
│   ├── add_ox_quizzes/      # OX 퀴즈 데이터 삽입 도구
│   └── add_qa_quizzes/      # QA 퀴즈 데이터 삽입 도구
├── internal/
│   ├── config/              # 환경 설정
│   ├── database/            # MongoDB 연결
│   ├── graph/               # GraphQL 리졸버 및 스키마
│   ├── models/              # 도메인 모델
│   ├── repository/          # 데이터 액세스 레이어
│   ├── service/             # 비즈니스 로직
│   ├── middleware/          # HTTP 미들웨어
│   ├── valkey/              # Valkey 클라이언트
│   └── websocket/           # WebSocket 핸들러
├── pkg/
│   ├── dictionary/          # 일본어 단어 사전 (174,709개)
│   └── utils/               # 유틸리티 함수
├── data/
│   └── japanese_words_jmdict.txt  # 일본어 단어 데이터
└── deployments/
    └── k8s/                 # Kubernetes 매니페스트
```

## 🗄️ 데이터베이스 스키마

### users 컬렉션
```javascript
{
  _id: ObjectID,
  UserName: String (unique),
  avatar_url: String,
  level: Number,
  credit: Number,
  created_at: Date,
  updated_at: Date
}
```

### ox_quizzes / qa_quizzes 컬렉션
```javascript
{
  _id: ObjectID,
  category: String,
  difficulty: String,
  question: String,
  answer: Boolean | String,  // OX: Boolean, QA: String
  options: Array<String>,    // QA만
  explanation: String,
  is_active: Boolean,
  created_at: Date
}
```

### user_stats 컬렉션
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

## 🧪 테스트

```bash
# 모든 테스트 실행
go test ./...

# 커버리지 확인
go test -cover ./...
```

## 🔧 개발

### GraphQL 스키마 재생성
```bash
# schema.graphqls 수정 후
go run github.com/99designs/gqlgen generate
```

### 코드 품질
```bash
# 포맷팅
go fmt ./...

# Lint
go vet ./...
```

## 📦 배포

### Kubernetes
```bash
cd deployments/k8s
.\deploy-all.ps1

# 상태 확인
kubectl get pods -n data

# 정리
.\cleanup.ps1
```

### Docker
```bash
docker build -t draw-and-guess-server:latest .
docker run -p 8080:8080 draw-and-guess-server:latest
```

## 📚 추가 문서

- [PRD.md](PRD.md) - 기능 명세 및 로드맵
- [SCHEMA_IMPROVEMENTS.md](docs/SCHEMA_IMPROVEMENTS.md) - GraphQL 스키마 개선사항
- [deployments/k8s/README.md](deployments/k8s/README.md) - Kubernetes 배포 가이드

## 🤝 기여 가이드

### 코드 스타일
- Standard Go Project Layout 준수
- `gofmt`, `goimports`로 포맷팅
- 테스트 코드 작성 권장

### 커밋 메시지
```
feat: 새로운 기능 추가
fix: 버그 수정
docs: 문서 업데이트
refactor: 리팩토링
test: 테스트 추가/수정
```

## 📝 라이센스

MIT
