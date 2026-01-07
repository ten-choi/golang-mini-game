# Draw and Guess Game - Backend Server

## 🎯 개요
GraphQL + WebSocket 기반의 멀티플레이어 그림 맞추기 게임 서버

**빅테크 표준 Go 프로젝트 구조** (Standard Go Project Layout)

## 📂 프로젝트 구조

```
.
├── api/                    # API 정의 파일 (GraphQL 스키마, OpenAPI 등)
│   └── graphql/           # GraphQL 스키마
├── build/                  # 빌드 및 패키징 파일
│   └── docker/            # Dockerfile, .dockerignore
├── cmd/                    # 애플리케이션 엔트리포인트
│   └── server/            # 메인 서버 애플리케이션
│       └── main.go        # ✅ main.go의 올바른 위치
├── deployments/            # 배포 설정 (Kubernetes, Helm 등)
│   └── k8s/               # Kubernetes manifests
├── internal/               # Private 애플리케이션 코드
│   ├── common/            # 공통 유틸리티 (에러, 로깅, 응답)
│   ├── config/            # 설정 관리
│   ├── database/          # 데이터베이스 연결
│   ├── graph/             # GraphQL 리졸버
│   ├── handlers/          # HTTP 핸들러
│   ├── middleware/        # HTTP 미들웨어
│   ├── models/            # 도메인 모델
│   ├── repository/        # 데이터 액세스 인터페이스
│   ├── routes/            # 라우팅
│   ├── service/           # 비즈니스 로직
│   ├── valkey/            # 캐시 클라이언트
│   └── websocket/         # WebSocket 핸들러
├── pkg/                    # 외부 공개 가능한 라이브러리
│   ├── constants/         # 공통 상수
│   └── utils/             # 유틸리티 함수
├── test/                   # 테스트 헬퍼 및 인테그레이션 테스트
│   ├── integration/       # 통합 테스트
│   └── mocks/             # Mock 객체
└── scripts/                # 빌드/배포 스크립트
```

> **Note**: `cmd/server/main.go`가 **표준 위치**입니다. `src/` 디렉토리는 Go 프로젝트에서 사용하지 않습니다.

## 📦 기술 스택
- **Language**: Go 1.24+
- **GraphQL**: gqlgen (schema-first)
- **Database**: MongoDB 7.0
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
# MongoDB & Valkey 배포
cd deployments/k8s
.\deploy-all.ps1

# MongoDB는 자동으로 데이터베이스를 생성합니다
```

### 3. 환경 변수 설정 (.env)
```env
# MongoDB (Kubernetes NodePort)
MONGO_URI=mongodb://admin:password@localhost:30017
MONGO_DB=draw_and_guess_db

# Valkey (Kubernetes NodePort)
VALKEY_ADDR=localhost:30379

# Server
SERVER_PORT=8080
GIN_MODE=debug
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

### GraphQL Endpoint
- **개발**: `http://localhost:8080/graphql`
- **프로덕션**: Apollo Studio 사용 권장

### REST Endpoints
| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | 헬스 체크 |
| POST | `/graphql` | GraphQL Query/Mutation |
| GET | `/graphql` | GraphQL Playground (개발용) |

### WebSocket Endpoints
| Protocol | Path | Description |
|----------|------|-------------|
| WS | `/ws` | WebSocket 연결 (Subscriptions) |

### GraphQL API 주요 기능

#### 1. User Management
```graphql
# 사용자 생성
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

# 사용자 조회
query {
  user(username: "player1") {
    id
    displayName
    avatarUrl
  }
}
```

#### 2. Game Room Management
```graphql
# 게임방 생성
mutation {
  createGameRoom(input: {
    name: "My Quiz Room"
    gameType: QA
    maxPlayers: 8
    totalRounds: 5
    hostUsername: "player1"
    isPrivate: false
  }) {
    id
    name
    gameType
    status
  }
}

# 게임방 참가
mutation {
  joinGameRoom(
    roomId: "room-uuid"
    username: "player2"
  ) {
    id
    players {
      username
      displayName
      score
      isReady
    }
  }
}
```

#### 3. Quiz Queries
```graphql
# OX 퀴즈 조회
query {
  randomOXQuiz {
    id
    question
    answer
    explanation
    difficulty
  }
}

# QA 퀴즈 조회 (4지선다)
query {
  randomQAQuiz {
    id
    question
    options
    answer
    explanation
  }
}
```

#### 4. Game Flow
```graphql
# 게임 시작 (방장만 가능)
mutation {
  startGame(roomId: "room-uuid") {
    id
    status
    currentRound
  }
}

# 정답 제출
mutation {
  submitAnswer(
    roomId: "room-uuid"
    username: "player1"
    answer: "1"  # OX: "true"/"false", QA: "0"-"3"
  )
}
```

#### 5. Real-time Subscriptions
```graphql
# 게임방 업데이트 구독
subscription {
  gameRoomUpdated(roomId: "room-uuid") {
    id
    status
    currentRound
    players {
      username
      score
      isReady
    }
  }
}

# 채팅 메시지 구독
subscription {
  chatMessage(roomId: "room-uuid") {
    username
    displayName
    message
    timestamp
  }
}
```

자세한 API 스키마는 [schema.graphqls](internal/graph/schema.graphqls)를 참고하세요.

## 🎮 게임 타입

| 게임 모드 | 설명 | 데이터 소스 |
|----------|------|-------------|
| **WORDCHAIN** | 한국어 끝말잇기 | `korean_words` 테이블 |
| **OX** | O/X 퀴즈 (참/거짓) | `ox_quizzes` 테이블 |
| **QA** | 4지선다 퀴즈 | `qa_quizzes` 테이블 |

### 게임 흐름
1. 방장이 게임방 생성 (게임 타입 선택)
2. 플레이어들이 입장 및 준비
3. 방장이 게임 시작 (최소 2명 필요)
4. 각 라운드마다 퀴즈 출제
5. 플레이어들이 답안 제출
6. 정답 공개 및 점수 집계
7. 모든 라운드 종료 후 최종 순위 발표

## 🧪 테스트

```bash
# 모든 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./internal/service/...

# 통합 테스트
go test ./test/integration/...

# 커버리지 확인
go test -cover ./...
```

## 🏗️ 아키텍처

### 레이어 구조
```
┌─────────────────────────────────────┐
│         GraphQL / WebSocket         │  API Layer
├─────────────────────────────────────┤
│     Resolvers / Handlers            │  Presentation
├─────────────────────────────────────┤
│     Service Layer (Interfaces)      │  Business Logic
├─────────────────────────────────────┤
│  Repository Layer (Interfaces)      │  Data Access
├─────────────────────────────────────┤
│  PostgreSQL          Valkey         │  Storage
└─────────────────────────────────────┘
```

### 주요 디자인 패턴
- **Interface-based Architecture**: 의존성 역전, 테스트 용이성
- **Repository Pattern**: 데이터 액세스 추상화
- **Service Layer**: 비즈니스 로직 캡슐화
- **Middleware Chain**: 공통 관심사 처리 (로깅, 에러 핸들링, CORS)
- **Graceful Shutdown**: 10초 타임아웃으로 안전한 종료

## 🎮 게임 타입

| 게임 모드 | 설명 | 데이터 소스 |
|----------|------|-------------|
| **WORDCHAIN** | 한국어 끝말잇기 | `korean_words` 테이블 |
| **OX** | O/X 퀴즈 (참/거짓) | `ox_quizzes` 테이블 |
| **QA** | 4지선다 퀴즈 | `qa_quizzes` 테이블 |

### 게임 흐름
1. 방장이 게임방 생성 (게임 타입 선택)
2. 플레이어들이 입장 및 준비
3. 방장이 게임 시작 (최소 2명 필요)
4. 각 라운드마다 퀴즈 출제
5. 플레이어들이 답안 제출
6. 정답 공개 및 점수 집계
7. 모든 라운드 종료 후 최종 순위 발표

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
# 개발 빌드
go build -o bin/server.exe ./cmd/server

# 프로덕션 빌드 (최적화)
go build -ldflags="-s -w" -o bin/server ./cmd/server
```

### GraphQL 스키마 재생성
```bash
# 스키마 파일 수정 후 (api/graphql/*.graphqls)
go run github.com/99designs/gqlgen generate
```

### Docker 빌드
```bash
docker build -f build/docker/Dockerfile -t draw-and-guess-server:latest .
```

### 코드 품질
```bash
# 포맷팅
go fmt ./...

# Lint
go vet ./...

# 테스트 커버리지
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📦 배포

### Kubernetes 배포
```bash
cd deployments/k8s
.\deploy-all.ps1

# 상태 확인
kubectl get pods -n data
kubectl get svc -n data

# 정리
.\cleanup.ps1
```

자세한 배포 가이드는 [deployments/k8s/README.md](deployments/k8s/README.md)를 참조하세요.

## 📚 추가 문서

- [PRD.md](PRD.md) - 제품 요구사항 문서
- [APOLLO_STUDIO_GUIDE.md](APOLLO_STUDIO_GUIDE.md) - Apollo Studio 사용 가이드
- [api/graphql/README.md](api/graphql/README.md) - GraphQL API 가이드
- [pkg/README.md](pkg/README.md) - 공용 패키지 가이드
- [test/integration/README.md](test/integration/README.md) - 테스트 가이드

## 🤝 기여 가이드

### 코드 스타일
- Standard Go Project Layout 준수
- `gofmt`, `goimports`로 포맷팅
- 인터페이스 기반 설계
- 테스트 코드 작성 필수

### 커밋 메시지
```
feat: 새로운 기능 추가
fix: 버그 수정
docs: 문서 업데이트
refactor: 코드 리팩토링
test: 테스트 추가/수정
chore: 빌드, 설정 변경
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
