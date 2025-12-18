import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';
import { GameType } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const CreateRoom: React.FC = () => {
  const navigate = useNavigate();
  const { t } = useLanguage();
  const [username] = useState(sessionStorage.getItem('username') || '');
  const [selectedGameType, setSelectedGameType] = useState<GameType>('guess');
  const [loading, setLoading] = useState(false);

  React.useEffect(() => {
    if (!username) {
      navigate('/');
    }
  }, [username, navigate]);

  const gameModes = [
    {
      type: 'guess' as GameType,
      title: '🎨 그림 맞추기',
      description: '그림을 그리고 다른 사람이 맞추는 게임',
      color: '#FF6B6B',
      icon: '🎨',
      rules: [
        '한 명씩 돌아가며 그림을 그립니다',
        '다른 플레이어들이 채팅으로 정답을 맞춥니다',
        '정답자 +2점, 그린 사람 +1점',
      ],
    },
    {
      type: 'ox' as GameType,
      title: '⭕❌ OX 퀴즈',
      description: '참/거짓을 빠르게 판단하는 게임',
      color: '#4ECDC4',
      icon: '⭕❌',
      rules: [
        '문제가 나오면 O 또는 X를 선택합니다',
        '가장 빠르게 정답을 맞춘 사람이 점수를 얻습니다',
        '정답 +2점',
      ],
    },
    {
      type: 'general' as GameType,
      title: '📚 일반 상식 퀴즈',
      description: '4개 중 정답을 고르는 객관식 게임',
      color: '#95E1D3',
      icon: '📚',
      rules: [
        '4개의 선택지 중 정답을 고릅니다',
        '가장 빠르게 정답을 맞춘 사람이 점수를 얻습니다',
        '정답 +2점',
      ],
    },
  ];

  const handleCreateRoom = async () => {
    if (!username) {
      alert('사용자 이름이 없습니다.');
      navigate('/');
      return;
    }

    setLoading(true);
    try {
      const result = await apiService.createGameRoom(username, selectedGameType);
      navigate(`/game/${result.room_id}`);
    } catch (error) {
      console.error('Failed to create room:', error);
      alert('방 생성에 실패했습니다.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.content}>
        <button onClick={() => navigate('/game-mode')} style={styles.backButton}>
          ← 뒤로
        </button>

        <h1 style={styles.title}>게임 방 만들기</h1>
        <p style={styles.subtitle}>
          플레이할 게임을 선택하세요 (최대 4명)
        </p>

        <div style={styles.modesGrid}>
          {gameModes.map((mode) => (
            <div
              key={mode.type}
              style={{
                ...styles.modeCard,
                borderColor: selectedGameType === mode.type ? mode.color : '#ddd',
                borderWidth: selectedGameType === mode.type ? '4px' : '2px',
                transform: selectedGameType === mode.type ? 'scale(1.02)' : 'scale(1)',
              }}
              onClick={() => setSelectedGameType(mode.type)}
            >
              <div style={{ ...styles.modeIcon, background: mode.color }}>
                {mode.icon}
              </div>
              <h3 style={styles.modeTitle}>{mode.title}</h3>
              <p style={styles.modeDescription}>{mode.description}</p>
              
              <div style={styles.rulesBox}>
                <div style={styles.rulesTitle}>게임 규칙</div>
                <ul style={styles.rulesList}>
                  {mode.rules.map((rule, index) => (
                    <li key={index} style={styles.ruleItem}>{rule}</li>
                  ))}
                </ul>
              </div>

              {selectedGameType === mode.type && (
                <div style={styles.selectedBadge}>선택됨 ✓</div>
              )}
            </div>
          ))}
        </div>

        <div style={styles.footer}>
          <div style={styles.infoBox}>
            <div style={styles.infoItem}>
              <span style={styles.infoIcon}>👥</span>
              <span style={styles.infoText}>최대 4명</span>
            </div>
            <div style={styles.infoItem}>
              <span style={styles.infoIcon}>🎯</span>
              <span style={styles.infoText}>3점 승리</span>
            </div>
            <div style={styles.infoItem}>
              <span style={styles.infoIcon}>🎮</span>
              <span style={styles.infoText}>3라운드</span>
            </div>
          </div>

          <button
            style={styles.createButton}
            onClick={handleCreateRoom}
            disabled={loading}
          >
            {loading ? '생성 중...' : '방 만들기'}
          </button>
        </div>
      </div>
    </div>
  );
};

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    padding: '20px',
  },
  content: {
    maxWidth: '1200px',
    width: '100%',
  },
  backButton: {
    padding: '10px 20px',
    fontSize: '16px',
    background: 'rgba(255, 255, 255, 0.2)',
    border: 'none',
    borderRadius: '8px',
    color: 'white',
    cursor: 'pointer',
    marginBottom: '20px',
    backdropFilter: 'blur(10px)',
  },
  title: {
    fontSize: '48px',
    fontWeight: 'bold',
    color: 'white',
    marginBottom: '10px',
    textAlign: 'center',
  },
  subtitle: {
    fontSize: '20px',
    color: 'rgba(255, 255, 255, 0.9)',
    marginBottom: '50px',
    textAlign: 'center',
  },
  modesGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
    gap: '30px',
    marginBottom: '40px',
  },
  modeCard: {
    background: 'white',
    borderRadius: '20px',
    padding: '30px',
    cursor: 'pointer',
    transition: 'all 0.3s ease',
    border: '2px solid',
    position: 'relative',
    boxShadow: '0 10px 30px rgba(0, 0, 0, 0.1)',
  },
  modeIcon: {
    width: '80px',
    height: '80px',
    borderRadius: '50%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: '40px',
    margin: '0 auto 20px',
  },
  modeTitle: {
    fontSize: '24px',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '10px',
    textAlign: 'center',
  },
  modeDescription: {
    fontSize: '16px',
    color: '#666',
    marginBottom: '20px',
    textAlign: 'center',
    lineHeight: '1.5',
  },
  rulesBox: {
    background: '#f8f9fa',
    borderRadius: '10px',
    padding: '15px',
    marginTop: '15px',
  },
  rulesTitle: {
    fontSize: '14px',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '10px',
  },
  rulesList: {
    margin: 0,
    paddingLeft: '20px',
    fontSize: '14px',
    color: '#666',
  },
  ruleItem: {
    marginBottom: '5px',
  },
  selectedBadge: {
    position: 'absolute',
    top: '20px',
    right: '20px',
    background: '#4CAF50',
    color: 'white',
    padding: '5px 15px',
    borderRadius: '20px',
    fontSize: '14px',
    fontWeight: 'bold',
  },
  footer: {
    background: 'white',
    borderRadius: '20px',
    padding: '30px',
    boxShadow: '0 10px 30px rgba(0, 0, 0, 0.1)',
  },
  infoBox: {
    display: 'flex',
    justifyContent: 'space-around',
    marginBottom: '25px',
    flexWrap: 'wrap',
    gap: '15px',
  },
  infoItem: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
  },
  infoIcon: {
    fontSize: '24px',
  },
  infoText: {
    fontSize: '16px',
    fontWeight: 'bold',
    color: '#333',
  },
  createButton: {
    width: '100%',
    padding: '18px',
    fontSize: '20px',
    fontWeight: 'bold',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    color: 'white',
    border: 'none',
    borderRadius: '12px',
    cursor: 'pointer',
    transition: 'transform 0.2s',
  },
};

export default CreateRoom;
