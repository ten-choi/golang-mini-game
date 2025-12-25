# Draw and Guess Game - Backend Server

## 🎯 개요
GraphQL + WebSocket 기반의 멀티플레이어 그림 맞추기 게임 서버

## 📦 기술 스택
- **Language**: Go 1.24+
- **GraphQL**: gqlgen (schema-first)
- **Database**: PostgreSQL 16
- **Cache**: Valkey (Redis 호환)
- **WebSocket**: gorilla/websocket
- **Web Framework**: Gin

## 🚀 빠른 시작

### 1. 의존성 설치
```bash
go mod download
```

### 2. 데이터베이스 준비 (Kubernetes)
```bash
# PostgreSQL & Valkey 배포
cd k8s
.\deploy-all.ps1

# 데이터베이스 생성
kubectl exec -n data postgres-0 -- psql -U postgres -c "CREATE DATABASE draw_guess_game;"
```

### 3. 환경 변수 설정 (.env)
```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=draw_guess_game
VALKEY_ADDR=localhost:6379
SERVER_PORT=8080
```

### 4. 서버 실행
```bash
# 빌드 & 실행
go build -o bin/server.exe ./cmd/server
.\bin\server.exe

# 또는 직접 실행
go run ./cmd/server/main.go
```

서버가 `http://localhost:8080`에서 실행됩니다.

## 📚 API 문서

### GraphQL Playground
- 개발용: `http://localhost:8080/api/v1/graphql`
- 프로덕션: [Apollo Studio](https://studio.apollographql.com) 사용 권장

### 주요 엔드포인트
- `GET  /health` - 헬스 체크
- `POST /api/v1/graphql` - GraphQL 쿼리/뮤테이션
- `GET  /api/v1/graphql` - GraphQL Playground (개발용)
- `WS   /api/v1/ws/lobby` - 로비 WebSocket
- `WS   /api/v1/ws/rooms/:id` - 게임방 WebSocket

## 🏗️ 프로젝트 구조
```
├── cmd/
│   └── server/           # 메인 애플리케이션 엔트리포인트
│       └── main.go
├── internal/
│   ├── config/          # 환경 설정
│   ├── database/        # PostgreSQL 연결
│   ├── valkey/          # Valkey(Redis) 클라이언트
│   ├── models/          # 데이터 모델
│   ├── repository/      # 데이터 접근 계층
│   ├── service/         # 비즈니스 로직
│   ├── graph/           # GraphQL 스키마 & 리졸버
│   ├── handlers/        # HTTP 핸들러
│   ├── routes/          # 라우팅 설정
│   └── websocket/       # WebSocket 핸들러
├── k8s/                 # Kubernetes 매니페스트
└── scripts/             # 유틸리티 스크립트
```

## 🎮 게임 타입
- **wordchain** (끝말잇기): 한국어 단어 체인 게임
- **ox** (OX 퀴즈): O/X 정답 맞추기
- **qa** (일반 퀴즈): 4지선다 퀴즈

## 📖 GraphQL API 예제

### 사용자 생성
```graphql
mutation {
  createUser(input: {
    username: "player1"
    displayName: "플레이어1"
    email: "player1@example.com"
  }) {
    id
    username
    displayName
    createdAt
  }
}
```

### 퀴즈 조회
```graphql
query {
  randomOXQuiz {
    id
    question
    answer
    explanation
  }
}
```

더 자세한 내용은 [APOLLO_STUDIO_GUIDE.md](APOLLO_STUDIO_GUIDE.md)를 참고하세요.

## 🔧 개발

### 빌드
```bash
go build -o bin/server.exe ./cmd/server
```

### GraphQL 스키마 재생성
```bash
gqlgen generate
```

### 테스트
```bash
go test ./...
```

## 📦 배포

### Docker 이미지 빌드
```bash
docker build -t draw-and-guess-server .
```

### Kubernetes 배포
```bash
cd k8s
kubectl apply -f .
```

## 📝 라이센스
MIT

- 엔드포인트: `ws://localhost:8080/app/ws`
- 메시지 형식: `{ "type": "subscribe|unsubscribe|message", "channel": "chat/<roomId>", "data": {...} }`
- 채널 예시
  - `draw/{roomId}`: 그림 스트로크 데이터
  - `chat/{roomId}`: 채팅/정답 피드백
  - `game/{roomId}`: 타이머 및 게임 상태
- 클라이언트는 `subscribe` 메시지를 보내 채널을 구독하고, 서버는 Valkey Pub/Sub을 통해 수신한 내용을 실시간으로 중계합니다.
