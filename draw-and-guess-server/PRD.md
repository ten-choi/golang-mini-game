# Draw & Guess 멀티플레이어 게임 서버 - 기능 명세

## 1. 개요
GraphQL + WebSocket 기반의 실시간 멀티플레이어 퀴즈 게임 플랫폼

**기술 스택**: Go 1.24+ | gqlgen | MongoDB | Valkey | WebSocket

---

## 2. 기능 요구사항

### 2.1 사용자 관리

#### 2.1.1 사용자 생성 (P0) ✅ 완료
- **기능**: 신규 사용자 계정 생성
- **입력**:
  - UserName (필수, 고유값)
  - avatarUrl (선택)
- **출력**: 생성된 사용자 정보 + ID
- **검증**:
  - UserName sparse 인덱스 (null 허용)
  - MongoDB unique constraint
- **에러 처리**:
  - Duplicate key error: 이미 존재하는 UserName
  - 400 Bad Request: 유효하지 않은 입력

#### 2.1.2 사용자 조회 (P0) ✅ 완료
- **단일 조회**: ID로 특정 사용자 정보 조회
- **전체 조회**: 모든 사용자 목록 조회
- **출력**: User 객체 또는 배열

#### 2.1.3 사용자 정보 수정 (P1) ✅ 완료
- **기능**: UserName, avatarUrl 업데이트
- **입력**: ID + 수정할 필드들
- **검증**: 존재하는 사용자만 수정 가능

### 2.2 게임 관리

#### 2.2.1 게임 타입 (P0) ✅ 완료
1. **WORDCHAIN (끝말잇기)**
   - 일본어 단어 체인 게임
   - japanese_words_jmdict.txt 기반 (174,709 단어)
   - 실시간 단어 제출 및 검증
   - Dictionary 패키지로 검증

2. **OX (O/X 퀴즈)**
   - True/False 질문 형식
   - ox_quizzes 컬렉션에서 랜덤 추출
   - 난이도/카테고리별 분류

3. **QA (객관식 퀴즈)**
   - 4지선다형 문제
   - qa_quizzes 컬렉션에서 랜덤 추출
   - 이미지 URL 지원

#### 2.2.2 게임방 생성 (P0) ✅ 완료
- **입력**:
  - name: 방 이름
  - gameType: WORDCHAIN | OX | QA
  - maxPlayers: 최대 인원 (2-10명)
  - totalRounds: 총 라운드 수 (1 Round = 1 문제)
  - hostUserName: 방장 UserName
  - isPrivate: 비공개 여부 (선택)
  - password: 비공개방 비밀번호 (선택)
- **비즈니스 로직**:
  - 방장은 자동으로 플레이어에 추가 (isReady: true)
  - 상태는 WAITING으로 시작
  - UUID 자동 생성
  - 인메모리 gameRooms 맵에 저장

#### 2.2.3 게임방 참가/퇴장 (P0) ✅ 완료
- **참가**:
  - maxPlayers 체크
  - 중복 참가 방지
  - GraphQL Subscription으로 실시간 알림
  - Valkey PubSub 기반
- **퇴장**:
  - LeaveGameRoom mutation
  - 플레이어 목록에서 제거
  - 실시간 업데이트

#### 2.2.4 게임 시작 (P0) ✅ 완료
- **권한**: 방장만 가능
- **조건**: WAITING 상태
- **동작**:
  - 상태를 PLAYING으로 변경
  - 게임 타입별 퀴즈 전체 로드 (총 라운드 수만큼)
  - currentRound를 1로 설정
  - GraphQL Subscription 이벤트 발행

#### 2.2.5 게임 진행 (P0) ✅ 완료
- **답변 제출**: submitAnswer mutation
  - 정답 검증 및 점수 부여 (정답 시 100점)
  - 모든 플레이어 답변 시 자동 라운드 전환
  - 실시간 피드백 (정답/오답 이벤트)
- **라운드 관리**:
  - 자동 라운드 시작 (startRound)
  - 마지막 라운드 후 게임 종료
- **게임 종료**: endGame mutation
  - 최종 점수 계산
  - FINISHED 상태로 전환

### 2.3 퀴즈 관리 ✅ 완료

- **OX 퀴즈**: randomOXQuiz query - 활성 상태 퀴즈 랜덤 반환
- **QA 퀴즈**: randomQAQuiz query - 4개 옵션 배열 포함
- **단어 검증**: isValidWord query - 일본어 단어 유효성 검증

### 2.4 통계 및 랭킹 ✅ 완료

- **개인 통계**: playerStats query - 게임 타입별 전적 조회
- **리더보드**: leaderboard query - 게임 타입별 상위 랭커 조회

---

## 3. 기술 스택

### 3.1 백엔드
- **언어**: Go 1.24+
- **프레임워크**: Gin
- **GraphQL**: gqlgen (schema-first)
- **WebSocket**: gorilla/websocket

### 3.2 데이터베이스
- **MongoDB 7.0+**: 주 저장소 (users, quizzes, stats)
- **Valkey 7.2**: 실시간 게임방 상태, Pub/Sub 메시징

### 3.3 API 엔드포인트

#### GraphQL
- **POST /graphql**: Query/Mutation 실행
- **GET /graphql**: WebSocket Subscription 업그레이드
- **GET /playground**: GraphQL Playground UI

#### WebSocket
- **WS /ws/lobby**: 로비 실시간 업데이트
- **WS /ws/rooms/:id**: 게임방 실시간 통신 (채팅, 그림)

#### REST
- **GET /health**: 헬스 체크

### 3.4 GraphQL Subscription
- `gameRoomUpdated`: 게임방 상태 변경
- `playerJoined`, `playerLeft`: 플레이어 입/퇴장
- `gameStarted`, `gameEnded`: 게임 시작/종료
- `roundStarted`, `roundEnded`: 라운드 시작/종료
- `chatMessage`: 채팅 메시지
- `gameEvent`: 게임 이벤트 (정답, 오답 등)

---

## 4. 데이터베이스 스키마

### users 컬렉션
```javascript
{
  _id: ObjectID,
  UserName: String (unique, sparse),
  avatar_url: String,
  level: Number (default: 1),
  credit: Number (default: 0),
  guild_id: ObjectID,
  created_at: Date,
  updated_at: Date
}
```

### ox_quizzes 컬렉션
```javascript
{
  _id: ObjectID,
  category: String,
  difficulty: String,
  question: String,
  answer: Boolean,
  explanation: String,
  is_active: Boolean (default: true),
  created_at: Date
}
```

### qa_quizzes 컬렉션
```javascript
{
  _id: ObjectID,
  category: String,
  difficulty: String,
  question: String,
  options: Array<String>,
  correct_answer_index: Number,
  explanation: String,
  image_url: String,
  is_active: Boolean (default: true),
  created_at: Date
}
```

### player_stats 컬렉션
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
  created_at: Date,
  updated_at: Date
}
```

---

## 5. 개발 로드맵

### ✅ Phase 1 (v1.0) - MVP 완료
- 사용자 관리 (CRUD)
- 게임 타입 정의 (WORDCHAIN, OX, QA)
- GraphQL API
- WebSocket 실시간 통신
- MongoDB + Valkey 인프라
- 일본어 단어 사전 (174,709개)

### ✅ Phase 2 (v1.1) - 게임 플로우 완료
- 게임방 생성/참가/퇴장
- 게임 시작/종료
- 답변 제출 및 점수 시스템
- 라운드 자동 전환
- GraphQL Subscription
- WebSocket 채팅
- 통계 및 리더보드

### 🚧 Phase 3 (v1.2) - 개선 및 최적화
- [ ] 제한 시간 타이머
- [ ] 속도 기반 보너스 점수
- [ ] 방장 권한 검증 강화
- [ ] 재접속 처리
- [ ] 성능 최적화

### 📋 Phase 4 (v1.3) - 상점 시스템
- [ ] 아이템 상점 (아바타, 테마 등)
- [ ] 사용자 인벤토리
- [ ] 거래 내역
- [ ] 재화 시스템 통합