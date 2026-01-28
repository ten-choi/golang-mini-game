import React, { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { GeneralQuiz, OXQuiz } from '../types';

type QuizType = 'general' | 'ox';

const QuizGame: React.FC = () => {
  const { type } = useParams<{ type: QuizType }>();
  const navigate = useNavigate();
  // const { t } = useLanguage();
  // const [username] = useState(sessionStorage.getItem('username') || 'Guest');
  
  // 게임 상태
  const [quizzes, setQuizzes] = useState<(GeneralQuiz | OXQuiz)[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [score, setScore] = useState(0);
  const [selectedAnswer, setSelectedAnswer] = useState<number | boolean | null>(null);
  const [showResult, setShowResult] = useState(false);
  const [isCorrect, setIsCorrect] = useState(false);
  const [gameFinished, setGameFinished] = useState(false);

  useEffect(() => {
    // 실제로는 API에서 퀴즈를 가져와야 함
    // 여기서는 테스트용 하드코딩
    if (type === 'general') {
      setQuizzes(getTestGeneralQuizzes());
    } else if (type === 'ox') {
      setQuizzes(getTestOXQuizzes());
    }
  }, [type]);

  const currentQuiz = quizzes[currentIndex];

  const handleAnswer = (answer: number | boolean) => {
    setSelectedAnswer(answer);
    const correct = currentQuiz.answer === answer;
    setIsCorrect(correct);
    setShowResult(true);
    
    if (correct) {
      setScore(score + 1);
    }
  };

  const handleNext = () => {
    if (currentIndex < quizzes.length - 1) {
      setCurrentIndex(currentIndex + 1);
      setSelectedAnswer(null);
      setShowResult(false);
      setIsCorrect(false);
    } else {
      setGameFinished(true);
    }
  };

  const handleRestart = () => {
    setCurrentIndex(0);
    setScore(0);
    setSelectedAnswer(null);
    setShowResult(false);
    setIsCorrect(false);
    setGameFinished(false);
  };

  if (quizzes.length === 0) {
    return <div style={styles.loading}>퀴즈 로딩 중...</div>;
  }

  if (gameFinished) {
    const percentage = Math.round((score / quizzes.length) * 100);
    return (
      <div style={styles.container}>
        <div style={styles.resultCard}>
          <h1 style={styles.resultTitle}>🎉 퀴즈 완료!</h1>
          <div style={styles.scoreDisplay}>
            <div style={styles.scoreBig}>{score}</div>
            <div style={styles.scoreLabel}>/ {quizzes.length} 문제</div>
          </div>
          <div style={styles.percentage}>{percentage}%</div>
          <p style={styles.resultMessage}>
            {percentage >= 80 ? '훌륭해요! 🌟' : percentage >= 60 ? '잘했어요! 👍' : '다시 도전해보세요! 💪'}
          </p>
          <div style={styles.buttonGroup}>
            <button style={styles.retryButton} onClick={handleRestart}>
              다시 하기
            </button>
            <button style={styles.homeButton} onClick={() => navigate('/game-mode')}>
              게임 선택
            </button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div style={styles.container}>
      <div style={styles.quizCard}>
        <div style={styles.header}>
          <button onClick={() => navigate('/game-mode')} style={styles.backButton}>
            ← 나가기
          </button>
          <div style={styles.progress}>
            문제 {currentIndex + 1} / {quizzes.length}
          </div>
          <div style={styles.score}>점수: {score}</div>
        </div>

        <div style={styles.quizContent}>
          <div style={styles.category}>{currentQuiz.category}</div>
          <h2 style={styles.question}>{currentQuiz.question}</h2>

          {type === 'general' && 'options' in currentQuiz && (
            <div style={styles.optionsGrid}>
              {currentQuiz.options.map((option, index) => (
                <button
                  key={index}
                  style={{
                    ...styles.optionButton,
                    ...(selectedAnswer === index ? styles.selectedOption : {}),
                    ...(showResult && index === currentQuiz.answer ? styles.correctOption : {}),
                    ...(showResult && selectedAnswer === index && !isCorrect ? styles.wrongOption : {}),
                  }}
                  onClick={() => !showResult && handleAnswer(index)}
                  disabled={showResult}
                >
                  {option}
                </button>
              ))}
            </div>
          )}

          {type === 'ox' && (
            <div style={styles.oxButtons}>
              <button
                style={{
                  ...styles.oxButton,
                  ...styles.oxTrue,
                  ...(selectedAnswer === true ? styles.selectedOX : {}),
                  ...(showResult && currentQuiz.answer === true ? styles.correctOX : {}),
                  ...(showResult && selectedAnswer === true && !isCorrect ? styles.wrongOX : {}),
                }}
                onClick={() => !showResult && handleAnswer(true)}
                disabled={showResult}
              >
                ⭕ O
              </button>
              <button
                style={{
                  ...styles.oxButton,
                  ...styles.oxFalse,
                  ...(selectedAnswer === false ? styles.selectedOX : {}),
                  ...(showResult && currentQuiz.answer === false ? styles.correctOX : {}),
                  ...(showResult && selectedAnswer === false && !isCorrect ? styles.wrongOX : {}),
                }}
                onClick={() => !showResult && handleAnswer(false)}
                disabled={showResult}
              >
                ❌ X
              </button>
            </div>
          )}

          {showResult && (
            <div style={styles.resultBox}>
              <div style={isCorrect ? styles.correctMessage : styles.wrongMessage}>
                {isCorrect ? '✅ 정답입니다!' : '❌ 틀렸습니다!'}
              </div>
              {currentQuiz.explanation && (
                <div style={styles.explanation}>
                  💡 {currentQuiz.explanation}
                </div>
              )}
              <button style={styles.nextButton} onClick={handleNext}>
                {currentIndex < quizzes.length - 1 ? '다음 문제' : '결과 보기'}
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

// 테스트용 퀴즈 데이터 (실제로는 API에서 가져와야 함)
const getTestGeneralQuizzes = (): GeneralQuiz[] => [
  {
    id: '1',
    category: '지리',
    difficulty: 'easy',
    question: '대한민국의 수도는?',
    options: ['서울', '부산', '대구', '인천'],
    answer: 0,
    explanation: '대한민국의 수도는 서울입니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as GeneralQuiz,
  {
    id: '2',
    category: '과학',
    difficulty: 'easy',
    question: '물의 화학식은?',
    options: ['H2O', 'CO2', 'O2', 'NaCl'],
    answer: 0,
    explanation: '물의 화학식은 H2O입니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as GeneralQuiz,
  {
    id: '3',
    category: '역사',
    difficulty: 'medium',
    question: '한글을 만든 사람은?',
    options: ['이순신', '세종대왕', '김구', '유관순'],
    answer: 1,
    explanation: '세종대왕이 훈민정음(한글)을 창제했습니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as GeneralQuiz,
  {
    id: '4',
    category: '동물',
    difficulty: 'medium',
    question: '가장 빠른 육지 동물은?',
    options: ['사자', '표범', '치타', '말'],
    answer: 2,
    explanation: '치타는 시속 최대 120km로 달릴 수 있는 가장 빠른 육지 동물입니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as GeneralQuiz,
  {
    id: '5',
    category: '과학',
    difficulty: 'medium',
    question: '태양계에서 가장 큰 행성은?',
    options: ['지구', '화성', '목성', '토성'],
    answer: 2,
    explanation: '목성은 태양계에서 가장 큰 행성입니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as GeneralQuiz,
];

const getTestOXQuizzes = (): OXQuiz[] => [
  {
    id: '1',
    category: '과일',
    difficulty: 'easy',
    question: '사과는 과일이다',
    answer: true,
    explanation: '사과는 과일입니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as OXQuiz,
  {
    id: '2',
    category: '동물',
    difficulty: 'easy',
    question: '펭귄은 날 수 있다',
    answer: false,
    explanation: '펭귄은 날지 못하는 새입니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as OXQuiz,
  {
    id: '3',
    category: '과학',
    difficulty: 'medium',
    question: '지구는 태양 주위를 돈다',
    answer: true,
    explanation: '지구는 태양 주위를 공전합니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as OXQuiz,
  {
    id: '4',
    category: '동물',
    difficulty: 'medium',
    question: '거미는 곤충이다',
    answer: false,
    explanation: '거미는 거미강에 속하며 곤충이 아닙니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as OXQuiz,
  {
    id: '5',
    category: '과학',
    difficulty: 'hard',
    question: '빛의 속도는 소리의 속도보다 느리다',
    answer: false,
    explanation: '빛의 속도는 소리의 속도보다 훨씬 빠릅니다.',
    usageCount: 0,
    isActive: true,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
  } as OXQuiz,
];

const styles: { [key: string]: React.CSSProperties } = {
  container: {
    minHeight: '100vh',
    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    padding: '20px',
  },
  loading: {
    minHeight: '100vh',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    fontSize: '24px',
    color: 'white',
  },
  quizCard: {
    background: 'white',
    borderRadius: '20px',
    padding: '30px',
    maxWidth: '800px',
    width: '100%',
    boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3)',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '30px',
  },
  backButton: {
    padding: '10px 20px',
    background: '#f0f0f0',
    border: 'none',
    borderRadius: '8px',
    cursor: 'pointer',
    fontSize: '14px',
  },
  progress: {
    fontSize: '18px',
    fontWeight: 'bold',
    color: '#667eea',
  },
  score: {
    fontSize: '18px',
    fontWeight: 'bold',
    color: '#764ba2',
  },
  quizContent: {
    marginTop: '20px',
  },
  category: {
    display: 'inline-block',
    padding: '5px 15px',
    background: '#e8eaf6',
    color: '#667eea',
    borderRadius: '20px',
    fontSize: '14px',
    marginBottom: '15px',
  },
  question: {
    fontSize: '28px',
    fontWeight: 'bold',
    color: '#333',
    marginBottom: '30px',
    lineHeight: '1.4',
  },
  optionsGrid: {
    display: 'grid',
    gridTemplateColumns: '1fr 1fr',
    gap: '15px',
    marginBottom: '20px',
  },
  optionButton: {
    padding: '20px',
    fontSize: '18px',
    border: '2px solid #ddd',
    borderRadius: '12px',
    background: 'white',
    cursor: 'pointer',
    transition: 'all 0.3s',
    textAlign: 'left',
  },
  selectedOption: {
    borderColor: '#667eea',
    background: '#e8eaf6',
  },
  correctOption: {
    borderColor: '#4caf50',
    background: '#e8f5e9',
  },
  wrongOption: {
    borderColor: '#f44336',
    background: '#ffebee',
  },
  oxButtons: {
    display: 'flex',
    gap: '20px',
    marginBottom: '20px',
  },
  oxButton: {
    flex: 1,
    padding: '40px 20px',
    fontSize: '48px',
    border: '3px solid',
    borderRadius: '15px',
    cursor: 'pointer',
    transition: 'all 0.3s',
    background: 'white',
  },
  oxTrue: {
    borderColor: '#4ECDC4',
  },
  oxFalse: {
    borderColor: '#FF6B6B',
  },
  selectedOX: {
    transform: 'scale(1.05)',
    boxShadow: '0 5px 15px rgba(0, 0, 0, 0.2)',
  },
  correctOX: {
    background: '#e8f5e9',
    borderColor: '#4caf50',
  },
  wrongOX: {
    background: '#ffebee',
    borderColor: '#f44336',
  },
  resultBox: {
    marginTop: '30px',
    padding: '20px',
    background: '#f9f9f9',
    borderRadius: '12px',
  },
  correctMessage: {
    fontSize: '24px',
    fontWeight: 'bold',
    color: '#4caf50',
    marginBottom: '15px',
  },
  wrongMessage: {
    fontSize: '24px',
    fontWeight: 'bold',
    color: '#f44336',
    marginBottom: '15px',
  },
  explanation: {
    fontSize: '16px',
    color: '#666',
    lineHeight: '1.6',
    marginBottom: '20px',
  },
  nextButton: {
    width: '100%',
    padding: '15px',
    fontSize: '18px',
    fontWeight: 'bold',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '10px',
    cursor: 'pointer',
  },
  resultCard: {
    background: 'white',
    borderRadius: '20px',
    padding: '60px 40px',
    maxWidth: '500px',
    textAlign: 'center',
    boxShadow: '0 20px 60px rgba(0, 0, 0, 0.3)',
  },
  resultTitle: {
    fontSize: '36px',
    color: '#333',
    marginBottom: '30px',
  },
  scoreDisplay: {
    marginBottom: '20px',
  },
  scoreBig: {
    fontSize: '72px',
    fontWeight: 'bold',
    color: '#667eea',
  },
  scoreLabel: {
    fontSize: '24px',
    color: '#999',
  },
  percentage: {
    fontSize: '48px',
    fontWeight: 'bold',
    color: '#764ba2',
    marginBottom: '20px',
  },
  resultMessage: {
    fontSize: '24px',
    color: '#666',
    marginBottom: '40px',
  },
  buttonGroup: {
    display: 'flex',
    gap: '15px',
  },
  retryButton: {
    flex: 1,
    padding: '15px',
    fontSize: '18px',
    fontWeight: 'bold',
    background: '#667eea',
    color: 'white',
    border: 'none',
    borderRadius: '10px',
    cursor: 'pointer',
  },
  homeButton: {
    flex: 1,
    padding: '15px',
    fontSize: '18px',
    fontWeight: 'bold',
    background: '#95E1D3',
    color: 'white',
    border: 'none',
    borderRadius: '10px',
    cursor: 'pointer',
  },
};

export default QuizGame;
