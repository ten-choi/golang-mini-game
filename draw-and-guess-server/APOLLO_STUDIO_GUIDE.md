# Apollo Studio Schema Registry 문서화 가이드

## 📚 개요

이 프로젝트는 **Apollo Studio Schema Registry**를 사용하여 모든 API를 문서화합니다:
- ✅ GraphQL Operations (Query, Mutation, Subscription)
- ✅ REST API Endpoints
- ✅ WebSocket Protocol

**HTTP만으로 가능**: HTTPS가 없어도 스키마를 Apollo Studio에 업로드하여 문서화할 수 있습니다!

## 🚀 빠른 시작

### 1. Rover CLI 설치

```bash
# npm을 사용하는 경우
npm install -g @apollo/rover

# 또는 직접 설치 (Linux/Mac)
curl -sSL https://rover.apollo.dev/nix/latest | sh

# Windows (PowerShell)
iwr 'https://rover.apollo.dev/win/latest' | iex
```

### 2. Apollo Studio 계정 생성

1. [Apollo Studio](https://studio.apollographql.com) 접속
2. 계정 생성 또는 로그인
3. 새 Graph 생성 (이름: `draw-and-guess`)
4. API Key 발급: [User Settings > API Keys](https://studio.apollographql.com/user-settings/api-keys)

### 3. API Key 설정

**Windows (PowerShell):**
```powershell
$env:APOLLO_KEY="your-apollo-key-here"
$env:APOLLO_GRAPH_REF="draw-and-guess@main"
```

**Linux/Mac (Bash):**
```bash
export APOLLO_KEY="your-apollo-key-here"
export APOLLO_GRAPH_REF="draw-and-guess@main"
```

### 4. 스키마 업로드

**Windows:**
```powershell
cd draw-and-guess-server
.\scripts\publish-schema.ps1
```

**Linux/Mac:**
```bash
cd draw-and-guess-server
chmod +x scripts/publish-schema.sh
./scripts/publish-schema.sh
```

## 📖 문서화된 내용

### GraphQL Operations
스키마는 다음을 포함합니다:
- **Queries**: `user`, `gameRooms`, `gameRoom`, `quizzes`, `quiz`
- **Mutations**: `createUser`, `createGameRoom`, `startGame` 등
- **Subscriptions**: `gameRoomUpdated`, `lobbyUpdated`, `chatMessage`

### REST API Endpoints
스키마 주석에 모든 REST 엔드포인트가 문서화되어 있습니다:

```graphql
"""
REST API Endpoints Documentation

Game Room Management:
  GET    /app/game/rooms              - Get all active game rooms
  POST   /app/game/room               - Create new game room
  POST   /app/game/room/:id/join      - Join game room
  ...
"""
```

### WebSocket Protocol
WebSocket 프로토콜 전체가 문서화되어 있습니다:

```graphql
"""
WebSocket Protocol Documentation

Connection: ws://host:port/app/ws

Message Types:
1. Subscribe: {"type": "subscribe", "channel": "lobby", ...}
2. Unsubscribe: {"type": "unsubscribe", ...}
3. Message: {"type": "message", ...}
"""
```

## 🔍 문서 확인

스키마를 업로드한 후:

1. Apollo Studio 접속: https://studio.apollographql.com
2. `draw-and-guess` 그래프 선택
3. **Schema** 탭에서 전체 API 문서 확인
4. **Explorer** 탭에서 GraphQL 쿼리 테스트

## 🛠️ 스키마 파일

- `src/graph/schema.graphqls` - 원본 GraphQL 스키마
- `src/graph/schema-complete.graphqls` - **완전한 문서화 스키마** (REST API + WebSocket 포함)

## 📝 업데이트 방법

API를 변경할 때마다:

1. `schema-complete.graphqls` 파일 수정
2. 스크립트 재실행:
   ```powershell
   .\scripts\publish-schema.ps1
   ```
3. Apollo Studio에서 변경사항 확인

## 🎯 장점

✅ **HTTP만으로 가능** - HTTPS 없이도 문서화
✅ **중앙 집중식 문서** - 팀 전체가 공유
✅ **자동 완성** - Apollo Studio Explorer
✅ **변경 추적** - 스키마 버전 관리
✅ **API 검증** - Breaking changes 감지

## 🔗 유용한 링크

- [Apollo Studio](https://studio.apollographql.com)
- [Rover CLI 문서](https://www.apollographql.com/docs/rover/)
- [Schema Registry 가이드](https://www.apollographql.com/docs/graphos/delivery/schema-registry/)
