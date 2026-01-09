/**
 * GraphQL Client Service
 * 
 * Provides GraphQL client for communicating with the backend GraphQL API
 */

import { GraphQLClient } from 'graphql-request';
import { env } from '../config/env';

// GraphQL endpoint
const GRAPHQL_ENDPOINT = `${env.apiBaseUrl}/graphql`;

// Create GraphQL client
export const graphqlClient = new GraphQLClient(GRAPHQL_ENDPOINT, {
  headers: {
    'Content-Type': 'application/json',
  },
});

// ============================================
// GraphQL Queries
// ============================================

// User Queries
export const GET_USER = `
  query GetUser($nickname: String!) {
    user(nickname: $nickname) {
      id
      nickname
      avatarUrl
      level
      credit
      hanCoin
      guildId
      createdAt
      updatedAt
    }
  }
`;

export const GET_USERS = `
  query GetUsers {
    users {
      id
      nickname
      avatarUrl
      level
      credit
      createdAt
    }
  }
`;

// Game Room Queries
export const GET_GAME_ROOMS = `
  query GetGameRooms($gameType: GameType) {
    gameRooms(gameType: $gameType) {
      id
      name
      gameType
      status
      currentRound
      totalRounds
      roundTimeLimit
      maxPlayers
      hostUsername
      isPrivate
      players {
        username
        displayName
        score
        isReady
      }
      createdAt
    }
  }
`;

export const GET_GAME_ROOM = `
  query GetGameRoom($id: ID!) {
    gameRoom(id: $id) {
      id
      name
      gameType
      status
      currentRound
      totalRounds
      roundTimeLimit
      maxPlayers
      hostUsername
      isPrivate
      usedQuizIds
      players {
        username
        displayName
        score
        isReady
      }
      createdAt
    }
  }
`;

// Quiz Queries
export const GET_RANDOM_OX_QUIZ = `
  query GetRandomOXQuiz($roomId: ID) {
    randomOXQuiz(roomId: $roomId) {
      id
      category
      difficulty
      question
      explanation
    }
  }
`;

export const GET_RANDOM_QA_QUIZ = `
  query GetRandomQAQuiz($roomId: ID) {
    randomQAQuiz(roomId: $roomId) {
      id
      category
      difficulty
      question
      options
      explanation
      imageUrl
    }
  }
`;

// Player Stats Queries
export const GET_PLAYER_STATS = `
  query GetPlayerStats($username: String!, $gameType: String) {
    playerStats(username: $username, gameType: $gameType) {
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
      maxPlayers
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
  query ValidateWord($word: String!) {
    validateWord(word: $word)
  }
`;

// ============================================
// GraphQL Mutations
// ============================================

// User Mutations
export const CREATE_USER = `
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      id
      nickname
      avatarUrl
      level
      credit
      hanCoin
      createdAt
      updatedAt
    }
  }
`;

export const UPDATE_USER = `
  mutation UpdateUser($nickname: String!, $input: UpdateUserInput!) {
    updateUser(nickname: $nickname, input: $input) {
      id
      nickname
      avatarUrl
      level
      credit
      hanCoin
      updatedAt
    }
  }
`;

export const DELETE_USER = `
  mutation DeleteUser($nickname: String!) {
    deleteUser(nickname: $nickname)
  }
`;

// Game Room Mutations
export const CREATE_GAME_ROOM = `
  mutation CreateGameRoom($input: CreateGameRoomInput!) {
    createGameRoom(input: $input) {
      id
      name
      gameType
      status
      maxPlayers
      totalRounds
      hostUsername
      isPrivate
      players {
        username
        displayName
        score
        isReady
      }
      createdAt
    }
  }
`;

export const UPDATE_GAME_ROOM = `
  mutation UpdateGameRoom($roomId: ID!, $input: UpdateGameRoomInput!) {
    updateGameRoom(roomId: $roomId, input: $input) {
      id
      name
      maxPlayers
      totalRounds
      isPrivate
    }
  }
`;

export const JOIN_GAME_ROOM = `
  mutation JoinGameRoom($roomId: ID!, $username: String!, $password: String) {
    joinGameRoom(roomId: $roomId, username: $username, password: $password) {
      id
      name
      gameType
      status
      currentRound
      totalRounds
      maxPlayers
      hostUsername
      players {
        username
        displayName
        score
        isReady
      }
    }
  }
`;

export const LEAVE_GAME_ROOM = `
  mutation LeaveGameRoom($roomId: ID!, $username: String!) {
    leaveGameRoom(roomId: $roomId, username: $username) {
      id
      players {
        username
        displayName
        score
        isReady
      }
      hostUsername
    }
  }
`;

export const READY_PLAYER = `
  mutation ReadyPlayer($roomId: ID!, $username: String!) {
    readyPlayer(roomId: $roomId, username: $username) {
      id
      players {
        username
        displayName
        score
        isReady
      }
    }
  }
`;

export const START_GAME = `
  mutation StartGame($roomId: ID!) {
    startGame(roomId: $roomId) {
      id
      status
      currentRound
    }
  }
`;

export const SUBMIT_ANSWER = `
  mutation SubmitAnswer($roomId: ID!, $username: String!, $answer: String!) {
    submitAnswer(roomId: $roomId, username: $username, answer: $answer) {
      isCorrect
      score
      message
    }
  }
`;

export const NEXT_ROUND = `
  mutation NextRound($roomId: ID!) {
    nextRound(roomId: $roomId) {
      id
      currentRound
      status
    }
  }
`;

export const END_GAME = `
  mutation EndGame($roomId: ID!) {
    endGame(roomId: $roomId) {
      id
      status
    }
  }
`;

export const DELETE_GAME_ROOM = `
  mutation DeleteGameRoom($roomId: ID!) {
    deleteGameRoom(roomId: $roomId)
  }
`;

// Invitation Mutations
export const SEND_INVITATION = `
  mutation SendInvitation($roomId: ID!, $inviterId: ID!, $inviteeId: ID!) {
    sendInvitation(roomId: $roomId, inviterId: $inviterId, inviteeId: $inviteeId) {
      id
      roomId
      inviterId
      inviteeId
      status
      createdAt
      expiresAt
    }
  }
`;

export const RESPOND_INVITATION = `
  mutation RespondInvitation($invitationId: ID!, $accept: Boolean!) {
    respondInvitation(invitationId: $invitationId, accept: $accept) {
      id
      status
    }
  }
`;

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

export const TRANSFER_HOST = `
  mutation TransferHost($roomId: ID!, $newHostUsername: String!) {
    transferHost(roomId: $roomId, newHostUsername: $newHostUsername) {
      id
      hostUsername
    }
  }
`;

export const SET_READY = `
  mutation SetReady($roomId: ID!, $username: String!, $ready: Boolean!) {
    setReady(roomId: $roomId, username: $username, ready: $ready)
  }
`;

export const START_ROUND = `
  mutation StartRound($roomId: ID!) {
    startRound(roomId: $roomId) {
      id
      currentRound
    }
  }
`;

// ============================================
// GraphQL Subscriptions
// ============================================

export const SUBSCRIBE_PLAYER_JOINED = `
  subscription PlayerJoined($roomId: ID!) {
    playerJoined(roomId: $roomId) {
      username
      displayName
      score
      isReady
    }
  }
`;

export const SUBSCRIBE_PLAYER_LEFT = `
  subscription PlayerLeft($roomId: ID!) {
    playerLeft(roomId: $roomId) {
      username
      displayName
    }
  }
`;

export const SUBSCRIBE_PLAYER_READY_UPDATED = `
  subscription PlayerReadyUpdated($roomId: ID!) {
    playerReadyUpdated(roomId: $roomId) {
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
      players {
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
      players {
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
      players {
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
