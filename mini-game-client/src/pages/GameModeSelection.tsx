import React from 'react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../i18n/LanguageContext';

const GameModeSelection: React.FC = () => {
  const navigate = useNavigate();
  useLanguage(); // Keep the hook for potential future use
  const username = sessionStorage.getItem('username');

  React.useEffect(() => {
    if (!username) {
      navigate('/');
    }
  }, [username, navigate]);

  const gameModes = [
    {
      id: 'create',
      title: '🎮 방 만들기',
      description: '게임 방을 만들어 친구들을 초대하세요',
      color: '#FF6B6B',
      route: '/create-room',
    },
    {
      id: 'join',
      title: '🚪 방 참가하기',
      description: '다른 사람이 만든 방에 참가하세요',
      color: '#4ECDC4',
      route: '/rooms',
    },
  ];

  return (
    <div style={styles.container}>
      <div style={styles.content}>
        <button onClick={() => navigate('/')} style={styles.backButton}>
          ← 홈으로
        </button>

        <h1 style={styles.title}>게임 모드 선택</h1>
        <p style={styles.subtitle}>플레이하고 싶은 게임을 선택하세요</p>

        <div style={styles.modeGrid}>
          {gameModes.map((mode) => (
            <div
              key={mode.id}
              style={{ ...styles.modeCard, borderColor: mode.color }}
              onClick={() => navigate(mode.route)}
            >
              <div style={{ ...styles.modeIcon, background: mode.color }}>
                {mode.title.split(' ')[0]}
              </div>
              <h3 style={styles.modeTitle}>{mode.title}</h3>
              <p style={styles.modeDescription}>{mode.description}</p>
              <button
                style={{ ...styles.playButton, background: mode.color }}
                onClick={(e) => {
                  e.stopPropagation();
                  navigate(mode.route);
                }}
              >
                플레이
              </button>
            </div>
          ))}
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
    maxWidth: '1000px',
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
  modeGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(280px, 1fr))',
    gap: '30px',
    marginTop: '40px',
  },
  modeCard: {
    background: 'white',
    borderRadius: '20px',
    padding: '40px 30px',
    textAlign: 'center',
    cursor: 'pointer',
    transition: 'all 0.3s ease',
    border: '4px solid',
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
  },
  modeDescription: {
    fontSize: '16px',
    color: '#666',
    marginBottom: '25px',
    lineHeight: '1.5',
  },
  playButton: {
    padding: '12px 40px',
    fontSize: '18px',
    fontWeight: 'bold',
    border: 'none',
    borderRadius: '10px',
    color: 'white',
    cursor: 'pointer',
    transition: 'transform 0.2s',
  },
};

export default GameModeSelection;
