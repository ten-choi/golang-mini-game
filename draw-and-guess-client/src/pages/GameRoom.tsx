import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import DrawingCanvas from '../components/DrawingCanvas';
import { wsService } from '../services/websocket';
import { apiService } from '../services/api';
import { GameRoom as GameRoomType, Player } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const GameRoom: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const { language: preferredLanguage, t } = useLanguage();
  const [username] = useState(sessionStorage.getItem('username') || 'Guest');
  const [selectedColor, setSelectedColor] = useState('#000000');
  const [room, setRoom] = useState<GameRoomType | null>(null);
  const [isDrawer, setIsDrawer] = useState(false);
  const [message, setMessage] = useState('');
  const [timeLeft, setTimeLeft] = useState(60);
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
      () => {}, // Connected
      (error) => console.error('WebSocket connection error:', error)
    );
    
    loadRoom();

    const gameSubscription = wsService.subscribe(`game/${roomId}`, (data) => {
      if (data.type === 'update') {
        // Update timer from WebSocket if available
        if (data.time_left !== undefined) {
          setTimeLeft(data.time_left);
        }
        if (data.game_status !== undefined && data.game_status === 'finished') {
          setMessage(t.gameRoom.gameFinished);
        }
        // Always reload full room data
        loadRoom();
      }
    });

    const chatSubscription = wsService.subscribe(`chat/${roomId}`, (data) => {
      setChatMessages((prev) => [...prev, {
        username: data.username || 'System',
        text: data.message,
        type: data.type || 'chat'
      }]);
    });

    return () => {
      if (gameSubscription) gameSubscription.unsubscribe();
      if (chatSubscription) chatSubscription.unsubscribe();
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [roomId]);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  useEffect(() => {
    // No longer need polling - timer updates come via WebSocket
    // Only load room initially and when game state changes
  }, [room?.game_status]);

  const loadRoom = async () => {
    if (!roomId) return;
    try {
      const rooms = await apiService.getGameRooms(roomId);
      if (rooms.length > 0) {
        const newRoom = rooms[0];
        
        // Check if round changed - clear canvas if so
        if (prevRoundRef.current !== newRoom.round_number) {
          setClearCanvasTrigger(prev => prev + 1);
          prevRoundRef.current = newRoom.round_number;
        }
        
        setRoom(newRoom);
        setIsDrawer(newRoom.drawer_user === username);
        setTimeLeft(newRoom.time_left);
      }
    } catch (error) {
      console.error('Failed to load room:', error);
    }
  };

  const handleStartGame = async () => {
    if (!roomId || !isDrawer) return;
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

    const chatMessage = {
      username: username,
      message: chatInput,
      type: 'chat'
    };

    // Send to WebSocket for real-time broadcasting (will show in both screens)
    wsService.sendMessage(roomId, 'chat', chatMessage);

    // If player (not drawer) and game is playing, check if it's correct answer
    if (room?.drawer_user !== username && room?.game_status === 'playing') {
      try {
        const response = await apiService.handleChatMessage(roomId, username, chatInput);
        
        if (response.result.is_correct) {
          setMessage(`${t.gameRoom.correctAnswer} 🎉`);
          wsService.sendMessage(roomId, 'chat', {
            username: t.gameRoom.system,
            message: `${username}${t.gameRoom.answeredCorrectly} 🎉`,
            type: 'system'
          });
          wsService.sendMessage(roomId, 'game', {
            type: 'update'
          });
          setTimeout(() => {
            setMessage('');
            loadRoom();
          }, 2000);
        }
      } catch (error) {
        console.error('Failed to check answer:', error);
      }
    }

    setChatInput('');
  };

  const handleLeaveRoom = () => {
    navigate('/rooms');
  };

  if (!roomId || !room) {
    return <div style={styles.loading}>{t.gameRoom.loading}</div>;
  }

  const myPlayer = room.players.find((p: Player) => p.username === username);
  const sortedPlayers = [...room.players].sort((a, b) => b.score - a.score);

  const localizedWord = room.current_word_translations?.[preferredLanguage] || room.current_word;

  return (
    <div style={styles.container}>
      <div style={styles.header}>
        <div>
          <h2 style={styles.title}>{t.gameRoom.title} - {username}</h2>
          <p style={styles.subtitle}>
            {isDrawer ? `🎨 ${t.gameRoom.host}` : `👀 ${t.gameRoom.player}`}
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
          <span style={styles.infoValue}>{room.round_number}/{room.max_rounds}</span>
        </div>
        <div style={styles.infoBox}>
          <span style={styles.infoLabel}>{t.gameRoom.timeLeft}</span>
          <span style={styles.infoValue}>{timeLeft}{t.gameRoom.seconds}</span>
        </div>
        <div style={styles.infoBox}>
          <span style={styles.infoLabel}>{t.gameRoom.status}</span>
          <span style={styles.infoValue}>
            {room.game_status === 'waiting' && t.gameRoom.statusValues.waiting}
            {room.game_status === 'playing' && t.gameRoom.statusValues.playing}
            {room.game_status === 'finished' && t.gameRoom.statusValues.finished}
          </span>
        </div>
      </div>

      {/* Current Word (Drawer only) */}
      {isDrawer && room.game_status === 'playing' && (
        <div style={styles.wordDisplay}>
          {t.gameRoom.currentWord}: <strong>{localizedWord}</strong>
          {localizedWord && localizedWord !== room.current_word && (
            <span style={styles.wordHint}> ({room.current_word})</span>
          )}
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
          allowDrawing={isDrawer && room.game_status === 'playing'}
          selectedColor={selectedColor}
          clearTrigger={clearCanvasTrigger}
        />          {/* Start Button (Drawer only) */}
          {isDrawer && room.game_status === 'waiting' && (
            <button onClick={handleStartGame} style={styles.startButton}>
              {t.gameRoom.gameStart}
            </button>
          )}
        </div>

        {/* Players Scoreboard */}
        <div style={styles.rightPanel}>
          <div style={styles.scoreHeader}>{t.gameRoom.players}</div>
          <div style={styles.playerList}>
            {sortedPlayers.map((player: Player, index: number) => (
              <div 
                key={player.username}
                style={{
                  ...styles.playerCard,
                  ...(player.username === username ? styles.myPlayer : {}),
                }}
              >
                <div style={styles.playerRank}>#{index + 1}</div>
                <div style={styles.playerInfo}>
                  <div style={styles.playerName}>
                    {player.username}
                    {player.username === room.drawer_user && ' 👑'}
                  </div>
                  <div style={styles.playerStats}>
                    {t.gameRoom.score}: {player.score} | {t.gameRoom.attempts}: {player.attempts}/3
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
          {room.game_status === 'finished' && (
            <div style={styles.resultSection}>
              <h3 style={styles.resultTitle}>🎊 {t.gameRoom.gameFinished}</h3>
              <div style={styles.winner}>
                {t.gameRoom.winner}: {sortedPlayers[0]?.username} ({sortedPlayers[0]?.score}{t.gameRoom.score})
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
  playerList: {
    padding: '12px',
    maxHeight: '200px',
    overflowY: 'auto',
    borderBottom: '2px solid #e2e8f0',
  },
  playerCard: {
    display: 'flex',
    alignItems: 'center',
    padding: '10px',
    marginBottom: '8px',
    background: '#f7fafc',
    borderRadius: '8px',
    border: '2px solid transparent',
  },
  myPlayer: {
    background: '#e6fffa',
    border: '2px solid #48bb78',
  },
  playerRank: {
    fontSize: 'clamp(18px, 4vw, 24px)',
    fontWeight: 'bold',
    color: '#667eea',
    marginRight: '10px',
    minWidth: '35px',
  },
  playerInfo: {
    flex: 1,
    minWidth: 0,
  },
  playerName: {
    fontSize: 'clamp(14px, 3vw, 16px)',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '2px',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  playerStats: {
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
