/**
 * GraphQL Client Service
 * 
 * Provides GraphQL client for communicating with the backend GraphQL API
 */

import { GraphQLClient } from 'graphql-request';
import { env } from '../config/env';

// GraphQL endpoint
const GRAPHQL_ENDPOINT = `${env.apiBaseUrl}/api/v1/graphql`;

// Create GraphQL client
export const graphqlClient = new GraphQLClient(GRAPHQL_ENDPOINT, {
  headers: {
    'Content-Type': 'application/json',
  },
});

// ============================================
// GraphQL Queries
// ============================================

export const GET_GAME_ROOMS = `
  query GetGameRooms($gameType: GameType) {
    gameRooms(gameType: $gameType) {
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
      maxPlayers
      hostUsername
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

export const GET_RANDOM_OX_QUIZ = `
  query GetRandomOXQuiz {
    randomOXQuiz {
      id
      category
      difficulty
      question
      answer
      explanation
    }
  }
`;

export const GET_RANDOM_QA_QUIZ = `
  query GetRandomQAQuiz {
    randomQAQuiz {
      id
      category
      difficulty
      question
      options
      answer
      explanation
      imageUrl
    }
  }
`;

// ============================================
// GraphQL Mutations
// ============================================

export const CREATE_GAME_ROOM = `
  mutation CreateGameRoom($input: CreateGameRoomInput!) {
    createGameRoom(input: $input) {
      id
      name
      gameType
      status
      maxPlayers
      hostUsername
      createdAt
    }
  }
`;

export const JOIN_GAME_ROOM = `
  mutation JoinGameRoom($roomId: ID!, $username: String!) {
    joinGameRoom(roomId: $roomId, username: $username) {
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

export const DELETE_GAME_ROOM = `
  mutation DeleteGameRoom($roomId: ID!) {
    deleteGameRoom(roomId: $roomId)
  }
`;

export const CREATE_USER = `
  mutation CreateUser($input: CreateUserInput!) {
    createUser(input: $input) {
      id
      username
      displayName
      createdAt
    }
  }
`;
