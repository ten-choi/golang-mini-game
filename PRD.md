
# Draw and Guess - 게임 룰

## 기본 규칙
- **플레이어**: 방장 1명 + 참가자 여러 명
- **목표**: 3점을 먼저 획득하거나 3라운드 후 최고점자 우승

## 게임 진행

### 1. 방 생성 & 참가
- 방장이 방 생성 → 대기 상태
- 참가자들 입장 (최소 2명 이상)

### 2. 라운드 시작 (60초)
- 방장 화면에만 **랜덤 동물 이름** 표시
- 방장만 그림 그리기 가능
- 참가자들은 그림을 보며 추측

### 3. 정답 제출
- 참가자는 채팅창에 정답을 입력
- 가장 먼저 맞힌 참가자: **+2점**
- 방장: **+1점**
- 정답/오답 여부가 채팅창에 실시간 표시

### 4. 채팅 시스템
- 모든 플레이어가 실시간 채팅 가능
- 채팅으로 정답 제출 (자동 검증)
- 오답/정답 여부 실시간 피드백
- Valkey Pub/Sub으로 모든 컨테이너 간 메시지 동기화

### 5. 다음 라운드
- 정답 공개 후 캔버스 초기화
- 새로운 제시어로 다음 라운드 시작

### 6. 게임 종료
- 누군가 **3점 달성** 시 즉시 종료
- 또는 **3라운드 종료** 후 최고점자 우승
- 결과 화면: 최종 점수 & 랭킹

## 점수 시스템
| 조건 | 점수 |
|------|------|
| 정답자 (최초) | +2점 |
| 방장 (정답 나올 시) | +1점 |
| 우승 조건 | 3점 선취 또는 3라운드 최고점 |

## 제약 사항
- 라운드 시간: **60초**
- 제시어: **동물 이름** (15개 중 랜덤)
- Valkey Pub/Sub으로 그리기/채팅 실시간 동기화

## 기술 스펙

### API
- `GET /app/game/rooms` - 방 목록 / 단일 방 조회(id 쿼리)
- `POST /app/game/room` - 방 생성 (`ldap_user`)
- `POST /app/game/room/:id/join` - 방 참가 (`ldap_user`)
- `POST /app/game/room/:id/leave` - 방 나가기 (`ldap_user`)
- `POST /app/game/room/:id/start` - 게임 시작 (방장)
- `POST /app/game/room/:id/answer` - 정답 제출 (`ldap_user`, `answer`)
- `POST /app/game/room/:id/chat` - 채팅 메시지/피드백 (`ldap_user`, `message`)
- `PATCH /app/game/room/:id` - 그림 데이터 업데이트 (스트로크 JSON)
- `DELETE /app/game/room/:id` - 방 삭제

### WebSocket
- 엔드포인트: `ws://<server>/app/ws`
- 메시지 형식: `{ "type": "subscribe|unsubscribe|message", "channel": "chat/<roomId>", "data": {...} }`
- 채널 구분
  - `draw/{roomId}`: 브러시 스트로크/캔버스 정보 공유
  - `game/{roomId}`: 타이머, 라운드, 점수, 게임 상태 브로드캐스트
  - `chat/{roomId}`: 채팅/정답 피드백
- 서버는 Valkey Pub/Sub과 연동해 모든 구독자에게 메시지를 중계

### 데이터
```
GameRoom:
  - DrawerUser: 방장
  - Players: [{name, score, attempts}]
  - CurrentWord: 현재 제시어
  - RoundNumber: 현재 라운드 (1-3)
  - TimeLeft: 남은 시간 (초)
  - Status: waiting | playing | finished
```

## 동물 리스트
사자, 호랑이, 코끼리, 기린, 펭귄, 돌고래, 독수리, 토끼, 여우, 판다, 캥거루, 코알라, 악어, 하마, 얼룩말