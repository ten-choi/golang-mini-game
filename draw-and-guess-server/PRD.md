# PRD: Draw & Guess 멀티플레이어 게임 플랫폼

## 📋 문서 정보
- **프로젝트명**: Draw & Guess Game Server
- **버전**: v1.0
- **작성일**: 2025-12-25
- **문서 타입**: Product Requirements Document

---

## 1. 제품 개요

### 1.1 목적
실시간 멀티플레이어 게임 서버로, 사용자들이 다양한 게임 모드(끝말잇기, OX퀴즈, 일반퀴즈)를 통해 경쟁하고 즐길 수 있는 플랫폼을 제공합니다.

### 1.2 핵심 가치
- **실시간 상호작용**: WebSocket 기반의 즉각적인 게임 경험
- **다양한 게임 모드**: 3가지 게임 타입으로 다양성 제공
- **확장 가능한 아키텍처**: GraphQL + 마이크로서비스 패턴
- **개발자 친화적**: Apollo Studio 통합으로 API 문서화 자동화

### 1.3 목표 사용자
- **Primary**: 캐주얼 게임을 즐기는 한국어 사용자
- **Secondary**: 게임 플랫폼 개발자 (API 활용)

---

## 2. 기능 요구사항

### 2.1 사용자 관리

#### 2.1.1 사용자 생성 (P0)
- **기능**: 신규 사용자 계정 생성
- **입력**:
  - username (필수, 고유값)
  - displayName (필수)
  - email (선택)
  - avatarURL (선택)
- **출력**: 생성된 사용자 정보 + ID
- **검증**:
  - username 중복 체크
  - 3-20자 길이 제한
- **에러 처리**:
  - 409 Conflict: 이미 존재하는 username
  - 400 Bad Request: 유효하지 않은 입력

#### 2.1.2 사용자 조회 (P0)
- **단일 조회**: username으로 특정 사용자 정보 조회
- **전체 조회**: 모든 사용자 목록 조회 (페이지네이션 미구현)
- **출력**: User 객체 또는 배열

#### 2.1.3 사용자 정보 수정 (P1)
- **기능**: displayName, email, avatarURL 업데이트
- **입력**: username + 수정할 필드들 (모두 선택)
- **검증**: 존재하는 사용자만 수정 가능

### 2.2 게임 관리

#### 2.2.1 게임 타입 (P0)
1. **wordchain (끝말잇기)**
   - 한국어 단어 체인 게임
   - korean_words 테이블 기반 검증
   - 실시간 단어 제출 및 검증

2. **ox (OX 퀴즈)**
   - True/False 질문 형식
   - ox_quizzes 테이블에서 랜덤 추출
   - 난이도/카테고리별 분류

3. **qa (일반 퀴즈)**
   - 4지선다 문제
   - qa_quizzes 테이블에서 랜덤 추출
   - 이미지 지원 (imageURL)

#### 2.2.2 게임방 생성 (P0)
- **입력**:
  - name: 방 이름
  - gameType: wordchain | ox | qa
  - maxPlayers: 최대 인원 (2-10명)
  - totalRounds: 총 라운드 수
  - hostUsername: 방장 username
- **비즈니스 로직**:
  - 방장은 자동으로 플레이어에 추가
  - 상태는 WAITING으로 시작
  - UUID 자동 생성

#### 2.2.3 게임방 참가/퇴장 (P0)
- **참가**:
  - maxPlayers 체크
  - 중복 참가 방지
  - WebSocket으로 실시간 알림
- **퇴장**:
  - 방장 퇴장 시 자동 위임 또는 방 삭제
  - 마지막 플레이어 퇴장 시 방 자동 삭제

#### 2.2.4 게임 시작 (P0)
- **권한**: 방장만 가능
- **조건**:
  - 최소 2명 이상
  - WAITING 상태
- **동작**:
  - 상태를 PLAYING으로 변경
  - 첫 라운드 시작
  - WebSocket으로 게임 시작 알림

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
- **조회**: username + gameType (선택)

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

#### 3.1.2 디렉토리 구조 (Clean Architecture + Standard Go Layout)
```
cmd/
  └── server/              # 🚀 애플리케이션 엔트리포인트 (thin)
      └── main.go

internal/
  ├── app/                 # 🔧 DI/Wiring, 서버 초기화
  │   └── server.go        # Gin 엔진 빌드, 의존성 주입
  │
  ├── config/              # ⚙️ 환경 설정
  │   └── config.go
  │
  ├── platform/            # ⚙️ 횡단 관심사 인프라스트럭처
  │   ├── logger/          # 구조화된 로깅
  │   │   └── logger.go
  │   ├── apperr/          # 표준화된 에러 처리
  │   │   └── errors.go
  │   ├── db/              # 데이터베이스 연결 관리
  │   │   └── postgres.go
  │   ├── cache/           # 캐시 클라이언트
  │   │   └── valkey.go
  │   ├── trace/           # 분산 추적
  │   │   └── trace.go
  │   └── id/              # ID 생성 (Snowflake)
  │       └── snowflake.go
  │
  ├── domain/              # 🧠 순수 비즈니스 모델 (프레임워크 독립)
  │   ├── user.go
  │   ├── quiz.go
  │   ├── player_stats.go
  │   └── game_room.go
  │
  ├── repository/          # 💾 데이터 접근 인터페이스
  │   ├── user_repository.go
  │   ├── quiz_repository.go
  │   └── player_stats_repository.go
  │
  ├── service/             # ✅ 비즈니스 로직 / Use Cases
  │   ├── user_service.go
  │   ├── quiz_service.go
  │   ├── player_stats_service.go
  │   └── game_room_store.go    # 인메모리 게임방 스토어
  │
  └── transport/           # 🌐 API 레이어 (프레젠테이션)
      ├── http/
      │   ├── router.go    # Gin 라우트 마운트 (REST/GraphQL/WS)
      │   └── middleware/
      │       ├── cors.go
      │       ├── error.go
      │       ├── logger.go
      │       └── trace.go
      │
      ├── rest/            # REST API 핸들러
      │   ├── health_handler.go
      │   ├── response.go
      │   └── RESPONSE_GUIDE.md
      │
      ├── graphql/         # GraphQL API
      │   ├── schema.graphqls
      │   ├── resolver.go
      │   ├── schema.resolvers.go
      │   ├── generated.go
      │   ├── pubsub.go
      │   └── model/
      │       └── models_gen.go
      │
      └── ws/              # WebSocket 실시간 통신
          ├── websocket.go
          ├── ws_message.go
          └── dto/
              └── websocket_dto.go

pkg/                       # 외부에 공개 가능한 라이브러리
  └── utils/
      ├── snowflake.go     # (deprecated, use platform/id/)
      └── strings.go
```

**아키텍처 설계 원칙**:
- **관심사의 분리**: 도메인 로직과 인프라 구현 분리
- **의존성 규칙**: 내부 레이어(domain, service)는 외부 레이어(transport)를 의존하지 않음
- **테스트 용이성**: 인터페이스 기반 설계로 모킹 가능
- **확장성**: 새로운 전송 프로토콜(gRPC 등) 추가 용이

### 3.2 기술 스택

#### 3.2.1 백엔드
- **언어**: Go 1.24+
- **웹 프레임워크**: Gin v1.11.0
- **GraphQL**: gqlgen v0.17.85 (schema-first)
- **WebSocket**: gorilla/websocket v1.5.1
- **Database Driver**: lib/pq v1.10.9
- **Cache Client**: go-redis/v9

#### 3.2.2 데이터베이스
- **주 저장소**: PostgreSQL 16
  - 구조적 데이터 (users, quizzes, stats)
  - ACID 트랜잭션 보장
- **캐시**: Valkey 7.2 (Redis 호환)
  - 실시간 게임방 상태
  - 세션 관리
  - Pub/Sub 메시징

#### 3.2.3 인프라
- **컨테이너**: Docker
- **오케스트레이션**: Kubernetes
  - StatefulSet (PostgreSQL, Valkey)
  - PersistentVolume (데이터 영속성)
- **Namespace**: data (데이터베이스)

### 3.3 API 설계

#### 3.3.1 GraphQL Endpoint
- **POST /api/v1/graphql**: 쿼리/뮤테이션 실행
- **GET /api/v1/graphql**: Apollo Sandbox UI
- **특징**:
  - Schema-first 접근
  - 타입 안정성
  - Apollo Studio 통합

#### 3.3.2 WebSocket Endpoint
- **WS /api/v1/ws/lobby**: 로비 실시간 업데이트
- **WS /api/v1/ws/rooms/:id**: 게임방 실시간 통신
- **메시지 타입**:
  - subscribe: 채널 구독
  - unsubscribe: 구독 해제
  - message: 데이터 전송

#### 3.3.3 REST Endpoint
- **GET /health**: 헬스 체크 (인프라용)

### 3.4 데이터베이스 스키마

#### 3.4.1 users 테이블
```sql
id SERIAL PRIMARY KEY
username VARCHAR(50) UNIQUE NOT NULL
display_name VARCHAR(100) NOT NULL
email VARCHAR(255)
avatar_url TEXT
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

#### 3.4.2 korean_words 테이블
```sql
id SERIAL PRIMARY KEY
word VARCHAR(50) UNIQUE NOT NULL
```

#### 3.4.3 ox_quizzes 테이블
```sql
id SERIAL PRIMARY KEY
category VARCHAR(50) NOT NULL
difficulty VARCHAR(20) NOT NULL
question TEXT NOT NULL
answer BOOLEAN NOT NULL
explanation TEXT
usage_count INTEGER DEFAULT 0
is_active BOOLEAN DEFAULT TRUE
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

#### 3.4.4 qa_quizzes 테이블
```sql
id SERIAL PRIMARY KEY
category VARCHAR(50) NOT NULL
difficulty VARCHAR(20) NOT NULL
question TEXT NOT NULL
options TEXT[] NOT NULL  -- 4개 선택지
answer INTEGER NOT NULL  -- 0-3 인덱스
explanation TEXT
image_url TEXT
usage_count INTEGER DEFAULT 0
is_active BOOLEAN DEFAULT TRUE
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
```

#### 3.4.5 player_stats 테이블
```sql
id SERIAL PRIMARY KEY
username VARCHAR(50) NOT NULL
game_type VARCHAR(20) NOT NULL
total_games INTEGER DEFAULT 0
total_wins INTEGER DEFAULT 0
total_score INTEGER DEFAULT 0
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
UNIQUE(username, game_type)
```

---

## 4. 비기능 요구사항

### 4.1 성능

#### 4.1.1 응답 시간
- **GraphQL 쿼리**: < 100ms (p95)
- **WebSocket 메시지**: < 50ms (p95)
- **데이터베이스 쿼리**: < 50ms (p95)

#### 4.1.2 동시성
- **동시 접속자**: 1,000명
- **동시 게임방**: 100개
- **WebSocket 연결**: 1,000개

#### 4.1.3 처리량
- **GraphQL QPS**: 1,000 queries/sec
- **WebSocket 메시지**: 5,000 messages/sec

### 4.2 안정성

#### 4.2.1 가용성
- **목표 Uptime**: 99.9% (월 43분 다운타임 허용)
- **Graceful Shutdown**: 10초 타임아웃
- **Health Check**: 5초 간격

#### 4.2.2 에러 처리
- **표준화된 에러 코드**:
  - 400: Bad Request
  - 401: Unauthorized
  - 404: Not Found
  - 409: Conflict
  - 500: Internal Server Error
- **에러 로깅**: 모든 에러는 로그에 기록
- **Panic Recovery**: 자동 복구 및 500 응답

#### 4.2.3 데이터 일관성
- **트랜잭션**: ACID 보장
- **Optimistic Locking**: 동시성 제어
- **Idempotency**: 중복 요청 방지

### 4.3 보안

#### 4.3.1 인증/인가 (P2 - 향후 구현)
- **인증 방식**: JWT 기반
- **세션 관리**: Valkey 세션 스토어
- **권한 관리**: RBAC (Role-Based Access Control)

#### 4.3.2 데이터 보호
- **전송 보안**: HTTPS/WSS (프로덕션)
- **SQL Injection 방지**: Prepared Statements
- **XSS 방지**: Input Sanitization
- **CORS**: 허용 도메인 화이트리스트

#### 4.3.3 Rate Limiting (P2 - 향후 구현)
- **API Rate Limit**: 100 req/min per IP
- **WebSocket Message Limit**: 50 msg/sec per connection

### 4.4 확장성

#### 4.4.1 수평 확장
- **Stateless 서버**: 모든 상태는 DB/Valkey에 저장
- **Load Balancing**: Kubernetes Service
- **Database Connection Pool**: 최대 100 연결

#### 4.4.2 캐싱 전략
- **게임방 상태**: Valkey (TTL 1시간)
- **퀴즈 데이터**: 메모리 캐시 (선택적)
- **CDN**: 정적 리소스 (프로덕션)

### 4.5 모니터링 및 로깅

#### 4.5.1 로깅
- **레벨**: INFO, WARN, ERROR, DEBUG
- **포맷**: 구조화된 로그 (JSON)
- **내용**:
  - 요청/응답 (method, path, status, latency)
  - 에러 스택 트레이스
  - 비즈니스 이벤트

#### 4.5.2 메트릭 (P2 - 향후 구현)
- **시스템 메트릭**: CPU, Memory, Disk
- **애플리케이션 메트릭**: QPS, Latency, Error Rate
- **비즈니스 메트릭**: Active Users, Active Games

---

## 5. 운영 요구사항

### 5.1 배포

#### 5.1.1 개발 환경
```bash
# 로컬 개발
go run ./cmd/server/main.go

# 빌드
go build -o bin/server.exe ./cmd/server
```

#### 5.1.2 프로덕션 배포
```bash
# Docker 이미지 빌드
docker build -t draw-and-guess-server:v1.0 .

# Kubernetes 배포
kubectl apply -f k8s/
```

#### 5.1.3 환경 변수
```env
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=<secret>
POSTGRES_DB=draw_guess_game
VALKEY_ADDR=localhost:6379
SERVER_PORT=8080
GIN_MODE=release  # 프로덕션
```

### 5.2 데이터베이스 관리

#### 5.2.1 마이그레이션
- **전략**: Schema 초기화 코드 실행
- **롤백**: 수동 SQL 스크립트
- **백업**: 일일 자동 백업 (Kubernetes CronJob)

#### 5.2.2 데이터 시딩
- **한국어 단어**: CSV 파일 import
- **샘플 퀴즈**: SQL INSERT 스크립트
- **테스트 사용자**: 개발 환경 전용

### 5.3 문서화

#### 5.3.1 API 문서
- **도구**: Apollo Studio
- **자동화**: gqlgen 스키마 자동 생성
- **업데이트**: 스키마 변경 시 자동

#### 5.3.2 개발자 문서
- **README.md**: 프로젝트 개요, 설치, 실행
- **APOLLO_STUDIO_GUIDE.md**: API 문서화 가이드
- **k8s/README.md**: 인프라 배포 가이드

---

## 6. 제약사항 및 가정

### 6.1 제약사항
- **한국어 전용**: 현재 한국어만 지원
- **인증 미구현**: v1.0에서는 인증 없음 (username만으로 식별)
- **페이지네이션 미구현**: 전체 조회 시 모든 데이터 반환
- **파일 업로드 미지원**: 아바타 이미지는 URL만 저장

### 6.2 가정
- **신뢰할 수 있는 클라이언트**: 악의적인 사용자 없음 (v1.0)
- **내부 네트워크**: Kubernetes 클러스터 내부 통신
- **소규모 사용자**: 초기 1,000명 이하
- **단일 리전**: 한국 서버만 운영

---

## 7. 성공 지표 (KPI)

### 7.1 기술 지표
- **서버 Uptime**: 99.9%
- **평균 응답 시간**: < 100ms
- **에러율**: < 1%
- **동시 접속자**: 100+

### 7.2 비즈니스 지표
- **일일 활성 사용자 (DAU)**: 50+
- **게임 완료율**: 80%
- **평균 게임 시간**: 5-10분
- **재방문율**: 60%

---

## 8. 로드맵

### Phase 1 (v1.0) - MVP ✅ 완료
- [x] 사용자 관리 (CRUD)
- [x] 게임 타입 정의 (wordchain, ox, qa)
- [x] 퀴즈 조회 (랜덤)
- [x] GraphQL API
- [x] WebSocket 실시간 통신
- [x] PostgreSQL + Valkey 인프라

### Phase 2 (v1.1) - 게임 플로우
- [ ] 게임방 생성/참가/퇴장
- [ ] 게임 시작/진행/종료
- [ ] 점수 시스템
- [ ] 타이머 관리
- [ ] 리더보드

### Phase 3 (v1.2) - 확장
- [ ] 인증/인가 (JWT)
- [ ] Rate Limiting
- [ ] Pagination
- [ ] 모니터링 (Prometheus + Grafana)
- [ ] 알림 시스템

### Phase 4 (v2.0) - 고급 기능
- [ ] 친구 시스템
- [ ] 채팅 필터링
- [ ] AI 기반 퀴즈 추천
- [ ] 멀티 리전 지원

---

## 9. 팀 및 리소스

### 9.1 개발 팀
- **백엔드 개발자**: 1명 (현재)
- **프론트엔드 개발자**: TBD
- **DevOps**: TBD

### 9.2 필요 리소스
- **개발 환경**: 로컬 머신 (Windows/Mac/Linux)
- **인프라**: Kubernetes 클러스터
  - PostgreSQL Pod: 1 core, 1GB RAM
  - Valkey Pod: 0.5 core, 512MB RAM
  - App Server: 2 cores, 2GB RAM

---

## 10. 리스크 및 완화 전략

### 10.1 기술 리스크
| 리스크 | 영향 | 확률 | 완화 전략 |
|--------|------|------|-----------|
| WebSocket 연결 끊김 | High | Medium | 자동 재연결 로직 |
| DB 성능 저하 | High | Low | Connection Pool 최적화, 인덱싱 |
| 캐시 데이터 불일치 | Medium | Medium | TTL 설정, Invalidation 로직 |

### 10.2 운영 리스크
| 리스크 | 영향 | 확률 | 완화 전략 |
|--------|------|------|-----------|
| 서버 다운타임 | High | Low | Health Check, Auto-restart |
| 데이터 손실 | High | Low | 자동 백업, Replication |
| 보안 취약점 | High | Medium | 정기 보안 업데이트 |

---

## 부록

### A. 용어 정의
- **끝말잇기**: 이전 단어의 마지막 글자로 시작하는 단어를 말하는 게임
- **Valkey**: Redis 호환 오픈소스 캐시/메시징 시스템
- **gqlgen**: Go용 GraphQL 서버 생성기 (schema-first)
- **StatefulSet**: Kubernetes에서 상태를 가진 애플리케이션을 관리하는 리소스

### B. 참고 자료
- [GraphQL 스펙](https://spec.graphql.org/)
- [gqlgen 문서](https://gqlgen.com/)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [12 Factor App](https://12factor.net/)

### C. 변경 이력
| 버전 | 날짜 | 변경 내용 | 작성자 |
|------|------|-----------|--------|
| 1.0 | 2025-12-25 | 초안 작성 | System |

---

**문서 승인**
- Product Owner: ___________
- Tech Lead: ___________
- 날짜: ___________
