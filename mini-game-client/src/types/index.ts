// ============================================
// Core Types - Aligned with q-connect-server GraphQL Schema
// ============================================

/** RFC3339 형식의 날짜/시간 문자열 */
export type Time = string;

// ============================================
// User Types
// ============================================

export interface User {
  /** 고유 식별자 (MongoDB ObjectID) */
  id: string;
  /** 로그인용 식별자 (3-20자, 로그인 전용) */
  hangeId: string;
  /** 고유 사용자명 (3-20자, 주요 식별자) */
  name: string;
  /** 프로필 이미지 URL */
  avatarUrl?: string;
  /** 플레이어 레벨 */
  level: number;
  /** 일반 재화 (게임 플레이로 획득) */
  credit: number;
  /** 소속 길드 ID (선택사항) */
  guildId?: string;
  /** 계정 생성 시각 */
  createdAt: Time;
  /** 마지막 업데이트 시각 */
  updatedAt: Time;
}

export interface CreateUserInput {
  /** 로그인용 식별자 (3-20자) */
  hangeId: string;
  /** 고유 사용자명 (3-20자) (선택사항) */
  name?: string;
  /** 프로필 이미지 URL (선택사항) */
  avatarUrl?: string;
}

export interface UpdateUserInput {
  /** 새로운 프로필 이미지 URL */
  avatarUrl?: string;
  /** 새로운 레벨 */
  level?: number;
  /** 새로운 일반 재화 */
  credit?: number;
}

// ============================================
// Game Room Types
// ============================================

/** 게임 모드 */
export type GameType = 'OX' | 'QA' | 'WORDCHAIN' | 'DRAWING';

/** 게임방 상태 */
export type GameStatus = 'WAITING' | 'PLAYING' | 'FINISHED';

/** 게임방 내 사용자 정보 */
export interface GameUser {
  /** 사용자의 고유 id */
  userId: string;
  /** 사용자의 고유 사용자명 */
  name: string;
  /** 현재 게임에서 획득한 점수 */
  score: number;
  /** 준비 완료 여부 */
  isReady: boolean;
}

/** 멀티플레이어 게임 세션 */
export interface GameRoom {
  /** 고유 게임방 식별자 (UUID) */
  id: string;
  /** 게임방 표시 이름 */
  name: string;
  /** 게임 모드: OX, QA, 또는 WORDCHAIN */
  gameType: GameType;
  /** 현재 게임 상태 */
  status: GameStatus;
  /** 현재 라운드 번호 (1부터 시작) */
  currentRound: number;
  /** 총 라운드 수 (기본값: 5) */
  totalRounds: number;
  /** 라운드당 제한 시간 (초 단위, 기본값: 30초) */
  roundTimeLimit: number;
  /** 게임방 내 사용자 목록 */
  users: GameUser[];
  /** 최대 사용자 수 (2-10명) */
  maxUsers: number;
  /** 방장 사용자 ID */
  hostUserId: string;
  /** 현재 게임에서 사용된 퀴즈 ID 목록 */
  usedQuizIds: string[];
  /** 비공개방 여부 */
  isPrivate: boolean;
  /** 게임방 비밀번호 (비공개방만) */
  password?: string;
  /** 게임방 생성 시각 */
  createdAt: Time;
  
  // 끝말잇기 게임 전용 필드
  /** 끝말잇기 게임 - 마지막 단어 */
  wordchainLastWord?: string;
  /** 끝말잇기 게임 - 사용된 단어 목록 */
  wordchainUsedWords?: string[];
  /** 끝말잇기 게임 - 현재 턴 플레이어 ID */
  currentTurnUserId?: string;
  /** 끝말잇기 게임 - 턴 시작 시간 */
  wordchainTurnStartTime?: Time;
}

export interface CreateGameRoomInput {
  /** 게임방 이름 (1-50자, 로비에 표시) */
  name: string;
  /** 게임 모드 (OX, QA, WORDCHAIN 중 선택) */
  gameType: GameType;
  /** 최대 플레이어 수 (2-10명, 권장: 4-6명) */
  maxUsers: number;
  /** 총 라운드 수 (1-20, 권장: 3-10) */
  totalRounds: number;
  /** 라운드당 제한 시간 (초 단위, 10-300초, 기본값: 30초) */
  roundTimeLimit?: number;
  /** 방장 사용자 ID (게임방 생성자) */
  hostUserId: string;
  /** 비공개 방 여부 */
  isPrivate?: boolean;
  /** 게임방 비밀번호 (isPrivate=true 시 필수, 4-20자) */
  password?: string;
}

export interface UpdateGameRoomInput {
  /** 새로운 게임방 이름 */
  name?: string;
  /** 새로운 최대 플레이어 수 */
  maxUsers?: number;
  /** 새로운 총 라운드 수 */
  totalRounds?: number;
  /** 새로운 라운드당 제한 시간 (초 단위, 10-300초) */
  roundTimeLimit?: number;
  /** 비공개 설정 변경 */
  isPrivate?: boolean;
  /** 새로운 비밀번호 */
  password?: string;
}

// ============================================
// Invitation Types
// ============================================

export type InviteStatus = 'PENDING' | 'ACCEPTED' | 'REJECTED' | 'EXPIRED';

export interface Invitation {
  /** 고유 초대 ID (UUID) */
  id: string;
  /** 초대 대상 게임방 ID */
  roomId: string;
  /** 초대 대상 게임방 정보 */
  room: GameRoom;
  /** 초대한 사용자 ID */
  inviterId: string;
  /** 초대한 사용자 정보 */
  inviter: User;
  /** 초대받은 사용자 ID */
  inviteeId: string;
  /** 초대받은 사용자 정보 */
  invitee: User;
  /** 현재 초대 상태 */
  status: InviteStatus;
  /** 초대 생성 시각 */
  createdAt: Time;
  /** 초대 만료 시각 (생성 후 24시간) */
  expiresAt: Time;
}

// ============================================
// Quiz Types
// ============================================

/** OX 퀴즈 (참/거짓 문제) */
export interface OXQuiz {
  /** 고유 퀴즈 ID (Snowflake) */
  id: string;
  /** 퀴즈 카테고리 */
  category: string;
  /** 난이도: easy, medium, hard */
  difficulty: string;
  /** 문제 텍스트 */
  question: string;
  /** 정답: true 또는 false */
  answer: boolean;
  /** 정답 설명 (선택사항) */
  explanation?: string;
  /** 이 퀴즈가 사용된 횟수 */
  usageCount: number;
  /** 퀴즈 활성화 여부 */
  isActive: boolean;
  /** 퀴즈 생성 시각 */
  createdAt: Time;
  /** 마지막 업데이트 시각 */
  updatedAt: Time;
}

/** 4지선다 퀴즈 */
export interface GeneralQuiz {
  /** 고유 퀴즈 ID (Snowflake) */
  id: string;
  /** 퀴즈 카테고리 */
  category: string;
  /** 난이도: easy, medium, hard */
  difficulty: string;
  /** 문제 텍스트 */
  question: string;
  /** 4개의 선택지 배열 */
  options: string[];
  /** 정답 인덱스 (0-3) */
  answer: number;
  /** 정답 설명 (선택사항) */
  explanation?: string;
  /** 시각적 문제를 위한 이미지 URL (선택사항) */
  imageUrl?: string;
  /** 이 퀴즈가 사용된 횟수 */
  usageCount: number;
  /** 퀴즈 활성화 여부 */
  isActive: boolean;
  /** 퀴즈 생성 시각 */
  createdAt: Time;
  /** 마지막 업데이트 시각 */
  updatedAt: Time;
}

/** OX 퀴즈의 선택지 */
export type OXChoice = 'O' | 'X';

// ============================================
// Game Config & Data Types
// ============================================

/** 게임 전역 설정 정보 */
export interface GameConfig {
  /** 최대 플레이어 수 제한 */
  maxUsers: number;
  /** 라운드당 제한 시간 (초) */
  roundDuration: number;
  /** 그리기 제한 시간 (초, 향후 기능) */
  drawingTime: number;
  /** 추측 제한 시간 (초, 향후 기능) */
  guessingTime: number;
  /** 게임당 기본 라운드 수 */
  roundsPerGame: number;
}

/** 끝말잇기 게임의 시작 단어 정보 */
export interface WordchainPrompt {
  /** 시작 단어 (히라가나 또는 카타카나, 2자 이상) */
  word: string;
  /** 힌트 (선택사항, 단어의 의미) */
  hint?: string;
}

// ============================================
// User Stats Types
// ============================================

/** 사용자의 게임 타입별 통계 정보 */
export interface UserStats {
  /** 사용자명 */
  username: string;
  /** 게임 타입 (OX, QA, WORDCHAIN) */
  gameType: string;
  /** 총 플레이한 게임 수 */
  totalGames: number;
  /** 총 승리 횟수 (1등 횟수) */
  totalWins: number;
  /** 총 획득 점수 (모든 게임 누적) */
  totalScore: number;
  /** 통계 생성 시각 */
  createdAt: Time;
  /** 마지막 업데이트 시각 */
  updatedAt: Time;
}

// ============================================
// Result Types
// ============================================

/** 답변 제출 결과 */
export interface AnswerResult {
  /** 제출 성공 여부 */
  success: boolean;
  /** 정답 여부 (제출자에게만 전달) */
  isCorrect: boolean;
  /** 획득한 점수 (정답 시 100점) */
  earnedScore: number;
  /** 현재 총 점수 */
  totalScore: number;
  /** 정답 (라운드 종료 후에만 제공) */
  correctAnswer?: string;
  /** 설명 (선택사항) */
  explanation?: string;
}

// ============================================
// Chat & Events Types
// ============================================

/** 게임방 채팅 메시지 */
export interface ChatMessage {
  /** 고유 메시지 ID */
  id: string;
  /** 메시지가 전송된 게임방 ID */
  roomId: string;
  /** 발신자 사용자명 */
  username: string;
  /** 발신자 표시 이름 */
  displayName: string;
  /** 메시지 내용 (최대 500자) */
  message: string;
  /** 전송 시각 */
  timestamp: Time;
}

/** 게임 내에서 발생하는 이벤트 타입 */
export type GameEventType =
  | 'user_JOINED'
  | 'user_LEFT'
  | 'user_READY'
  | 'HOST_CHANGED'
  | 'ROUND_STARTED'
  | 'ROUND_ENDED'
  | 'ANSWER_SUBMITTED'
  | 'CORRECT_ANSWER'
  | 'WRONG_ANSWER'
  | 'GAME_ENDED';

/** 게임 이벤트 상세 정보 */
export interface GameEvent {
  /** 이벤트 타입 */
  type: GameEventType;
  /** 이벤트가 발생한 게임방 ID */
  roomId: string;
  /** 이벤트 관련 사용자명 (선택사항) */
  username?: string;
  /** 이벤트 관련 사용자 표시 이름 (선택사항) */
  displayName?: string;
  /** 추가 데이터 (JSON 문자열) */
  data?: string;
  /** 이벤트 발생 시각 */
  timestamp: Time;
}

/** 에러 이벤트 정보 */
export interface ErrorEvent {
  /** 에러 코드 */
  code: string;
  /** 사용자에게 표시할 에러 메시지 */
  message: string;
  /** 에러 발생 시각 */
  timestamp: Time;
}

/** 사용자의 연결 상태 정보 */
export interface UserConnection {
  /** 사용자명 */
  username: string;
  /** 연결 상태 */
  isConnected: boolean;
  /** 마지막 접속 시각 */
  lastSeen: Time;
  /** 핑 (밀리초) */
  ping?: number;
}

// ============================================
// WebSocket Types (aligned with backend websocket_dto.go)
// ============================================

export interface WebSocketRequest {
  type: 'subscribe' | 'unsubscribe' | 'message' | 'identify';
  channel: string;
  data?: any;
}

export interface WebSocketResponse {
  type: string;
  channel: string;
  data: any;
}

// WebSocket Message Data Types from websocket_dto.go
export interface WSChatMessageData {
  room_id: string;
  user_id: string;
  username: string;
  message: string;
}

/** Chat message payload for WebSocket handlers (camelCase for frontend) */
export interface ChatMessagePayload {
  roomId: string;
  userId: string;
  username: string;
  message: string;
}

/** Game event payload for WebSocket handlers (camelCase for frontend) */
export interface GameEventPayload {
  type: GameEventType;
  roomId: string;
  userId?: string;
  username?: string;
  data?: any;
}

export interface WSGameStateData {
  room_id: string;
  current_round: number;
  drawer?: string;
  time_left: number;
  users: Array<{
    username: string;
    score: number;
    is_ready: boolean;
  }>;
}

export interface WSDrawingData {
  room_id: string;
  action: 'draw' | 'clear' | 'undo';
  points?: Array<{ x: number; y: number }>;
  color?: string;
  width?: number;
}

export interface WSAnswerSubmitData {
  room_id: string;
  user_id: string;
  username: string;
  answer: string;
}

export interface WSCorrectAnswerData {
  room_id: string;
  user_id: string;
  username: string;
  answer: string;
  score: number;
}

export interface WSRoundStartData {
  room_id: string;
  round: number;
  drawer?: string;
  topic?: string;
  topic_hint?: string;
  time_limit: number;
  quiz?: OXQuiz | GeneralQuiz; // 서버에서 퀴즈 데이터 전송
}

export interface WSRoundEndData {
  room_id: string;
  round: number;
  correct_answer: string;
  winners: string[];
  scoreboard: Array<{
    username: string;
    score: number;
  }>;
}

export interface WSGameEndData {
  room_id: string;
  winner: {
    username: string;
    score: number;
  };
  final_scores: Array<{
    username: string;
    score: number;
  }>;
}

// ============================================
// API Response Types
// ============================================

export interface GraphQLResponse<T = any> {
  data?: T;
  errors?: Array<{
    message: string;
    extensions?: any;
  }>;
}

export interface ApiError extends Error {
  statusCode?: number;
  code?: string;
  originalError?: any;
}

// ============================================
// Drawing Types
// ============================================

export interface Point {
  x: number;
  y: number;
}

export interface Stroke {
  points: Point[];
  color: string;
  width?: number;
}

export interface DrawingData {
  room_id: string;
  action: 'draw' | 'clear' | 'undo' | 'start' | 'end';
  points?: Point[];
  color?: string;
  width?: number;
  stroke_id?: string;
  user_id?: string;
  line_width?: number;
}

/** Drawing event payload for WebSocket messages (camelCase for frontend) */
export interface DrawingEventPayload {
  roomId: string;
  userId: string;
  action: 'draw' | 'clear' | 'undo' | 'start' | 'end';
  points?: Point[];
  color?: string;
  width?: number;
  strokeId?: string;
  lineWidth?: number;
}

// ============================================
// UI State Types
// ============================================

export interface GameState {
  roomId: string;
  isDrawer: boolean;
  currentWord?: string;
  users: GameUser[];
  roundNumber: number;
  timeLeft: number;
  gameStatus: GameStatus;
}

export interface CanvasState {
  isDrawing: boolean;
  currentStroke: Point[];
  strokes: Stroke[];
  selectedColor: string;
  lineWidth: number;
}

// ============================================
// Language Types
// ============================================

export type SupportedLanguage = 'ko' | 'en' | 'ja' | 'zh';

export interface LanguageOption {
  code: SupportedLanguage;
  label: string;
  flag?: string;
}
