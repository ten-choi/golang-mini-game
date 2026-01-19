/**
 * WebSocket Service
 * 
 * Manages WebSocket connections with automatic reconnection,
 * subscription management, and typed message handling.
 * Aligned with backend WebSocket DTO (websocket_dto.go)
 */

import type { 
  WebSocketRequest, 
  WebSocketResponse, 
  WSChatMessageData,
  WSGameStateData,
  WSDrawingData,
  WSAnswerSubmitData,
  WSCorrectAnswerData,
  WSRoundStartData,
  WSRoundEndData,
  WSGameEndData
} from '../types';

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
        // Use lobby WebSocket by default
        const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws/lobby';
        
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
   * @param username - Username
   * @param roomId - Room ID (optional)
   */
  identify(username: string, roomId?: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot identify: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'identify',
      channel: '', // Not used for identify
      data: {
        username: username,
        roomId: roomId || ''
      }
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Identified as ${username} in room ${roomId || 'lobby'}`);
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
    const chatData: WSChatMessageData = {
      room_id: roomId,
      user_id: userId,
      username: username,
      message: message
    };

    this.sendMessage(`game/room-${roomId}`, {
      type: 'chat',
      data: chatData
    });
  }

  /**
   * Send a drawing event (aligned with DrawingData DTO)
   * @param roomId - Room ID
   * @param action - Drawing action
   * @param points - Drawing points
   * @param color - Drawing color
   * @param width - Line width
   */
  sendDrawingEvent(
    roomId: string, 
    action: 'draw' | 'clear' | 'undo',
    points?: Array<{ x: number; y: number }>,
    color?: string,
    width?: number
  ): void {
    const drawingData: WSDrawingData = {
      room_id: roomId,
      action: action,
      points: points,
      color: color,
      width: width
    };

    this.sendMessage(`game/room-${roomId}`, {
      type: 'drawing',
      data: drawingData
    });
  }

  /**
   * Submit an answer (aligned with AnswerSubmitData DTO)
   * @param roomId - Room ID
   * @param userId - User ID
   * @param username - Username
   * @param answer - Answer text
   */
  submitAnswer(roomId: string, userId: string, username: string, answer: string): void {
    const answerData: WSAnswerSubmitData = {
      room_id: roomId,
      user_id: userId,
      username: username,
      answer: answer
    };

    this.sendMessage(`game/room-${roomId}`, {
      type: 'answer',
      data: answerData
    });
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
      const message: WebSocketResponse = JSON.parse(event.data);
      
      // Handle lobby chat history (special case - sent directly by server)
      if (message.type === 'LOBBY_CHAT_HISTORY') {
        console.log('[WebSocket] Received LOBBY_CHAT_HISTORY:', message);
        const callback = this.subscriptions.get('lobby');
        if (callback) {
          callback(message);
        }
        return;
      }
      
      // Handle lobby chat messages (special case)
      if (message.type === 'LOBBY_CHAT') {
        console.log('[WebSocket] Received LOBBY_CHAT:', message);
        const callback = this.subscriptions.get('lobby');
        if (callback) {
          callback(message);
        }
        return;
      }
      
      // Handle standard channel messages
      if (message.type === 'message' && message.channel) {
        // Call channel-specific callback
        const callback = this.subscriptions.get(message.channel);
        if (callback) {
          callback(message.data);
        }

        // Handle typed message data based on WebSocket DTO types
        if (message.data && typeof message.data === 'object') {
          this.handleTypedMessage(message.channel, message.data);
        }
      }
      // Handle error responses
      else if (message.type === 'error') {
        console.error('[WebSocket] Server error:', message.data);
        const errorHandlers = this.messageHandlers.get('ERROR');
        if (errorHandlers) {
          errorHandlers.forEach(handler => handler(message.data));
        }
      }
    } catch (error) {
      console.error('[WebSocket] Error parsing message:', error);
    }
  }

  /**
   * Handle typed messages from server based on message structure
   */
  private handleTypedMessage(_channel: string, data: any): void {
    // Detect message type from data structure
    if ('type' in data) {
      const messageType = data.type;
      const handlers = this.messageHandlers.get(messageType);
      if (handlers) {
        handlers.forEach(handler => handler(data.data || data));
      }

      // Handle specific WebSocket DTO message types
      switch (messageType) {
        case 'game_state':
          this.handleGameState(data.data as WSGameStateData);
          break;
        case 'chat':
          this.handleChatMessage(data.data as WSChatMessageData);
          break;
        case 'drawing':
          this.handleDrawing(data.data as WSDrawingData);
          break;
        case 'round_start':
          this.handleRoundStart(data.data as WSRoundStartData);
          break;
        case 'round_end':
          this.handleRoundEnd(data.data as WSRoundEndData);
          break;
        case 'correct_answer':
          this.handleCorrectAnswer(data.data as WSCorrectAnswerData);
          break;
        case 'game_end':
          this.handleGameEnd(data.data as WSGameEndData);
          break;
      }
    }
  }

  private handleGameState(data: WSGameStateData): void {
    console.log(`[WebSocket] Game state update: round ${data.current_round}, time left ${data.time_left}s`);
    const handlers = this.messageHandlers.get('game_state');
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  private handleChatMessage(data: WSChatMessageData): void {
    console.log(`[WebSocket] Chat from ${data.username}: ${data.message}`);
    const handlers = this.messageHandlers.get('chat_message');
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  private handleDrawing(data: WSDrawingData): void {
    console.log(`[WebSocket] Drawing event: ${data.action}`);
    const handlers = this.messageHandlers.get('drawing');
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  private handleRoundStart(data: WSRoundStartData): void {
    console.log(`[WebSocket] Round ${data.round} started`);
    const handlers = this.messageHandlers.get('round_start');
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  private handleRoundEnd(data: WSRoundEndData): void {
    console.log(`[WebSocket] Round ${data.round} ended`);
    const handlers = this.messageHandlers.get('round_end');
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  private handleCorrectAnswer(data: WSCorrectAnswerData): void {
    console.log(`[WebSocket] Correct answer from ${data.username}: +${data.score} points`);
    const handlers = this.messageHandlers.get('correct_answer');
    if (handlers) {
      handlers.forEach(handler => handler(data));
    }
  }

  private handleGameEnd(data: WSGameEndData): void {
    console.log(`[WebSocket] Game ended, winner: ${data.winner.username}`);
    const handlers = this.messageHandlers.get('game_end');
    if (handlers) {
      handlers.forEach(handler => handler(data));
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
