import axios from 'axios';
import { GameRoom } from '../types';

const API_BASE_URL = `${import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'}/app`;

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/x-www-form-urlencoded',
  },
});

export const apiService = {
  // Get game rooms
  getGameRooms: async (uuid?: string): Promise<GameRoom[]> => {
    const params = uuid ? { id: uuid } : {};
    const response = await api.get('/game/rooms', { params });
    return response.data.result || [];
  },

  // Create game room
  createGameRoom: async (
    ldapUser: string
  ): Promise<{ room_id: string; current_word: string; current_word_translations?: Record<string, string> }> => {
    const formData = new URLSearchParams();
    formData.append('ldap_user', ldapUser);
    const response = await api.post('/game/room', formData);
    return response.data.result;
  },

  // Join game room
  joinGameRoom: async (roomId: string, username: string) => {
    const formData = new URLSearchParams();
    formData.append('username', username);
    const response = await api.post(`/game/room/${roomId}/join`, formData);
    return response.data;
  },

  // Start game
  startGame: async (roomId: string) => {
    const response = await api.post(`/game/room/${roomId}/start`, new URLSearchParams());
    return response.data;
  },

  // Submit answer
  submitAnswer: async (roomId: string, username: string, answer: string) => {
    const formData = new URLSearchParams();
    formData.append('username', username);
    formData.append('answer', answer);
    const response = await api.post(`/game/room/${roomId}/answer`, formData);
    return response.data;
  },

  // Handle chat message (also checks for correct answer)
  handleChatMessage: async (roomId: string, username: string, message: string) => {
    const response = await api.post(`/game/room/${roomId}/chat`, {
      username,
      message,
    }, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    return response.data;
  },

  // Advance to next round
  advanceRound: async (roomId: string) => {
    const formData = new URLSearchParams();
    formData.append('action', 'advance_round');
    const response = await api.patch(`/game/room/${roomId}`, formData);
    return response.data;
  },

  // Update game room
  updateGameRoom: async (
    uuid: string,
    data: { user_count?: number; topic?: string; isCorrect?: boolean; is_active?: boolean }
  ) => {
    const formData = new URLSearchParams();
    if (data.user_count !== undefined) formData.append('user_count', data.user_count.toString());
    if (data.topic) formData.append('topic', data.topic);
    if (data.isCorrect !== undefined) formData.append('isCorrect', data.isCorrect.toString());
    if (data.is_active !== undefined) formData.append('is_active', data.is_active.toString());
    const response = await api.patch(`/game/room/${uuid}`, formData);
    return response.data;
  },

  // Delete game room
  deleteGameRoom: async (uuid: string) => {
    const response = await api.delete(`/game/room/${uuid}`);
    return response.data;
  },
};
