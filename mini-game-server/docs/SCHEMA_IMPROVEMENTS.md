# GraphQL Schema Improvements for Apollo Client

이 문서는 Apollo Client 개발자를 위해 GraphQL Schema에 추가된 개선 사항을 설명합니다.

## 📋 개선 요약

### 1. Query 개선사항

#### ✅ Pagination 및 Filtering 추가

**`users` Query**
```graphql
users(
  limit: Int          # 기본값: 50, 최대: 100
  offset: Int         # 기본값: 0
  search: String      # 닉네임 부분 검색 (대소문자 구분 없음)
  minLevel: Int       # 최소 레벨 필터
): [User!]!
```

**사용 예시:**
```graphql
# 페이지네이션
query {
  users(limit: 20, offset: 40) {
    UserName
    level
  }
}

# 검색
query {
  users(search: "user", minLevel: 10) {
    UserName
    level
  }
}
```

**`gameRooms` Query**
```graphql
gameRooms(
  gameType: GameType         # 게임 타입 필터
  status: GameStatus         # 상태 필터 (WAITING, PLAYING)
  includePrivate: Boolean    # 비공개방 포함 여부 (기본: false)
  hasSpace: Boolean          # 입장 가능한 방만 (기본: true)
  limit: Int                 # 반환 개수 제한
  sortBy: String             # 정렬 (CREATED_DESC, CREATED_ASC, userS_DESC)
): [GameRoom!]!
```

**사용 예시:**
```graphql
# 입장 가능한 OX 게임방만 조회
query {
  gameRooms(
    gameType: OX
    status: WAITING
    hasSpace: true
    sortBy: "CREATED_DESC"
  ) {
    id
    name
    users { UserName }
  }
}
```

#### ✅ 새로운 Query 추가

**`myCurrentRoom` - 현재 참여 중인 게임방 조회**
```graphql
query {
  myCurrentRoom(UserName: "user123") {
    id
    name
    gameType
    status
    currentRound
    users {
      UserName
      score
    }
  }
}
```
- **사용 시점**: 앱 재시작 시 진행 중인 게임 복구, 다른 기기에서 접속 시

**`myInvitations` - 받은 초대 목록 조회**
```graphql
query {
  myInvitations(userId: "user123", status: PENDING) {
    id
    room {
      name
      gameType
    }
    inviter {
      UserName
    }
    createdAt
    expiresAt
  }
}
```
- **사용 시점**: 초대 알림 화면, 대기 중인 초대 개수 표시

**`invitation` - 특정 초대 정보 조회**
```graphql
query {
  invitation(id: "inv-123") {
    id
    status
    room {
      name
      users { UserName }
    }
    inviter { UserName }
  }
}
```
- **사용 시점**: 초대 링크를 통한 정보 확인

---

### 2. Mutation 개선사항

#### ✅ 반환 타입 개선 (Boolean → 구체적 타입)

**`setReady` Mutation**
```graphql
# 이전: Boolean 반환
setReady(roomId: ID!, UserName: String!, ready: Boolean!): Boolean!

# 개선: GameRoom 반환 (Apollo Cache 자동 업데이트)
setReady(roomId: ID!, UserName: String!, ready: Boolean!): GameRoom!
```

**Apollo Client 장점:**
```typescript
const [setReady] = useMutation(SET_READY, {
  // update 함수 불필요! Apollo가 자동으로 cache 업데이트
});
```

**`submitAnswer` Mutation - AnswerResult 타입 추가**
```graphql
type AnswerResult {
  success: Boolean!           # 제출 성공 여부
  isCorrect: Boolean!         # 정답 여부 (본인에게만)
  earnedScore: Int!           # 획득한 점수
  totalScore: Int!            # 현재 총 점수
  correctAnswer: String       # 정답 (즉시 피드백용)
  explanation: String         # 설명
}

mutation {
  submitAnswer(
    roomId: ID!
    UserName: String!
    answer: String!
  ): AnswerResult!  # Boolean 대신 상세 정보 반환
}
```

**Optimistic UI 예시:**
```typescript
const [submitAnswer] = useMutation(SUBMIT_ANSWER, {
  optimisticResponse: {
    submitAnswer: {
      __typename: 'AnswerResult',
      success: true,
      isCorrect: true,  // 낙관적으로 정답 가정
      earnedScore: 100,
      totalScore: myCurrentScore + 100,
      correctAnswer: null,
      explanation: null,
    },
  },
});
```

#### ✅ Batch Mutation 추가

**`inviteUsers` - 여러 사용자를 한 번에 초대**
```graphql
mutation {
  inviteUsers(
    roomId: "room-123"
    inviteeUserNames: ["friend1", "friend2", "friend3"]
  ) {
    id
    invitee { UserName }
    status
  }
}
```

**장점:**
- 단일 네트워크 요청으로 여러 초대 처리
- 성능 향상 및 네트워크 비용 절감

---

### 3. Subscription 개선사항

#### ✅ 플레이어별 개인화 이벤트

**`myEvents` - 본인에게만 전송되는 이벤트**
```graphql
subscription {
  myEvents(UserName: "user123") {
    type    # CORRECT_ANSWER, WRONG_ANSWER, LEVEL_UP, etc.
    data
    timestamp
  }
}
```

**사용 예시:**
- 정답/오답 개인 알림
- 레벨 업 알림
- 업적 해금 알림

**`invitationReceived` - 초대 알림**
```graphql
subscription {
  invitationReceived(userId: "user123") {
    id
    room {
      name
      gameType
    }
    inviter {
      UserName
    }
    expiresAt
  }
}
```

**사용 예시:**
- 실시간 초대 알림
- 초대 배지 업데이트

#### ✅ 연결 상태 모니터링

**`userConnectionStatus` - 플레이어 온라인/오프라인 상태**
```graphql
type UserConnection {
  UserName: String!
  isConnected: Boolean!
  lastSeen: Time!
  ping: Int  # 네트워크 지연시간 (ms)
}

subscription {
  userConnectionStatus(roomId: "room-123") {
    UserName
    isConnected
    lastSeen
    ping
  }
}
```

**사용 예시:**
- 플레이어 온라인 상태 표시
- 네트워크 품질 모니터링
- 재연결 처리

---

### 4. 새로운 타입 정의

#### ✅ Result Types

**`AnswerResult`** - 답변 제출 결과
```graphql
type AnswerResult {
  success: Boolean!
  isCorrect: Boolean!
  earnedScore: Int!
  totalScore: Int!
  correctAnswer: String
  explanation: String
}
```

**`UserConnection`** - 플레이어 연결 상태
```graphql
type UserConnection {
  UserName: String!
  isConnected: Boolean!
  lastSeen: Time!
  ping: Int
}
```

---

## 🎯 Apollo Client 활용 팁

### 1. Fragment 사용 (클라이언트에 이미 적용됨)

```typescript
// Fragments로 코드 재사용성 향상
export const user_FIELDS = `
  fragment UserFields on User {
    UserName
    score
    isReady
  }
`;

export const GAME_ROOM_DETAIL_FIELDS = `
  fragment GameRoomDetailFields on GameRoom {
    ...GameRoomFields
    users {
      ...UserFields
    }
  }
  ${GAME_ROOM_FIELDS}
  ${user_FIELDS}
`;
```

### 2. Optimistic UI

```typescript
// submitAnswer에서 즉각적인 UI 업데이트
const [submitAnswer] = useMutation(SUBMIT_ANSWER, {
  optimisticResponse: {
    submitAnswer: {
      __typename: 'AnswerResult',
      success: true,
      isCorrect: true,
      earnedScore: 100,
      totalScore: currentScore + 100,
    },
  },
  // Apollo가 자동으로 cache 업데이트
});
```

### 3. Cache 관리

```typescript
// setReady mutation - 자동 cache 업데이트
const [setReady] = useMutation(SET_READY);

// GameRoom이 반환되므로 Apollo가 자동으로
// GET_GAME_ROOM 쿼리 결과를 업데이트함
```

### 4. Subscription 활용

```typescript
// 실시간 게임방 업데이트
const { data } = useSubscription(GAME_ROOM_UPDATED, {
  variables: { roomId },
  onData: ({ data }) => {
    // Apollo가 자동으로 cache 업데이트
    // 별도 처리 불필요
  },
});

// 개인 이벤트 수신
const { data } = useSubscription(MY_EVENTS, {
  variables: { UserName },
  onData: ({ data }) => {
    const event = data.data.myEvents;
    if (event.type === 'CORRECT_ANSWER') {
      showSuccessAnimation();
    }
  },
});
```

---

## 📊 성능 최적화

### 1. Pagination으로 대량 데이터 처리
```typescript
// 무한 스크롤 구현
const { data, fetchMore } = useQuery(GET_USERS, {
  variables: { limit: 20, offset: 0 },
});

const loadMore = () => {
  fetchMore({
    variables: { offset: data.users.length },
  });
};
```

### 2. Batch Mutation으로 네트워크 요청 감소
```typescript
// 한 번에 여러 친구 초대
const [inviteUsers] = useMutation(INVITE_USERS);

inviteUsers({
  variables: {
    roomId,
    inviteeUserNames: selectedFriends.map(f => f.UserName),
  },
});
```

### 3. Subscription 필터링
```typescript
// 필요한 이벤트만 구독
const { data } = useSubscription(MY_EVENTS, {
  variables: { UserName: currentUser },
  skip: !isInGame,  // 게임 중일 때만 구독
});
```

---

## 🔧 구현 필요 사항 (서버 측)

현재 Schema는 업데이트되었지만, 다음 resolver 구현이 필요합니다:

1. **Query Resolvers:**
   - `users` - pagination 및 filtering
   - `gameRooms` - 추가 필터 옵션
   - `myCurrentRoom` - 사용자의 현재 게임방 조회
   - `myInvitations` - 초대 목록 조회
   - `invitation` - 특정 초대 조회

2. **Mutation Resolvers:**
   - `setReady` - GameRoom 반환하도록 수정
   - `submitAnswer` - AnswerResult 반환하도록 수정
   - `inviteUsers` - 배치 초대 처리

3. **Subscription Resolvers:**
   - `myEvents` - 플레이어별 이벤트
   - `invitationReceived` - 초대 알림
   - `userConnectionStatus` - 연결 상태

---

## 📝 마이그레이션 가이드

### 기존 코드 업데이트 필요 사항:

#### 1. setReady mutation
```typescript
// Before
const [setReady] = useMutation(SET_READY);
// returns: Boolean

// After
const [setReady] = useMutation(SET_READY);
// returns: GameRoom (자동 cache 업데이트)
```

#### 2. submitAnswer mutation
```typescript
// Before
const [submitAnswer] = useMutation(SUBMIT_ANSWER);
// returns: Boolean

// After
const [submitAnswer] = useMutation(SUBMIT_ANSWER);
// returns: AnswerResult { success, isCorrect, earnedScore, ... }

// 사용 예시:
const result = await submitAnswer({ variables: { ... } });
if (result.data.submitAnswer.isCorrect) {
  showCorrectAnimation();
  updateScore(result.data.submitAnswer.totalScore);
}
```

#### 3. gameRooms query에 필터 추가
```typescript
// Before
const { data } = useQuery(GET_GAME_ROOMS, {
  variables: { gameType: 'OX' },
});

// After (옵션)
const { data } = useQuery(GET_GAME_ROOMS, {
  variables: {
    gameType: 'OX',
    status: 'WAITING',
    hasSpace: true,
    sortBy: 'CREATED_DESC',
    limit: 20,
  },
});
```

---

## 🚀 다음 단계

1. **서버 Resolver 구현** - 새로운 쿼리/뮤테이션/구독의 실제 로직 구현
2. **테스트 작성** - 새로운 기능에 대한 단위/통합 테스트
3. **클라이언트 UI 업데이트** - 새로운 기능 활용한 UI 개선
4. **성능 모니터링** - Pagination, Subscription 성능 측정

---

## 📚 참고 자료

- [Apollo Client Documentation](https://www.apollographql.com/docs/react/)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [Optimistic UI Guide](https://www.apollographql.com/docs/react/performance/optimistic-ui/)
- [Apollo Cache Management](https://www.apollographql.com/docs/react/caching/cache-configuration/)
