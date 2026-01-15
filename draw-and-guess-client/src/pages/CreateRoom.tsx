import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';
import { GameType } from '../types';
import { useLanguage } from '../i18n/LanguageContext';

const CreateRoom: React.FC = () => {
  const navigate = useNavigate();
  const [username] = useState(sessionStorage.getItem('username') || '');
  const [selectedGameType, setSelectedGameType] = useState<GameType>('WORDCHAIN');
  const [maxPlayers, setMaxPlayers] = useState(2);
  const [totalRounds, setTotalRounds] = useState(3);
  const [roundTimeLimit, setRoundTimeLimit] = useState(15);
  const [loading, setLoading] = useState(false);

  React.useEffect(() => {
    if (!username) {
      navigate('/');
    }
  }, [username, navigate]);

  const gameModes = [
    {
      type: 'WORDCHAIN' as GameType,
      title: '🔤 끝말잇기',
      description: '일본어 단어로 끝말잇기 게임',
      color: '#FF6B6B',
      icon: '🔤',
      rules: [
        '일본어 히라가나/카타카나 단어를 입력합니다',
        '이전 단어의 마지막 글자로 시작해야 합니다',
        'ん으로 끝나는 단어는 사용 불가',
        '정답 +100점',
      ],
    },
    {
      type: 'OX' as GameType,
      title: '⭕❌ OX 퀴즈',
      description: '참/거짓을 빠르게 판단하는 게임',
      color: '#4ECDC4',
      icon: '⭕❌',
      rules: [
        '문제가 나오면 O 또는 X를 선택합니다',
        '가장 빠르게 정답을 맞춘 사람이 점수를 얻습니다',
        '정답 +100점',
      ],
    },
    {
      type: 'QA' as GameType,
      title: '📚 일반 상식 퀴즈',
      description: '4개 중 정답을 고르는 객관식 게임',
      color: '#95E1D3',
      icon: '📚',
      rules: [
        '4개의 선택지 중 정답을 고릅니다',
        '가장 빠르게 정답을 맞춘 사람이 점수를 얻습니다',
        '정답 +100점',
      ],
    },
    {
      type: 'DRAWING' as GameType,
      title: '🎨 그림 맞추기',
      description: '그림을 그리고 정답을 맞추는 게임',
      color: '#F38181',
      icon: '🎨',
      rules: [
        '제시된 단어를 그림으로 표현합니다',
        '다른 플레이어가 정답을 맞춥니다',
        '정답을 맞춘 사람과 그린 사람 모두 점수 획득',
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
      const result = await apiService.createGameRoom({
        name: `${username}의 게임방`,
        gameType: selectedGameType,
        maxPlayers: maxPlayers,
        totalRounds: totalRounds,
        roundTimeLimit: roundTimeLimit,
        hostUsername: username,
        isPrivate: false
      });
      navigate(`/game/${result.id}`);
    } catch (error) {
      console.error('Failed to create room:', error);
      alert(`방 생성에 실패했습니다: ${error instanceof Error ? error.message : '알 수 없는 오류'}`);
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
              <div style={styles.playerSelector}>
                <button 
                  style={styles.playerButton}
                  onClick={() => setMaxPlayers(Math.max(2, maxPlayers - 1))}
                  disabled={maxPlayers <= 2}
                >
                  -
                </button>
                <span style={styles.infoText}>{maxPlayers}명</span>
                <button 
                  style={styles.playerButton}
                  onClick={() => setMaxPlayers(Math.min(4, maxPlayers + 1))}
                  disabled={maxPlayers >= 4}
                >
                  +
                </button>
              </div>
            </div>
            <div style={styles.infoItem}>
              <span style={styles.infoIcon}>�</span>
              <div style={styles.playerSelector}>
                <button 
                  style={styles.playerButton}
                  onClick={() => setTotalRounds(Math.max(1, totalRounds - 1))}
                  disabled={totalRounds <= 1}
                >
                  -
                </button>
                <span style={styles.infoText}>{totalRounds}라운드</span>
                <button 
                  style={styles.playerButton}
                  onClick={() => setTotalRounds(Math.min(10, totalRounds + 1))}
                  disabled={totalRounds >= 10}
                >
                  +
                </button>
              </div>
            </div>
            <div style={styles.infoItem}>
              <span style={styles.infoIcon}>⏱️</span>
              <div style={styles.playerSelector}>
                <button 
                  style={styles.playerButton}
                  onClick={() => setRoundTimeLimit(Math.max(10, roundTimeLimit - 5))}
                  disabled={roundTimeLimit <= 10}
                >
                  -
                </button>
                <span style={styles.infoText}>{roundTimeLimit}초</span>
                <button 
                  style={styles.playerButton}
                  onClick={() => setRoundTimeLimit(Math.min(60, roundTimeLimit + 5))}
                  disabled={roundTimeLimit >= 60}
                >
                  +
                </button>
              </div>
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
  playerSelector: {
    display: 'flex',
    alignItems: 'center',
    gap: '10px',
  },
  playerButton: {
    width: '30px',
    height: '30px',
    borderRadius: '50%',
    border: '2px solid #667eea',
    background: 'white',
    color: '#667eea',
    fontSize: '18px',
    fontWeight: 'bold',
    cursor: 'pointer',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
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
