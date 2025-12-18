
# Draw and Guess - 제품 명세서 (PRD)

## 📋 프로젝트 개요

**Draw and Guess**는 실시간 멀티플레이어 그림 맞추기 게임입니다. 한 명의 플레이어가 주어진 단어를 그림으로 표현하면, 다른 플레이어들이 채팅으로 정답을 맞추는 방식입니다.

### 핵심 가치
- 실시간 멀티플레이어 경험
- 직관적인 그리기 인터페이스
- 다국어 지원 (한국어, 영어, 일본어, 중국어)
- 빠른 게임 진행 (60초 라운드)

---

## 🎮 게임 규칙

### 기본 규칙
- **플레이어**: 방장(Drawer) 1명 + 참가자(Players) 여러 명
- **목표**: 
  - 3점을 먼저 획득하거나
  - 3라운드 종료 후 최고점자 우승
- **라운드 시간**: 각 라운드당 60초

### 게임 진행 흐름

#### 1. 방 생성 & 참가
- 방장이 방 생성 → `waiting` 상태
- 참가자들 입장 (최소 2명 이상 권장)
- 각 플레이어는 고유한 username으로 참여

#### 2. 게임 시작
- 방장이 게임 시작 버튼 클릭
- 상태가 `waiting` → `playing`으로 변경
- 첫 번째 라운드 시작

#### 3. 라운드 진행 (60초)
- **방장(Drawer) 화면**:
  - 랜덤 단어가 화면에 표시 (다국어 지원)
  - 캔버스에 그림 그리기 가능
  - 그리기 데이터는 실시간으로 다른 플레이어에게 전송
  
- **참가자(Players) 화면**:
  - 방장이 그리는 그림을 실시간으로 확인
  - 채팅창에 정답 입력
  - 정답/오답 여부 실시간 피드백

#### 4. 정답 처리
- 참가자가 채팅으로 정답 제출
- 서버에서 자동으로 정답 검증 (대소문자 구분 없음)
- **점수 부여**:
  - 정답자 (최초): **+2점**
  - 방장: **+1점**
- 정답이 나오면 즉시 다음 라운드로 진행

#### 5. 다음 라운드
- 이전 라운드 우승자가 다음 라운드의 Drawer가 됨
- 캔버스 자동 초기화
- 새로운 단어 선택 (이전에 사용하지 않은 단어)
- 타이머 60초로 리셋

#### 6. 게임 종료
- **종료 조건**:
  1. 누군가 **3점 달성** 시 즉시 종료
  2. **3라운드 종료** 후 최고점자 우승
- 최종 점수 및 랭킹 표시
- 상태가 `playing` → `finished`로 변경

---

## 📊 점수 시스템

| 조건 | 점수 | 비고 |
|------|------|------|
| 정답자 (최초) | +2점 | 가장 먼저 정답을 맞춘 플레이어 |
| Drawer (정답 나올 시) | +1점 | 그림을 잘 그려서 정답자가 나온 경우 |
| 우승 조건 | 3점 선취 | 또는 3라운드 종료 후 최고점 |

### 점수 계산 예시
```
라운드 1: Alice가 그림, Bob이 정답
  - Bob: +2점 (정답자)
  - Alice: +1점 (Drawer)

라운드 2: Bob이 그림, Charlie가 정답
  - Charlie: +2점 (정답자)
  - Bob: +1점 (Drawer)
  총점: Bob 3점, Charlie 2점, Alice 1점
  → Bob 우승 (3점 달성)
```

---

## ⚙️ 시스템 제약 사항

| 항목 | 값 | 설명 |
|------|-----|------|
| 라운드 시간 | 60초 | 각 라운드의 제한 시간 |
| 최대 라운드 | 3라운드 | 게임당 최대 진행 라운드 |
| 우승 점수 | 3점 | 이 점수에 도달하면 즉시 게임 종료 |
| 제시어 출처 | MongoDB | 동물 이름 15개 (다국어) |
| 방 유효 시간 | 1시간 | Valkey TTL (inactive room 자동 삭제) |

### 제시어 목록
사자, 호랑이, 코끼리, 기린, 펭귄, 돌고래, 독수리, 토끼, 여우, 판다, 캥거루, 코알라, 악어, 하마, 얼룩말

### 실시간 동기화
- **Valkey Pub/Sub**을 사용한 실시간 메시지 브로드캐스팅
- 모든 서버 인스턴스 간 동기화 (수평 확장 가능)
- 채널 구조:
  - `draw/{roomId}`: 그리기 데이터
  - `game/{roomId}`: 게임 상태 업데이트
  - `chat/{roomId}`: 채팅 메시지

---

## 🔧 기술 스펙

### REST API 엔드포인트

Base URL: `http://localhost:8080/app`

#### 1. 게임 방 조회
```http
GET /game/rooms?id={roomId}
```
- **Query Parameters**:
  - `id` (optional): 특정 방 UUID
- **Response**:
  ```json
  {
    "status": true,
    "message": "success",
    "result": [{
      "uuid": "abc-123",
      "is_active": true,
      "drawer_user": "Alice",
      "players": [{"username": "Bob", "score": 2, "attempts": 1}],
      "current_word": "lion",
      "current_word_translations": {
        "ko": "사자",
        "en": "lion",
        "ja": "ライオン",
        "zh": "狮子"
      },
      "round_number": 2,
      "time_left": 45,
      "game_status": "playing",
      "max_rounds": 3,
      "winning_score": 3
    }]
  }
  ```

#### 2. 게임 방 생성
```http
POST /game/room
Content-Type: application/x-www-form-urlencoded

ldap_user=Alice
```
- **Form Data**:
  - `ldap_user`: 방장의 username
- **Response**:
  ```json
  {
    "status": true,
    "result": {
      "room_id": "uuid-here",
      "current_word": "lion",
      "current_word_translations": {...}
    }
  }
  ```

#### 3. 게임 방 참가
```http
POST /game/room/:id/join
Content-Type: application/x-www-form-urlencoded

username=Bob
```
- **URL Parameters**: `id` - 방 UUID
- **Form Data**: `username` - 참가자 이름

#### 4. 게임 시작
```http
POST /game/room/:id/start
```
- **권한**: 방장만 실행 가능
- **효과**: `game_status`를 `waiting` → `playing`으로 변경

#### 5. 정답 제출 (채팅)
```http
POST /game/room/:id/chat
Content-Type: application/json

{
  "username": "Bob",
  "message": "사자"
}
```
- **Response**:
  ```json
  {
    "status": true,
    "result": {
      "is_correct": true,
      "message": "정답입니다!"
    }
  }
  ```

#### 6. 방 삭제
```http
DELETE /game/room/:id
```
- **효과**: Valkey에서 방 데이터 삭제

---

### WebSocket 프로토콜

#### 연결
```javascript
ws://localhost:8080/app/ws
```

#### 메시지 형식
```typescript
// 클라이언트 → 서버
interface WebSocketRequest {
  type: 'subscribe' | 'unsubscribe' | 'message';
  channel: string;  // 예: "game/room-uuid"
  data?: any;       // message 타입인 경우
}

// 서버 → 클라이언트
interface WebSocketResponse {
  type: string;
  channel: string;
  data: any;
}
```

#### 채널 구조
| 채널 | 용도 | 메시지 예시 |
|------|------|-------------|
| `draw/{roomId}` | 그리기 데이터 전송 | `{action: "draw", points: [...], color: "#000"}` |
| `game/{roomId}` | 게임 상태 업데이트 | `{type: "timer_update", time_left: 45}` |
| `chat/{roomId}` | 채팅 메시지 | `{username: "Bob", message: "hello", type: "chat"}` |

#### 구독 예시
```javascript
// 게임 채널 구독
ws.send(JSON.stringify({
  type: 'subscribe',
  channel: 'game/abc-123'
}));

// 그리기 데이터 전송
ws.send(JSON.stringify({
  type: 'message',
  channel: 'draw/abc-123',
  data: {
    action: 'draw',
    points: [{x: 100, y: 200}, {x: 105, y: 205}],
    color: '#FF0000',
    width: 5
  }
}));
```

#### Valkey Pub/Sub 통합
- 모든 WebSocket 메시지는 Valkey Pub/Sub으로 중계
- 여러 서버 인스턴스 간 메시지 동기화
- 수평 확장 시에도 모든 클라이언트가 동일한 메시지 수신

---

### 주요 데이터 구조

#### GameRoom
```go
type GameRoom struct {
    UUID                    string              // 방 고유 ID
    IsActive                bool                // 활성 상태
    DrawerUser              string              // 현재 그림 그리는 사람
    RoomCreator             string              // 방 생성자
    LastRoundWinner         string              // 이전 라운드 우승자
    Players                 []Player            // 참가자 목록
    CurrentWord             string              // 현재 제시어 (영어)
    CurrentWordTranslations map[string]string   // 다국어 번역
    RoundNumber             int                 // 현재 라운드 (1-3)
    TimeLeft                int                 // 남은 시간 (초)
    GameStatus              string              // waiting | playing | finished
    UsedWords               []string            // 사용된 단어들
    MaxRounds               int                 // 최대 라운드 (3)
    WinningScore            int                 // 우승 점수 (3)
    GameType                GameType            // "guess"
    CreatedAt               time.Time
    UpdatedAt               time.Time
}
```

#### Player
```go
type Player struct {
    Username string  // 플레이어 이름
    Score    int     // 현재 점수
    Attempts int     // 이번 라운드 제출 횟수
}
```

#### GameTopic (MongoDB)
```go
type GameTopic struct {
    Canonical    string              // 기준 단어 (영어)
    Translations map[string]string   // {"ko": "사자", "en": "lion", ...}
}
```

---

## 🌐 다국어 지원

지원 언어: 한국어(ko), 영어(en), 일본어(ja), 중국어(zh)

### 제시어 다국어 목록
| 영어 (canonical) | 한국어 | 일본어 | 중국어 |
|-----------------|--------|--------|--------|
| lion | 사자 | ライオン | 狮子 |
| tiger | 호랑이 | トラ | 老虎 |
| elephant | 코끼리 | ゾウ | 大象 |
| giraffe | 기린 | キリン | 长颈鹿 |
| penguin | 펭귄 | ペンギン | 企鹅 |
| dolphin | 돌고래 | イルカ | 海豚 |
| eagle | 독수리 | ワシ | 鹰 |
| rabbit | 토끼 | ウサギ | 兔子 |
| fox | 여우 | キツネ | 狐狸 |
| panda | 판다 | パンダ | 熊猫 |
| kangaroo | 캥거루 | カンガルー | 袋鼠 |
| koala | 코알라 | コアラ | 考拉 |
| crocodile | 악어 | ワニ | 鳄鱼 |
| hippopotamus | 하마 | カバ | 河马 |
| zebra | 얼룩말 | シマウマ | 斑马 |

---

## 🏗️ 아키텍처 개요

### 기술 스택
- **Backend**: Go (Gin Framework)
- **Frontend**: React + TypeScript + Vite
- **Database**: MongoDB (퀴즈 데이터)
- **Cache**: Valkey (게임 상태, Pub/Sub)
- **Protocol**: REST API + WebSocket

### 배포 구조
```
┌─────────────┐
│   Client    │
│ (React App) │
└──────┬──────┘
       │ HTTP/WS
       ▼
┌─────────────┐      ┌─────────────┐
│ Go Server 1 │◄────►│   Valkey    │
└─────────────┘      │  (Pub/Sub)  │
       │             └─────────────┘
┌─────────────┐             ▲
│ Go Server 2 │─────────────┘
└─────────────┘
       │
       ▼
┌─────────────┐
│   MongoDB   │
└─────────────┘
```