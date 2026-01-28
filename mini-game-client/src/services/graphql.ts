/**
 * GraphQL Client Service
 * Aligned with q-connect-server GraphQL Schema
 */

import { GraphQLClient } from 'graphql-request';
import { env } from '../config/env';

// GraphQL endpoint from environment
const GRAPHQL_ENDPOINT = env.graphqlUrl;

// Create GraphQL client
export const graphqlClient = new GraphQLClient(GRAPHQL_ENDPOINT, {
  headers: {
    'Content-Type': 'application/json',
  },
});

// ============================================
// GraphQL Fragments (for code reuse & Apollo Cache)
// ============================================

export const USER_FIELDS = `
  fragment UserFields on User {
    id
    hangeId
    name
    avatarUrl
    level
    credit
    guildId
    createdAt
    updatedAt
  }
`;

export const GAME_USER_FIELDS = `
  fragment GameUserFields on GameUser {
    userId
    name
    score
    isReady
  }
`;

export const GAME_ROOM_FIELDS = `
  fragment GameRoomFields on GameRoom {
    id
    name
    gameType
    status
    currentRound
    totalRounds
    roundTimeLimit
    maxUsers
    hostUserId
    isPrivate
    createdAt
  }
`;

export const GAME_ROOM_DETAIL_FIELDS = `
  fragment GameRoomDetailFields on GameRoom {
    ...GameRoomFields
    usedQuizIds
    users {
      ...GameUserFields
    }
    wordchainLastWord
    wordchainUsedWords
    currentTurnUserId
    wordchainTurnStartTime
  }
  ${GAME_ROOM_FIELDS}
  ${GAME_USER_FIELDS}
`;

export const OX_QUIZ_FIELDS = `
  fragment OXQuizFields on OXQuiz {
    id
    category
    difficulty
    question
    answer
    explanation
    usageCount
    isActive
    createdAt
    updatedAt
  }
`;

export const GENERAL_QUIZ_FIELDS = `
  fragment GeneralQuizFields on GeneralQuiz {
    id
    category
    difficulty
    question
    options
    answer
    explanation
    imageUrl
    usageCount
    isActive
    createdAt
    updatedAt
  }
`;

// ============================================
// GraphQL Queries
// ============================================

// User Queries
export const GET_USER_BY_HANGE_ID = `
  ${USER_FIELDS}
  query GetUserByHangeId($hangeId: String!) {
    userByHangeId(hangeId: $hangeId) {
      ...UserFields
    }
  }
`;

export const GET_USER_BY_NAME = `
  ${USER_FIELDS}
  query GetUserByName($name: String!) {
    userByName(name: $name) {
      ...UserFields
    }
  }
`;

// Deprecated: 하위 호환성을 위해 유지
export const GET_USER = GET_USER_BY_NAME;

export const GET_USERS = `
  query GetUsers($limit: Int, $offset: Int, $search: String, $minLevel: Int) {
    users(limit: $limit, offset: $offset, search: $search, minLevel: $minLevel) {
      id
      name
      avatarUrl
      level
      credit
      createdAt
    }
  }
`;

// Game Room Queries
export const GET_GAME_ROOMS = `
  ${GAME_ROOM_FIELDS}
  ${GAME_USER_FIELDS}
  query GetGameRooms(
    $gameType: GameType
    $status: GameRoomStatus
    $includePrivate: Boolean
    $hasSpace: Boolean
    $limit: Int
    $sortBy: String
  ) {
    gameRooms(
      gameType: $gameType
      status: $status
      includePrivate: $includePrivate
      hasSpace: $hasSpace
      limit: $limit
      sortBy: $sortBy
    ) {
      ...GameRoomFields
      users {
        ...GameUserFields
      }
    }
  }
`;

export const GET_GAME_ROOM = `
  ${GAME_ROOM_DETAIL_FIELDS}
  query GetGameRoom($id: ID!) {
    gameRoom(id: $id) {
      ...GameRoomDetailFields
    }
  }
`;

export const GET_MY_CURRENT_ROOM = `
  ${GAME_ROOM_DETAIL_FIELDS}
  query GetMyCurrentRoom($username: String!) {
    myCurrentRoom(username: $username) {
      ...GameRoomDetailFields
    }
  }
`;

// Invitation Queries
export const GET_MY_INVITATIONS = `
  query GetMyInvitations($userId: ID!, $status: InviteStatus) {
    myInvitations(userId: $userId, status: $status) {
      id
      status
      createdAt
      expiresAt
      room {
        id
        name
        gameType
        maxUsers
        users {
          name
        }
      }
      inviter {
        name
        avatarUrl
      }
    }
  }
`;

export const GET_INVITATION = `
  query GetInvitation($id: ID!) {
    invitation(id: $id) {
      id
      status
      createdAt
      expiresAt
      room {
        id
        name
        gameType
        status
        users {
          name
          score
        }
      }
      inviter {
        name
        avatarUrl
      }
    }
  }
`;

// Quiz Queries
// NOTE: These queries are kept for debugging/testing purposes only.
// In production, quizzes are pre-loaded by the server at game start
// and pushed automatically via WebSocket. No manual quiz requests needed.
export const GET_RANDOM_OX_QUIZ = `
  ${OX_QUIZ_FIELDS}
  query GetRandomOXQuiz($roomId: ID) {
    randomOXQuiz(roomId: $roomId) {
      ...OXQuizFields
    }
  }
`;

export const GET_RANDOM_QA_QUIZ = `
  ${GENERAL_QUIZ_FIELDS}
  query GetRandomQAQuiz($roomId: ID) {
    randomQAQuiz(roomId: $roomId) {
      ...GeneralQuizFields
    }
  }
`;

// User Stats Queries
export const GET_USER_STATS = `
  query GetUserStats($userId: ID!, $gameType: String) {
    userStats(userId: $userId, gameType: $gameType) {
      username
      gameType
      totalGames
      totalWins
      totalScore
      createdAt
      updatedAt
    }
  }
`;

export const GET_LEADERBOARD = `
  query GetLeaderboard($gameType: String!, $limit: Int) {
    leaderboard(gameType: $gameType, limit: $limit) {
      username
      gameType
      totalGames
      totalWins
      totalScore
    }
  }
`;

// Game Config Queries
export const GET_GAME_CONFIG = `
  query GetGameConfig {
    gameConfig {
      maxUsers
      roundDuration
      drawingTime
      guessingTime
      roundsPerGame
    }
  }
`;

export const GET_RANDOM_WORDCHAIN_PROMPT = `
  query GetRandomWordchainPrompt {
    randomWordchainPrompt {
      word
      hint
    }
  }
`;

export const VALIDATE_WORD = `
  query IsValidWord($word: String!) {
    isValidWord(word: $word)
  }
`;

// ============================================
// GraphQL Mutations
// ============================================

// User Mutations
export const CREATE_USER = `
  ${USER_FIELDS}
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      ...UserFields
    }
  }
`;

export const UPDATE_USER = `
  ${USER_FIELDS}
  mutation UpdateUser($userName: String!, $input: UpdateUserInput!) {
    updateUser(userName: $userName, input: $input) {
      ...UserFields
    }
  }
`;

export const DELETE_USER = `
  mutation DeleteUser($userName: String!) {
    deleteUser(userName: $userName)
  }
`;

// Game Room Mutations
export const CREATE_GAME_ROOM = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation CreateGameRoom($input: CreateGameRoomInput!) {
    createGameRoom(input: $input) {
      ...GameRoomDetailFields
    }
  }
`;

export const UPDATE_GAME_ROOM = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation UpdateGameRoom($roomId: ID!, $input: UpdateGameRoomInput!) {
    updateGameRoom(roomId: $roomId, input: $input) {
      ...GameRoomDetailFields
    }
  }
`;

export const JOIN_GAME_ROOM = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation JoinGameRoom($roomId: ID!, $userId: ID!, $password: String) {
    joinGameRoom(roomId: $roomId, userId: $userId, password: $password) {
      ...GameRoomDetailFields
    }
  }
`;

export const LEAVE_GAME_ROOM = `
  mutation LeaveGameRoom($roomId: ID!, $userId: ID!) {
    leaveGameRoom(roomId: $roomId, userId: $userId) {
      id
      hostUserId
      gameType
      status
      maxUsers
      currentRound
      totalRounds
      roundTimeLimit
      users {
        userId
        name
        score
        isReady
      }
    }
  }
`;

export const DELETE_GAME_ROOM = `
  mutation DeleteGameRoom($roomId: ID!) {
    deleteGameRoom(roomId: $roomId)
  }
`;

export const SET_READY = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation SetReady($roomId: ID!, $userId: ID!, $ready: Boolean!) {
    setReady(roomId: $roomId, userId: $userId, ready: $ready) {
      ...GameRoomDetailFields
    }
  }
`;

export const START_GAME = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation StartGame($roomId: ID!) {
    startGame(roomId: $roomId) {
      ...GameRoomDetailFields
    }
  }
`;

export const START_ROUND = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation StartRound($roomId: ID!) {
    startRound(roomId: $roomId) {
      ...GameRoomDetailFields
    }
  }
`;

export const SUBMIT_ANSWER = `
  mutation SubmitAnswer($roomId: ID!, $userId: ID!, $answer: String!) {
    submitAnswer(roomId: $roomId, userId: $userId, answer: $answer) {
      success
      isCorrect
      earnedScore
      totalScore
      correctAnswer
      explanation
    }
  }
`;

export const NEXT_ROUND = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation NextRound($roomId: ID!) {
    nextRound(roomId: $roomId) {
      ...GameRoomDetailFields
    }
  }
`;

export const END_GAME = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation EndGame($roomId: ID!) {
    endGame(roomId: $roomId) {
      ...GameRoomDetailFields
    }
  }
`;

export const TRANSFER_HOST = `
  ${GAME_ROOM_DETAIL_FIELDS}
  mutation TransferHost($roomId: ID!, $newHostUserId: ID!) {
    transferHost(roomId: $roomId, newHostUserId: $newHostUserId) {
      ...GameRoomDetailFields
    }
  }
`;

// Chat Mutation
export const SEND_CHAT = `
  mutation SendChat($roomId: ID!, $username: String!, $message: String!) {
    sendChat(roomId: $roomId, username: $username, message: $message) {
      id
      roomId
      username
      displayName
      message
      timestamp
    }
  }
`;

// ============================================
// Wordchain-specific operations (using submitAnswer mutation)
// ============================================

// Note: For wordchain game, use SUBMIT_ANSWER mutation with word as answer
// No separate SUBMIT_WORDCHAIN_WORD mutation exists in schema

// ============================================
// GraphQL Subscriptions
// ============================================

export const SUBSCRIBE_user_JOINED = `
  subscription userJoined($roomId: ID!) {
    userJoined(roomId: $roomId) {
      username
      displayName
      score
      isReady
    }
  }
`;

export const SUBSCRIBE_user_LEFT = `
  subscription userLeft($roomId: ID!) {
    userLeft(roomId: $roomId) {
      username
      displayName
    }
  }
`;

export const SUBSCRIBE_user_READY_UPDATED = `
  subscription userReadyUpdated($roomId: ID!) {
    userReadyUpdated(roomId: $roomId) {
      username
      displayName
      isReady
    }
  }
`;

export const SUBSCRIBE_HOST_CHANGED = `
  subscription HostChanged($roomId: ID!) {
    hostChanged(roomId: $roomId) {
      username
      displayName
    }
  }
`;

export const SUBSCRIBE_GAME_STARTED = `
  subscription GameStarted($roomId: ID!) {
    gameStarted(roomId: $roomId) {
      id
      status
      currentRound
      users {
        username
        displayName
        score
        isReady
      }
    }
  }
`;

export const SUBSCRIBE_ROUND_STARTED = `
  subscription RoundStarted($roomId: ID!) {
    roundStarted(roomId: $roomId) {
      id
      currentRound
      totalRounds
    }
  }
`;

export const SUBSCRIBE_ROUND_ENDED = `
  subscription RoundEnded($roomId: ID!) {
    roundEnded(roomId: $roomId) {
      id
      currentRound
      users {
        username
        displayName
        score
      }
    }
  }
`;

export const SUBSCRIBE_GAME_ENDED = `
  subscription GameEnded($roomId: ID!) {
    gameEnded(roomId: $roomId) {
      id
      status
      users {
        username
        displayName
        score
      }
    }
  }
`;

export const SUBSCRIBE_CHAT_MESSAGE = `
  subscription ChatMessage($roomId: ID!) {
    chatMessage(roomId: $roomId) {
      id
      username
      displayName
      message
      timestamp
    }
  }
`;

export const SUBSCRIBE_GAME_EVENT = `
  subscription GameEvent($roomId: ID!) {
    gameEvent(roomId: $roomId) {
      type
      roomId
      username
      displayName
      data
      timestamp
    }
  }
`;

export const SUBSCRIBE_ERROR = `
  subscription Error($roomId: ID!) {
    error(roomId: $roomId) {
      code
      message
      timestamp
    }
  }
`;
