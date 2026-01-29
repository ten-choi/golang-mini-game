# 멀티플레이어 게임 서버

## 개요
GraphQL + WebSocket 기반 실시간 퀴즈 게임 플랫폼

**스택**: Go 1.24+ | gqlgen | MongoDB | Valkey | gorilla/websocket

---

## 기능

### 사용자 관리 ✅
- **생성**: HangeID, Name(선택), AvatarURL(선택)
- **조회**: ID/HangeID/Name으로 조회, 전체 목록
- **수정**: AvatarURL, Level, Credit 업데이트
- **로그인**: HangeID로 조회 → WebSocket 인증

### 게임 타입 ✅
1. **WORDCHAIN (끝말잇기)**: 일본어 단어 174,709개, 중복 체크, 턴제
2. **OX (O/X 퀴즈)**: True/False, 난이도 1-5, 점수 차등
3. **QA (객관식)**: 4지선다, 이미지 지원, 정답 인덱스

### 게임방 관리 ✅
- **생성**: 이름, 게임타입, 최대인원(2-10), 라운드수, 비공개/비밀번호
- **참가/퇴장**: 인원 체크, WebSocket 실시간 알림
- **준비**: SetReady mutation, 방장은 항상 준비 상태
- **시작**: 방장 전용, 모든 퀴즈 사전 로드, PLAYING 상태 전환
- **초대**: InviteUser mutation, Valkey 상태 확인, 개인 채널 알림

### 게임 진행 ✅
- **퀴즈**: 라운드별 자동 전송, currentRound/totalRounds 포함, 중복 방지
- **타이머**: 라운드별 제한시간, 1초 단위 WebSocket 브로드캐스트
- **답변**: WebSocket으로 제출, 점수 차등(난이도 × 50), pending 처리
- **라운드 종료**: 타이머 종료 시 정답/설명/스코어보드 공개, 자동 다음 라운드
- **게임 종료**: 마지막 라운드 후 FINISHED, 통계 업데이트

### 실시간 통신 ✅
- **WebSocket**: `/ws` 엔드포인트, 로그인 후 인증
- **채널**: `lobby`, `user/{userId}`, `game/{roomId}`
- **Valkey Pub/Sub**: 메시지 브로드캐스팅
- **프로토콜**: `{"type":"...", "channel":"...", "data":{...}}`

---

## 데이터베이스

### users
```js
{
  _id: ObjectID,
  hange_id: String (unique, 로그인용),
  name: String (unique, 게임 표시명),
  avatar_url: String,
  level: Number,
  credit: Number,
  created_at: Date,
  updated_at: Date
}
```

### ox_quizzes
```js
{
  _id: ObjectID,
  category: String,
  difficulty: Number (1-5),
  question: String,
  answer: Boolean,
  explanation: String
}
```

### qa_quizzes
```js
{
  _id: ObjectID,
  category: String,
  difficulty: Number (1-5),
  question: String,
  options: [String, String, String, String],
  answer: Number (0-3, 정답 인덱스),
  explanation: String,
  image_url: String
}
```

### user_stats
```js
{
  _id: ObjectID,
  user_id: String (unique),
  wordchain_stats: { games, wins, score },
  ox_stats: { games, wins, score },
  qa_stats: { games, wins, score },
  created_at: Date,
  updated_at: Date
}
```

---

## 기술 스택

- **Backend**: Go 1.24+, Gin, gqlgen
- **DB**: MongoDB 7.0+ (users, quizzes, stats)
- **Cache**: Valkey 7.2 (게임방 상태, Pub/Sub)
- **Realtime**: gorilla/websocket, Valkey Pub/Sub

### API
- `POST /graphql` - GraphQL Mutation/Query
- `GET /graphql` - GraphQL Playground
- `WS /ws` - WebSocket 실시간 통신
- `GET /health` - 헬스 체크

---

## 데이터베이스

### users
```js
{
  _id: ObjectID,
  hange_id: String (unique, 로그인용),
  name: String (unique, 게임 표시명),
  avatar_url: String,
  level: Number,
  credit: Number,
  created_at: Date,
  updated_at: Date
}
```

### ox_quizzes
```js
{
  _id: ObjectID,
  category: String,
  difficulty: Number (1-5),
  question: String,
  answer: Boolean,
  explanation: String
}
```

### qa_quizzes
```js
{
  _id: ObjectID,
  category: String,
  difficulty: Number (1-5),
  question: String,
  options: [String, String, String, String],
  answer: Number (0-3, 정답 인덱스),
  explanation: String,
  image_url: String
}
```

### user_stats
```js
{
  _id: ObjectID,
  user_id: String (unique),
  wordchain_stats: { games, wins, score },
  ox_stats: { games, wins, score },
  qa_stats: { games, wins, score },
  created_at: Date,
  updated_at: Date
}
```

---

## 기술 스택

- **Backend**: Go 1.24+, Gin, gqlgen
- **DB**: MongoDB 7.0+ (users, quizzes, stats)
- **Cache**: Valkey 7.2 (게임방 상태, Pub/Sub)
- **Realtime**: gorilla/websocket, Valkey Pub/Sub

### API
- `POST /graphql` - GraphQL Mutation/Query
- `GET /graphql` - GraphQL Playground
- `WS /ws` - WebSocket 실시간 통신
- `GET /health` - 헬스 체크

- [ ] 성능 최적화

### 📋 Phase 4 (v1.3) - 상점 시스템
- [ ] 아이템 상점 (아바타, 테마 등)
- [ ] 사용자 인벤토리
- [ ] 거래 내역
- [ ] 재화 시스템 통합