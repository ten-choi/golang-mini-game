# 🎨 Draw and Guess - 실시간 멀티플레이어 그림 맞추기 게임

> WebSocket 기반 실시간 그림 맞추기 게임 - Golang + React + Valkey

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18.2-61DAFB?style=flat&logo=react)](https://react.dev/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.3-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 📝 프로젝트 소개

**Draw and Guess**는 실시간으로 여러 명의 플레이어가 함께 즐길 수 있는 그림 맞추기 게임입니다. 한 명의 플레이어가 주어진 단어를 그림으로 표현하면, 다른 플레이어들이 채팅으로 정답을 맞추는 방식입니다.

### ✨ 주요 기능

- 🎮 **실시간 멀티플레이어**: WebSocket을 통한 실시간 게임 진행
- 🎨 **직관적인 그리기 도구**: HTML5 Canvas 기반 드로잉
- 🌍 **다국어 지원**: 한국어, 영어, 일본어, 중국어 지원
- 💬 **실시간 채팅**: 게임 중 실시간 채팅 및 정답 제출
- 🏆 **점수 시스템**: 선착순 정답자에게 높은 점수 부여
- ⚡ **수평 확장 가능**: Valkey Pub/Sub을 통한 멀티 서버 지원

## 🏗️ 기술 스택

### Backend
- **언어**: Go 1.21+
- **프레임워크**: Gin Web Framework
- **데이터베이스**: MongoDB (게임 토픽 저장)
- **캐시/메시징**: Valkey (게임 상태 관리 & Pub/Sub)
- **프로토콜**: REST API + WebSocket

### Frontend
- **언어**: TypeScript 5.3
- **프레임워크**: React 18.2
- **빌드 도구**: Vite 5.0
- **스타일링**: CSS Modules
- **상태 관리**: React Hooks

### Infrastructure
- **컨테이너화**: Docker & Docker Compose
- **오케스트레이션**: Kubernetes (k8s 매니페스트 포함)

## � API 문서화

이 프로젝트는 **Apollo Studio Schema Registry**를 사용하여 모든 API를 문서화합니다:

- ✅ **GraphQL Operations** - Query, Mutation, Subscription
- ✅ **REST API Endpoints** - 19개 엔드포인트 전체
- ✅ **WebSocket Protocol** - 실시간 통신 프로토콜

### 문서 확인 방법

1. **Apollo Studio** (권장): [상세 가이드](draw-and-guess-server/APOLLO_STUDIO_GUIDE.md)
   ```powershell
   # 스키마 업로드 후 Apollo Studio에서 확인
   cd draw-and-guess-server
   .\scripts\publish-schema.ps1
   ```

2. **GraphQL Playground**: http://localhost:8080/graphql/playground
3. **Apollo Sandbox**: http://localhost:8080/graphql/sandbox

## 🚀 빠른 시작

### 사전 요구사항

- Docker & Docker Compose
- Go 1.21+ (로컬 개발 시)
- Node.js 18+ (로컬 개발 시)
- Rover CLI (Apollo 스키마 업로드 시)

### 1. 저장소 클론

```bash
git clone https://github.com/yourusername/golang-mini-game.git
cd golang-mini-game
```

### 2. Docker Compose로 실행

```bash
# 백엔드 의존성 시작 (MongoDB, Valkey)
docker-compose up -d

# 백엔드 서버 실행
cd draw-and-guess-server
go run src/main.go

# 프론트엔드 개발 서버 실행 (새 터미널)
cd draw-and-guess-client
npm install
npm run dev
```

### 3. 접속

- **프론트엔드**: http://localhost:5173
- **백엔드 API**: http://localhost:8080
- **Swagger UI**: http://localhost:8080/swagger/index.html (개발 시)

## 📁 프로젝트 구조

```
golang-mini-game/
├── draw-and-guess-server/    # Go 백엔드 서버
│   ├── src/
│   │   ├── main.go           # 서버 진입점
│   │   ├── config/           # 설정 관리
│   │   ├── handlers/         # HTTP 핸들러
│   │   ├── models/           # 데이터 모델
│   │   ├── websocket/        # WebSocket 관리
│   │   ├── valkey/           # Valkey 클라이언트
│   │   └── database/         # MongoDB 연결
│   ├── k8s/                  # Kubernetes 매니페스트
│   └── Dockerfile
│
├── draw-and-guess-client/    # React 프론트엔드
│   ├── src/
│   │   ├── App.tsx           # 앱 진입점
│   │   ├── pages/            # 페이지 컴포넌트
│   │   ├── components/       # 재사용 컴포넌트
│   │   ├── services/         # API & WebSocket 서비스
│   │   ├── types/            # TypeScript 타입 정의
│   │   └── i18n/             # 다국어 지원
│   └── package.json
│
├── docker-compose.yml        # 로컬 개발 환경
├── PRD.md                    # 제품 명세서
├── ARCHITECTURE.md           # 아키텍처 문서
└── README.md                 # 이 파일
```

## 🎮 게임 방법

### 게임 진행

1. **방 생성**: 플레이어가 게임 방을 생성하고 방장이 됩니다
2. **참가자 입장**: 다른 플레이어들이 방에 참가합니다
3. **게임 시작**: 방장이 게임을 시작하면 첫 라운드가 시작됩니다
4. **그림 그리기**: 방장은 화면에 표시된 단어를 그림으로 표현합니다
5. **정답 맞추기**: 참가자들은 채팅으로 정답을 제출합니다
6. **다음 라운드**: 정답자가 다음 라운드의 그림 그리는 사람이 됩니다
7. **게임 종료**: 3점 달성 또는 3라운드 종료 시 게임이 끝납니다

### 점수 시스템

- **정답자 (최초)**: +2점
- **그림 그린 사람**: +1점
- **우승 조건**: 3점 선취 또는 3라운드 후 최고점

## 🔌 API 문서

### REST API

Base URL: `http://localhost:8080/app`

| Method | Endpoint | 설명 |
|--------|----------|------|
| GET | `/game/rooms` | 게임 방 목록 조회 |
| POST | `/game/room` | 게임 방 생성 |
| POST | `/game/room/:id/join` | 방 참가 |
| POST | `/game/room/:id/start` | 게임 시작 |
| POST | `/game/room/:id/chat` | 채팅/정답 제출 |
| DELETE | `/game/room/:id` | 방 삭제 |

자세한 API 문서는 [PRD.md](PRD.md)를 참조하세요.

### WebSocket

**연결**: `ws://localhost:8080/app/ws`

**채널**:
- `game/{roomId}` - 게임 상태 업데이트
- `draw/{roomId}` - 그리기 데이터
- `chat/{roomId}` - 채팅 메시지

## 🛠️ 개발 가이드

### 백엔드 개발

```bash
cd draw-and-guess-server

# 의존성 설치
go mod download

# 개발 서버 실행
go run src/main.go

# 빌드
go build -o server src/main.go

# 테스트
go test ./...
```

### 프론트엔드 개발

```bash
cd draw-and-guess-client

# 의존성 설치
npm install

# 개발 서버 실행
npm run dev

# 프로덕션 빌드
npm run build

# 빌드 미리보기
npm run preview
```

### 환경 변수

#### 백엔드 (.env)
```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=draw_and_guess_db
VALKEY_ADDR=localhost:6379
SERVER_PORT=8080
```

#### 프론트엔드 (.env)
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080
```

## 📦 배포

### Docker로 백엔드 빌드

```bash
cd draw-and-guess-server
docker build -t draw-and-guess-server:latest .
docker run -p 8080:8080 draw-and-guess-server:latest
```

### Kubernetes 배포

```bash
cd draw-and-guess-server/k8s
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
```

## 🤝 기여하기

프로젝트에 기여하고 싶으시다면:

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 라이선스

이 프로젝트는 MIT 라이선스 하에 배포됩니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 📚 추가 문서

- [PRD.md](PRD.md) - 제품 명세서 및 상세 API 문서
- [ARCHITECTURE.md](ARCHITECTURE.md) - 시스템 아키텍처 설명
- [draw-and-guess-server/README.md](draw-and-guess-server/README.md) - 백엔드 상세 문서
- [draw-and-guess-server/GRAPHQL_TEST_GUIDE.md](draw-and-guess-server/GRAPHQL_TEST_GUIDE.md) - GraphQL 테스트 가이드

## 🐛 버그 리포트 & 기능 요청

이슈가 있거나 새로운 기능을 제안하고 싶으시다면 [GitHub Issues](https://github.com/yourusername/golang-mini-game/issues)를 이용해주세요.

## 👥 제작자

Your Name - [@yourhandle](https://github.com/yourhandle)

프로젝트 링크: [https://github.com/yourusername/golang-mini-game](https://github.com/yourusername/golang-mini-game)

---

⭐ 이 프로젝트가 도움이 되셨다면 별을 눌러주세요!