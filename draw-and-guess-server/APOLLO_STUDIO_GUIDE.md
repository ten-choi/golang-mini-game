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

#### Queries (데이터 조회)
- `user(username)` - 사용자 단건 조회
- `users` - 모든 사용자 목록
- `gameRoom(id)` - 게임방 단건 조회
- `gameRooms(gameType?)` - 게임방 목록 (타입 필터링 가능)
- `randomOXQuiz(roomId?)` - 랜덤 OX 퀴즈
- `randomQAQuiz(roomId?)` - 랜덤 QA 퀴즈
- `playerStats(username, gameType?)` - 플레이어 통계
- `leaderboard(gameType, limit?)` - 리더보드
- `gameConfig` - 게임 설정
- `randomWordchainPrompt` - 끝말잇기 단어

#### Mutations (데이터 변경)
- `createUser(input)` - 사용자 생성
- `updateUser(username, input)` - 사용자 정보 수정
- `createGameRoom(input)` - 게임방 생성
- `joinGameRoom(roomId, username, password?)` - 게임방 참가
- `leaveGameRoom(roomId, username)` - 게임방 퇴장
- `startGame(roomId)` - 게임 시작
- `submitAnswer(roomId, username, answer)` - 답안 제출
- `sendChat(roomId, username, message)` - 채팅 메시지 전송

#### Subscriptions (실시간 업데이트)
- `lobbyUpdated` - 로비 변경 알림
- `gameRoomUpdated(roomId)` - 게임방 상태 변경
- `playerJoined(roomId)` - 플레이어 입장 알림
- `playerLeft(roomId)` - 플레이어 퇴장 알림
- `gameStarted(roomId)` - 게임 시작 알림
- `roundStarted(roomId)` - 라운드 시작 알림
- `chatMessage(roomId)` - 채팅 메시지
- `gameEvent(roomId)` - 게임 이벤트 (정답/오답 등)

### 실제 사용 예제

자세한 예제는 [README.md](README.md)의 "API 문서" 섹션을 참고하세요.

### REST API Endpoints

현재 프로젝트는 GraphQL 우선 설계로, REST 엔드포인트는 다음만 제공합니다:

- `GET /health` - 헬스 체크
- `POST /graphql` - GraphQL 쿼리/뮤테이션
- `GET /graphql` - GraphQL Playground (개발용)

### WebSocket Protocol

WebSocket은 GraphQL Subscriptions을 통해 자동으로 처리됩니다:

```
Connection: ws://localhost:8080/graphql
Protocol: graphql-transport-ws
```

## 🔍 문서 확인

스키마를 업로드한 후:

1. Apollo Studio 접속: https://studio.apollographql.com
2. `draw-and-guess` 그래프 선택
3. **Schema** 탭에서 전체 API 문서 확인
4. **Explorer** 탭에서 GraphQL 쿼리 테스트

## 🛠️ 스키마 파일

- `internal/graph/schema.graphqls` - GraphQL 스키마 원본
- `internal/transport/graphql/schema.graphqls` - 트랜스포트 레이어 스키마 (동일 파일)

> **Note**: 현재 프로젝트는 `internal/graph/` 디렉토리를 사용하고 있습니다.

## 📝 업데이트 방법

API를 변경할 때마다:

1. `internal/graph/schema.graphqls` 파일 수정
2. GraphQL 코드 재생성:
   ```bash
   go run github.com/99designs/gqlgen generate
   ```
3. 스크립트 재실행:
   ```powershell
   .\scripts\publish-schema.ps1
   ```
4. Apollo Studio에서 변경사항 확인

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
