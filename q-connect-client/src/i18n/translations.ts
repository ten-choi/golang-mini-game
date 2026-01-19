export type Language = 'en' | 'ko' | 'ja';

export interface Translations {
  home: {
    title: string;
    subtitle: string;
    enterUsername: string;
    usernamePlaceholder: string;
    selectLanguage: string;
    createRoom: string;
    joinRoom: string;
    viewRooms: string;
    rulesTitle: string;
    rules: string[];
  };
  roomList: {
    title: string;
    subtitle: string;
    availableRooms: string;
    noRooms: string;
    join: string;
    backToHome: string;
    round: string;
    status: {
      waiting: string;
      playing: string;
      finished: string;
    };
  };
  gameRoom: {
    title: string;
    host: string;
    player: string;
    loading: string;
    leave: string;
    round: string;
    timeLeft: string;
    seconds: string;
    status: string;
    statusValues: {
      waiting: string;
      playing: string;
      finished: string;
    };
    currentWord: string;
    gameStart: string;
    players: string;
    score: string;
    attempts: string;
    chat: string;
    messagePlaceholder: string;
    send: string;
    gameFinished: string;
    winner: string;
    correctAnswer: string;
    system: string;
    answeredCorrectly: string;
  };
}

export const translations: Record<Language, Translations> = {
  ko: {
    home: {
      title: '그림 맞추기 게임',
      subtitle: '친구들과 함께 그림을 그리고 맞춰보세요!',
      enterUsername: '사용자 이름',
      usernamePlaceholder: '이름을 입력하세요',
      selectLanguage: '언어 선택',
      createRoom: '방 만들기',
      joinRoom: '방 참가',
      viewRooms: '방 목록 보기',
      rulesTitle: '게임 규칙',
      rules: [
        '👑 호스트가 제시된 동물을 그립니다',
        '👥 플레이어들이 정답을 맞춥니다',
        '⏱️ 각 라운드는 60초입니다',
        '🎯 플레이어당 3번의 시도 기회',
        '⭐ 정답 시 플레이어 +2점, 호스트 +1점',
        '🏆 3점 먼저 획득 또는 3라운드 후 승자 결정',
      ],
    },
    roomList: {
      title: '그림 맞추기 게임',
      subtitle: '참가할 방을 선택하세요',
      availableRooms: '사용 가능한 방',
      noRooms: '현재 활성화된 방이 없습니다',
      join: '참가',
      backToHome: '홈으로',
      round: '라운드',
      status: {
        waiting: '대기 중',
        playing: '플레이 중',
        finished: '종료',
      },
    },
    gameRoom: {
      title: '그림 맞추기',
      host: '호스트 (그림 그리기)',
      player: '플레이어 (정답 맞추기)',
      loading: '로딩 중...',
      leave: '나가기',
      round: '라운드',
      timeLeft: '남은 시간',
      seconds: '초',
      status: '상태',
      statusValues: {
        waiting: '대기 중',
        playing: '플레이 중',
        finished: '종료',
      },
      currentWord: '제시어',
      gameStart: '게임 시작',
      players: '플레이어 점수',
      score: '점수',
      attempts: '시도',
      chat: '채팅',
      messagePlaceholder: '메시지를 입력하세요...',
      send: '전송',
      gameFinished: '게임 종료!',
      winner: '우승',
      correctAnswer: '정답!',
      system: '시스템',
      answeredCorrectly: '님이 정답을 맞혔습니다!',
    },
  },
  en: {
    home: {
      title: 'Draw and Guess Game',
      subtitle: 'Draw and guess with friends!',
      enterUsername: 'Username',
      usernamePlaceholder: 'Enter your name',
      selectLanguage: 'Select Language',
      createRoom: 'Create Room',
      joinRoom: 'Join Room',
      viewRooms: 'View Rooms',
      rulesTitle: 'Game Rules',
      rules: [
        '👑 Host draws the presented animal',
        '👥 Players guess the answer',
        '⏱️ Each round is 60 seconds',
        '🎯 3 attempts per player',
        '⭐ Correct answer: Player +2 points, Host +1 point',
        '🏆 Winner: First to 3 points or highest after 3 rounds',
      ],
    },
    roomList: {
      title: 'Draw and Guess Game',
      subtitle: 'Select a room to join',
      availableRooms: 'Available Rooms',
      noRooms: 'No active rooms available',
      join: 'Join',
      backToHome: 'Back to Home',
      round: 'Round',
      status: {
        waiting: 'Waiting',
        playing: 'Playing',
        finished: 'Finished',
      },
    },
    gameRoom: {
      title: 'Draw and Guess',
      host: 'Host (Drawing)',
      player: 'Player (Guessing)',
      loading: 'Loading...',
      leave: 'Leave',
      round: 'Round',
      timeLeft: 'Time Left',
      seconds: 'sec',
      status: 'Status',
      statusValues: {
        waiting: 'Waiting',
        playing: 'Playing',
        finished: 'Finished',
      },
      currentWord: 'Word',
      gameStart: 'Start Game',
      players: 'Player Scores',
      score: 'Score',
      attempts: 'Attempts',
      chat: 'Chat',
      messagePlaceholder: 'Type a message...',
      send: 'Send',
      gameFinished: 'Game Finished!',
      winner: 'Winner',
      correctAnswer: 'Correct!',
      system: 'System',
      answeredCorrectly: 'answered correctly!',
    },
  },
  ja: {
    home: {
      title: 'お絵描きゲーム',
      subtitle: '友達と絵を描いて当てよう！',
      enterUsername: 'ユーザー名',
      usernamePlaceholder: '名前を入力',
      selectLanguage: '言語選択',
      createRoom: 'ルーム作成',
      joinRoom: 'ルーム参加',
      viewRooms: 'ルーム一覧',
      rulesTitle: 'ゲームルール',
      rules: [
        '👑 ホストが提示された動物を描きます',
        '👥 プレイヤーが答えを当てます',
        '⏱️ 各ラウンドは60秒です',
        '🎯 プレイヤーごとに3回の試行',
        '⭐ 正解時：プレイヤー+2点、ホスト+1点',
        '🏆 勝者：3点先取または3ラウンド後最高得点',
      ],
    },
    roomList: {
      title: 'お絵描きゲーム',
      subtitle: '参加するルームを選択',
      availableRooms: '利用可能なルーム',
      noRooms: 'アクティブなルームがありません',
      join: '参加',
      backToHome: 'ホームへ',
      round: 'ラウンド',
      status: {
        waiting: '待機中',
        playing: 'プレイ中',
        finished: '終了',
      },
    },
    gameRoom: {
      title: 'お絵描きゲーム',
      host: 'ホスト（描画）',
      player: 'プレイヤー（推測）',
      loading: '読み込み中...',
      leave: '退出',
      round: 'ラウンド',
      timeLeft: '残り時間',
      seconds: '秒',
      status: 'ステータス',
      statusValues: {
        waiting: '待機中',
        playing: 'プレイ中',
        finished: '終了',
      },
      currentWord: 'お題',
      gameStart: 'ゲーム開始',
      players: 'プレイヤースコア',
      score: 'スコア',
      attempts: '試行',
      chat: 'チャット',
      messagePlaceholder: 'メッセージを入力...',
      send: '送信',
      gameFinished: 'ゲーム終了！',
      winner: '優勝',
      correctAnswer: '正解！',
      system: 'システム',
      answeredCorrectly: 'さんが正解しました！',
    },
  },
};
