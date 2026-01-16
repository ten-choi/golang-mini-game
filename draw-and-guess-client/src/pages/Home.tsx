import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { apiService } from '../services/api';

const Home: React.FC = () => {
  const [username, setUsername] = useState('');
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const handleLogin = async () => {
    if (!username.trim()) {
      alert('닉네임을 입력해주세요');
      return;
    }
    
    setLoading(true);
    try {
      console.log('[Home] === 로그인 시작 ===');
      console.log('[Home] 입력된 HangeId:', username.trim());
      
      // 로그인은 HangeId로 조회
      const user = await apiService.getUserByHangeId(username.trim());
      console.log('[Home] getUserByHangeId 결과:', user);
      
      if (!user) {
        console.log('[Home] ❌ 사용자를 찾을 수 없습니다.');
        alert('존재하지 않는 HangeId입니다. 회원가입을 해주세요.');
        return;
      }
      
      console.log('[Home] ✅ 기존 사용자 로그인:', user);
      // 실제 username(name)을 sessionStorage에 저장
      sessionStorage.setItem('username', user.name);
      navigate('/rooms');
    } catch (error: any) {
      console.error('[Home] ❌ 로그인 에러:', error);
      alert(`로그인 실패: ${error?.message || '알 수 없는 오류'}`);
    } finally {
      setLoading(false);
    }
  };

  const handleSignup = async () => {
    if (!username.trim()) {
      alert('닉네임을 입력해주세요');
      return;
    }
    
    setLoading(true);
    try {
      console.log('[Home] === 회원가입 시작 ===');
      console.log('[Home] 입력된 HangeId:', username.trim());
      
      // 먼저 중복 확인 - HangeId로 확인
      const existingUser = await apiService.getUserByHangeId(username.trim());
      if (existingUser) {
        console.log('[Home] ❌ 이미 존재하는 HangeId');
        alert('이미 사용 중인 HangeId입니다. 다른 HangeId를 사용해주세요.');
        return;
      }
      
      // 새 사용자 생성
      const newUser = await apiService.createUser({
        hangeId: username.trim(),
        name: username.trim(),
      });
      console.log('[Home] ✅ 회원가입 완료:', newUser);
      
      alert(`환영합니다, ${newUser.name}님!`);
      // 실제 username(name)을 sessionStorage에 저장
      sessionStorage.setItem('username', newUser.name);
      navigate('/rooms');
    } catch (error: any) {
      console.error('[Home] ❌ 회원가입 에러:', error);
      alert(`회원가입 실패: ${error?.message || '알 수 없는 오류'}`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={styles.container}>
      <div style={styles.content}>
        <h1 style={styles.title}>🎨 그림 맞추기 게임</h1>
        
        <div style={styles.inputGroup}>
          <input
            type="text"
            placeholder="HangeId 입력 (로그인용 ID)"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleLogin()}
            style={styles.input}
            autoFocus
          />
        </div>

        <div style={styles.buttonGroup}>
          <button
            onClick={handleLogin}
            style={{...styles.button, ...styles.loginButton}}
            disabled={loading}
          >
            {loading ? '로딩 중...' : '🔑 로그인'}
          </button>
          
          <button
            onClick={handleSignup}
            style={{...styles.button, ...styles.signupButton}}
            disabled={loading}
          >
            {loading ? '로딩 중...' : '✨ 회원가입'}
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
    padding: '40px',
    borderRadius: '12px',
    boxShadow: '0 10px 25px rgba(0,0,0,0.2)',
    maxWidth: '400px',
    width: '100%',
  },
  title: {
    textAlign: 'center',
    marginBottom: '30px',
    color: '#333',
    fontSize: '28px',
  },
  inputGroup: {
    marginBottom: '20px',
  },
  input: {
    width: '100%',
    padding: '12px',
    fontSize: '16px',
    border: '2px solid #ddd',
    borderRadius: '8px',
    boxSizing: 'border-box',
  },
  buttonGroup: {
    display: 'flex',
    flexDirection: 'column',
    gap: '12px',
  },
  button: {
    padding: '14px',
    fontSize: '18px',
    fontWeight: 'bold',
    color: 'white',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    transition: 'background 0.3s',
    width: '100%',
  },
  loginButton: {
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  },
  signupButton: {
    background: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
  },
};

export default Home;
