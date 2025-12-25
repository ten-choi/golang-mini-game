/**
 * WebSocket Service
 * 
 * Manages WebSocket connections with automatic reconnection,
 * subscription management, and typed message handling.
 * Aligned with backend WebSocket message formats (ws_message.go)
 */

import type { 
  WebSocketRequest, 
  WebSocketResponse, 
  WSSuccessMessage,
  WSErrorMessage,
  GameEventPayload,
  ChatMessagePayload,
  DrawingEventPayload
} from '../types';

// ============================================
// Types
// ============================================

type ChannelType = 'chat' | 'draw' | 'game';
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
        const wsUrl = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/api/v1/ws/lobby';
        
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
   * Send a message to a specific channel
   * @param roomId - Room UUID
   * @param type - Channel type (chat, draw, game)
   * @param data - Message data
   */
  sendMessage(roomId: string, type: ChannelType, data: any): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send message: not connected');
      return;
    }

    // Validate input
    if (!roomId || !type || !data) {
      console.error('[WebSocket] Invalid message parameters');
      return;
    }

    const channel = `${type}/${roomId}`;
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
   * Send a chat message (aligned with server format)
   * @param roomId - Room UUID
   * @param playerId - Player ID
   * @param playerName - Player name
   * @param message - Message content
   */
  sendChatMessage(roomId: string, playerId: string, playerName: string, message: string): void {
    const chatData: ChatMessagePayload = {
      roomId,
      playerId,
      playerName,
      message
    };

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send chat message: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'chat',
      channel: `game/${roomId}`,
      data: chatData
    };

    try {
      this.ws.send(JSON.stringify(request));
      console.log(`[WebSocket] Sent chat message from ${playerName}`);
    } catch (error) {
      console.error('[WebSocket] Failed to send chat message:', error);
    }
  }

  /**
   * Send a drawing event (aligned with server format)
   * @param roomId - Room UUID
   * @param playerId - Player ID
   * @param action - Drawing action
   * @param data - Drawing data (points, color, etc.)
   */
  sendDrawingEvent(roomId: string, playerId: string, action: 'start' | 'draw' | 'end' | 'clear', data?: any): void {
    const drawingData: DrawingEventPayload = {
      roomId,
      playerId,
      action,
      ...data
    };

    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('[WebSocket] Cannot send drawing event: not connected');
      return;
    }

    const request: WebSocketRequest = {
      type: 'drawing',
      channel: `game/${roomId}`,
      data: drawingData
    };

    try {
      this.ws.send(JSON.stringify(request));
    } catch (error) {
      console.error('[WebSocket] Failed to send drawing event:', error);
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
      const message: WebSocketResponse = JSON.parse(event.data);
      
      // Handle standard channel messages
      if (message.type === 'message' && message.channel) {
        // Call channel-specific callback
        const callback = this.subscriptions.get(message.channel);
        if (callback) {
          callback(message.data);
        }

        // Handle typed message data
        if (message.data && typeof message.data === 'object') {
          this.handleTypedMessage(message.data);
        }
      }
      // Handle server success messages (WSSuccessMessage format)
      else if (message.type && message.type !== 'message') {
        this.handleServerMessage(message as any);
      }
    } catch (error) {
      console.error('[WebSocket] Error parsing message:', error);
    }
  }

  /**
   * Handle typed messages from server
   */
  private handleTypedMessage(data: any): void {
    if ('type' in data) {
      const handlers = this.messageHandlers.get(data.type);
      if (handlers) {
        handlers.forEach(handler => handler(data));
      }
    }
  }

  /**
   * Handle server messages in WSSuccessMessage/WSErrorMessage format
   */
  private handleServerMessage(message: WSSuccessMessage | WSErrorMessage): void {
    if (message.type === 'ERROR') {
      const errorMsg = message as WSErrorMessage;
      console.error(`[WebSocket] Server error: ${errorMsg.code} - ${errorMsg.message}`);
      
      // Trigger error handlers
      const errorHandlers = this.messageHandlers.get('ERROR');
      if (errorHandlers) {
        errorHandlers.forEach(handler => handler(errorMsg));
      }
      return;
    }

    // Handle success messages by type
    const successMsg = message as WSSuccessMessage;
    const handlers = this.messageHandlers.get(successMsg.type);
    if (handlers) {
      handlers.forEach(handler => handler(successMsg.payload));
    }

    // Special handling for specific message types
    switch (successMsg.type) {
      case 'GAME_EVENT':
        this.handleGameEvent(successMsg.payload as GameEventPayload);
        break;
      case 'CHAT_MESSAGE':
        this.handleChatMessage(successMsg.payload as ChatMessagePayload);
        break;
      case 'DRAW_EVENT':
        this.handleDrawEvent(successMsg.payload as DrawingEventPayload);
        break;
    }
  }

  private handleGameEvent(payload: GameEventPayload): void {
    console.log(`[WebSocket] Game event: ${payload.eventType} in room ${payload.roomId}`);
    const handlers = this.messageHandlers.get('game_event');
    if (handlers) {
      handlers.forEach(handler => handler(payload));
    }
  }

  private handleChatMessage(payload: ChatMessagePayload): void {
    console.log(`[WebSocket] Chat from ${payload.playerName}: ${payload.message}`);
    const handlers = this.messageHandlers.get('chat_message');
    if (handlers) {
      handlers.forEach(handler => handler(payload));
    }
  }

  private handleDrawEvent(payload: DrawingEventPayload): void {
    console.log(`[WebSocket] Draw event: ${payload.action} by ${payload.playerId}`);
    const handlers = this.messageHandlers.get('draw_event');
    if (handlers) {
      handlers.forEach(handler => handler(payload));
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
