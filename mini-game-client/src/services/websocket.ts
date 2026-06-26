/**
 * WebSocket Service
 * 
 * Manages WebSocket connections with automatic reconnection,
 * subscription management, and typed message handling.
 * Aligned with backend WebSocket DTO (websocket_dto.go)
 */

import type { 
  WebSocketRequest
} from '../types';
import { env } from '../config/env';

// ============================================
// Types
// ============================================

type MessageCallback = (data: any) => void;

interface Subscription {
  unsubscribe: () => void;
}

interface WebSocketConfig {
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
}

// ============================================
// WebSocket Service
// ============================================

export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private heartbeatTimer: NodeJS.Timeout | null = null;
  private subscriptions: Map<string, MessageCallback> = new Map();
  private messageHandlers: Map<string, Set<MessageCallback>> = new Map();
  private isConnecting = false;
  private reconnectAttempts = 0;
  private config: Required<WebSocketConfig>;

  constructor(config: WebSocketConfig = {}) {
    this.config = {
      reconnectInterval: config.reconnectInterval || 3000,
      maxReconnectAttempts: config.maxReconnectAttempts || 10,
      heartbeatInterval: config.heartbeatInterval || 30000,
    };
  }

  /**
   * Connect to WebSocket server
   * @param onConnect - Callback when connection is established
   * @param onError - Callback when connection error occurs
   */
  connect(onConnect?: () => void, onError?: (error: any) => void): Promise<void> {
    // Already connected
    if (this.ws?.readyState === WebSocket.OPEN) {
      if (onConnect) onConnect();
      return Promise.resolve();
    }

    // Connection in progress
    if (this.isConnecting) {
      return new Promise((resolve) => {
        const checkInterval = setInterval(() => {
          if (this.ws?.readyState === WebSocket.OPEN) {
            clearInterval(checkInterval);
            if (onConnect) onConnect();
            resolve();
          }
        }, 100);
      });
    }

    this.isConnecting = true;

    return new Promise((resolve, reject) => {
      try {
        // Import env configuration
        const wsUrl = this.normalizeWsUrl(env.wsUrl || 'ws://localhost:8080/ws');
        
        console.log(`[WebSocket] Connecting to ${wsUrl}...`);
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log('[WebSocket] Connected successfully');
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          
          // Resubscribe to all channels
          this.resubscribeAll();
          
          // Start heartbeat
          this.startHeartbeat();

          if (onConnect) onConnect();
          resolve();
        };

        this.ws.onerror = (event) => {
          console.error('[WebSocket] Error:', event);
          this.isConnecting = false;
          if (onError) onError(event);
          reject(event);
        };

        this.ws.onclose = (event) => {
          console.log(`[WebSocket] Connection closed (code: ${event.code})`);
          this.isConnecting = false;
          this.ws = null;
          this.stopHeartbeat();
          
          // Attempt to reconnect
          this.scheduleReconnect(onConnect, onError);
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event);
        };
      } catch (error) {
        console.error('[WebSocket] Connection failed:', error);
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  private normalizeWsUrl(url: string): string {
    if (!url) return 'ws://localhost:8080/ws';
    // If URL contains /ws/extra or /ws/lobby, normalize to /ws
    const match = url.match(/^(wss?:\/\/[^/]+\/ws)(?:\/.*)?$/);
    if (match) {
      return match[1];
    }
    return url;
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    console.log('[WebSocket] Disconnecting...');
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    
    this.stopHeartbeat();
    
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    
    this.subscriptions.clear();
    this.reconnectAttempts = 0;
  }

  /**
   * Subscribe to a channel
   * @param channel - Channel name (e.g., \"game/room-123\")
   * @param callback - Callback to handle incoming messages
   * @returns Subscription object with unsubscribe method
   */
  subscribe(channel: string, callback: MessageCallback): Subscription {
    console.log(`[WebSocket] Subscribing to channel: ${channel}`);
    this.subscriptions.set(channel, callback);
    this.sendSubscribe(channel);
    
    // Extract roomId from game/ channels and set it
    if (channel.startsWith('game/')) {
      const roomId = channel.substring(5); // Remove 'game/' prefix
      console.log(`[WebSocket] Extracted roomId: ${roomId}`);
    }
    
    return {
      unsubscribe: () => this.unsubscribe(channel),
    };
  }

  /**
   * Unsubscribe from a channel
   * @param channel - Channel name
   */
  unsubscribe(channel: string): void {
    console.log(`[WebSocket] Unsubscribing from channel: ${channel}`);
    this.subscriptions.delete(channel);
    
    if (this.ws?.readyState === WebSocket.OPEN) {
      const request: WebSocketRequest = {
        type: 'unsubscribe',
        channel: channel,
      };
      this.ws.send(JSON.stringify(request));
    }
  }

  /**
   * Identify client with username and roomId
   * This helps the server track which user is in which room for proper cleanup
   * @param userId - User ID (MongoDB ObjectID)
   * @param roomId - Room ID (optional)
   */
  identify(userId: string, roomId?: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot identify: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'set_user_id',
      channel: '', // Not used for set_user_id
      data: {
        userId: userId,
        roomId: roomId || ''
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Identified as userId ${userId} in room ${roomId || 'lobby'}`);
    } catch (error) {
      console.error('[WebSocket] Failed to identify:', error);
    }
  }

  /**
   * Send a message to a specific channel
   * @param channel - Channel name (e.g., "game/room-123")
   * @param data - Message data
   */
  sendMessage(channel: string, data: any): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send message: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'message',
      channel: channel,
      data: data,
    };
    
    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Sent message to ${channel}:`, data);
    } catch (error) {
      console.error('[WebSocket] Failed to send message:', error);
    }
  }

  /**
   * Send a chat message (aligned with ChatMessageData DTO)
   * @param roomId - Room ID
   * @param userId - User ID
   * @param username - Username
   * @param message - Message content
   */
  sendChatMessage(roomId: string, userId: string, username: string, message: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send chat message: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'chat',
      channel: '', // Not used for chat
      data: {
        roomId: roomId,
        userId: userId,
        username: username,
        message: message
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Sent chat message to room ${roomId}`);
    } catch (error) {
      console.error('[WebSocket] Failed to send chat message:', error);
    }
  }

  /**
   * Send a drawing event (aligned with DrawingData DTO)
   * @param roomId - Room ID
   * @param userId - User ID
   * @param action - Drawing action
   * @param drawData - Drawing data (points, color, lineWidth, strokeId, etc.)
   */
  sendDrawingEvent(
    roomId: string, 
    userId: string,
    action: 'draw' | 'clear' | 'undo' | 'start' | 'end',
    drawData?: any
  ): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send drawing event: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'drawing',
      channel: '', // Not used for drawing
      data: {
        roomId: roomId,
        userId: userId,
        action: action,
        ...drawData
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Sent drawing event to room ${roomId}:`, action);
    } catch (error) {
      console.error('[WebSocket] Failed to send drawing event:', error);
    }
  }

  /*if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot submit answer: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'answer',
      channel: '', // Not used for answer
      data: {
        roomId: roomId,
        userId: userId,
        username: username,
        answer: answer
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Submitted answer to room ${roomId}`);
    } catch (error) {
      console.error('[WebSocket] Failed to submit answer:', error);
    }ser_id: userId,
      username: username,
      answer: answer
    };

    this.sendMessage(`game/room-${roomId}`, {
      type: 'answer',
      data: answerData
    });
  }

  /**
   * Send lobby chat message
   * @param userId - User ID
   * @param username - Username
   * @param message - Message content
   */
  sendLobbyChatMessage(userId: string, username: string, message: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send lobby chat: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'lobby_chat',
      channel: 'lobby',
      data: {
        userId: userId,
        username: username,
        message: message
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log('[WebSocket] Sent lobby chat message');
    } catch (error) {
      console.error('[WebSocket] Failed to send lobby chat:', error);
    }
  }

  /**
   * Send game action (e.g., start game, ready, etc.)
   * @param roomId - Room ID
   * @param action - Action type
   * @param data - Additional data
   */
  sendGameAction(roomId: string, action: string, data?: any): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send game action: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'game_action',
      channel: '', // Not used
      data: {
        roomId: roomId,
        action: action,
        ...data
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Sent game action to room ${roomId}:`, action);
    } catch (error) {
      console.error('[WebSocket] Failed to send game action:', error);
    }
  }

  /**
   * Submit quiz answer
   * @param roomId - Room ID
   * @param userId - User ID
   * @param answer - Answer (number for index, boolean for OX)
   * @param quizId - Quiz ID
   */
  sendQuizAnswer(roomId: string, userId: string, answer: number | boolean, quizId: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send quiz answer: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'quiz_answer',
      channel: '', // Not used
      data: {
        roomId: roomId,
        userId: userId,
        answer: answer,
        quizId: quizId
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Sent quiz answer to room ${roomId}:`, answer);
    } catch (error) {
      console.error('[WebSocket] Failed to send quiz answer:', error);
    }
  }

  /**
   * Send wordchain word submission
   * @param roomId - Room ID
   * @param userId - User ID
   * @param word - Submitted word
   * @param lastWord - Previous word
   */
  sendWordchainSubmit(roomId: string, userId: string, word: string, lastWord: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send wordchain submit: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'wordchain_submit',
      channel: '', // Not used
      data: {
        userId: userId,
        word: word,
        lastWord: lastWord
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Sent wordchain submit to room ${roomId}: ${word}`);
    } catch (error) {
      console.error('[WebSocket] Failed to send wordchain submit:', error);
    }
  }

  /**
   * Request lobby chat history
   */
  requestLobbyChatHistory(): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot request lobby chat: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'get_lobby_messages',
      channel: 'lobby',
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log('[WebSocket] Requested lobby chat history');
    } catch (error) {
      console.error('[WebSocket] Failed to request lobby chat:', error);
    }
  }

  /**
   * Check if WebSocket is connected
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  // ============================================
  // Private Methods
  // ============================================

  private sendSubscribe(channel: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      const request: WebSocketRequest = {
        type: 'subscribe',
        channel: channel,
      };
      this.ws.send(JSON.stringify(request));
    }
  }

  private resubscribeAll(): void {
    this.subscriptions.forEach((_, channel) => {
      this.sendSubscribe(channel);
    });
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const message = JSON.parse(event.data);
      
      console.log('[WebSocket] Received message:', message);
      
      // Handle server message format: { "type": "message", "channel": "...", "data": "..." }
      if (message.type === 'message' && message.channel) {
        let parsedData = message.data;
        
        // Parse data if it's a JSON string
        if (typeof message.data === 'string') {
          try {
            parsedData = JSON.parse(message.data);
          } catch (e) {
            console.warn('[WebSocket] Failed to parse message.data as JSON:', message.data);
          }
        }

        console.log('[WebSocket] Parsed data:', parsedData);

        // Get callback for this channel
        const callback = this.subscriptions.get(message.channel);
        if (callback) {
          callback(parsedData);
        }

        // Also handle typed messages
        if (parsedData && typeof parsedData === 'object' && 'type' in parsedData) {
          const messageType = String(parsedData.type);
          const normalized = messageType.toLowerCase();
          const upper = messageType.toUpperCase();

          const handlersToCall = new Set<MessageCallback>();
          const exactHandlers = this.messageHandlers.get(messageType);
          const normalizedHandlers = this.messageHandlers.get(normalized);
          const upperHandlers = this.messageHandlers.get(upper);

          exactHandlers?.forEach(handler => handlersToCall.add(handler));
          normalizedHandlers?.forEach(handler => handlersToCall.add(handler));
          upperHandlers?.forEach(handler => handlersToCall.add(handler));

          if (handlersToCall.size > 0) {
            handlersToCall.forEach(handler => handler(parsedData.data || parsedData));
          }
        }
      }
      // Handle direct format messages (success, error, etc.)
      else if (message.type) {
        const messageType = message.type;
        
        switch (messageType) {
          case 'success':
            console.log('[WebSocket] Success:', message.message);
            break;
          case 'error':
            console.error('[WebSocket] Error:', message.message);
            const errorHandlers = this.messageHandlers.get('ERROR');
            if (errorHandlers) {
              errorHandlers.forEach(handler => handler(message));
            }
            break;
          case 'unsubscribed':
            console.log('[WebSocket] Unsubscribed from:', message.channel);
            break;
          default:
            console.log(`[WebSocket] Unhandled message type: ${messageType}`, message);
            break;
        }
      }
    } catch (error) {
      console.error('[WebSocket] Error parsing message:', error);
    }
  }



  /**
   * Add a message handler for a specific message type
   * @param type - Message type (e.g., 'room_update', 'chat')
   * @param callback - Callback function
   */
  addMessageHandler(type: string, callback: MessageCallback): void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set());
    }
    this.messageHandlers.get(type)!.add(callback);
  }

  /**
   * Remove a message handler
   * @param type - Message type
   * @param callback - Callback function to remove
   */
  removeMessageHandler(type: string, callback: MessageCallback): void {
    const handlers = this.messageHandlers.get(type);
    if (handlers) {
      handlers.delete(callback);
      if (handlers.size === 0) {
        this.messageHandlers.delete(type);
      }
    }
  }

  private scheduleReconnect(onConnect?: () => void, onError?: (error: any) => void): void {
    if (this.reconnectAttempts >= this.config.maxReconnectAttempts) {
      console.error('[WebSocket] Max reconnection attempts reached');
      return;
    }

    this.reconnectAttempts++;
    console.log(
      `[WebSocket] Reconnecting in ${this.config.reconnectInterval}ms ` +
      `(attempt ${this.reconnectAttempts}/${this.config.maxReconnectAttempts})`
    );

    if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
    this.reconnectTimer = setTimeout(() => {
      this.connect(onConnect, onError).catch(console.error);
    }, this.config.reconnectInterval);
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        // Send ping (can be implemented if backend supports it)
        // For now, we just check the connection state
        console.log('[WebSocket] Heartbeat: connection alive');
      }
    }, this.config.heartbeatInterval);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }
}

// ============================================
// Singleton Instance
// ============================================

export const wsService = new WebSocketService();
