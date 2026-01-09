/**
 * API Service
 * 
 * This module provides a centralized API client for communicating with the backend.
 * All API calls use GraphQL for game-related operations.
 */

import { graphqlClient, 
  // Queries
  GET_GAME_ROOMS, GET_GAME_ROOM, GET_USER, GET_USERS,
  GET_RANDOM_OX_QUIZ, GET_RANDOM_QA_QUIZ, 
  GET_PLAYER_STATS, GET_LEADERBOARD, GET_GAME_CONFIG,
  GET_RANDOM_WORDCHAIN_PROMPT, VALIDATE_WORD,
  // Mutations
  CREATE_GAME_ROOM, UPDATE_GAME_ROOM, JOIN_GAME_ROOM, 
  LEAVE_GAME_ROOM, DELETE_GAME_ROOM, START_GAME,
  SUBMIT_ANSWER, READY_PLAYER, NEXT_ROUND, END_GAME,
  CREATE_USER, UPDATE_USER, DELETE_USER,
  SEND_CHAT, TRANSFER_HOST, SET_READY, START_ROUND,
  SEND_INVITATION, RESPOND_INVITATION
} from './graphql';
import type { 
  GameRoom, 
  User,
  Player,
  CreateGameRoomInput,
  UpdateGameRoomInput,
  CreateUserInput,
  UpdateUserInput,
  GeneralQuiz,
  OXQuiz,
  PlayerStats,
  GameConfig,
  WordchainPrompt,
  GraphQLResponse
} from '../types';

// ============================================
// Error Handling
// ============================================

export class ApiError extends Error {
  constructor(
    message: string,
    public statusCode?: number,
    public code?: string,
    public originalError?: any
  ) {
    super(message);
    this.name = 'ApiError';
    Object.setPrototypeOf(this, ApiError.prototype);
  }
}

const handleGraphQLError = (error: any, context?: string): never => {
  const contextPrefix = context ? `[${context}] ` : '';
  
  console.error(`${contextPrefix}GraphQL Error:`, error);
  
  if (error.response?.errors) {
    const firstError = error.response.errors[0];
    throw new ApiError(
      firstError.message || 'GraphQL error',
      error.response.status,
      firstError.extensions?.code,
      error
    );
  }
  
  throw new ApiError(
    error.message || 'Unknown GraphQL error',
    undefined,
    'UNKNOWN_ERROR',
    error
  );
};


// ============================================
// API Service Methods
// ============================================

export const apiService = {
  // ===== User Methods =====
  
  async getUser(nickname: string): Promise<User | null> {
    try {
      console.log('[API] Getting user:', nickname);
      const data: any = await graphqlClient.request(GET_USER, { nickname });
      console.log('[API] User found:', data.user);
      return data.user;
    } catch (error: any) {
      // 사용자가 없는 경우 null 반환 (에러를 던지지 않음)
      console.log('[API] User not found or error:', error?.message || error);
      return null;
    }
  },

  async getUsers(): Promise<User[]> {
    try {
      const data: any = await graphqlClient.request(GET_USERS);
      return data.users || [];
    } catch (error) {
      return handleGraphQLError(error, 'getUsers');
    }
  },

  async createUser(input: CreateUserInput): Promise<User> {
    try {
      const data: any = await graphqlClient.request(CREATE_USER, { input });
      console.log('[API] Created user:', data.createUser.nickname);
      return data.createUser;
    } catch (error) {
      return handleGraphQLError(error, 'createUser');
    }
  },

  async updateUser(nickname: string, input: UpdateUserInput): Promise<User> {
    try {
      const data: any = await graphqlClient.request(UPDATE_USER, { nickname, input });
      console.log('[API] Updated user:', nickname);
      return data.updateUser;
    } catch (error) {
      return handleGraphQLError(error, 'updateUser');
    }
  },

  async deleteUser(nickname: string): Promise<boolean> {
    try {
      const data: any = await graphqlClient.request(DELETE_USER, { nickname });
      console.log('[API] Deleted user:', nickname);
      return data.deleteUser;
    } catch (error) {
      return handleGraphQLError(error, 'deleteUser');
    }
  },

  // ===== Game Room Methods =====
  
  async getGameRooms(gameType?: 'OX' | 'QA' | 'WORDCHAIN'): Promise<GameRoom[]> {
    try {
      const data: any = await graphqlClient.request(GET_GAME_ROOMS, { gameType });
      console.log('[API] Get game rooms:', data.gameRooms?.length || 0, 'rooms');
      return data.gameRooms || [];
    } catch (error) {
      return handleGraphQLError(error, 'getGameRooms');
    }
  },

  async getGameRoom(id: string): Promise<GameRoom | null> {
    try {
      const data: any = await graphqlClient.request(GET_GAME_ROOM, { id });
      console.log('[API] Get game room:', id);
      return data.gameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'getGameRoom');
    }
  },

  async createGameRoom(input: CreateGameRoomInput): Promise<GameRoom> {
    try {
      if (!input.hostUsername || input.hostUsername.trim() === '') {
        throw new ApiError('Host username is required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(CREATE_GAME_ROOM, { input });
      console.log('[API] Created game room:', data.createGameRoom.id);
      return data.createGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'createGameRoom');
    }
  },

  async updateGameRoom(roomId: string, input: UpdateGameRoomInput): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(UPDATE_GAME_ROOM, { roomId, input });
      console.log('[API] Updated game room:', roomId);
      return data.updateGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'updateGameRoom');
    }
  },

  async joinGameRoom(roomId: string, username: string, password?: string): Promise<GameRoom> {
    try {
      if (!roomId || !username || username.trim() === '') {
        throw new ApiError('Room ID and username are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(JOIN_GAME_ROOM, {
        roomId,
        username,
        password
      });
      
      console.log('[API] Joined game room:', roomId, 'as', username);
      return data.joinGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'joinGameRoom');
    }
  },

  async leaveGameRoom(roomId: string, username: string): Promise<GameRoom> {
    try {
      if (!roomId || !username) {
        throw new ApiError('Room ID and username are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(LEAVE_GAME_ROOM, {
        roomId,
        username
      });
      
      console.log('[API] Left game room:', roomId, 'user:', username);
      return data.leaveGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'leaveGameRoom');
    }
  },

  async deleteGameRoom(roomId: string): Promise<boolean> {
    try {
      if (!roomId) {
        throw new ApiError('Room ID is required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(DELETE_GAME_ROOM, { roomId });
      console.log('[API] Deleted game room:', roomId);
      return data.deleteGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'deleteGameRoom');
    }
  },

  async transferHost(roomId: string, newHostUsername: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(TRANSFER_HOST, { roomId, newHostUsername });
      console.log('[API] Transferred host in room:', roomId, 'to', newHostUsername);
      return data.transferHost;
    } catch (error) {
      return handleGraphQLError(error, 'transferHost');
    }
  },

  async setReady(roomId: string, username: string, ready: boolean): Promise<boolean> {
    try {
      const data: any = await graphqlClient.request(SET_READY, { roomId, username, ready });
      console.log('[API] Set ready:', username, ready);
      return data.setReady;
    } catch (error) {
      return handleGraphQLError(error, 'setReady');
    }
  },

  async readyPlayer(roomId: string, username: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(READY_PLAYER, { roomId, username });
      console.log('[API] Ready player:', username);
      return data.readyPlayer;
    } catch (error) {
      return handleGraphQLError(error, 'readyPlayer');
    }
  },

  // ===== Game Flow Methods =====
  
  async startGame(roomId: string): Promise<GameRoom> {
    try {
      if (!roomId) {
        throw new ApiError('Room ID is required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(START_GAME, { roomId });
      console.log('[API] Started game in room:', roomId);
      return data.startGame;
    } catch (error) {
      return handleGraphQLError(error, 'startGame');
    }
  },

  async startRound(roomId: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(START_ROUND, { roomId });
      console.log('[API] Started round in room:', roomId);
      return data.startRound;
    } catch (error) {
      return handleGraphQLError(error, 'startRound');
    }
  },

  async submitAnswer(roomId: string, username: string, answer: string): Promise<{ isCorrect: boolean; score: number; message: string }> {
    try {
      const data: any = await graphqlClient.request(SUBMIT_ANSWER, { roomId, username, answer });
      console.log('[API] Submitted answer:', answer, 'correct:', data.submitAnswer.isCorrect);
      return data.submitAnswer;
    } catch (error) {
      return handleGraphQLError(error, 'submitAnswer');
    }
  },

  async nextRound(roomId: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(NEXT_ROUND, { roomId });
      console.log('[API] Next round in room:', roomId);
      return data.nextRound;
    } catch (error) {
      return handleGraphQLError(error, 'nextRound');
    }
  },

  async endGame(roomId: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(END_GAME, { roomId });
      console.log('[API] Ended game in room:', roomId);
      return data.endGame;
    } catch (error) {
      return handleGraphQLError(error, 'endGame');
    }
  },

  // ===== Quiz Methods =====
  
  async getRandomOXQuiz(roomId?: string): Promise<OXQuiz | null> {
    try {
      const data: any = await graphqlClient.request(GET_RANDOM_OX_QUIZ, { roomId });
      return data.randomOXQuiz;
    } catch (error) {
      return handleGraphQLError(error, 'getRandomOXQuiz');
    }
  },

  async getRandomQAQuiz(roomId?: string): Promise<GeneralQuiz | null> {
    try {
      const data: any = await graphqlClient.request(GET_RANDOM_QA_QUIZ, { roomId });
      return data.randomQAQuiz;
    } catch (error) {
      return handleGraphQLError(error, 'getRandomQAQuiz');
    }
  },

  // ===== Player Stats Methods =====
  
  async getPlayerStats(username: string, gameType?: string): Promise<PlayerStats[]> {
    try {
      const data: any = await graphqlClient.request(GET_PLAYER_STATS, { username, gameType });
      return data.playerStats || [];
    } catch (error) {
      return handleGraphQLError(error, 'getPlayerStats');
    }
  },

  async getLeaderboard(gameType: string, limit?: number): Promise<PlayerStats[]> {
    try {
      const data: any = await graphqlClient.request(GET_LEADERBOARD, { gameType, limit });
      return data.leaderboard || [];
    } catch (error) {
      return handleGraphQLError(error, 'getLeaderboard');
    }
  },

  // ===== Game Config Methods =====
  
  async getGameConfig(): Promise<GameConfig> {
    try {
      const data: any = await graphqlClient.request(GET_GAME_CONFIG);
      return data.gameConfig;
    } catch (error) {
      return handleGraphQLError(error, 'getGameConfig');
    }
  },

  async getRandomWordchainPrompt(): Promise<WordchainPrompt> {
    try {
      const data: any = await graphqlClient.request(GET_RANDOM_WORDCHAIN_PROMPT);
      return data.randomWordchainPrompt;
    } catch (error) {
      return handleGraphQLError(error, 'getRandomWordchainPrompt');
    }
  },

  async validateWord(word: string): Promise<boolean> {
    try {
      const data: any = await graphqlClient.request(VALIDATE_WORD, { word });
      return data.validateWord;
    } catch (error) {
      return handleGraphQLError(error, 'validateWord');
    }
  },

  // ===== Chat Methods =====
  
  async sendChat(roomId: string, username: string, message: string): Promise<any> {
    try {
      if (!roomId || !username || !message || message.trim() === '') {
        throw new ApiError('All fields are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(SEND_CHAT, { roomId, username, message });
      console.log('[API] Chat message sent:', message);
      return data.sendChat;
    } catch (error) {
      return handleGraphQLError(error, 'sendChat');
    }
  },

  // ===== Invitation Methods =====
  
  async sendInvitation(roomId: string, inviterId: string, inviteeId: string): Promise<any> {
    try {
      const data: any = await graphqlClient.request(SEND_INVITATION, { roomId, inviterId, inviteeId });
      console.log('[API] Sent invitation to:', inviteeId);
      return data.sendInvitation;
    } catch (error) {
      return handleGraphQLError(error, 'sendInvitation');
    }
  },

  async respondInvitation(invitationId: string, accept: boolean): Promise<any> {
    try {
      const data: any = await graphqlClient.request(RESPOND_INVITATION, { invitationId, accept });
      console.log('[API] Responded to invitation:', accept);
      return data.respondInvitation;
    } catch (error) {
      return handleGraphQLError(error, 'respondInvitation');
    }
  },
};
