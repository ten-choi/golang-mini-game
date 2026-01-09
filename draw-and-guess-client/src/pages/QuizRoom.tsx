import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { wsService } from '../services/websocket';
import { apiService } from '../services/api';
import { GameRoom, Player, OXQuiz, GeneralQuiz } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const QuizRoom: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [username] = useState(sessionStorage.getItem('username') || 'Guest');
  const [room, setRoom] = useState<GameRoom | null>(null);
  const [currentQuiz, setCurrentQuiz] = useState<OXQuiz | GeneralQuiz | null>(null);
  const [selectedAnswer, setSelectedAnswer] = useState<number | null>(null);
  const [answered, setAnswered] = useState(false);
  const [timeLeft, setTimeLeft] = useState<number | null>(null);
  const [chatMessages, setChatMessages] = useState<Array<{username: string, text: string, type: 'chat' | 'system'}>>([]);
  const timerRef = useRef<NodeJS.Timeout | null>(null);
  const chatEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!roomId) return;
    
    wsService.connect(
      () => {
        console.log('[QuizRoom] WebSocket connected');
        // Identify client with username and roomId
        wsService.identify(username, roomId);
      }, 
      (error) => console.error('[QuizRoom] WebSocket connection error:', error)
    );
    
    loadRoom();

    // Subscribe to game channel for room updates
    const gameSubscription = wsService.subscribe(`game/${roomId}`, (data) => {
      console.log('[QuizRoom] Received game message:', data);
      console.log('[QuizRoom] Message type:', data.type);
      console.log('[QuizRoom] Message data:', data.data);
      
      if (data.type === 'update' || data.type === 'room_update') {
        // If we have the full room data, update it directly
        if (data.data) {
          console.log('[QuizRoom] Updating room from WebSocket. Players count:', data.data.players?.length);
          setRoom(data.data);
        } else {
          // Fallback: reload room data from API
          console.log('[QuizRoom] No data in message, reloading from API');
          loadRoom();
        }
      }
      
      // Handle quiz message
      if (data.type === 'quiz') {
        console.log('[QuizRoom] ========== QUIZ RECEIVED FROM SUBSCRIPTION ==========');
        handleQuiz(data);
      }
      
      // Handle timer message
      if (data.type === 'timer') {
        handleTimer(data);
      }
    });

    const handleChat = (data: any) => {
      setChatMessages(prev => [...prev, {
        username: data.username,
        text: data.message,
        type: data.type || 'chat'
      }]);
    };

    const handleQuiz = (data: any) => {
      console.log('[QuizRoom] ========== QUIZ RECEIVED ==========');
      console.log('[QuizRoom] Raw data:', JSON.stringify(data, null, 2));
      // Server sends quiz data in 'data' field
      const quizData = data.data || data;
      console.log('[QuizRoom] Extracted quiz data:', JSON.stringify(quizData, null, 2));
      console.log('[QuizRoom] Quiz type:', quizData.type);
      console.log('[QuizRoom] Quiz question:', quizData.question);
      console.log('[QuizRoom] Quiz options:', quizData.options);
      console.log('[QuizRoom] Quiz options type:', typeof quizData.options);
      console.log('[QuizRoom] Quiz options is array:', Array.isArray(quizData.options));
      if (quizData.options) {
        console.log('[QuizRoom] Options length:', quizData.options.length);
        console.log('[QuizRoom] Options content:', quizData.options);
      }
      console.log('[QuizRoom] =====================================');
      setCurrentQuiz(quizData);
      setSelectedAnswer(null);
      setAnswered(false);
      // Don't set timeLeft here - wait for timer message from server
    };

    const handleTimer = (data: any) => {
      console.log('[QuizRoom] Received timer update:', data);
      // Server sends { data: { timeLeft: X } }
      const timeValue = data.data?.timeLeft !== undefined ? data.data.timeLeft : data.timeLeft;
      console.log('[QuizRoom] Time left:', timeValue);
      setTimeLeft(timeValue);
      
      // When timer reaches 0, reset answer state but keep quiz visible
      if (timeValue === 0) {
        console.log('[QuizRoom] Timer reached 0, waiting for next quiz...');
        setSelectedAnswer(null);
        setAnswered(false);
      }
    };

    wsService.addMessageHandler('chat', handleChat);
    wsService.addMessageHandler('quiz', handleQuiz);
    wsService.addMessageHandler('timer', handleTimer);

    return () => {
      if (gameSubscription) gameSubscription.unsubscribe();
      wsService.removeMessageHandler('chat', handleChat);
      wsService.removeMessageHandler('quiz', handleQuiz);
      wsService.removeMessageHandler('timer', handleTimer);
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [roomId]);

  // Remove client-side timer - now managed by server
  // useEffect(() => {
  //   if (room?.status === 'PLAYING' && timeLeft > 0) {
  //     timerRef.current = setInterval(() => {
  //       setTimeLeft(prev => Math.max(0, prev - 1));
  //     }, 1000);
  //   } else if (timerRef.current) {
  //     clearInterval(timerRef.current);
  //   }

  //   return () => {
  //     if (timerRef.current) clearInterval(timerRef.current);
  //   };
  // }, [room?.status, timeLeft]);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  const loadRoom = async () => {
    try {
      if (!roomId) return;
      const gameRoom = await apiService.getGameRoom(roomId);
      if (gameRoom) {
        setRoom(gameRoom);
        // Set initial timeLeft when room is loaded
        setTimeLeft(gameRoom.roundTimeLimit);
        
        // Auto-rejoin if not in players list (e.g., after refresh)
        const isInRoom = gameRoom.players.some(p => p.username === username);
        if (!isInRoom) {
          console.log('[QuizRoom] Player not in room, auto-rejoining...');
          try {
            await apiService.joinGameRoom(roomId, username);
            // Reload room to get updated player list
            const updatedRoom = await apiService.getGameRoom(roomId);
            if (updatedRoom) {
              setRoom(updatedRoom);
            }
          } catch (joinError) {
            console.error('[QuizRoom] Failed to auto-rejoin:', joinError);
          }
        }
      } else {
        console.error('[QuizRoom] Room not found:', roomId);
        alert('게임방을 찾을 수 없습니다.');
        navigate('/room-list');
      }
    } catch (error) {
      console.error('Failed to load room:', error);
    }
  };

  const handleStartGame = async () => {
    if (!roomId || room?.hostUsername !== username) return;
    try {
      await apiService.startGame(roomId);
      loadRoom();
    } catch (error) {
      console.error('Failed to start game:', error);
    }
  };

  const handleAnswerSelect = (answerIndex: number) => {
    if (answered || !roomId || !currentQuiz) return;
    setSelectedAnswer(answerIndex);
  };

  const handleSubmitAnswer = () => {
    if (answered || selectedAnswer === null || !roomId || !currentQuiz) return;
    
    setAnswered(true);

    // Send answer to WebSocket
    wsService.sendMessage(roomId, 'answer', {
      username: username,
      answer: selectedAnswer,
      quiz_id: (currentQuiz as any)._id
    });

    // Add system message
    setChatMessages(prev => [...prev, {
      username: t.gameRoom.system,
      text: `${username}님이 답을 제출했습니다.`,
      type: 'system'
    }]);
  };

  const handleLeaveRoom = async () => {
    if (!roomId) return;
    
    try {
      await apiService.leaveGameRoom(roomId, username);
    } catch (error) {
      console.error('[QuizRoom] Failed to leave room:', error);
    }
    
    navigate('/rooms');
  };

  if (!roomId || !room) {
    return <div style={styles.loading}>{t.gameRoom.loading}</div>;
  }

  const isHost = room.hostUsername === username;
  const sortedPlayers = [...room.players].sort((a, b) => b.score - a.score);
  const isOXQuiz = room.gameType === 'OX';

  return (
    <div style={styles.container}>
      {/* Header */}
      <div style={styles.header}>
        <div>
          <h2 style={styles.title}>
            {isOXQuiz ? '⭕❌ OX 퀴즈' : '📚 4지선다 상식퀴즈'}
          </h2>
          <p style={styles.subtitle}>
            {isHost ? '👑 호스트' : '👤 플레이어'} - {username}
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
          <span style={styles.infoLabel}>👥 플레이어</span>
          <span style={styles.infoValue}>{room.players.length}/4</span>
        </div>
      </div>

      <div style={styles.mainContent}>
        {/* Left: Quiz Area */}
        <div style={styles.leftPanel}>
          {room.status === 'WAITING' ? (
            <div style={styles.waitingArea}>
              <h3 style={styles.waitingTitle}>대기 중...</h3>
              <p style={styles.waitingText}>호스트가 게임을 시작할 때까지 기다려주세요</p>
              {isHost && (
                <button 
                  onClick={handleStartGame} 
                  style={styles.startButton}
                  disabled={room.players.length < 2}
                >
                  {room.players.length < 2 ? '최소 2명 필요' : '게임 시작'}
                </button>
              )}
            </div>
          ) : room.status === 'FINISHED' ? (
            <div style={styles.finishedArea}>
              <h3 style={styles.finishedTitle}>🏆 게임 종료!</h3>
              <div style={styles.finalRanking}>
                {sortedPlayers.map((player, index) => (
                  <div key={player.username} style={styles.rankItem}>
                    <span style={styles.rank}>
                      {index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : `${index + 1}위`}
                    </span>
                    <span style={styles.rankName}>{player.username}</span>
                    <span style={styles.rankScore}>{player.score}점</span>
                  </div>
                ))}
              </div>
              <button onClick={handleLeaveRoom} style={styles.backToLobbyButton}>
                로비로 돌아가기
              </button>
            </div>
          ) : currentQuiz ? (
            <div style={styles.quizArea}>
              <div style={styles.quizQuestion}>
                <div style={styles.questionLabel}>문제</div>
                <div style={styles.questionText}>{currentQuiz.question}</div>
              </div>

              {isOXQuiz ? (
                <div style={styles.oxAnswers}>
                  <button
                    style={{
                      ...styles.oxButton,
                      ...styles.oButton,
                      ...(selectedAnswer === 1 ? styles.selectedO : {}),
                      ...(answered ? styles.disabledButton : {})
                    }}
                    onClick={() => handleAnswerSelect(1)}
                    disabled={answered}
                  >
                    ⭕ O (참)
                  </button>
                  <button
                    style={{
                      ...styles.oxButton,
                      ...styles.xButton,
                      ...(selectedAnswer === 0 ? styles.selectedX : {}),
                      ...(answered ? styles.disabledButton : {})
                    }}
                    onClick={() => handleAnswerSelect(0)}
                    disabled={answered}
                  >
                    ❌ X (거짓)
                  </button>
                </div>
              ) : (
                <div style={styles.generalAnswers}>
                  {currentQuiz && (currentQuiz as any).options && Array.isArray((currentQuiz as any).options) ? (
                    (currentQuiz as any).options.map((option: string, index: number) => (
                      <button
                        key={index}
                        style={{
                          ...styles.generalButton,
                          ...(selectedAnswer === index ? styles.selectedGeneral : {}),
                          ...(answered ? styles.disabledButton : {})
                        }}
                        onClick={() => handleAnswerSelect(index)}
                        disabled={answered}
                      >
                        <span style={styles.optionNumber}>{index + 1}</span>
                        <span style={styles.optionText}>{option}</span>
                      </button>
                    ))
                  ) : (
                    <div style={styles.noOptions}>
                      <p>선택지를 불러오는 중...</p>
                      <p style={{fontSize: '12px', color: '#999'}}>
                        {JSON.stringify(currentQuiz)}
                      </p>
                    </div>
                  )}
                </div>
              )}

              {!answered && selectedAnswer !== null && (
                <button
                  style={styles.submitButton}
                  onClick={handleSubmitAnswer}
                >
                  정답 제출하기
                </button>
              )}

              {answered && (
                <div style={styles.answeredMessage}>
                  ✓ 답변 제출됨! 결과를 기다리는 중...
                </div>
              )}
            </div>
          ) : (
            <div style={styles.noQuiz}>
              <p>다음 문제를 준비 중입니다...</p>
            </div>
          )}
        </div>

        {/* Right: Players & Chat */}
        <div style={styles.rightPanel}>
          {/* Players */}
          <div style={styles.playersSection}>
            <h3 style={styles.sectionTitle}>👥 플레이어 ({room.players.length}/4)</h3>
            <div style={styles.playersList}>
              {sortedPlayers.map((player) => (
                <div key={player.username} style={styles.playerCard}>
                  <div style={styles.playerInfo}>
                    <span style={styles.playerName}>
                      {player.username === room.hostUsername && '👑 '}
                      {player.username}
                    </span>
                  </div>
                  <span style={styles.playerScore}>{player.score}점</span>
                </div>
              ))}
            </div>
          </div>

          {/* Chat */}
          <div style={styles.chatSection}>
            <h3 style={styles.sectionTitle}>💬 채팅</h3>
            <div style={styles.chatMessages}>
              {chatMessages.map((msg, index) => (
                <div
                  key={index}
                  style={msg.type === 'system' ? styles.systemMessage : styles.chatMessage}
                >
                  {msg.type === 'system' ? (
                    <span style={styles.systemText}>{msg.text}</span>
                  ) : (
                    <>
                      <span style={styles.chatUsername}>{msg.username}:</span>
                      <span style={styles.chatText}>{msg.text}</span>
                    </>
                  )}
                </div>
              ))}
              <div ref={chatEndRef} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    padding: '20px',
  },
  loading: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    fontSize: '24px',
    color: 'white',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    background: 'white',
    padding: '20px',
    borderRadius: '12px',
    marginBottom: '20px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
  },
  title: {
    margin: 0,
    fontSize: '24px',
    color: '#333',
  },
  subtitle: {
    margin: '5px 0 0 0',
    fontSize: '14px',
    color: '#666',
  },
  leaveButton: {
    padding: '10px 20px',
    background: '#FF6B6B',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontWeight: 'bold',
  },
  gameInfo: {
    display: 'flex',
    gap: '15px',
    marginBottom: '20px',
    flexWrap: 'wrap',
  },
  infoBox: {
    background: 'white',
    padding: '15px 20px',
    borderRadius: '8px',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    flex: 1,
    minWidth: '120px',
  },
  infoLabel: {
    fontSize: '12px',
    color: '#666',
    marginBottom: '5px',
  },
  infoValue: {
    fontSize: '20px',
    fontWeight: 'bold',
    color: '#333',
  },
  mainContent: {
    display: 'grid',
    gridTemplateColumns: '2fr 1fr',
    gap: '20px',
    height: 'calc(100vh - 250px)',
  },
  leftPanel: {
    background: 'white',
    borderRadius: '12px',
    padding: '20px',
    display: 'flex',
    flexDirection: 'column',
  },
  rightPanel: {
    display: 'flex',
    flexDirection: 'column',
    gap: '20px',
  },
  waitingArea: {
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100%',
  },
  waitingTitle: {
    fontSize: '32px',
    color: '#333',
    marginBottom: '10px',
  },
  waitingText: {
    fontSize: '18px',
    color: '#666',
    marginBottom: '30px',
  },
  startButton: {
    padding: '15px 40px',
    background: '#4CAF50',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
  },
  finishedArea: {
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100%',
  },
  finishedTitle: {
    fontSize: '36px',
    color: '#333',
    marginBottom: '30px',
  },
  finalRanking: {
    width: '100%',
    maxWidth: '400px',
    marginBottom: '30px',
  },
  rankItem: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '15px',
    background: '#f8f9fa',
    borderRadius: '8px',
    marginBottom: '10px',
  },
  rank: {
    fontSize: '24px',
    width: '60px',
  },
  rankName: {
    flex: 1,
    fontSize: '18px',
    fontWeight: 'bold',
  },
  rankScore: {
    fontSize: '20px',
    fontWeight: 'bold',
    color: '#667eea',
  },
  backToLobbyButton: {
    padding: '15px 40px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
  },
  quizArea: {
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
  },
  quizQuestion: {
    marginBottom: '30px',
  },
  questionLabel: {
    fontSize: '14px',
    color: '#666',
    marginBottom: '10px',
    fontWeight: 'bold',
  },
  questionText: {
    fontSize: '24px',
    fontWeight: 'bold',
    color: '#333',
    lineHeight: 1.5,
  },
  oxAnswers: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '20px',
    flex: 1,
  },
  oxButton: {
    fontSize: '32px',
    fontWeight: 'bold',
    border: '3px solid',
    borderRadius: '12px',
    cursor: 'pointer',
    transition: 'all 0.3s',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  oButton: {
    background: '#E8F5E9',
    borderColor: '#4CAF50',
    color: '#2E7D32',
  },
  xButton: {
    background: '#FFEBEE',
    borderColor: '#F44336',
    color: '#C62828',
  },
  selectedO: {
    background: '#4CAF50',
    color: 'white',
    transform: 'scale(1.05)',
  },
  selectedX: {
    background: '#F44336',
    color: 'white',
    transform: 'scale(1.05)',
  },
  generalAnswers: {
    display: 'flex',
    flexDirection: 'column',
    gap: '15px',
    flex: 1,
  },
  generalButton: {
    display: 'flex',
    alignItems: 'center',
    gap: '15px',
    padding: '20px',
    background: '#f8f9fa',
    border: '3px solid #e9ecef',
    borderRadius: '12px',
    fontSize: '18px',
    cursor: 'pointer',
    transition: 'all 0.3s',
    textAlign: 'left',
  },
  selectedGeneral: {
    background: '#667eea',
    color: 'white',
    borderColor: '#667eea',
    transform: 'scale(1.02)',
  },
  disabledButton: {
    opacity: 0.6,
    cursor: 'not-allowed',
  },
  optionNumber: {
    width: '40px',
    height: '40px',
    borderRadius: '50%',
    background: '#667eea',
    color: 'white',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontWeight: 'bold',
    fontSize: '20px',
    flexShrink: 0,
  },
  optionText: {
    flex: 1,
  },
  submitButton: {
    marginTop: '20px',
    padding: '15px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
    transition: 'all 0.3s',
  },
  answeredMessage: {
    padding: '15px',
    background: '#E8F5E9',
    color: '#2E7D32',
    borderRadius: '8px',
    textAlign: 'center',
    fontWeight: 'bold',
    marginTop: '20px',
  },
  noQuiz: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    height: '100%',
    fontSize: '20px',
    color: '#666',
  },
  playersSection: {
    background: 'white',
    borderRadius: '12px',
    padding: '20px',
    maxHeight: '40%',
    overflow: 'auto',
  },
  chatSection: {
    background: 'white',
    borderRadius: '12px',
    padding: '20px',
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    minHeight: 0,
  },
  sectionTitle: {
    margin: '0 0 15px 0',
    fontSize: '16px',
    color: '#333',
  },
  playersList: {
    display: 'flex',
    flexDirection: 'column',
    gap: '10px',
  },
  playerCard: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '12px',
    background: '#f8f9fa',
    borderRadius: '8px',
  },
  playerInfo: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
  },
  playerName: {
    fontWeight: 'bold',
    fontSize: '14px',
  },
  playerScore: {
    fontWeight: 'bold',
    fontSize: '16px',
    color: '#667eea',
  },
  chatMessages: {
    flex: 1,
    overflowY: 'auto',
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
    minHeight: 0,
  },
  chatMessage: {
    padding: '8px 12px',
    background: '#f8f9fa',
    borderRadius: '8px',
    fontSize: '14px',
  },
  systemMessage: {
    padding: '8px 12px',
    background: '#E3F2FD',
    borderRadius: '8px',
    fontSize: '13px',
    textAlign: 'center',
  },
  chatUsername: {
    fontWeight: 'bold',
    marginRight: '8px',
    color: '#667eea',
  },
  chatText: {
    color: '#333',
  },
  systemText: {
    color: '#1976D2',
    fontStyle: 'italic',
  },
};

export default QuizRoom;
