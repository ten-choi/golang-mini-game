import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLanguage } from '../i18n/LanguageContext';

const Home: React.FC = () => {
  const [username, setUsername] = useState('');
  const { language, setLanguage, t } = useLanguage();
  const navigate = useNavigate();

  const languageOptions = [
    { code: 'ko', label: '한국어' },
    { code: 'en', label: 'English' },
    { code: 'ja', label: '日本語' },
  ];

  return (
    <div style={styles.container}>
      <div style={styles.content}>
        <h1 style={styles.title}>🎨 {t.home.title}</h1>
        
        <p style={styles.subtitle}>{t.home.subtitle}</p>

        <div style={styles.rulesBox}>
          <h3 style={styles.rulesTitle}>{t.home.rulesTitle}</h3>
          <ul style={styles.rulesList}>
            {t.home.rules.map((rule, index) => (
              <li key={index}>{rule}</li>
            ))}
          </ul>
        </div>

        <div style={styles.inputGroup}>
          <input
            type="text"
            placeholder={t.home.usernamePlaceholder}
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            style={styles.input}
          />
        </div>

        <div style={styles.languageGroup}>
          <span style={styles.languageLabel}>{t.home.selectLanguage}</span>
          <div style={styles.languageButtons}>
            {languageOptions.map((option) => (
              <button
                key={option.code}
                type="button"
                onClick={() => setLanguage(option.code)}
                style={{
                  ...styles.languageButton,
                  ...(language === option.code ? styles.languageButtonActive : {}),
                }}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div style={styles.buttonGroup}>
          <button
            onClick={() => {
              if (!username.trim()) {
                alert(t.home.enterUsername);
                return;
              }
              sessionStorage.setItem('username', username.trim());
              navigate('/game-mode');
            }}
            style={styles.button}
          >
            🎯 게임 시작
          </button>
        </div>
      </div>
    </div>
  );
};

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    padding: '10px',
  },
  content: {
    background: 'white',
    padding: 'clamp(20px, 5vw, 40px)',
    borderRadius: '12px',
    boxShadow: '0 10px 25px rgba(0,0,0,0.2)',
    maxWidth: '500px',
    width: '100%',
  },
  title: {
    textAlign: 'center',
    marginBottom: '15px',
    color: '#333',
    fontSize: 'clamp(20px, 5vw, 28px)',
  },
  subtitle: {
    textAlign: 'center',
    marginBottom: '20px',
    color: '#666',
    fontSize: 'clamp(14px, 3vw, 16px)',
  },
  rulesBox: {
    background: '#f7fafc',
    padding: 'clamp(12px, 3vw, 20px)',
    borderRadius: '8px',
    marginBottom: '20px',
    border: '2px solid #e2e8f0',
  },
  rulesTitle: {
    margin: '0 0 10px 0',
    color: '#667eea',
    fontSize: 'clamp(14px, 3.5vw, 18px)',
  },
  rulesList: {
    margin: '0',
    paddingLeft: '20px',
    color: '#555',
    fontSize: 'clamp(12px, 2.5vw, 14px)',
    lineHeight: '1.8',
  },
  inputGroup: {
    marginBottom: '15px',
  },
  languageGroup: {
    marginBottom: '20px',
  },
  languageLabel: {
    display: 'block',
    marginBottom: '8px',
    fontSize: '14px',
    fontWeight: 600,
    color: '#2d3748',
  },
  languageButtons: {
    display: 'flex',
    gap: '8px',
    flexWrap: 'wrap',
  },
  languageButton: {
    flex: '1 1 30%',
    minWidth: '90px',
    padding: '10px',
    borderRadius: '999px',
    border: '2px solid #cbd5f5',
    background: '#fff',
    color: '#4c51bf',
    fontWeight: 600,
    cursor: 'pointer',
    transition: 'background 0.2s, color 0.2s, border 0.2s',
  },
  languageButtonActive: {
    background: '#667eea',
    color: '#fff',
    borderColor: '#667eea',
  },
  input: {
    width: '100%',
    padding: 'clamp(10px, 2vw, 12px)',
    fontSize: 'clamp(14px, 3vw, 16px)',
    border: '2px solid #ddd',
    borderRadius: '8px',
    boxSizing: 'border-box',
  },
  buttonGroup: {
    display: 'flex',
    flexDirection: 'column',
    gap: '10px',
  },
  button: {
    padding: 'clamp(10px, 2.5vw, 14px)',
    fontSize: 'clamp(14px, 3.5vw, 18px)',
    fontWeight: 'bold',
    color: 'white',
    background: '#667eea',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'background 0.3s',
    width: '100%',
  },
  secondaryButton: {
    background: '#48bb78',
  },
};

export default Home;
