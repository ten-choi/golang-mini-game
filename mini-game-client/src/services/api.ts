/**
 * API Service
 * 
 * This module provides a centralized API client for communicating with the backend.
 * All API calls use GraphQL for game-related operations.
 * Aligned with q-connect-server GraphQL Schema
 */

import { graphqlClient, 
  // Queries
  GET_GAME_ROOMS, GET_GAME_ROOM, GET_USER_BY_HANGE_ID, GET_USER_BY_NAME, GET_USERS,
  GET_user_STATS, GET_LEADERBOARD, GET_GAME_CONFIG,
  GET_RANDOM_WORDCHAIN_PROMPT, VALIDATE_WORD,
  // Mutations
  CREATE_GAME_ROOM, UPDATE_GAME_ROOM, JOIN_GAME_ROOM, 
  LEAVE_GAME_ROOM, DELETE_GAME_ROOM, START_GAME, START_ROUND,
  SUBMIT_ANSWER, NEXT_ROUND, END_GAME, READY_user, TRANSFER_HOST,
  CREATE_USER, UPDATE_USER, DELETE_USER,
  SEND_CHAT,
  SUBMIT_WORDCHAIN_WORD, SKIP_WORDCHAIN_TURN
} from './graphql';
import type { 
  GameRoom, 
  User,
  CreateGameRoomInput,
  UpdateGameRoomInput,
  CreateUserInput,
  UpdateUserInput,
  UserStats,
  GameConfig,
  WordchainPrompt,
  ChatMessage
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
  
  /**
   * 로그인용: HangeId로 사용자 조회
   */
  async getUserByHangeId(hangeId: string): Promise<User | null> {
    try {
      console.log('[API] Getting user by hangeId:', hangeId);
      const data: any = await graphqlClient.request(GET_USER_BY_HANGE_ID, { hangeId: hangeId });
      console.log('[API] User found by hangeId:', data.userByHangeId);
      return data.userByHangeId;
    } catch (error: any) {
      // 사용자가 없는 경우 null 반환 (에러를 던지지 않음)
      console.log('[API] User not found by hangeId or error:', error?.message || error);
      return null;
    }
  },

  /**
   * 사용자명으로 사용자 조회
   */
  async getUserByName(userName: string): Promise<User | null> {
    try {
      console.log('[API] Getting user by name:', userName);
      const data: any = await graphqlClient.request(GET_USER_BY_NAME, { name: userName });
      console.log('[API] User found by name:', data.userByName);
      return data.userByName;
    } catch (error: any) {
      // 사용자가 없는 경우 null 반환 (에러를 던지지 않음)
      console.log('[API] User not found by name or error:', error?.message || error);
      return null;
    }
  },

  /**
   * @deprecated 하위 호환성을 위해 유지. getUserByName 사용 권장
   */
  async getUser(userName: string): Promise<User | null> {
    return this.getUserByName(userName);
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
      console.log('[API] Created user:', data.createUser.name);
      return data.createUser;
    } catch (error) {
      return handleGraphQLError(error, 'createUser');
    }
  },

  async updateUser(userId: string, input: UpdateUserInput): Promise<User> {
    try {
      const data: any = await graphqlClient.request(UPDATE_USER, { id: userId, input });
      console.log('[API] Updated user:', userId);
      return data.updateUser;
    } catch (error) {
      return handleGraphQLError(error, 'updateUser');
    }
  },

  async deleteUser(userId: string): Promise<boolean> {
    try {
      const data: any = await graphqlClient.request(DELETE_USER, { id: userId });
      console.log('[API] Deleted user:', userId);
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
      if (!input.hostUserId || input.hostUserId.trim() === '') {
        throw new ApiError('Host user ID is required', 400, 'INVALID_INPUT');
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
      const data: any = await graphqlClient.request(UPDATE_GAME_ROOM, { id: roomId, input });
      console.log('[API] Updated game room:', roomId);
      return data.updateGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'updateGameRoom');
    }
  },

  async joinGameRoom(roomId: string, userId: string, password?: string): Promise<GameRoom> {
    try {
      if (!roomId || !userId || userId.trim() === '') {
        throw new ApiError('Room ID and user ID are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(JOIN_GAME_ROOM, {
        roomId,
        userId,
        password
      });
      
      console.log('[API] Joined game room:', roomId, 'user:', userId);
      return data.joinGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'joinGameRoom');
    }
  },

  async leaveGameRoom(roomId: string, userId: string): Promise<GameRoom | null> {
    try {
      if (!roomId || !userId) {
        throw new ApiError('Room ID and user ID are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(LEAVE_GAME_ROOM, {
        roomId,
        userId
      });
      
      console.log('[API] Left game room:', roomId, 'user:', userId, 'room deleted:', !data.leaveGameRoom);
      return data.leaveGameRoom || null;
    } catch (error) {
      return handleGraphQLError(error, 'leaveGameRoom');
    }
  },

  async deleteGameRoom(roomId: string): Promise<boolean> {
    try {
      if (!roomId) {
        throw new ApiError('Room ID is required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(DELETE_GAME_ROOM, { id: roomId });
      console.log('[API] Deleted game room:', roomId);
      return data.deleteGameRoom;
    } catch (error) {
      return handleGraphQLError(error, 'deleteGameRoom');
    }
  },

  async transferHost(roomId: string, currentHostUserId: string, newHostUserId: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(TRANSFER_HOST, { roomId, currentHostUserId, newHostUserId });
      console.log('[API] Transferred host in room:', roomId, 'to', newHostUserId);
      return data.transferHost;
    } catch (error) {
      return handleGraphQLError(error, 'transferHost');
    }
  },

  async readyUser(roomId: string, userId: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(READY_user, { roomId, userId });
      console.log('[API] Ready user:', userId);
      return data.readyuser;
    } catch (error) {
      return handleGraphQLError(error, 'readyuser');
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

  async submitAnswer(roomId: string, userId: string, answer: string): Promise<{ success: boolean; isCorrect: boolean; earnedScore: number; totalScore: number; correctAnswer?: string; explanation?: string }> {
    try {
      const data: any = await graphqlClient.request(SUBMIT_ANSWER, { roomId, userId, answer });
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

  // ===== Chat Methods =====
  
  async sendChat(roomId: string, userId: string, message: string): Promise<ChatMessage> {
    try {
      if (!roomId || !userId || !message || message.trim() === '') {
        throw new ApiError('All fields are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(SEND_CHAT, { roomId, userId, message });
      console.log('[API] Chat message sent:', message);
      return data.sendChat;
    } catch (error) {
      return handleGraphQLError(error, 'sendChat');
    }
  },

  // ===== Wordchain Methods =====
  
  async submitWordchainWord(roomId: string, userId: string, word: string): Promise<{ success: boolean; isCorrect: boolean; earnedScore: number; totalScore: number; correctAnswer?: string; explanation?: string }> {
    try {
      const data: any = await graphqlClient.request(SUBMIT_WORDCHAIN_WORD, { roomId, userId, word });
      return data.submitWordchainWord;
    } catch (error) {
      return handleGraphQLError(error, 'submitWordchainWord');
    }
  },

  async skipWordchainTurn(roomId: string, userId: string): Promise<GameRoom> {
    try {
      const data: any = await graphqlClient.request(SKIP_WORDCHAIN_TURN, { roomId, userId });
      return data.skipWordchainTurn;
    } catch (error) {
      return handleGraphQLError(error, 'skipWordchainTurn');
    }
  },

  // ===== Quiz Methods =====
  // Note: Quiz questions are now pre-loaded by server at game start
  // and pushed automatically via WebSocket each round.
  // No manual quiz request methods needed - all quiz data comes via WebSocket.

  // ===== User Stats Methods =====
  
  async getUserStats(username: string, gameType?: string): Promise<UserStats[]> {
    try {
      const data: any = await graphqlClient.request(GET_user_STATS, { username, gameType });
      return data.userStats || [];
    } catch (error) {
      return handleGraphQLError(error, 'getUserStats');
    }
  },

  async getLeaderboard(gameType: string, limit?: number): Promise<UserStats[]> {
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
};
