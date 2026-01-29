import React, { useState, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { wsService } from '../services/websocket';
import { apiService } from '../services/api';
import { GameRoom, WordchainPrompt, User } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const WordchainRoom: React.FC = () => {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [user] = useState<User | null>(() => {
    const userData = sessionStorage.getItem('user');
    return userData ? JSON.parse(userData) : null;
  });
  const [username] = useState(user?.name || sessionStorage.getItem('username') || 'Guest');
  const [room, setRoom] = useState<GameRoom | null>(null);
  const [currentPrompt, setCurrentPrompt] = useState<WordchainPrompt | null>(null);
  const [wordInput, setWordInput] = useState('');
  const [timeLeft, setTimeLeft] = useState<number | null>(null);
  const [chatMessages, setChatMessages] = useState<Array<{username: string, text: string, type: 'chat' | 'system' | 'correct' | 'wrong'}>>([]);
  const [lastWord, setLastWord] = useState('');
  const [roundEndInfo, setRoundEndInfo] = useState<{round: number, reason: string, countdown: number} | null>(null);
  const [gameEndInfo, setGameEndInfo] = useState<{users: Array<{name: string, score: number}>, message: string} | null>(null);
  const chatEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!roomId) return;
    
    wsService.connect(
      () => {
        console.log('[WordchainRoom] WebSocket connected');
        // Identify with userId
        if (user) {
          wsService.identify(user.id, roomId);
        }
      }, 
      (error) => console.error('[WordchainRoom] WebSocket connection error:', error)
    );
    
    loadRoom();

    const gameSubscription = wsService.subscribe(`game/${roomId}`, (data) => {
      console.log('[WordchainRoom] Received game message:', data);
      
      // Handle messages from graph package (format: {type: "word_result", userName: ..., word: ..., correct: ...})
      if (data.type === 'word_result') {
        console.log('[WordchainRoom] Word result:', data);
        if (data.correct) {
          setChatMessages(prev => [...prev, {
            username: t.gameRoom.system,
            text: `${data.userName}님이 정답! "${data.word}" (+100점)`,
            type: 'correct'
          }]);
          setLastWord(data.word);
        } else {
          setChatMessages(prev => [...prev, {
            username: t.gameRoom.system,
            text: `${data.userName}: "${data.word}" - ${data.reason}`,
            type: 'wrong'
          }]);
        }
      } else if (data.type === 'wordchain_prompt') {
        console.log('[WordchainRoom] Received wordchain prompt:', data);
        setCurrentPrompt(data.prompt);
        setLastWord(data.lastWord || '');
        setWordInput('');
      } else if (data.type === 'round_end' || data.type === 'ROUND_END') {
        console.log('[WordchainRoom] Round ended:', data);
        const roundData = data.data || data;
        setRoundEndInfo({
          round: roundData.round,
          reason: roundData.reason,
          countdown: 3
        });
        
        let count = 3;
        const countdownInterval = setInterval(() => {
          count--;
          setRoundEndInfo(prev => prev ? {...prev, countdown: count} : null);
          if (count <= 0) {
            clearInterval(countdownInterval);
            setRoundEndInfo(null);
            loadRoom();
          }
        }, 1000);
      } else if (data.type === 'game_end' || data.type === 'GAME_END') {
        console.log('[WordchainRoom] Game ended:', data);
        setRoundEndInfo(null);
        const gameData = data.data || data;
        const sortedusers = [...gameData.users].sort((a, b) => b.score - a.score);
        setGameEndInfo({
          users: sortedusers,
          message: gameData.message
        });
      }
      // Handle room_deleted message
      else if (data.type === 'room_deleted') {
        console.log('[WordchainRoom] Room deleted:', data.data);
        alert('방이 삭제되었습니다.');
        navigate('/room-list');
      }
      // Handle websocket handler messages (format: {type: "CHAT_MESSAGE", payload: {...}})
      else if (data.type === 'CHAT_MESSAGE' && data.payload) {
        const payload = data.payload;
        setChatMessages(prev => [...prev, {
          username: payload.username,
          text: payload.message,
          type: 'chat'
        }]);
      }
      // Handle room update messages
      else if (data.type === 'update' || data.type === 'room_update' || data.type === 'ROOM_UPDATE') {
        if (data.data) {
          setRoom(data.data);
        } else if (data.payload) {
          setRoom(data.payload);
        } else {
          loadRoom();
        }
      }
      // Handle timer updates
      else if (data.type === 'timer' || data.type === 'TIMER') {
        setTimeLeft(data.timeLeft);
      }
    });

    const handleChat = (data: any) => {
      setChatMessages(prev => [...prev, {
        username: data.username,
        text: data.message,
        type: data.type || 'chat'
      }]);
    };

    const handleWordchainPrompt = (data: any) => {
      console.log('[WordchainRoom] Received wordchain prompt:', data);
      setCurrentPrompt(data.prompt);
      setLastWord(data.lastWord || '');
      setWordInput('');
    };

    const handleTimer = (data: any) => {
      setTimeLeft(data.timeLeft);
    };

    const handleWordResult = (data: any) => {
      console.log('[WordchainRoom] Word result:', data);
      if (data.correct) {
        setChatMessages(prev => [...prev, {
          username: t.gameRoom.system,
          text: `${data.username}님이 정답! "${data.word}" (+100점)`,
          type: 'correct'
        }]);
        setLastWord(data.word);
      } else {
        setChatMessages(prev => [...prev, {
          username: t.gameRoom.system,
          text: `${data.username}: "${data.word}" - ${data.reason}`,
          type: 'wrong'
        }]);
      }
    };

    const handleRoundEnd = (data: any) => {
      console.log('[WordchainRoom] Round ended:', data);
      setRoundEndInfo({
        round: data.round,
        reason: data.reason,
        countdown: 3
      });
      
      // 3초 카운트다운
      let count = 3;
      const countdownInterval = setInterval(() => {
        count--;
        setRoundEndInfo(prev => prev ? {...prev, countdown: count} : null);
        if (count <= 0) {
          clearInterval(countdownInterval);
          setRoundEndInfo(null);
        }
      }, 1000);
    };

    const handleGameEnd = (data: any) => {
      console.log('[WordchainRoom] Game ended:', data);
      // 라운드 종료 팝업 닫기
      setRoundEndInfo(null);
      
      // 점수 정렬 (높은 점수 순)
      const sortedusers = [...data.users].sort((a, b) => b.score - a.score);
      
      setGameEndInfo({
        users: sortedusers,
        message: data.message
      });
    };

    // Handle room deletion
    const handleRoomDeleted = (data: any) => {
      console.log('[WordchainRoom] Room has been deleted:', data);
      alert('방이 삭제되었습니다.');
      navigate('/room-list');
    };

    wsService.addMessageHandler('chat', handleChat);
    wsService.addMessageHandler('wordchain_prompt', handleWordchainPrompt);
    wsService.addMessageHandler('timer', handleTimer);
    wsService.addMessageHandler('word_result', handleWordResult);
    wsService.addMessageHandler('round_end', handleRoundEnd);
    wsService.addMessageHandler('game_end', handleGameEnd);
    wsService.addMessageHandler('room_deleted', handleRoomDeleted);

    return () => {
      if (gameSubscription) gameSubscription.unsubscribe();
      wsService.removeMessageHandler('chat', handleChat);
      wsService.removeMessageHandler('wordchain_prompt', handleWordchainPrompt);
      wsService.removeMessageHandler('timer', handleTimer);
      wsService.removeMessageHandler('word_result', handleWordResult);
      wsService.removeMessageHandler('round_end', handleRoundEnd);
      wsService.removeMessageHandler('game_end', handleGameEnd);
      wsService.removeMessageHandler('room_deleted', handleRoomDeleted);
    };
  }, [roomId]);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  // Update timeLeft when room is loaded or status changes
  useEffect(() => {
    if (room) {
      // Always sync timeLeft with room's roundTimeLimit when not playing
      if (room.status !== 'PLAYING') {
        setTimeLeft(room.roundTimeLimit);
      }
    }
  }, [room?.roundTimeLimit, room?.status]);

  const loadRoom = async () => {
    try {
      if (!roomId || !user) return;
      const gameRoom = await apiService.getGameRoom(roomId);
      if (gameRoom) {
        setRoom(gameRoom);
        // Set initial timeLeft immediately when room is loaded
        setTimeLeft(gameRoom.roundTimeLimit);
        
        // Auto-rejoin if not in users list (e.g., after refresh)
        const isInRoom = gameRoom.users.some(p => p.userId === user.id);
        if (!isInRoom) {
          console.log('[WordchainRoom] User not in room, auto-rejoining...');
          try {
            await apiService.joinGameRoom(roomId, user.id);
            // Auto-set ready after joining
            await apiService.setReady(roomId, user.id, true);
            // Reload room to get updated user list
            const updatedRoom = await apiService.getGameRoom(roomId);
            if (updatedRoom) {
              setRoom(updatedRoom);
            }
          } catch (joinError) {
            console.error('[WordchainRoom] Failed to auto-rejoin:', joinError);
          }
        } else {
          // Check if user is already ready, if not, set ready
          const currentUser = gameRoom.users.find(p => p.userId === user.id);
          if (currentUser && !currentUser.isReady && gameRoom.status === 'WAITING') {
            console.log('[WordchainRoom] Auto-setting user ready...');
            try {
              await apiService.setReady(roomId, user.id, true);
              const updatedRoom = await apiService.getGameRoom(roomId);
              if (updatedRoom) {
                setRoom(updatedRoom);
              }
            } catch (readyError) {
              console.error('[WordchainRoom] Failed to set ready:', readyError);
            }
          }
        }
      } else {
        console.error('[WordchainRoom] Room not found:', roomId);
        alert('게임방을 찾을 수 없습니다.');
        navigate('/room-list');
      }
    } catch (error) {
      console.error('Failed to load room:', error);
    }
  };

  const handleStartGame = async () => {
    if (!roomId || !user || room?.hostUserId !== user.id) return;
    try {
      await apiService.startGame(roomId);
      loadRoom();
    } catch (error) {
      console.error('Failed to start game:', error);
    }
  };

  const handleSubmitWord = () => {
    if (!wordInput.trim() || !roomId) return;

    wsService.sendWordchainSubmit(roomId, username, wordInput.trim(), lastWord);
    setWordInput('');
  };

  const handleLeaveRoom = async () => {
    if (!roomId || !user) return;
    
    try {
      const result = await apiService.leaveGameRoom(roomId, user.id);
      console.log('[WordchainRoom] Leave room result:', result ? 'room updated' : 'room deleted');
      // 방을 나가면 로비 채팅 히스토리 요청
      wsService.requestLobbyChatHistory();
      navigate('/rooms');
    } catch (error) {
      console.error('[WordchainRoom] Failed to leave room:', error);
      // 에러가 발생해도 로비로 이동
      navigate('/rooms');
    }
  };

  if (!roomId || !room) {
    return <div style={styles.loading}>로딩 중...</div>;
  }

  const isHost = user && room.hostUserId === user.id;
  const sortedusers = [...room.users].sort((a, b) => b.score - a.score);

  return (
    <div style={styles.container}>
      {/* Header */}
      <div style={styles.header}>
        <div>
          <h2 style={styles.title}>🔤 끝말잇기</h2>
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
          <span style={styles.infoValue}>{room.users.length}/{room.maxUsers}</span>
        </div>
        {room.status === 'PLAYING' && room.currentTurnUserId && (
          <div style={{...styles.infoBox, background: '#4CAF50', color: 'white'}}>
            <span style={styles.infoLabel}>🎯 현재 턴</span>
            <span style={styles.infoValue}>
              {room.users.find(u => u.userId === room.currentTurnUserId)?.name || 'Unknown'}
            </span>
          </div>
        )}
      </div>

      <div style={styles.mainContent}>
        {/* Left: Game Area */}
        <div style={styles.leftPanel}>
          {room.status === 'WAITING' ? (
            <div style={styles.waitingArea}>
              <h3 style={styles.waitingTitle}>대기 중...</h3>
              <p style={styles.waitingText}>호스트가 게임을 시작할 때까지 기다려주세요</p>
              <div style={styles.rules}>
                <h4 style={styles.rulesTitle}>게임 규칙</h4>
                <ul style={styles.rulesList}>
                  <li>일본어 히라가나/카타카나 단어를 입력합니다</li>
                  <li>이전 단어의 마지막 글자로 시작해야 합니다</li>
                  <li>ん으로 끝나는 단어는 사용 불가</li>
                  <li>정답 +100점</li>
                </ul>
              </div>
              {isHost && (
                <button 
                  onClick={handleStartGame} 
                  style={styles.startButton}
                  disabled={room.users.length < 2}
                >
                  {room.users.length < 2 ? '최소 2명 필요' : '게임 시작'}
                </button>
              )}
            </div>
          ) : room.status === 'FINISHED' ? (
            <div style={styles.finishedArea}>
              <h3 style={styles.finishedTitle}>🏆 게임 종료!</h3>
              <div style={styles.finalRanking}>
                {sortedusers.map((user, index) => (
                  <div key={user.name} style={styles.rankItem}>
                    <span style={styles.rank}>
                      {index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : `${index + 1}위`}
                    </span>
                    <span style={styles.rankName}>{user.name}</span>
                    <span style={styles.rankScore}>{user.score}점</span>
                  </div>
                ))}
              </div>
              <button onClick={handleLeaveRoom} style={styles.backToLobbyButton}>
                로비로 돌아가기
              </button>
            </div>
          ) : (
            <div style={styles.gameArea}>
              {/* Word History */}
              {room && (room as any).wordchainUsedWords && (room as any).wordchainUsedWords.length > 0 && (
                <div style={styles.wordHistory}>
                  <div style={styles.wordHistoryTitle}>📝 사용된 단어</div>
                  <div style={styles.wordHistoryList}>
                    {(room as any).wordchainUsedWords.map((word: string, index: number) => (
                      <span key={index} style={styles.wordHistoryItem}>
                        {word}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {lastWord && (
                <div style={styles.lastWordDisplay}>
                  <div style={styles.lastWordLabel}>마지막 단어</div>
                  <div style={styles.lastWordText}>{lastWord}</div>
                  <div style={styles.lastWordHint}>
                    "{lastWord[lastWord.length - 1]}"로 시작하는 단어를 입력하세요
                  </div>
                </div>
              )}

              {/* Turn indicator */}
              {room.currentTurnUserId && (
                <div style={{
                  ...styles.turnIndicator,
                  background: room.currentTurnUserId === user?.id ? '#4CAF50' : '#FF9800'
                }}>
                  {room.currentTurnUserId === user?.id
                    ? '🎯 당신의 차례입니다!' 
                    : `⏳ ${room.users.find(u => u.userId === room.currentTurnUserId)?.name || 'Unknown'}의 차례를 기다리는 중...`}
                </div>
              )}

              <div style={styles.inputArea}>
                <input
                  type="text"
                  value={wordInput}
                  onChange={(e) => setWordInput(e.target.value)}
                  onKeyPress={(e) => e.key === 'Enter' && handleSubmitWord()}
                  placeholder={room.currentTurnUserId === user?.id
                    ? "일본어 단어를 입력하세요 (히라가나/카타카나)"
                    : "다른 플레이어의 차례입니다..."}
                  style={styles.wordInput}
                  disabled={room.status !== 'PLAYING' || room.currentTurnUserId !== user?.id}
                />
                <button
                  onClick={handleSubmitWord}
                  style={styles.submitButton}
                  disabled={!wordInput.trim() || room.status !== 'PLAYING' || room.currentTurnUserId !== user?.id}
                >
                  제출
                </button>
              </div>

              {currentPrompt && (
                <div style={styles.promptInfo}>
                  <div style={styles.promptLabel}>참고</div>
                  <div style={styles.promptText}>{currentPrompt.hint || '힌트 없음'}</div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Right: Users & Chat */}
        <div style={styles.rightPanel}>
          {/* Users */}
          <div style={styles.usersSection}>
            <h3 style={styles.sectionTitle}>👥 플레이어 ({room.users.length}/{room.maxUsers})</h3>
            <div style={styles.usersList}>
              {sortedusers.map((gameUser) => (
                <div key={gameUser.userId} style={styles.userCard}>
                  <div style={styles.userInfo}>
                    <span style={styles.userName}>
                      {gameUser.userId === room.hostUserId && '👑 '}
                      {gameUser.name}
                    </span>
                  </div>
                  <span style={styles.userScore}>{gameUser.score}점</span>
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
                  style={
                    msg.type === 'system' ? styles.systemMessage :
                    msg.type === 'correct' ? styles.correctMessage :
                    msg.type === 'wrong' ? styles.wrongMessage :
                    styles.chatMessage
                  }
                >
                  {msg.type === 'system' || msg.type === 'correct' || msg.type === 'wrong' ? (
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

      {/* Round End Popup */}
      {roundEndInfo && (
        <div style={styles.modalOverlay}>
          <div style={styles.roundEndModal}>
            <div style={styles.roundEndIcon}>🏁</div>
            <h2 style={styles.roundEndTitle}>
              {roundEndInfo.round}라운드 종료
            </h2>
            <p style={styles.roundEndReason}>
              {roundEndInfo.reason}
            </p>
            <div style={styles.roundEndCountdown}>
              <div style={styles.countdownCircle}>
                {roundEndInfo.countdown}
              </div>
              <p style={styles.countdownText}>
                {roundEndInfo.countdown > 0 
                  ? `다음 라운드 시작까지 ${roundEndInfo.countdown}초`
                  : '다음 라운드 시작!'}
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Game End Popup */}
      {gameEndInfo && (
        <div style={styles.modalOverlay}>
          <div style={styles.gameEndModal}>
            <div style={styles.gameEndIcon}>🎉</div>
            <h2 style={styles.gameEndTitle}>
              게임 종료!
            </h2>
            <p style={styles.gameEndMessage}>
              {gameEndInfo.message}
            </p>
            <div style={styles.finalScores}>
              <h3 style={styles.scoresTitle}>최종 점수</h3>
              {gameEndInfo.users.map((gameUser, index) => (
                <div key={gameUser.name} style={{
                  ...styles.scoreRow,
                  background: index === 0 ? 'linear-gradient(135deg, #FFD700, #FFA500)' : '#f8f9fa'
                }}>
                  <div style={styles.scoreRank}>
                    {index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : `${index + 1}위`}
                  </div>
                  <div style={styles.scoreName}>{gameUser.name}</div>
                  <div style={styles.scorePoints}>{gameUser.score}점</div>
                </div>
              ))}
            </div>
            <button 
              onClick={() => {
                setGameEndInfo(null);
                loadRoom(); // 방 정보 새로고침
                if (room) {
                  setTimeLeft(room.roundTimeLimit || 30); // 타이머 리셋
                }
              }}
              style={styles.gameEndButton}
            >
              확인
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #FF6B6B 0%, #FF8E53 100%)',
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
  },
  infoBox: {
    flex: 1,
    background: 'white',
    padding: '15px',
    borderRadius: '12px',
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  },
  infoLabel: {
    fontSize: '14px',
    color: '#666',
    marginBottom: '5px',
  },
  infoValue: {
    fontSize: '24px',
    fontWeight: 'bold',
    color: '#FF6B6B',
  },
  mainContent: {
    display: 'grid',
    gridTemplateColumns: '2fr 1fr',
    gap: '20px',
    height: 'calc(100vh - 200px)',
  },
  leftPanel: {
    background: 'white',
    borderRadius: '12px',
    padding: '30px',
    boxShadow: '0 4px 6px rgba(0,0,0,0.1)',
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
    alignItems: 'center',
    justifyContent: 'center',
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
  rules: {
    background: '#f8f9fa',
    padding: '20px',
    borderRadius: '12px',
    marginBottom: '30px',
    width: '100%',
    maxWidth: '500px',
  },
  rulesTitle: {
    fontSize: '18px',
    color: '#333',
    marginBottom: '15px',
  },
  rulesList: {
    fontSize: '16px',
    color: '#666',
    lineHeight: '1.8',
    paddingLeft: '20px',
  },
  startButton: {
    padding: '15px 40px',
    fontSize: '18px',
    fontWeight: 'bold',
    background: '#FF6B6B',
    color: 'white',
    border: 'none',
    borderRadius: '12px',
    cursor: 'pointer',
  },
  finishedArea: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100%',
  },
  finishedTitle: {
    fontSize: '36px',
    color: '#333',
    marginBottom: '30px',
  },
  finalRanking: {
    width: '100%',
    maxWidth: '500px',
    marginBottom: '30px',
  },
  rankItem: {
    display: 'flex',
    alignItems: 'center',
    padding: '15px',
    background: '#f8f9fa',
    borderRadius: '12px',
    marginBottom: '10px',
  },
  rank: {
    fontSize: '24px',
    marginRight: '15px',
  },
  rankName: {
    flex: 1,
    fontSize: '18px',
    color: '#333',
  },
  rankScore: {
    fontSize: '20px',
    fontWeight: 'bold',
    color: '#FF6B6B',
  },
  backToLobbyButton: {
    padding: '15px 40px',
    fontSize: '18px',
    fontWeight: 'bold',
    background: '#FF6B6B',
    color: 'white',
    border: 'none',
    borderRadius: '12px',
    cursor: 'pointer',
  },
  gameArea: {
    display: 'flex',
    flexDirection: 'column',
    gap: '20px',
    flex: 1,
  },
  wordHistory: {
    background: '#e7f3ff',
    padding: '15px',
    borderRadius: '12px',
    border: '2px solid #90caf9',
  },
  wordHistoryTitle: {
    fontSize: '16px',
    fontWeight: 'bold',
    color: '#1976d2',
    marginBottom: '10px',
  },
  wordHistoryList: {
    display: 'flex',
    flexWrap: 'wrap',
    gap: '8px',
  },
  wordHistoryItem: {
    background: 'white',
    padding: '6px 12px',
    borderRadius: '16px',
    fontSize: '14px',
    color: '#1976d2',
    border: '1px solid #90caf9',
  },
  lastWordDisplay: {
    background: '#fff3cd',
    padding: '20px',
    borderRadius: '12px',
    textAlign: 'center',
  },
  lastWordLabel: {
    fontSize: '14px',
    color: '#856404',
    marginBottom: '10px',
  },
  lastWordText: {
    fontSize: '48px',
    fontWeight: 'bold',
    color: '#FF6B6B',
    marginBottom: '10px',
  },
  lastWordHint: {
    fontSize: '16px',
    color: '#856404',
  },
  turnIndicator: {
    padding: '15px',
    borderRadius: '12px',
    textAlign: 'center',
    fontSize: '18px',
    fontWeight: 'bold',
    color: 'white',
  },
  inputArea: {
    display: 'flex',
    gap: '10px',
  },
  wordInput: {
    flex: 1,
    padding: '15px',
    fontSize: '18px',
    border: '2px solid #ddd',
    borderRadius: '8px',
  },
  submitButton: {
    padding: '15px 30px',
    fontSize: '18px',
    fontWeight: 'bold',
    background: '#FF6B6B',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
  },
  promptInfo: {
    background: '#e7f3ff',
    padding: '15px',
    borderRadius: '8px',
  },
  promptLabel: {
    fontSize: '14px',
    color: '#0c5460',
    marginBottom: '5px',
  },
  promptText: {
    fontSize: '16px',
    color: '#004085',
  },
  usersSection: {
    background: 'white',
    borderRadius: '12px',
    padding: '20px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
  },
  sectionTitle: {
    fontSize: '18px',
    color: '#333',
    marginBottom: '15px',
  },
  usersList: {
    display: 'flex',
    flexDirection: 'column',
    gap: '10px',
  },
  userCard: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '12px',
    background: '#f8f9fa',
    borderRadius: '8px',
  },
  userInfo: {
    display: 'flex',
    alignItems: 'center',
  },
  userName: {
    fontSize: '16px',
    color: '#333',
  },
  userScore: {
    fontSize: '16px',
    fontWeight: 'bold',
    color: '#FF6B6B',
  },
  chatSection: {
    background: 'white',
    borderRadius: '12px',
    padding: '20px',
    boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
  },
  chatMessages: {
    flex: 1,
    overflowY: 'auto',
    display: 'flex',
    flexDirection: 'column',
    gap: '10px',
  },
  chatMessage: {
    padding: '10px',
    background: '#f8f9fa',
    borderRadius: '8px',
  },
  systemMessage: {
    padding: '10px',
    background: '#d1ecf1',
    borderRadius: '8px',
    textAlign: 'center',
  },
  correctMessage: {
    padding: '10px',
    background: '#d4edda',
    borderRadius: '8px',
    textAlign: 'center',
  },
  wrongMessage: {
    padding: '10px',
    background: '#f8d7da',
    borderRadius: '8px',
    textAlign: 'center',
  },
  modalOverlay: {
    position: 'fixed',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    background: 'rgba(0, 0, 0, 0.7)',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 1000,
  },
  roundEndModal: {
    background: 'white',
    borderRadius: '20px',
    padding: '40px',
    maxWidth: '500px',
    width: '90%',
    textAlign: 'center',
    boxShadow: '0 10px 40px rgba(0,0,0,0.3)',
    animation: 'slideIn 0.3s ease-out',
  },
  roundEndIcon: {
    fontSize: '80px',
    marginBottom: '20px',
  },
  roundEndTitle: {
    fontSize: '32px',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '15px',
  },
  roundEndReason: {
    fontSize: '18px',
    color: '#666',
    marginBottom: '30px',
    lineHeight: '1.6',
  },
  roundEndCountdown: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: '15px',
  },
  countdownCircle: {
    width: '80px',
    height: '80px',
    borderRadius: '50%',
    background: 'linear-gradient(135deg, #FF6B6B, #FF8E53)',
    color: 'white',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    fontSize: '40px',
    fontWeight: 'bold',
    boxShadow: '0 4px 15px rgba(255, 107, 107, 0.4)',
  },
  countdownText: {
    fontSize: '16px',
    color: '#666',
    margin: 0,
  },
  gameEndModal: {
    background: 'white',
    borderRadius: '20px',
    padding: '40px',
    maxWidth: '600px',
    width: '90%',
    textAlign: 'center',
    boxShadow: '0 10px 40px rgba(0,0,0,0.3)',
    animation: 'slideIn 0.3s ease-out',
  },
  gameEndIcon: {
    fontSize: '80px',
    marginBottom: '20px',
  },
  gameEndTitle: {
    fontSize: '36px',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '10px',
  },
  gameEndMessage: {
    fontSize: '18px',
    color: '#666',
    marginBottom: '30px',
  },
  finalScores: {
    marginBottom: '30px',
  },
  scoresTitle: {
    fontSize: '20px',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '15px',
  },
  scoreRow: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '15px 20px',
    borderRadius: '12px',
    marginBottom: '10px',
    transition: 'transform 0.2s',
  },
  scoreRank: {
    fontSize: '20px',
    fontWeight: 'bold',
    minWidth: '60px',
  },
  scoreName: {
    fontSize: '18px',
    fontWeight: 'bold',
    flex: 1,
    textAlign: 'left',
    marginLeft: '15px',
  },
  scorePoints: {
    fontSize: '20px',
    fontWeight: 'bold',
    color: '#FF6B6B',
  },
  gameEndButton: {
    padding: '15px 40px',
    background: 'linear-gradient(135deg, #FF6B6B, #FF8E53)',
    color: 'white',
    border: 'none',
    borderRadius: '12px',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
    boxShadow: '0 4px 15px rgba(255, 107, 107, 0.4)',
    transition: 'transform 0.2s',
  },
  chatUsername: {
    fontWeight: 'bold',
    color: '#FF6B6B',
    marginRight: '8px',
  },
  chatText: {
    color: '#333',
  },
  systemText: {
    color: '#333',
    fontSize: '14px',
  },
};

export default WordchainRoom;
