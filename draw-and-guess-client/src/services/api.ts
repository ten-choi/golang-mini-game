/**
 * API Service
 * 
 * This module provides a centralized API client for communicating with the backend.
 * All API calls are typed and include error handling.
 * Uses GraphQL for all game-related operations.
 */

import axios, { AxiosError } from 'axios';
import { graphqlClient, CREATE_GAME_ROOM, GET_GAME_ROOMS, GET_GAME_ROOM, JOIN_GAME_ROOM, LEAVE_GAME_ROOM, START_GAME, DELETE_GAME_ROOM } from './graphql';
import type { 
  GameRoom, 
  ApiResult, 
  CreateRoomResponse, 
  AnswerSubmitResponse 
} from '../types';

// ============================================
// Configuration
// ============================================

import { env } from '../config/env';

const API_BASE_URL = `${env.apiBaseUrl}/api/v1`;

export const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
  },
});

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
    
    // Ensure proper prototype chain for instanceof checks
    Object.setPrototypeOf(this, ApiError.prototype);
  }
}

const handleApiError = (error: unknown, context?: string): never => {
  const contextPrefix = context ? `[${context}] ` : '';
  
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{code?: string, message?: string, error?: string}>;
    const statusCode = axiosError.response?.status;
    const errorData = axiosError.response?.data;
    const message = errorData?.message || errorData?.error || axiosError.message || 'Unknown API error';
    const code = errorData?.code;
    
    console.error(`${contextPrefix}API Error:`, {
      statusCode,
      code,
      message,
      url: axiosError.config?.url,
      method: axiosError.config?.method
    });
    
    throw new ApiError(message, statusCode, code, error);
  }
  
  console.error(`${contextPrefix}Unknown error:`, error);
  throw new ApiError('Unknown error occurred', undefined, 'UNKNOWN_ERROR', error);
};

// ============================================
// API Service Methods
// ============================================

export const apiService = {
  /**
   * Get all game rooms or a specific room by UUID
   * @param uuid - Optional room UUID
   * @returns Array of game rooms
   */
  getGameRooms: async (uuid?: string): Promise<GameRoom[]> => {
    try {
      if (uuid) {
        // Get single room
        const data: any = await graphqlClient.request(GET_GAME_ROOM, { id: uuid });
        console.log('[API] Get game room:', uuid);
        return data.gameRoom ? [mapGraphQLRoomToGameRoom(data.gameRoom)] : [];
      } else {
        // Get all rooms
        const data: any = await graphqlClient.request(GET_GAME_ROOMS);
        console.log('[API] Get game rooms:', data.gameRooms?.length || 0, 'rooms');
        return (data.gameRooms || []).map(mapGraphQLRoomToGameRoom);
      }
    } catch (error) {
      return handleApiError(error, 'getGameRooms');
    }
  },

  /**
   * Create a new game room
   * @param ldapUser - Username of the room creator
   * @param gameType - Type of game (ox, general, guess)
   * @returns Room creation result with room ID and word (for guess type)
   */
  createGameRoom: async (ldapUser: string, gameType: 'ox' | 'general' | 'guess' = 'guess'): Promise<CreateRoomResponse> => {
    try {
      if (!ldapUser || ldapUser.trim() === '') {
        throw new ApiError('Username is required', 400, 'INVALID_INPUT');
      }
      
      // Map game type to GraphQL GameType enum
      const gqlGameType = gameType === 'ox' ? 'OX' : gameType === 'general' ? 'QA' : 'WORDCHAIN';
      
      const data: any = await graphqlClient.request(CREATE_GAME_ROOM, {
        input: {
          name: `${ldapUser}'s Game`,
          gameType: gqlGameType,
          maxPlayers: 8,
          totalRounds: 3,
          hostUsername: ldapUser
        }
      });
      
      console.log('[API] Created game room:', data.createGameRoom.id);
      return {
        room_id: data.createGameRoom.id,
        current_word: '', // Will be set by game logic
        current_word_translations: {}
      };
    } catch (error) {
      return handleApiError(error, 'createGameRoom');
    }
  },

  /**
   * Join an existing game room
   * @param roomId - Room UUID
   * @param username - Username of the player joining
   */
  joinGameRoom: async (roomId: string, username: string): Promise<any> => {
    try {
      if (!roomId || !username || username.trim() === '') {
        throw new ApiError('Room ID and username are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(JOIN_GAME_ROOM, {
        roomId,
        username
      });
      
      console.log('[API] Joined game room:', roomId, 'as', username);
      return { success: true, room: mapGraphQLRoomToGameRoom(data.joinGameRoom) };
    } catch (error) {
      return handleApiError(error, 'joinGameRoom');
    }
  },

  /**
   * Start the game (only host can call this)
   * @param roomId - Room UUID
   */
  startGame: async (roomId: string): Promise<any> => {
    try {
      if (!roomId) {
        throw new ApiError('Room ID is required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(START_GAME, { roomId });
      console.log('[API] Started game in room:', roomId);
      return { success: true, room: data.startGame };
    } catch (error) {
      return handleApiError(error, 'startGame');
    }
  },

  /**
   * Submit an answer (deprecated - use handleChatMessage instead)
   * @param roomId - Room UUID
   * @param username - Username of the player
   * @param answer - The answer to submit
   */
  submitAnswer: async (roomId: string, username: string, answer: string): Promise<any> => {
    try {
      const formData = new URLSearchParams();
      formData.append('username', username);
      formData.append('answer', answer);
      const response = await api.post(`/game/room/${roomId}/answer`, formData);
      return response.data;
    } catch (error) {
      return handleApiError(error, 'submitAnswer');
    }
  },

  /**
   * Send a chat message (also checks if it's a correct answer)
   * @param roomId - Room UUID
   * @param username - Username of the sender
   * @param message - Message content
   * @returns Result indicating if the message was a correct answer
   */
  handleChatMessage: async (
    roomId: string, 
    username: string, 
    message: string
  ): Promise<any> => {
    try {
      if (!roomId || !username || !message || message.trim() === '') {
        throw new ApiError('All fields are required', 400, 'INVALID_INPUT');
      }
      
      const response = await api.post(
        `/game/room/${roomId}/chat`,
        { username, message },
        { headers: { 'Content-Type': 'application/json' } }
      );
      console.log('[API] Chat message sent:', message, 'correct:', response.data?.result?.is_correct);
      return response.data;
    } catch (error) {
      return handleApiError(error, 'handleChatMessage');
    }
  },

  /**
   * Advance to the next round (internal use)
   * @param roomId - Room UUID
   */
  advanceRound: async (roomId: string): Promise<any> => {
    try {
      const formData = new URLSearchParams();
      formData.append('action', 'advance_round');
      const response = await api.patch(`/game/room/${roomId}`, formData);
      console.log('[API] Advanced round in room:', roomId);
      return response.data;
    } catch (error) {
      return handleApiError(error, 'advanceRound');
    }
  },

  /**
   * Update game room settings
   * @param uuid - Room UUID
   * @param data - Data to update
   */
  updateGameRoom: async (
    uuid: string,
    data: { 
      user_count?: number; 
      topic?: string; 
      isCorrect?: boolean; 
      is_active?: boolean;
    }
  ): Promise<any> => {
    try {
      if (!uuid) {
        throw new ApiError('Room UUID is required', 400, 'INVALID_INPUT');
      }
      
      const formData = new URLSearchParams();
      if (data.user_count !== undefined) formData.append('user_count', data.user_count.toString());
      if (data.topic) formData.append('topic', data.topic);
      if (data.isCorrect !== undefined) formData.append('isCorrect', data.isCorrect.toString());
      if (data.is_active !== undefined) formData.append('is_active', data.is_active.toString());
      const response = await api.patch(`/game/room/${uuid}`, formData);
      console.log('[API] Updated game room:', uuid);
      return response.data;
    } catch (error) {
      return handleApiError(error, 'updateGameRoom');
    }
  },

  /**
   * Delete a game room
   * @param uuid - Room UUID
   */
  deleteGameRoom: async (uuid: string): Promise<any> => {
    try {
      if (!uuid) {
        throw new ApiError('Room UUID is required', 400, 'INVALID_INPUT');
      }
      
      const response = await api.delete(`/game/room/${uuid}`);
      console.log('[API] Deleted game room:', uuid);
      return response.data;
    } catch (error) {
      return handleApiError(error, 'deleteGameRoom');
    }
  },

  /**
   * Leave a game room
   * @param roomId - Room UUID
   * @param username - Username of the player leaving
   */
  leaveGameRoom: async (roomId: string, username: string): Promise<any> => {
    try {
      if (!roomId || !username) {
        throw new ApiError('Room ID and username are required', 400, 'INVALID_INPUT');
      }
      
      const data: any = await graphqlClient.request(LEAVE_GAME_ROOM, {
        roomId,
        username
      });
      
      console.log('[API] Left game room:', roomId, 'user:', username);
      return { success: true, room: data.leaveGameRoom };
    } catch (error) {
      return handleApiError(error, 'leaveGameRoom');
    }
  },
};

// ============================================
// Helper Functions
// ============================================

/**
 * Map GraphQL GameRoom to client GameRoom type
 */
function mapGraphQLRoomToGameRoom(gqlRoom: any): GameRoom {
  return {
    id: gqlRoom.id,
    uuid: gqlRoom.id,
    is_active: gqlRoom.status === 'PLAYING',
    drawer_user: gqlRoom.hostUsername,
    room_creator: gqlRoom.hostUsername,
    players: gqlRoom.players?.map((p: any) => ({
      username: p.username,
      score: p.score,
      attempts: 0
    })) || [],
    current_word: '',
    current_word_translations: {},
    round_number: gqlRoom.currentRound || 1,
    time_left: 60,
    game_status: gqlRoom.status?.toLowerCase() || 'waiting',
    used_words: [],
    max_rounds: gqlRoom.totalRounds || 3,
    winning_score: 100,
    game_type: gqlRoom.gameType?.toLowerCase() === 'ox' ? 'ox' : gqlRoom.gameType?.toLowerCase() === 'qa' ? 'general' : 'guess',
    created_at: gqlRoom.createdAt,
    updated_at: gqlRoom.createdAt
  };
}
