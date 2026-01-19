# PRD: Draw & Guess 멀티플레이어 게임 플랫폼

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
   - 한국어 단어 체인 게임
   - japanese_words_jmdict.txt 기반 (174,709 단어)
   - 실시간 단어 제출 및 검증
   - Dictionary 패키지로 검증

2. **OX (OX 퀴즈)**
   - True/False 질문 형식
   - ox_quizzes 컬렉션에서 랜덤 추출
   - 난이도/카테고리별 분류
   - MongoDB 기반

3. **QA (일반 퀴즈)**
   - 주관식 문제
   - qa_quizzes 컬렉션에서 랜덤 추출
   - MongoDB 기반

#### 2.2.2 게임방 생성 (P0) ✅ 완료
- **입력**:
  - name: 방 이름
  - gameType: WORDCHAIN | OX | QA
  - maxusers: 최대 인원 (2-10명)
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
  - maxusers 체크
  - 중복 참가 방지
  - GraphQL Subscription으로 실시간 알림
  - Valkey PubSub 기반
- **퇴장**:
  - LeaveGameRoom mutation
  - 플레이어 목록에서 제거
  - 실시간 업데이트

#### 2.2.4 게임 시작 (P0) ✅ 완료
- **권한**: 방장만 가능 (현재 검증 미구현)
- **조건**:
  - WAITING 상태
- **동작**:
  - 상태를 PLAYING으로 변경
  - currentRound를 1로 설정
  - GraphQL Subscription으로 게임 시작 알림
  - gameStarted, gameRoomUpdated 이벤트 발행

#### 2.2.5 게임 진행 (P0)
- **라운드 관리**:
  - currentRound 추적
  - 제한 시간 관리 (타이머)
  - 자동 라운드 전환
- **점수 시스템**:
  - 정답 시 점수 부여
  - 속도에 따른 보너스 점수
  - 실시간 리더보드 업데이트

### 2.3 퀴즈 관리

#### 2.3.1 랜덤 퀴즈 조회 (P0)
- **OX 퀴즈**: GET randomOXQuiz
  - 활성 상태(is_active=true) 퀴즈만 반환
  - RANDOM() 사용
  - usage_count 자동 증가 (선택적)

- **일반 퀴즈**: GET randomQAQuiz
  - 활성 퀴즈 랜덤 반환
  - 4개 옵션 배열 포함
  - 이미지 URL 포함 가능

#### 2.3.2 단어 검증 (P1)
- **기능**: 끝말잇기용 한국어 단어 유효성 검증
- **입력**: word (string)
- **출력**: boolean
- **데이터**: korean_words 테이블 기반

### 2.4 통계 및 랭킹 (P2)

#### 2.4.1 플레이어 통계 (P2)
- **개인 통계**:
  - 총 게임 수
  - 승리 수
  - 총 점수
  - 게임 타입별 분리
- **조회**: UserName + gameType (선택)

#### 2.4.2 리더보드 (P2)
- **기능**: 게임 타입별 상위 랭커 조회
- **입력**:
  - gameType: wordchain | ox | qa
  - limit: 반환할 순위 수 (기본 10)
- **정렬**: totalScore DESC

---

## 3. 기술 요구사항

### 3.1 아키텍처

#### 3.1.1 레이어 구조
```
Client (Frontend)
    ↓
API Gateway (GraphQL + WebSocket)
    ↓
Service Layer (Business Logic)
    ↓
Repository Layer (Data Access)
    ↓
Database (PostgreSQL) + Cache (Valkey)
```

#### 3.1.2 디렉토리 구조 (현재 실제 구조)
```
cmd/
  ├── extract_words/        # 한국어 단어 추출 도구
  ├── insert_quiz/         # 퀴즈 데이터 삽입 도구
  └── server/              # 🚀 애플리케이션 엔트리포인트
      └── main.go

internal/
  ├── app/                 # (deprecated, 사용 안 함)
  ├── common/              # 회세 관심사
  │   ├── errors.go        # 표준화된 에러 처리
  │   ├── logger.go        # 구조화된 로깅
  │   ├── response.go      # HTTP 응답 헬퍼
  │   └── RESPONSE_GUIDE.md
  ├── config/              # ⚙️ 환경 설정
  │   └── config.go        # .env 파일 로딩
  ├── database/            # 💾 MongoDB 연결 및 스키마
  │   └── mongodb.go
  ├── graph/               # 🌐 GraphQL API
  │   ├── schema.graphqls   # GraphQL 스키마
  │   ├── resolver.go       # 리졸버 루트
  │   ├── resolver_*.go     # 엔티티별 리졸버
  │   ├── generated.go      # gqlgen 자동 생성
  │   ├── pubsub.go         # Subscription 구현
  │   ├── game_room_store.go # 인메모리 게임방 저장소
  │   └── model/
  │       └── models_gen.go # GraphQL 모델
  ├── handlers/            # REST API 핸들러
  │   └── health_handler.go
  ├── middleware/          # Gin 미들웨어
  │   ├── cors.go
  │   ├── error.go
  │   ├── logger.go
  │   └── trace.go
  ├── models/              # 🧠 도메인 모델
  │   ├── user.go
  │   ├── quiz.go
  │   ├── user_stats.go
  │   ├── game_room.go     # WebSocket용 레거시 모델
  │   ├── shop.go
  │   ├── transaction.go
  │   └── websocket_dto.go
  ├── repository/          # 💾 데이터 접근 레이어
  │   ├── user_repository.go
  │   ├── quiz_repository.go
  │   ├── user_stats_repository.go
  │   ├── shop_repository.go
  │   └── transaction_repository.go
  ├── routes/              # 🛣️ 라우팅 설정
  │   └── router.go         # Gin 라우트 마운트
  ├── service/             # ✅ 비즈니스 로직
  │   ├── user_service.go
  │   ├── quiz_service.go
  │   ├── user_stats_service.go
  │   ├── shop_service.go
  │   └── transaction_service.go
  ├── transport/           # 🌐 API 레이어
  │   ├── graphql/         # (deprecated, graph/ 사용)
  │   ├── rest/            # REST API
  │   │   ├── health_handler.go
  │   │   ├── response.go
  │   │   └── RESPONSE_GUIDE.md
  │   └── ws/              # WebSocket 통신
  │       ├── websocket.go
  │       ├── ws_message.go
  │       └── dto/
  │           └── websocket_dto.go
  ├── valkey/              # Valkey/Redis 클라이언트
  │   └── valkey.go
  └── websocket/           # WebSocket 핸들러 (레거시)
      ├── websocket.go
      └── ws_message.go

pkg/                       # 외부 공개 라이브러리
  ├── dictionary/          # 한국어 단어 사전 (174,709 단어)
  │   └── dictionary.go
  └── utils/
      ├── snowflake.go     # Snowflake ID 생성기
      ├── strings.go
      └── trace.go

data/                      # 데이터 파일
  └── japanese_words_jmdict.txt # JMDict 한국어 단어 (174,709개)

deployments/               # 배포 스크립트
  └── k8s/                # Kubernetes 매니페스트
      ├── deployment.yaml
      ├── service.yaml
      ├── mongodb.yaml
      ├── valkey.yaml
      └── pv-*.yaml
```

**아키텍처 특징**:
- **혼합 구조**: Clean Architecture + Standard Go Layout
- **레거시 코드**: internal/websocket/ 과 internal/models/game_room.go
- **새 코드**: internal/transport/ws/ 와 internal/graph/
- **GraphQL**: gqlgen으로 스키마 기반 자동 생성

### 3.2 기술 스택

#### 3.2.1 백엔드
- **언어**: Go 1.24+
- **웹 프레임워크**: Gin v1.11.0
- **GraphQL**: gqlgen v0.17.85 (schema-first)
- **WebSocket**: gorilla/websocket v1.5.1
- **Database Driver**: lib/pq v1.10.9
- **Cache Client**: go-redis/v9

#### 3.2.2 데이터베이스
- **주 저장소**: MongoDB 7.0+
  - 문서 기반 데이터 (users, quizzes, stats)
  - 유연한 스키마
  - 인덱스: UserName, user_id, category 등
- **캐시**: Valkey 7.2 (Redis 호환)
  - 실시간 게임방 상태 (인메모리)
  - Pub/Sub 메시징 (WebSocket 및 Subscription)
  - 채널: "lobby", "game/{roomID}"

#### 3.2.3 인프라
- **컨테이너**: Docker
- **오케스트레이션**: Kubernetes
  - StatefulSet (PostgreSQL, Valkey)
  - PersistentVolume (데이터 영속성)
- **Namespace**: data (데이터베이스)

### 3.3 API 설계

#### 3.3.1 GraphQL Endpoint ✅ 완료
- **POST /graphql**: 쿼리/뮤테이션 실행
- **GET /graphql**: WebSocket Subscription 업그레이드
- **GET /playground**: GraphQL Playground UI
- **GET /sandbox**: Apollo Sandbox UI
- **특징**:
  - Schema-first 접근 (gqlgen)
  - 타입 안정성
  - WebSocket transport 지원 (graphql-ws 프로토콜)

#### 3.3.2 WebSocket Endpoint ✅ 완료
- **WS /ws/lobby**: 로비 실시간 업데이트
  - 로비 채팅
  - 게임방 목록 업데이트
- **WS /ws/rooms/:id**: 게임방 실시간 통신
  - 체팅 메시지
  - 그림 그리기 데이터 (drawing)
  - 게임 액션 (game_action)
- **메시지 타입**:
  - subscribe: 채널 구독
  - unsubscribe: 구독 해제
  - message: 데이터 전송
  - chat: 채팅 메시지
  - drawing: 그림 데이터
  - game_action: 게임 액션

#### 3.3.3 GraphQL Subscription ✅ 완료
- **gameRoomUpdated**: 게임방 상태 변경 구독
- **userJoined**: 플레이어 입장 이벤트
- **gameStarted**: 게임 시작 이벤트
- **gameEnded**: 게임 종료 이벤트
- **Valkey PubSub 기반**:
  - 각 Subscription은 Valkey 채널에 매핑
  - startValkeySubscription으로 메시지 수신
  - WebSocket 클라이언트에게 브로드캐스트

#### 3.3.4 REST Endpoint ✅ 완료
- **GET /health**: 헬스 체크 (인프라용)

### 3.4 데이터베이스 스키마 (MongoDB)

#### 3.4.1 users 컬렉션
```javascript
{
  _id: ObjectID,
  UserName: String (unique, sparse),
  avatar_url: String,
  level: Number (default: 1),
  credit: Number (default: 0),      // 일반 재화
  han_coin: Number (default: 0),    // 프리미엄 재화
  guild_id: ObjectID,
  created_at: Date,
  updated_at: Date
}
// Indexes: username_1 (unique, sparse), created_at_-1
```

#### 3.4.2 korean_words 컬렉션 (사용 안 함)
- **대체**: data/japanese_words_jmdict.txt (174,709 단어)
- **로딩**: pkg/dictionary/dictionary.go
- **검색**: IsValidWord(word string)

#### 3.4.3 ox_quizzes 컬렉션
```javascript
{
  _id: ObjectID,
  category: String,
  difficulty: String,
  question: String,
  answer: Boolean,
  explanation: String,
  usage_count: Number (default: 0),
  is_active: Boolean (default: true),
  created_at: Date,
  updated_at: Date
}
// Indexes: category_1, is_active_1, difficulty_1
```

#### 3.4.4 qa_quizzes 컬렉션
```javascript
{
  _id: ObjectID,
  category: String,
  difficulty: String,
  question: String,
  answer: String,              // 정답 텍스트
  explanation: String,
  usage_count: Number (default: 0),
  is_active: Boolean (default: true),
  created_at: Date,
  updated_at: Date
}
// Indexes: category_1, is_active_1, difficulty_1
```

#### 3.4.5 user_stats 컬렉션
```javascript
{
  _id: ObjectID,
  user_id: String (unique),
  total_games: Number (default: 0),
  total_wins: Number (default: 0),
  total_score: Number (default: 0),
  wordchain_games: Number (default: 0),
  ox_games: Number (default: 0),
  qa_games: Number (default: 0),
  created_at: Date,
  updated_at: Date
}
// Indexes: user_id_1 (unique), total_score_-1
```

#### 3.4.6 shop_items 컬렉션 (추가됨)
```javascript
{
  _id: ObjectID,
  item_type: String,           // "avatar", "theme", "effect"
  name: String,
  description: String,
  price: Number,
  currency_type: String,       // "credit" or "han_coin"
  is_available: Boolean,
  is_featured: Boolean,
  image_url: String,
  tags: Array<String>,
  created_at: Date,
  updated_at: Date
}
// Indexes: item_type_1, is_available_1, is_featured_1, price_1, tags_1
```

#### 3.4.7 user_inventory 컬렉션 (추가됨)
```javascript
{
  _id: ObjectID,
  user_id: String,
  item_id: ObjectID,
  item_type: String,
  is_equipped: Boolean (default: false),
  purchased_at: Date
}
// Indexes: [user_id_1, item_id_1] (unique), [user_id_1, item_type_1], [user_id_1, is_equipped_1]
```

#### 3.4.8 transactions 컬렉션 (추가됨)
```javascript
{
  _id: ObjectID,
  user_id: String,
  type: String,                // "purchase", "reward", "spend"
  amount: Number,
  currency_type: String,       // "credit" or "han_coin"
  item_id: ObjectID,
  description: String,
  status: String,              // "pending", "completed", "failed"
  created_at: Date
}
// Indexes: [user_id_1, created_at_-1], [user_id_1, type_1], [user_id_1, status_1], created_at_-1
```
 

## 8. 로드맵

### Phase 1 (v1.0) - MVP ✅ 완료
- [x] 사용자 관리 (CRUD)
- [x] 게임 타입 정의 (WORDCHAIN, OX, QA)
- [x] 퀴즈는 한번에 전체조회하고 정답대조를하자 (랜덤으로게임 라운드만큼의 문제를 전달)
- [x] GraphQL API (gqlgen)
- [x] WebSocket 실시간 통신
- [x] MongoDB + Valkey 인프라
- [x] 일본어 단어 사전 (174,709개)

### Phase 2 (v1.1) - 게임 플로우 🚧 진행 중
- [x] 게임방 생성/참가/퇴장
- [x] 게임 시작 (StartGame)
- [x] GraphQL Subscription (gameRoomUpdated, userJoined, gameStarted, gameEnded)
- [x] Valkey PubSub 통합
- [x] WebSocket 채팅 (lobby, room) 
- [ ] 게임 진행 로직 (라운드 관리)
- [ ] 정답 제출 및 점수 시스템 (SubmitAnswer)
- [ ] 타이머 관리
- [ ] 다음 라운드/게임 종료
- [ ] 리더보드
 
 
### Phase 4 (v1.3) - 상점 시스템 🏪
- [x] 상점 아이템 (shop_items 컬렉션)
- [x] 사용자 인벤토리 (user_inventory 컬렉션)
- [x] 거래 내역 (transactions 컬렉션)
- [x] 재화 시스템 (credit, han_coin)
- [ ] GraphQL Mutation: purchaseItem, equipItem
- [ ] GraphQL Query: shopItems, userInventory
 