import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';
import { wsService } from '../services/websocket';
import { GameRoom, GameType } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const RoomList: React.FC = () => {
  const [rooms, setRooms] = useState<GameRoom[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [selectedGameType, setSelectedGameType] = useState<GameType>('WORDCHAIN');
  const [maxUsers, setMaxUsers] = useState(4);
  const [totalRounds, setTotalRounds] = useState(5);
  const [roundTimeLimit, setRoundTimeLimit] = useState(30);
  const [creatingRoom, setCreatingRoom] = useState(false);
  const [lobbyChatMessages, setLobbyChatMessages] = useState<Array<{username: string, message: string}>>([]);
  const [chatInput, setChatInput] = useState('');
  const [user, setUser] = useState<{id: string; name: string} | null>(null);
  const chatEndRef = React.useRef<HTMLDivElement>(null);
  const navigate = useNavigate();
  const { t } = useLanguage();
  const username = sessionStorage.getItem('username') || '';

  useEffect(() => {
    if (!username) {
      navigate('/');
      return;
    }
    
    // Load user data from sessionStorage
    const loadUser = () => {
      try {
        const userDataStr = sessionStorage.getItem('user');
        if (userDataStr) {
          const userData = JSON.parse(userDataStr);
          setUser({ id: userData.id, name: userData.name });
          console.log('[RoomList] ✅ User loaded from session:', userData.name);
        } else {
          console.log('[RoomList] ❌ No user in session');
          alert('사용자 정보를 찾을 수 없습니다. 다시 로그인해주세요.');
          navigate('/');
        }
      } catch (error) {
        console.error('[RoomList] Failed to load user from session:', error);
        navigate('/');
      }
    };
    loadUser();
    
    loadRooms();

    // WebSocket 연결 및 lobby 채널 구독
    wsService.connect(
      () => {
        console.log('WebSocket connected for room list');
        // 연결되면 로비 채팅 히스토리 요청
        wsService.requestLobbyChatHistory();
      },
      (error) => console.error('WebSocket connection error:', error)
    );

    // lobby 채널 구독하여 방 목록 업데이트 수신
    const subscription = wsService.subscribe('lobby', (data) => {
      console.log('[RoomList] Lobby update received:', data);
      
      // 서버에서 보낸 lobby_update 메시지 처리
      if (data.type === 'lobby_update' || data.type === 'room_list_update') {
        if (data.data && Array.isArray(data.data)) {
          console.log('[RoomList] Updating rooms from WebSocket:', data.data.length, 'rooms');
          console.log('[RoomList] Updated room details:', data.data.map(r => ({ id: r.id, users: r.users?.length || 0, status: r.status })));
          setRooms(data.data);
        } else {
          // 데이터가 없으면 API에서 다시 로드
          console.warn('[RoomList] No data in lobby update, reloading from API');
          loadRooms();
        }
      }
      
      // 로비 채팅 히스토리 수신
      if (data.type === 'LOBBY_CHAT_HISTORY' && data.messages) {
        console.log('[RoomList] Received chat history:', data.messages.length, 'messages');
        setLobbyChatMessages(data.messages.map((msg: any) => ({
          username: msg.userName || msg.username || 'Unknown',
          message: msg.message
        })));
      }
      
      // 로비 채팅 메시지 처리
      if (data.type === 'LOBBY_CHAT' && data.payload) {
        const { username, message } = data.payload;
        console.log('[RoomList] Received lobby chat:', username, message);
        setLobbyChatMessages(prev => [...prev, { username: username || 'Unknown', message }]);
      }
    });

    return () => {
      subscription.unsubscribe();
    };
  }, []);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [lobbyChatMessages]);

  const loadRooms = async () => {
    try {
      const roomList = await apiService.getGameRooms();
      console.log('[RoomList] Loaded rooms:', roomList);
      setRooms(roomList);
    } catch (error) {
      console.error('Failed to load rooms:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleJoinRoom = async (room: GameRoom) => {
    if (!user) {
      alert(t.home.enterUsername);
      navigate('/');
      return;
    }

    try {
      await apiService.joinGameRoom(room.id, user.id);
      navigate(`/game/${room.id}`);
    } catch (error) {
      console.error('Failed to join room:', error);
      alert('Failed to join room');
    }
  };

  const handleRefresh = () => {
    setLoading(true);
    loadRooms();
  };

  const handleCreateRoom = async () => {
    if (!user) {
      alert('사용자 정보가 없습니다.');
      navigate('/');
      return;
    }

    setCreatingRoom(true);
    try {
      const result = await apiService.createGameRoom({
        name: `${user.name}의 게임방`,
        gameType: selectedGameType,
        maxUsers: maxUsers,
        totalRounds: totalRounds,
        roundTimeLimit: roundTimeLimit,
        hostUserId: user.id,
        isPrivate: false
      });
      setShowCreateModal(false);
      navigate(`/game/${result.id}`);
    } catch (error) {
      console.error('Failed to create room:', error);
      alert(`방 생성에 실패했습니다: ${error instanceof Error ? error.message : '알 수 없는 오류'}`);
    } finally {
      setCreatingRoom(false);
    }
  };

  const gameModes = [
    {
      type: 'WORDCHAIN' as GameType,
      title: '🔤 끝말잇기',
      description: '일본어 단어로 끝말잇기',
      color: '#FF6B6B',
    },
    {
      type: 'OX' as GameType,
      title: '⭕❌ OX 퀴즈',
      description: '참/거짓 판단 게임',
      color: '#4ECDC4',
    },
    {
      type: 'QA' as GameType,
      title: '📚 일반 상식',
      description: '4지선다 퀴즈',
      color: '#95E1D3',
    },
    {
      type: 'DRAWING' as GameType,
      title: '🎨 그림 맞추기',
      description: '그림으로 표현하기',
      color: '#F38181',
    },
  ];

  const handleSendLobbyChat = () => {
    if (!chatInput.trim() || !user) return;

    // WebSocket을 통해 직접 lobby_chat 메시지 전송
    if (wsService['ws'] && wsService['ws'].readyState === WebSocket.OPEN) {
      wsService['ws'].send(JSON.stringify({
        type: 'lobby_chat',
        channel: 'lobby',
        data: {
          userId: user.id,
          username: user.name,
          message: chatInput.trim()
        }
      }));
      console.log('[RoomList] Sent lobby chat from', user.name, ':', chatInput.trim());
    }

    setChatInput('');
  };

  if (loading) {
    return (
      <div style={styles.container}>
        <div style={styles.loadingText}>{t.gameRoom.loading}</div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <div style={styles.mainLayout}>
        {/* Left: Room List */}
        <div style={styles.content}>
          <div style={styles.header}>
            <button onClick={() => navigate('/')} style={styles.backButton}>
              ← {t.roomList.backToHome}
            </button>
            <h1 style={styles.title}>{t.roomList.availableRooms}</h1>
            <button onClick={() => setShowCreateModal(true)} style={styles.createButton}>
              ➕ 방 만들기
            </button>
          </div>

          {rooms.length === 0 ? (
            <div style={styles.emptyMessage}>
              {t.roomList.noRooms}
            </div>
          ) : (
            <div style={styles.roomList}>
              {rooms.map((room) => (
                <div key={room.id} style={styles.roomCard}>
                  <div style={styles.roomInfo}>
                    <div style={styles.roomHeader}>
                      <div style={styles.roomTitle}>
                        👑 호스트 ID: {room.hostUserId.substring(0, 8)}...
                      </div>
                      <div style={{
                        ...styles.gameTypeBadge,
                        background: room.gameType === 'WORDCHAIN' ? '#FF6B6B' : room.gameType === 'OX' ? '#4ECDC4' : room.gameType === 'QA' ? '#95E1D3' : '#F38181'
                      }}>
                        {room.gameType === 'WORDCHAIN' ? '🔤 끝말잇기' : room.gameType === 'OX' ? '⭕❌ OX' : room.gameType === 'QA' ? '📚 상식' : '🎨 그림'}
                      </div>
                    </div>
                    <div style={styles.roomStats}>
                      <span>👥 {room.users.length}/{room.maxUsers}</span>
                      {room.status === 'PLAYING' && (
                        <>
                          <span style={styles.separator}>•</span>
                          <span>🎮 {t.roomList.round} {room.currentRound}/{room.totalRounds}</span>
                        </>
                      )}
                    </div>
                    <div style={styles.gameStatus}>
                      {room.status === 'WAITING' && (
                        <span style={styles.statusWaiting}>{t.roomList.status.waiting}</span>
                      )}
                      {room.status === 'PLAYING' && (
                        <span style={styles.statusPlaying}>{t.roomList.status.playing}</span>
                      )}
                      {room.status === 'FINISHED' && (
                        <span style={styles.statusFinished}>{t.roomList.status.finished}</span>
                      )}
                    </div>
                  </div>
                  <button
                    onClick={() => handleJoinRoom(room)}
                    style={
                      room.status === 'PLAYING' || room.status === 'FINISHED'
                        ? { ...styles.joinButton, ...styles.joinButtonDisabled }
                        : styles.joinButton
                    }
                    disabled={room.status === 'PLAYING' || room.status === 'FINISHED'}
                  >
                    {room.status === 'FINISHED'
                      ? t.roomList.status.finished
                      : room.status === 'PLAYING'
                      ? t.roomList.status.playing
                      : t.roomList.join}
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Right: Lobby Chat */}
        <div style={styles.lobbyChatContainer}>
          <h3 style={styles.lobbyChatTitle}>💬 로비 채팅</h3>
          <div style={styles.lobbyChatMessages}>
            {lobbyChatMessages.length === 0 ? (
              <div style={styles.noChatMessage}>채팅이 없습니다</div>
            ) : (
              lobbyChatMessages.map((msg, index) => (
                <div key={index} style={styles.chatMessageItem}>
                  <span style={styles.chatUserName}>{msg.username}:</span>
                  <span style={styles.chatMessageText}>{msg.message}</span>
                </div>
              ))
            )}
            <div ref={chatEndRef} />
          </div>
          <div style={styles.lobbyChatInputArea}>
            <input
              type="text"
              value={chatInput}
              onChange={(e) => setChatInput(e.target.value)}
              onKeyPress={(e) => e.key === 'Enter' && handleSendLobbyChat()}
              placeholder="메시지를 입력하세요..."
              style={styles.lobbyChatInput}
            />
            <button onClick={handleSendLobbyChat} style={styles.lobbyChatSendButton}>
              전송
            </button>
          </div>
        </div>
      </div>

      {/* Create Room Modal */}
      {showCreateModal && (
        <div style={styles.modalOverlay} onClick={() => setShowCreateModal(false)}>
          <div style={styles.modalContent} onClick={(e) => e.stopPropagation()}>
            <div style={styles.modalHeader}>
              <h2 style={styles.modalTitle}>🎮 게임 방 만들기</h2>
              <button style={styles.modalCloseButton} onClick={() => setShowCreateModal(false)}>
                ✕
              </button>
            </div>
            
            <p style={styles.modalSubtitle}>플레이할 게임을 선택하세요</p>
            
            <div style={styles.gameModeGrid}>
              {gameModes.map((mode) => (
                <div
                  key={mode.type}
                  style={{
                    ...styles.gameModeCard,
                    borderColor: selectedGameType === mode.type ? mode.color : '#ddd',
                    background: selectedGameType === mode.type ? `${mode.color}15` : 'white',
                  }}
                  onClick={() => setSelectedGameType(mode.type)}
                >
                  <div style={styles.gameModeTitle}>{mode.title}</div>
                  <div style={styles.gameModeDescription}>{mode.description}</div>
                  {selectedGameType === mode.type && (
                    <div style={styles.selectedCheck}>✓</div>
                  )}
                </div>
              ))}
            </div>

            <div style={styles.settingsContainer}>
              <div style={styles.settingRow}>
                <label style={styles.settingLabel}>👥 최대 인원</label>
                <select 
                  style={styles.settingSelect}
                  value={maxUsers}
                  onChange={(e) => setMaxUsers(Number(e.target.value))}
                >
                  {[2, 3, 4, 5, 6, 7, 8].map(num => (
                    <option key={num} value={num}>{num}명</option>
                  ))}
                </select>
              </div>

              <div style={styles.settingRow}>
                <label style={styles.settingLabel}>🎯 라운드 수</label>
                <select 
                  style={styles.settingSelect}
                  value={totalRounds}
                  onChange={(e) => setTotalRounds(Number(e.target.value))}
                >
                  {[2, 3, 4, 5, 6, 7, 8, 9, 10].map(num => (
                    <option key={num} value={num}>{num}라운드</option>
                  ))}
                </select>
              </div>

              <div style={styles.settingRow}>
                <label style={styles.settingLabel}>⏱️ 문제당 시간</label>
                <select 
                  style={styles.settingSelect}
                  value={roundTimeLimit}
                  onChange={(e) => setRoundTimeLimit(Number(e.target.value))}
                >
                  {[5, 10, 15, 20, 25, 30].map(sec => (
                    <option key={sec} value={sec}>{sec}초</option>
                  ))}
                </select>
              </div>
            </div>

            <div style={styles.modalFooter}>
              <button style={styles.cancelButton} onClick={() => setShowCreateModal(false)}>
                취소
              </button>
              <button 
                style={styles.confirmButton} 
                onClick={handleCreateRoom}
                disabled={creatingRoom}
              >
                {creatingRoom ? '생성 중...' : '방 만들기'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    padding: '10px',
  },
  mainLayout: {
    display: 'grid',
    gridTemplateColumns: '1fr 350px',
    gap: '10px',
    maxWidth: '1400px',
    margin: '0 auto',
  },
  content: {
    background: 'white',
    borderRadius: '12px',
    padding: 'clamp(15px, 4vw, 20px)',
    boxShadow: '0 10px 25px rgba(0,0,0,0.2)',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '15px',
    gap: '8px',
    flexWrap: 'wrap',
  },
  title: {
    margin: 0,
    color: '#333',
    fontSize: 'clamp(18px, 4.5vw, 24px)',
    flex: '1 0 100%',
    textAlign: 'center',
    order: -1,
  },
  backButton: {
    padding: 'clamp(6px, 1.5vw, 8px) clamp(12px, 3vw, 16px)',
    background: '#718096',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
  },
  createButton: {
    padding: 'clamp(6px, 1.5vw, 8px) clamp(12px, 3vw, 16px)',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
    fontWeight: 'bold',
  },
  refreshButton: {
    padding: 'clamp(6px, 1.5vw, 8px) clamp(12px, 3vw, 16px)',
    background: '#48bb78',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
  },
  loadingText: {
    textAlign: 'center',
    color: 'white',
    fontSize: 'clamp(16px, 4vw, 20px)',
  },
  emptyMessage: {
    textAlign: 'center',
    padding: 'clamp(30px, 8vw, 40px)',
    color: '#666',
    fontSize: 'clamp(14px, 3vw, 16px)',
  },
  roomList: {
    display: 'flex',
    flexDirection: 'column',
    gap: '10px',
  },
  roomCard: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 'clamp(12px, 3vw, 16px)',
    background: '#f7fafc',
    borderRadius: '8px',
    border: '1px solid #e2e8f0',
    gap: '10px',
    flexWrap: 'wrap',
  },
  roomInfo: {
    flex: '1 1 auto',
    minWidth: '200px',
  },
  roomHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '4px',
    gap: '10px',
  },
  roomTitle: {
    fontSize: 'clamp(14px, 3vw, 16px)',
    fontWeight: 'bold',
    color: '#333',
  },
  gameTypeBadge: {
    fontSize: 'clamp(10px, 2vw, 12px)',
    padding: '4px 10px',
    borderRadius: '12px',
    color: 'white',
    fontWeight: 'bold',
    whiteSpace: 'nowrap',
  },
  roomStats: {
    fontSize: 'clamp(12px, 2.5vw, 14px)',
    color: '#666',
    display: 'flex',
    alignItems: 'center',
    gap: '6px',
    flexWrap: 'wrap',
  },
  separator: {
    color: '#cbd5e0',
  },
  gameStatus: {
    marginTop: '6px',
  },
  statusWaiting: {
    fontSize: 'clamp(11px, 2vw, 12px)',
    padding: '3px 6px',
    background: '#bee3f8',
    color: '#2c5282',
    borderRadius: '4px',
    fontWeight: 'bold',
    display: 'inline-block',
  },
  statusPlaying: {
    fontSize: 'clamp(11px, 2vw, 12px)',
    padding: '3px 6px',
    background: '#c6f6d5',
    color: '#22543d',
    borderRadius: '4px',
    fontWeight: 'bold',
    display: 'inline-block',
  },
  statusFinished: {
    fontSize: 'clamp(11px, 2vw, 12px)',
    padding: '3px 6px',
    background: '#fed7d7',
    color: '#742a2a',
    borderRadius: '4px',
    fontWeight: 'bold',
    display: 'inline-block',
  },
  joinButton: {
    padding: 'clamp(8px, 2vw, 10px) clamp(16px, 4vw, 20px)',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '6px',
    cursor: 'pointer',
    fontWeight: 'bold',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
    whiteSpace: 'nowrap',
  },
  joinButtonDisabled: {
    background: '#cbd5e0',
    cursor: 'not-allowed',
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
  modalContent: {
    background: 'white',
    borderRadius: '16px',
    padding: '24px',
    maxWidth: '500px',
    width: '90%',
    maxHeight: '80vh',
    overflowY: 'auto',
    boxShadow: '0 20px 60px rgba(0,0,0,0.3)',
  },
  modalHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '16px',
  },
  modalTitle: {
    margin: 0,
    fontSize: '24px',
    color: '#333',
  },
  modalCloseButton: {
    background: 'none',
    border: 'none',
    fontSize: '24px',
    cursor: 'pointer',
    color: '#999',
    padding: '0',
    width: '32px',
    height: '32px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  },
  modalSubtitle: {
    margin: '0 0 20px 0',
    fontSize: '14px',
    color: '#666',
  },
  gameModeGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, 1fr)',
    gap: '12px',
    marginBottom: '24px',
  },
  gameModeCard: {
    padding: '16px',
    borderRadius: '12px',
    border: '3px solid',
    cursor: 'pointer',
    transition: 'all 0.2s',
    position: 'relative',
  },
  gameModeTitle: {
    fontSize: '18px',
    fontWeight: 'bold',
    marginBottom: '4px',
    color: '#333',
  },
  gameModeDescription: {
    fontSize: '12px',
    color: '#666',
  },
  selectedCheck: {
    position: 'absolute',
    top: '8px',
    right: '8px',
    background: '#48bb78',
    color: 'white',
    borderRadius: '50%',
    width: '24px',
    height: '24px',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '14px',
    fontWeight: 'bold',
  },
  modalFooter: {
    display: 'flex',
    gap: '12px',
    justifyContent: 'flex-end',
  },
  cancelButton: {
    padding: '12px 24px',
    background: '#e2e8f0',
    color: '#333',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontSize: '14px',
    fontWeight: 'bold',
  },
  confirmButton: {
    padding: '12px 24px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontSize: '14px',
    fontWeight: 'bold',
  },
  settingsContainer: {
    marginBottom: '20px',
    display: 'flex',
    flexDirection: 'column',
    gap: '12px',
  },
  settingRow: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '12px',
    background: '#f7fafc',
    borderRadius: '8px',
  },
  settingLabel: {
    fontSize: '14px',
    fontWeight: 'bold',
    color: '#333',
  },
  settingSelect: {
    padding: '8px 12px',
    fontSize: '14px',
    border: '2px solid #e2e8f0',
    borderRadius: '6px',
    background: 'white',
    cursor: 'pointer',
    minWidth: '100px',
  },
  lobbyChatContainer: {
    background: 'white',
    borderRadius: '12px',
    padding: '15px',
    boxShadow: '0 10px 25px rgba(0,0,0,0.2)',
    display: 'flex',
    flexDirection: 'column',
    height: 'calc(100vh - 20px)',
  },
  lobbyChatTitle: {
    margin: '0 0 15px 0',
    fontSize: '18px',
    color: '#333',
    borderBottom: '2px solid #667eea',
    paddingBottom: '10px',
  },
  lobbyChatMessages: {
    flex: 1,
    overflowY: 'auto',
    display: 'flex',
    flexDirection: 'column',
    gap: '8px',
    marginBottom: '15px',
    padding: '10px',
    background: '#f7fafc',
    borderRadius: '8px',
  },
  noChatMessage: {
    textAlign: 'center',
    color: '#999',
    fontSize: '14px',
    padding: '20px',
  },
  chatMessageItem: {
    padding: '8px',
    background: 'white',
    borderRadius: '6px',
    fontSize: '14px',
  },
  chatUserName: {
    fontWeight: 'bold',
    color: '#667eea',
    marginRight: '8px',
  },
  chatMessageText: {
    color: '#333',
  },
  lobbyChatInputArea: {
    display: 'flex',
    gap: '8px',
  },
  lobbyChatInput: {
    flex: 1,
    padding: '10px',
    fontSize: '14px',
    border: '2px solid #e2e8f0',
    borderRadius: '8px',
  },
  lobbyChatSendButton: {
    padding: '10px 20px',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontWeight: 'bold',
    fontSize: '14px',
  },
};

export default RoomList;
