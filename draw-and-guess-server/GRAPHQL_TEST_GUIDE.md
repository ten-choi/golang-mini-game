# GraphQL API 테스트 가이드

## GraphQL Playground 접속
http://localhost:8080/graphql

---

## 1. User API 테스트

### 사용자 생성
```graphql
mutation {
  createUser(nickname: "player1") {
    id
    nickname
    profileImage
    money
    createdAt
    updatedAt
  }
}
```

### 사용자 조회
```graphql
query {
  user(id: "YOUR_USER_ID") {
    id
    nickname
    profileImage
    money
    createdAt
  }
}
```

### 사용자 업데이트
```graphql
mutation {
  updateUser(
    id: "YOUR_USER_ID"
    nickname: "newNickname"
    profileImage: "https://example.com/avatar.png"
  ) {
    id
    nickname
    profileImage
    updatedAt
  }
}
```

---

## 2. Game Room API 테스트

### 게임방 생성
```graphql
mutation {
  createGameRoom(
    hostUsername: "player1"
    maxRounds: 3
  ) {
    id
    name
    hostId
    players {
      userId
      username
      score
      isReady
    }
    maxPlayers
    currentRound
    totalRounds
    status
    createdAt
  }
}
```

### 모든 게임방 조회
```graphql
query {
  gameRooms {
    id
    name
    hostId
    players {
      username
      score
    }
    status
    currentRound
    totalRounds
  }
}
```

### 특정 게임방 조회
```graphql
query {
  gameRoom(id: "YOUR_ROOM_ID") {
    id
    name
    hostId
    players {
      username
      score
      isReady
    }
    status
    currentRound
    totalRounds
    createdAt
  }
}
```

### 게임방 참가
```graphql
mutation {
  joinGameRoom(
    roomId: "YOUR_ROOM_ID"
    username: "player2"
  ) {
    id
    players {
      username
      score
    }
  }
}
```

### 게임방 나가기
```graphql
mutation {
  leaveGameRoom(
    roomId: "YOUR_ROOM_ID"
    username: "player2"
  ) {
    id
    players {
      username
    }
  }
}
```

### 게임 시작
```graphql
mutation {
  startGame(roomId: "YOUR_ROOM_ID") {
    id
    status
    currentRound
    players {
      username
      score
    }
  }
}
```

### 게임방 삭제
```graphql
mutation {
  deleteGameRoom(roomId: "YOUR_ROOM_ID")
}
```

---

## 3. Subscription (실시간 이벤트 구독)

### 게임 이벤트 구독
```graphql
subscription {
  gameEvents(roomId: "YOUR_ROOM_ID") {
    type
    roomId
    data
    timestamp
  }
}
```

**이벤트 타입:**
- `player_joined` - 플레이어 입장
- `player_left` - 플레이어 퇴장
- `round_start` - 라운드 시작
- `round_end` - 라운드 종료
- `timer_update` - 타이머 업데이트
- `game_end` - 게임 종료

### 채팅 메시지 구독
```graphql
subscription {
  chatMessages(roomId: "YOUR_ROOM_ID") {
    roomId
    userId
    username
    message
    timestamp
  }
}
```

### 그림 그리기 구독
```graphql
subscription {
  drawingUpdates(roomId: "YOUR_ROOM_ID") {
    roomId
    action
    data
    timestamp
  }
}
```

**그리기 액션:**
- `draw` - 그림 그리기 (points, color, width 포함)
- `clear` - 캔버스 지우기
- `undo` - 마지막 액션 취소

---

## 4. WebSocket 정보 조회

### WebSocket 연결 정보
```graphql
query {
  webSocketInfo {
    url
    description
    channels
  }
}
```

**결과:**
```json
{
  "url": "ws://localhost:8080/app/ws",
  "description": "WebSocket endpoint for real-time game communication",
  "channels": [
    "game/room-{roomId} - Game events",
    "chat/room-{roomId} - Chat messages",
    "draw/room-{roomId} - Drawing data"
  ]
}
```

---

## 5. GraphQL Subscription vs 직접 WebSocket 비교

### GraphQL Subscription 방식 (권장)
```javascript
// Apollo Client 사용
import { ApolloClient, InMemoryCache, split, HttpLink } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import { getMainDefinition } from '@apollo/client/utilities';

// HTTP link for queries and mutations
const httpLink = new HttpLink({
  uri: 'http://localhost:8080/graphql'
});

// WebSocket link for subscriptions
const wsLink = new GraphQLWsLink(createClient({
  url: 'ws://localhost:8080/graphql',
}));

// Split based on operation type
const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return (
      definition.kind === 'OperationDefinition' &&
      definition.operation === 'subscription'
    );
  },
  wsLink,
  httpLink,
);

const client = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache()
});

// 사용 예시
const GAME_EVENTS_SUBSCRIPTION = gql`
  subscription OnGameEvent($roomId: String!) {
    gameEvents(roomId: $roomId) {
      type
      data
      timestamp
    }
  }
`;

client.subscribe({
  query: GAME_EVENTS_SUBSCRIPTION,
  variables: { roomId: 'YOUR_ROOM_ID' }
}).subscribe({
  next: ({ data }) => console.log('Game event:', data),
  error: (err) => console.error('Error:', err)
});
```

### 직접 WebSocket 방식 (레거시)
```javascript
// 기존 방식 (여전히 사용 가능)
const ws = new WebSocket('ws://localhost:8080/app/ws');
ws.send(JSON.stringify({
  type: 'subscribe',
  channel: 'game/room-YOUR_ROOM_ID'
}));
```

**GraphQL Subscription의 장점:**
- ✅ 타입 안정성 (TypeScript 지원)
- ✅ 자동 재연결
- ✅ 쿼리 언어 사용 (필요한 필드만 선택)
- ✅ Apollo Client와 완벽한 통합
- ✅ 에러 처리 표준화

---

## 6. 직접 WebSocket 연결 테스트 (레거시)

### JavaScript에서 WebSocket 연결
```javascript
// WebSocket 연결
const ws = new WebSocket('ws://localhost:8080/app/ws');

ws.onopen = () => {
  console.log('Connected to WebSocket');
  
  // 게임 채널 구독
  ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'game/room-YOUR_ROOM_ID'
  }));
  
  // 채팅 채널 구독
  ws.send(JSON.stringify({
    type: 'subscribe',
    channel: 'chat/room-YOUR_ROOM_ID'
  }));
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
};

// 채팅 메시지 전송
const sendChat = (message) => {
  ws.send(JSON.stringify({
    type: 'message',
    channel: 'chat/room-YOUR_ROOM_ID',
    data: {
      room_id: 'YOUR_ROOM_ID',
      user_id: 'user-001',
      username: 'player1',
      message: message
    }
  }));
};

// 그림 그리기 데이터 전송
const sendDrawing = (points, color, width) => {
  ws.send(JSON.stringify({
    type: 'message',
    channel: 'draw/room-YOUR_ROOM_ID',
    data: {
      room_id: 'YOUR_ROOM_ID',
      action: 'draw',
      points: points,
      color: color,
      width: width
    }
  }));
};
```

---

## 7. 전체 플로우 테스트 (GraphQL Subscription 사용)

### Step 1: 사용자 생성
```graphql
mutation {
  createUser(nickname: "player1") {
    id
    nickname
  }
}
```

### Step 2: 게임방 생성
```graphql
mutation {
  createGameRoom(hostUsername: "player1", maxRounds: 3) {
    id
    status
  }
}
```

### Step 3: 다른 플레이어 참가
```graphql
mutation {
  joinGameRoom(roomId: "ROOM_ID", username: "player2") {
    id
    players {
      username
    }
  }
}
```

### Step 4: Subscription으로 실시간 이벤트 구독
```graphql
# GraphQL Playground에서 실행
subscription {
  gameEvents(roomId: "ROOM_ID") {
    type
    data
    timestamp
  }
}

subscription {
  chatMessages(roomId: "ROOM_ID") {
    username
    message
    timestamp
  }
}
```

### Step 5: 게임 시작
```graphql
mutation {
  startGame(roomId: "ROOM_ID") {
    id
    status
    currentRound
  }
}
```

### Step 6: 실시간 채팅/그림 그리기
```javascript
// 채팅
ws.send(JSON.stringify({
  type: 'message',
  channel: 'chat/room-ROOM_ID',
  data: {
    room_id: 'ROOM_ID',
    user_id: 'player1',
    username: 'player1',
    message: '안녕하세요!'
  }
}));

// 그림 그리기
ws.send(JSON.stringify({
  type: 'message',
  channel: 'draw/room-ROOM_ID',
  data: {
    room_id: 'ROOM_ID',
    action: 'draw',
    points: [{x: 100, y: 100}, {x: 150, y: 150}],
    color: '#FF0000',
    width: 5
  }
}));
```

---

## 8. 에러 처리

모든 GraphQL 요청은 다음과 같은 에러 형식을 반환합니다:

```json
{
  "errors": [
    {
      "message": "user not found",
      "path": ["user"]
    }
  ]
}
```

---

## 9. REST API도 여전히 사용 가능

GraphQL 외에도 기존 REST API도 사용 가능합니다:

- `GET /app/game/rooms` - 게임방 목록
- `POST /app/game/room` - 게임방 생성
- `GET /app/user/:id` - 사용자 조회
- `POST /app/user` - 사용자 생성
- `GET /app/ws` - WebSocket 연결

---

## 10. 프론트엔드 통합 예제

### React + Apollo Client (Query, Mutation, Subscription 모두 포함)
```javascript
import { ApolloClient, InMemoryCache, split, HttpLink, gql, useSubscription } from '@apollo/client';
import { GraphQLWsLink } from '@apollo/client/link/subscriptions';
import { createClient } from 'graphql-ws';
import { getMainDefinition } from '@apollo/client/utilities';

// Apollo Client 설정
const httpLink = new HttpLink({
  uri: 'http://localhost:8080/graphql'
});

const wsLink = new GraphQLWsLink(createClient({
  url: 'ws://localhost:8080/graphql',
}));

const splitLink = split(
  ({ query }) => {
    const definition = getMainDefinition(query);
    return definition.kind === 'OperationDefinition' && definition.operation === 'subscription';
  },
  wsLink,
  httpLink,
);

const client = new ApolloClient({
  link: splitLink,
  cache: new InMemoryCache()
});

// Mutation: 게임방 생성
const CREATE_ROOM = gql`
  mutation CreateRoom($hostUsername: String!, $maxRounds: Int!) {
    createGameRoom(hostUsername: $hostUsername, maxRounds: $maxRounds) {
      id
      status
    }
  }
`;

client.mutate({
  mutation: CREATE_ROOM,
  variables: { hostUsername: 'player1', maxRounds: 3 }
});

// Subscription: 게임 이벤트 구독 (React Hook)
const GAME_EVENTS = gql`
  subscription GameEvents($roomId: String!) {
    gameEvents(roomId: $roomId) {
      type
      data
      timestamp
    }
  }
`;

function GameRoom({ roomId }) {
  const { data, loading } = useSubscription(GAME_EVENTS, {
    variables: { roomId }
  });

  if (loading) return <p>Connecting...</p>;
  if (data) {
    console.log('Game event:', data.gameEvents);
    // UI 업데이트
  }

  return <div>Game Room</div>;
}
```

### Vanilla JavaScript Fetch
```javascript
fetch('http://localhost:8080/graphql', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    query: `
      mutation {
        createGameRoom(hostUsername: "player1", maxRounds: 3) {
          id
          status
        }
      }
    `
  })
})
.then(res => res.json())
.then(data => console.log(data));
```

---

## 테스트 시작!

1. 서버가 실행 중인지 확인: `http://localhost:8080/graphql`
2. GraphQL Playground에서 위 쿼리들을 복사해서 테스트
3. WebSocket 연결은 브라우저 콘솔에서 JavaScript 코드로 테스트

**Happy Testing! 🎮**
