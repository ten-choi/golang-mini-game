import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import DrawingCanvas from '../components/DrawingCanvas';
import { wsService } from '../services/websocket';
import { apiService } from '../services/api';
import { 
  GameRoom as GameRoomType, 
  GameUser,
  User,
  ChatMessagePayload,
  GameEventPayload 
} from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const GameRoom: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [user] = useState<User | null>(() => {
    const userData = sessionStorage.getItem('user');
    return userData ? JSON.parse(userData) : null;
  });
  const [username] = useState(user?.name || sessionStorage.getItem('username') || 'Guest');
  const [selectedColor, setSelectedColor] = useState('#000000');
  const [room, setRoom] = useState<GameRoomType | null>(null);
  const [isDrawer, setIsDrawer] = useState(false);
  const [message, setMessage] = useState('');
  const [timeLeft, setTimeLeft] = useState<number | null>(null);
  const [chatMessages, setChatMessages] = useState<Array<{username: string, text: string, type: 'chat' | 'system' | 'answer'}>>([]);
  const [chatInput, setChatInput] = useState('');
  const [clearCanvasTrigger, setClearCanvasTrigger] = useState(0);
  const [isMobile, setIsMobile] = useState(window.innerWidth <= 768);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const chatEndRef = useRef<HTMLDivElement>(null);
  const prevRoundRef = useRef<number>(1);

  const colors = [
    '#000000', '#FF0000', '#00FF00', '#0000FF',
    '#FFFF00', '#FF00FF', '#00FFFF', '#FFA500',
  ];

  useEffect(() => {
    const handleResize = () => {
      setIsMobile(window.innerWidth <= 768);
    };
    window.addEventListener('resize', handleResize);
    return () => window.removeEventListener('resize', handleResize);
  }, []);

  useEffect(() => {
    if (!roomId) return;
    
    // Initialize WebSocket connection
    wsService.connect(
      () => {
        console.log('[GameRoom] WebSocket connected');
        // Identify client with userId and roomId
        if (user) {
          wsService.identify(user.id, roomId);
        }
      }, 
      (error) => console.error('[GameRoom] WebSocket connection error:', error)
    );
    
    loadRoom();

    // Subscribe to game channel - handle new server format
    const gameSubscription = wsService.subscribe(`game/${roomId}`, (data) => {
      console.log('[GameRoom] Received game message:', data);
      
      // Handle new server format: { type: "MESSAGE_TYPE", payload: {...} }
      if (data.type === 'CHAT_MESSAGE' && data.payload) {
        const payload = data.payload;
        setChatMessages((prev) => [...prev, {
          username: payload.username || 'System',
          text: payload.message,
          type: 'chat'
        }]);
      } else if (data.type === 'GAME_EVENT' && data.payload) {
        const payload = data.payload;
        console.log('[GameRoom] Game event:', payload.eventType);
        
        switch (payload.eventType) {
          case 'user_joined':
          case 'user_left':
          case 'game_started':
          case 'round_started':
          case 'round_ended':
          case 'game_ended':
            loadRoom();
            break;
        }
      } else if (data.type === 'GAME_ACTION' && data.payload) {
        const payload = data.payload;
        if (payload.action === 'update') {
          loadRoom();
        }
      } else if (data.type === 'DRAW_EVENT' && data.payload) {
        // Drawing events are handled by DrawingCanvas component
        console.log('[GameRoom] Draw event received:', data.payload.action);
      }
      // Handle legacy format for backward compatibility
      else if (data.type === 'update' || data.type === 'room_update') {
        if (data.data) {
          console.log('[GameRoom] Updating room from WebSocket:', data.data);
          setRoom(data.data);
          
          if (data.data.hostUserId) {
            setIsDrawer(data.data.hostUserId === user?.id);
          }
        } else {
          loadRoom();
        }
        
        if (data.time_left !== undefined) {
          setTimeLeft(data.time_left);
        }
        if (data.status !== undefined && data.status === 'FINISHED') {
          setMessage(t.gameRoom.gameFinished);
        }
      } else if (data.type === 'timer_update') {
        if (data.data) {
          setTimeLeft(data.data.time_left);
          if (data.data.status === 'FINISHED') {
            setMessage(t.gameRoom.gameFinished);
            loadRoom();
          }
        }
      }
    });

    // Subscribe to chat channel
    const chatSubscription = wsService.subscribe(`chat/${roomId}`, (data) => {
      setChatMessages((prev) => [...prev, {
        username: data.username || 'System',
        text: data.message,
        type: data.type || 'chat'
      }]);
    });

    // Handle GAME_EVENT messages
    const gameEventHandler = (payload: GameEventPayload) => {
      console.log('[GameRoom] Game event handler:', payload.eventType);
      
      switch (payload.eventType) {
        case 'user_joined':
        case 'user_left':
        case 'game_started':
        case 'round_started':
        case 'round_ended':
        case 'game_ended':
          loadRoom();
          break;
      }
    };

    // Handle CHAT_MESSAGE messages
    const chatMessageHandler = (payload: ChatMessagePayload) => {
      setChatMessages((prev) => [...prev, {
        username: payload.username || 'System',
        text: payload.message,
        type: 'chat'
      }]);
    };

    // Handle ERROR messages
    const errorHandler = (error: any) => {
      console.error('[GameRoom] WebSocket error:', error);
      setMessage(`Error: ${error.message || 'Unknown error'}`);
    };

    // Handle room deletion
    const roomDeletedHandler = (data: any) => {
      console.log('[GameRoom] Room has been deleted:', data);
      alert('방이 삭제되었습니다.');
      navigate('/room-list');
    };

    // Register message handlers
    wsService.addMessageHandler('game_event', gameEventHandler);
    wsService.addMessageHandler('chat_message', chatMessageHandler);
    wsService.addMessageHandler('ERROR', errorHandler);
    wsService.addMessageHandler('room_deleted', roomDeletedHandler);

    return () => {
      if (gameSubscription) gameSubscription.unsubscribe();
      if (chatSubscription) chatSubscription.unsubscribe();
      if (timerRef.current) clearInterval(timerRef.current);
      
      // Remove message handlers
      wsService.removeMessageHandler('game_event', gameEventHandler);
      wsService.removeMessageHandler('chat_message', chatMessageHandler);
      wsService.removeMessageHandler('ERROR', errorHandler);
      wsService.removeMessageHandler('room_deleted', roomDeletedHandler);
    };
  }, [roomId, t.gameRoom.gameFinished]);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  // Update timeLeft when room is loaded or status changes
  useEffect(() => {
    if (room && room.status !== 'PLAYING') {
      setTimeLeft(room.roundTimeLimit);
    }
  }, [room?.roundTimeLimit, room?.status]);

  useEffect(() => {
    // No longer need polling - timer updates come via WebSocket
    // Only load room initially and when game state changes
  }, [room?.status]);

  const loadRoom = async () => {
    if (!roomId) return;
    try {
      const newRoom = await apiService.getGameRoom(roomId);
      if (!newRoom) {
        console.error('[GameRoom] Room not found:', roomId);
        alert('게임방을 찾을 수 없습니다.');
        navigate('/room-list');
        return;
      }
        
      // If it's a quiz game, redirect to QuizRoom
      if (newRoom.gameType === 'OX' || newRoom.gameType === 'QA') {
        navigate(`/quiz-room/${roomId}`, { replace: true });
        return;
      }

      // If it's a wordchain game, redirect to WordchainRoom
      if (newRoom.gameType === 'WORDCHAIN') {
        navigate(`/wordchain-room/${roomId}`, { replace: true });
        return;
      }

      // This component only handles DRAWING type
      if (newRoom.gameType !== 'DRAWING') {
        console.error('[GameRoom] Invalid game type for this component:', newRoom.gameType);
        alert('잘못된 게임 타입입니다.');
        navigate('/rooms');
        return;
      }
      
      // Check if round changed - clear canvas if so
      if (prevRoundRef.current !== newRoom.currentRound) {
        setClearCanvasTrigger(prev => prev + 1);
        prevRoundRef.current = newRoom.currentRound;
      }
      
      setRoom(newRoom);
      // Set isDrawer based on hostUsername
      setIsDrawer(newRoom.hostUserId === user?.id);
      
      // Set initial timeLeft when room is loaded
      setTimeLeft(newRoom.roundTimeLimit);
      
      // Auto-rejoin if not in users list (e.g., after refresh)
      const isInRoom = newRoom.users.some(p => p.userId === user?.id);
      if (!isInRoom) {
        console.log('[GameRoom] User not in room, auto-rejoining...');
        try {
          await apiService.joinGameRoom(roomId, user?.id || '');
          // Reload room to get updated user list
          const updatedRoom = await apiService.getGameRoom(roomId);
          if (updatedRoom) {
            setRoom(updatedRoom);
            setIsDrawer(updatedRoom.hostUserId === user?.id);
          }
        } catch (joinError) {
          console.error('[GameRoom] Failed to auto-rejoin:', joinError);
        }
      }
    } catch (error) {
      console.error('Failed to load room:', error);
    }
  };

  const handleStartGame = async () => {
    if (!roomId || !isDrawer || !user) return;
    try {
      await apiService.startGame(roomId);
      setMessage(t.gameRoom.gameStart);
      loadRoom();
    } catch (error) {
      console.error('Failed to start game:', error);
    }
  };

  const handleSendChat = async () => {
    if (!chatInput.trim() || !roomId) return;

    // Use new WebSocket chat message format
    const userId = user?.id || username;
    wsService.sendChatMessage(roomId, userId, username, chatInput);

    setChatInput('');
  };

  const handleLeaveRoom = async () => {
    if (!roomId || !user) return;
    
    try {
      const result = await apiService.leaveGameRoom(roomId, user.id);
      console.log('[GameRoom] Leave room result:', result ? 'room updated' : 'room deleted');
      // 방을 나가면 로비 채팅 히스토리 요청
      wsService.requestLobbyChatHistory();
      navigate('/rooms');
    } catch (error) {
      console.error('[GameRoom] Failed to leave room:', error);
      // 에러가 발생해도 로비로 이동
      navigate('/rooms');
    }
  };

  if (!roomId || !room) {
    return <div style={styles.loading}>{t.gameRoom.loading}</div>;
  }

  // const myuser = room.users.find((p: GameUser) => p.userId === user?.id);
  const sortedusers = [...room.users].sort((a, b) => b.score - a.score);

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <div>
          <h2 style={styles.title}>{t.gameRoom.title} - {username}</h2>
          <p style={styles.subtitle}>
            {isDrawer ? `🎨 ${t.gameRoom.host}` : `👀 ${t.gameRoom.user}`}
          </p>
        </div>
        <button onClick={handleLeaveRoom} style={styles.leaveButton}>
          {t.gameRoom.leave}
        </button>
      </div>

      {/* Game Info */}
      <div style={styles.gameInfo}>
        <div style={styles.infoBox}>
          <span style={styles.infoLabel}>{t.gameRoom.round}</span>
          <span style={styles.infoValue}>
            {room.status === 'WAITING' ? `0/${room.totalRounds}` : `${room.currentRound}/${room.totalRounds}`}
          </span>
        </div>
        <div style={styles.infoBox}>
          <span style={styles.infoLabel}>{t.gameRoom.timeLeft}</span>
          <span style={styles.infoValue}>{timeLeft ?? room.roundTimeLimit}{t.gameRoom.seconds}</span>
        </div>
        <div style={styles.infoBox}>
          <span style={styles.infoLabel}>{t.gameRoom.status}</span>
          <span style={styles.infoValue}>
            {room.status === 'WAITING' && t.gameRoom.statusValues.waiting}
            {room.status === 'PLAYING' && t.gameRoom.statusValues.playing}
            {room.status === 'FINISHED' && t.gameRoom.statusValues.finished}
          </span>
        </div>
      </div>

      {/* Current Word - Only for WORDCHAIN game (not yet implemented) */}
      {isDrawer && room.status === 'PLAYING' && room.gameType === 'WORDCHAIN' && (
        <div style={styles.wordDisplay}>
          {t.gameRoom.currentWord}: <strong>{/* Word display to be implemented */}</strong>
        </div>
      )}

      {/* Message */}
      {message && <div style={styles.message}>{message}</div>}

      <div style={{...styles.content, flexDirection: isMobile ? 'column' : 'row'}}>
        {/* Drawing Area */}
        <div style={styles.leftPanel}>
          {isDrawer && (
            <div style={styles.colorPalette}>
              {colors.map((color) => (
                <div
                  key={color}
                  onClick={() => setSelectedColor(color)}
                  style={{
                    ...styles.colorBox,
                    backgroundColor: color,
                    border: selectedColor === color ? '3px solid #333' : '1px solid #ccc',
                  }}
                />
              ))}
            </div>
          )}
          
        <DrawingCanvas
          uuid={roomId}
          allowDrawing={isDrawer && room.status === 'PLAYING'}
          selectedColor={selectedColor}
          username={username}
          clearTrigger={clearCanvasTrigger}
        />          {/* Start Button (Drawer only) */}
          {isDrawer && room.status === 'WAITING' && (
            <button onClick={handleStartGame} style={styles.startButton}>
              {t.gameRoom.gameStart}
            </button>
          )}
        </div>

        {/* Users Scoreboard */}
        <div style={styles.rightPanel}>
          <div style={styles.scoreHeader}>{t.gameRoom.users}</div>
          <div style={styles.userList}>
            {sortedusers.map((gameUser: GameUser, index: number) => (
              <div 
                key={gameUser.userId}
                style={{
                  ...styles.userCard,
                  ...(gameUser.userId === user?.id ? styles.myUser : {}),
                }}
              >
                <div style={styles.userRank}>#{index + 1}</div>
                <div style={styles.userInfo}>
                  <div style={styles.userName}>
                    {gameUser.name}
                    {gameUser.userId === room.hostUserId && ' 👑'}
                  </div>
                  <div style={styles.userStats}>
                    {t.gameRoom.score}: {gameUser.score}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Chat Section */}
          <div style={styles.chatSection}>
            <div style={styles.chatHeader}>{t.gameRoom.chat}</div>
            <div style={styles.chatMessages}>
              {chatMessages.map((msg, index) => (
                <div 
                  key={index}
                  style={{
                    ...styles.chatMessage,
                    ...(msg.type === 'system' ? styles.systemMessage : {}),
                    ...(msg.type === 'answer' ? styles.answerMessage : {}),
                    ...(msg.username === username && msg.type === 'chat' ? styles.myMessage : {}),
                  }}
                >
                  <span style={styles.chatUsername}>{msg.username}:</span>
                  <span style={styles.chatText}>{msg.text}</span>
                </div>
              ))}
              <div ref={chatEndRef} />
            </div>
            <div style={styles.chatInputSection}>
              <input
                type="text"
                placeholder={t.gameRoom.messagePlaceholder}
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && handleSendChat()}
                style={styles.chatInput}
              />
              <button onClick={handleSendChat} style={styles.chatSendButton}>
                {t.gameRoom.send}
              </button>
            </div>
          </div>

          {/* Game Result */}
          {room.status === 'FINISHED' && (
            <div style={styles.resultSection}>
              <h3 style={styles.resultTitle}>🎊 {t.gameRoom.gameFinished}</h3>
              <div style={styles.winner}>
                {t.gameRoom.winner}: {sortedusers[0]?.name} ({sortedusers[0]?.score}{t.gameRoom.score})
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    padding: '10px',
  },
  loading: {
    minHeight: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '24px',
    color: 'white',
  },
  header: {
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    maxWidth: '1200px',
    margin: '0 auto 10px',
    padding: '15px',
    background: 'white',
    borderRadius: '12px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
  },
  title: {
    margin: '0 0 5px 0',
    color: '#333',
    fontSize: 'clamp(16px, 4vw, 24px)',
  },
  subtitle: {
    margin: 0,
    color: '#666',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
  },
  leaveButton: {
    padding: '8px 16px',
    background: '#e53e3e',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontWeight: 'bold',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
    whiteSpace: 'nowrap',
  },
  gameInfo: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '10px',
    maxWidth: '1200px',
    margin: '0 auto 10px',
    justifyContent: 'center',
  },
  infoBox: {
    background: 'white',
    padding: '10px 20px',
    borderRadius: '8px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    minWidth: '80px',
  },
  infoLabel: {
    fontSize: 'clamp(10px, 2vw, 12px)',
    color: '#666',
    marginBottom: '5px',
  },
  infoValue: {
    fontSize: 'clamp(16px, 4vw, 20px)',
    fontWeight: 'bold',
    color: '#333',
  },
  wordDisplay: {
    maxWidth: '1200px',
    margin: '0 auto 10px',
    padding: '15px',
    background: '#ffd700',
    borderRadius: '12px',
    textAlign: 'center',
    fontSize: 'clamp(16px, 4vw, 24px)',
    fontWeight: 'bold',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
  },
  wordHint: {
    fontWeight: 'normal',
    fontSize: 'clamp(12px, 2.5vw, 16px)',
    color: '#553c00',
    marginLeft: '6px',
  },
  message: {
    maxWidth: '1200px',
    margin: '0 auto 10px',
    padding: '12px',
    background: '#48bb78',
    color: 'white',
    borderRadius: '8px',
    textAlign: 'center',
    fontWeight: 'bold',
    fontSize: 'clamp(14px, 3vw, 18px)',
  },
  content: {
    display: 'flex',
    flexDirection: 'row',
    gap: '10px',
    maxWidth: '1200px',
    margin: '0 auto',
  },
  leftPanel: {
    flex: 1,
    background: 'white',
    padding: '15px',
    borderRadius: '12px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
    minWidth: 0,
  },
  colorPalette: {
    display: 'flex',
    gap: '6px',
    marginBottom: '15px',
    justifyContent: 'center',
    flexWrap: 'wrap',
  },
  colorBox: {
    width: 'clamp(30px, 8vw, 40px)',
    height: 'clamp(30px, 8vw, 40px)',
    borderRadius: '6px',
    cursor: 'pointer',
    transition: 'transform 0.2s',
    flexShrink: 0,
  },
  answerSection: {
    marginTop: '15px',
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
  },
  answerInput: {
    flex: 1,
    minWidth: '150px',
    padding: '10px',
    fontSize: 'clamp(14px, 3vw, 16px)',
    border: '2px solid #ddd',
    borderRadius: '6px',
  },
  submitButton: {
    padding: '10px 20px',
    background: '#48bb78',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontWeight: 'bold',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
  },
  startButton: {
    marginTop: '15px',
    width: '100%',
    padding: '12px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontWeight: 'bold',
    fontSize: 'clamp(14px, 3vw, 18px)',
  },
  rightPanel: {
    width: '100%',
    maxWidth: '350px',
    background: 'white',
    borderRadius: '12px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
    display: 'flex',
    flexDirection: 'column',
    minHeight: '400px',
  },
  scoreHeader: {
    padding: '12px',
    background: '#667eea',
    color: 'white',
    fontWeight: 'bold',
    fontSize: 'clamp(14px, 3vw, 18px)',
    borderRadius: '12px 12px 0 0',
    textAlign: 'center',
  },
  userList: {
    padding: '12px',
    maxHeight: '200px',
    overflowY: 'auto',
    borderBottom: '2px solid #e2e8f0',
  },
  userCard: {
    display: 'flex',
    alignItems: 'center',
    padding: '10px',
    marginBottom: '8px',
    background: '#f7fafc',
    borderRadius: '8px',
    border: '2px solid transparent',
  },
  myUser: {
    background: '#e6fffa',
    border: '2px solid #48bb78',
  },
  userRank: {
    fontSize: 'clamp(18px, 4vw, 24px)',
    fontWeight: 'bold',
    color: '#667eea',
    marginRight: '10px',
    minWidth: '35px',
  },
  userInfo: {
    flex: 1,
    minWidth: 0,
  },
  userName: {
    fontSize: 'clamp(14px, 3vw, 16px)',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '2px',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  userStats: {
    fontSize: 'clamp(10px, 2vw, 12px)',
    color: '#666',
  },
  chatSection: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    borderTop: '2px solid #e2e8f0',
    minHeight: '200px',
  },
  chatHeader: {
    padding: '10px 12px',
    background: '#f7fafc',
    fontWeight: 'bold',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
    color: '#333',
  },
  chatMessages: {
    flex: 1,
    padding: '10px',
    overflowY: 'auto',
    maxHeight: '250px',
    minHeight: '150px',
    background: '#fafafa',
  },
  chatMessage: {
    padding: '6px 10px',
    marginBottom: '6px',
    background: 'white',
    borderRadius: '8px',
    fontSize: 'clamp(11px, 2.5vw, 13px)',
    wordWrap: 'break-word',
    overflowWrap: 'break-word',
  },
  systemMessage: {
    background: '#fff5e1',
    color: '#d69e2e',
    fontStyle: 'italic',
  },
  answerMessage: {
    background: '#e6f7ff',
    borderLeft: '3px solid #1890ff',
  },
  myMessage: {
    background: '#e6fffa',
    borderLeft: '3px solid #48bb78',
  },
  chatUsername: {
    fontWeight: 'bold',
    marginRight: '6px',
    color: '#667eea',
    fontSize: 'clamp(11px, 2.5vw, 13px)',
  },
  chatText: {
    color: '#333',
    fontSize: 'clamp(11px, 2.5vw, 13px)',
  },
  chatInputSection: {
    display: 'flex',
    gap: '6px',
    padding: '10px',
    borderTop: '2px solid #e2e8f0',
    background: 'white',
    flexWrap: 'wrap',
  },
  chatInput: {
    flex: 1,
    minWidth: '120px',
    padding: '8px 10px',
    border: '1px solid #ddd',
    borderRadius: '6px',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
  },
  chatSendButton: {
    padding: '8px 14px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontWeight: 'bold',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
    whiteSpace: 'nowrap',
  },
  resultSection: {
    padding: '15px',
    borderTop: '2px solid #e2e8f0',
    textAlign: 'center',
  },
  resultTitle: {
    margin: '0 0 8px 0',
    color: '#333',
    fontSize: 'clamp(16px, 4vw, 20px)',
  },
  winner: {
    fontSize: 'clamp(14px, 3vw, 18px)',
    fontWeight: 'bold',
    color: '#667eea',
  },
};

export default GameRoom;
