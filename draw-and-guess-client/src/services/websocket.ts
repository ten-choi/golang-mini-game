export class WebSocketService {
  private ws: WebSocket | null = null;
  private reconnectTimer: NodeJS.Timeout | null = null;
  private subscriptions: Map<string, (data: any) => void> = new Map();
  private isConnecting = false;

  connect(onConnect?: () => void, onError?: (error: any) => void): Promise<void> {
    if (this.ws?.readyState === WebSocket.OPEN) {
      if (onConnect) onConnect();
      return Promise.resolve();
    }

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
        const wsBaseUrl = import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080';
        this.ws = new WebSocket(`${wsBaseUrl}/app/ws`);

        this.ws.onopen = () => {
          console.log('WebSocket connected');
          this.isConnecting = false;
          
          // Resubscribe to all channels
          this.subscriptions.forEach((callback, channel) => {
            this.sendSubscribe(channel);
          });

          if (onConnect) onConnect();
          resolve();
        };

        this.ws.onerror = (event) => {
          console.error('WebSocket error:', event);
          this.isConnecting = false;
          if (onError) onError(event);
          reject(event);
        };

        this.ws.onclose = () => {
          console.log('WebSocket closed, attempting to reconnect...');
          this.isConnecting = false;
          this.ws = null;
          
          // Auto reconnect after 3 seconds
          if (this.reconnectTimer) clearTimeout(this.reconnectTimer);
          this.reconnectTimer = setTimeout(() => {
            this.connect(onConnect, onError);
          }, 3000);
        };

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data);
            if (message.type === 'message' && message.channel) {
              const callback = this.subscriptions.get(message.channel);
              if (callback) {
                callback(message.data);
              }
            }
          } catch (error) {
            console.error('Error parsing WebSocket message:', error);
          }
        };
      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    
    this.subscriptions.clear();
  }

  private sendSubscribe(channel: string) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'subscribe',
        channel: channel,
      }));
    }
  }

  subscribe(channel: string, callback: (data: any) => void) {
    this.subscriptions.set(channel, callback);
    this.sendSubscribe(channel);
    
    return {
      unsubscribe: () => {
        this.unsubscribe(channel);
      }
    };
  }

  unsubscribe(channel: string) {
    this.subscriptions.delete(channel);
    
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({
        type: 'unsubscribe',
        channel: channel,
      }));
    }
  }

  sendMessage(uuid: string, type: 'chat' | 'draw' | 'game', data: any) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected');
      return;
    }

    const channel = `${type}/${uuid}`;
    
    this.ws.send(JSON.stringify({
      type: 'message',
      channel: channel,
      data: data,
    }));
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const wsService = new WebSocketService();
