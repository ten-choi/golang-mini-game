/**
 * API Service
 * 
 * This module provides a centralized API client for communicating with the backend.
 * All API calls are typed and include error handling.
 */

import axios, { AxiosError } from 'axios';
import type { 
  GameRoom, 
  ApiResult, 
  CreateRoomResponse, 
  AnswerSubmitResponse 
} from '../types';

// ============================================
// Configuration
// ============================================

const API_BASE_URL = `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/app`;

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
    public originalError?: any
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

const handleApiError = (error: unknown): never => {
  if (axios.isAxiosError(error)) {
    const axiosError = error as AxiosError<{code?: string, message?: string}>;
    const message = axiosError.response?.data?.message || axiosError.message;
    const statusCode = axiosError.response?.status;
    throw new ApiError(message, statusCode, error);
  }
  throw new ApiError('Unknown error occurred', undefined, error);
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
      const params = uuid ? { id: uuid } : {};
      const response = await api.get<GameRoom[]>('/game/rooms', { params });
      return response.data || [];
    } catch (error) {
      return handleApiError(error);
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
      const formData = new URLSearchParams();
      formData.append('ldap_user', ldapUser);
      formData.append('game_type', gameType);
      const response = await api.post<CreateRoomResponse>('/game/room', formData);
      return response.data;
    } catch (error) {
      return handleApiError(error);
    }
  },

  /**
   * Join an existing game room
   * @param roomId - Room UUID
   * @param username - Username of the player joining
   */
  joinGameRoom: async (roomId: string, username: string): Promise<any> => {
    try {
      const formData = new URLSearchParams();
      formData.append('username', username);
      const response = await api.post(`/game/room/${roomId}/join`, formData);
      return response.data;
    } catch (error) {
      return handleApiError(error);
    }
  },

  /**
   * Start the game (only host can call this)
   * @param roomId - Room UUID
   */
  startGame: async (roomId: string): Promise<any> => {
    try {
      const response = await api.post(`/game/room/${roomId}/start`, new URLSearchParams());
      return response.data;
    } catch (error) {
      return handleApiError(error);
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
      return handleApiError(error);
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
      const response = await api.post(
        `/game/room/${roomId}/chat`,
        { username, message },
        { headers: { 'Content-Type': 'application/json' } }
      );
      return response.data;
    } catch (error) {
      return handleApiError(error);
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
      return response.data;
    } catch (error) {
      return handleApiError(error);
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
      const formData = new URLSearchParams();
      if (data.user_count !== undefined) formData.append('user_count', data.user_count.toString());
      if (data.topic) formData.append('topic', data.topic);
      if (data.isCorrect !== undefined) formData.append('isCorrect', data.isCorrect.toString());
      if (data.is_active !== undefined) formData.append('is_active', data.is_active.toString());
      const response = await api.patch(`/game/room/${uuid}`, formData);
      return response.data;
    } catch (error) {
      return handleApiError(error);
    }
  },

  /**
   * Delete a game room
   * @param uuid - Room UUID
   */
  deleteGameRoom: async (uuid: string): Promise<any> => {
    try {
      const response = await api.delete(`/game/room/${uuid}`);
      return response.data;
    } catch (error) {
      return handleApiError(error);
    }
  },

  /**
   * Leave a game room
   * @param roomId - Room UUID
   * @param username - Username of the player leaving
   */
  leaveGameRoom: async (roomId: string, username: string): Promise<any> => {
    try {
      const formData = new URLSearchParams();
      formData.append('username', username);
      const response = await api.post(`/game/room/${roomId}/leave`, formData);
      return response.data;
    } catch (error) {
      return handleApiError(error);
    }
  },
};
