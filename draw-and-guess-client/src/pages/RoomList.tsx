import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';
import { GameRoom } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const RoomList: React.FC = () => {
  const [rooms, setRooms] = useState<GameRoom[]>([]);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();
  const { t } = useLanguage();

  useEffect(() => {
    loadRooms();
  }, []);

  const loadRooms = async () => {
    try {
      const roomList = await apiService.getGameRooms();
      setRooms(roomList.filter((room) => room.is_active));
    } catch (error) {
      console.error('Failed to load rooms:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleJoinRoom = async (room: GameRoom) => {
    const username = sessionStorage.getItem('username');
    if (!username) {
      alert(t.home.enterUsername);
      navigate('/');
      return;
    }

    try {
      await apiService.joinGameRoom(room.uuid, username);
      navigate(`/game/${room.uuid}`);
    } catch (error) {
      console.error('Failed to join room:', error);
      alert('Failed to join room');
    }
  };

  const handleRefresh = () => {
    setLoading(true);
    loadRooms();
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
      <div style={styles.content}>
        <div style={styles.header}>
          <button onClick={() => navigate('/')} style={styles.backButton}>
            ← {t.roomList.backToHome}
          </button>
          <h1 style={styles.title}>{t.roomList.availableRooms}</h1>
          <button onClick={handleRefresh} style={styles.refreshButton}>
            🔄
          </button>
        </div>

        {rooms.length === 0 ? (
          <div style={styles.emptyMessage}>
            {t.roomList.noRooms}
          </div>
        ) : (
          <div style={styles.roomList}>
            {rooms.map((room) => (
              <div key={room.uuid} style={styles.roomCard}>
                <div style={styles.roomInfo}>
                  <div style={styles.roomHeader}>
                    <div style={styles.roomTitle}>
                      👑 호스트: {room.drawer_user}
                    </div>
                    <div style={{
                      ...styles.gameTypeBadge,
                      background: room.game_type === 'guess' ? '#FF6B6B' : room.game_type === 'ox' ? '#4ECDC4' : '#95E1D3'
                    }}>
                      {room.game_type === 'guess' ? '🎨 그림' : room.game_type === 'ox' ? '⭕❌ OX' : '📚 상식'}
                    </div>
                  </div>
                  <div style={styles.roomStats}>
                    <span>👥 {room.players.length}/4</span>
                    {room.game_status === 'playing' && (
                      <>
                        <span style={styles.separator}>•</span>
                        <span>🎮 {t.roomList.round} {room.round_number}/{room.max_rounds}</span>
                      </>
                    )}
                  </div>
                  <div style={styles.gameStatus}>
                    {room.game_status === 'waiting' && (
                      <span style={styles.statusWaiting}>{t.roomList.status.waiting}</span>
                    )}
                    {room.game_status === 'playing' && (
                      <span style={styles.statusPlaying}>{t.roomList.status.playing}</span>
                    )}
                    {room.game_status === 'finished' && (
                      <span style={styles.statusFinished}>{t.roomList.status.finished}</span>
                    )}
                  </div>
                </div>
                <button
                  onClick={() => handleJoinRoom(room)}
                  style={
                    room.game_status === 'playing' || room.game_status === 'finished'
                      ? { ...styles.joinButton, ...styles.joinButtonDisabled }
                      : styles.joinButton
                  }
                  disabled={room.game_status === 'playing' || room.game_status === 'finished'}
                >
                  {room.game_status === 'finished'
                    ? t.roomList.status.finished
                    : room.game_status === 'playing'
                    ? t.roomList.status.playing
                    : t.roomList.join}
                </button>
              </div>
            ))}
          </div>
        )}
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
  content: {
    maxWidth: '700px',
    margin: '0 auto',
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
};

export default RoomList;
