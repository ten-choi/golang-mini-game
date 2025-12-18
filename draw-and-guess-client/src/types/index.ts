// ============================================
// Core Game Types (aligned with backend)
// ============================================

export interface GameRoom {
  id: string;
  uuid: string;
  is_active: boolean;
  drawer_user: string;
  room_creator: string;
  last_round_winner?: string;
  players: Player[];
  current_word: string;
  current_word_translations?: Record<string, string>;
  round_number: number;
  time_left: number;
  game_status: GameStatus;
  used_words: string[];
  max_rounds: number;
  winning_score: number;
  game_type?: GameType;
  created_at: string;
  updated_at: string;
}

export type GameStatus = 'waiting' | 'playing' | 'finished';
export type GameType = 'ox' | 'general' | 'guess';

// Quiz Types
export interface GeneralQuiz {
  id: string;
  category: string;
  difficulty: string;
  question: string;
  options: string[];
  answer: number;
  explanation?: string;
  image_url?: string;
}

export interface OXQuiz {
  id: string;
  category: string;
  difficulty: string;
  question: string;
  answer: boolean;
  explanation?: string;
}

export interface Player {
  username: string;
  score: number;
  attempts: number;
}

export interface GameTopic {
  canonical: string;
  translations: Record<string, string>;
}

// ============================================
// Drawing Types
// ============================================

export interface Stroke {
  points: Point[];
  color: string;
  width?: number;
}

export interface Point {
  x: number;
  y: number;
}

export interface DrawingData {
  room_id: string;
  action: 'draw' | 'clear' | 'undo';
  points?: Point[];
  color?: string;
  width?: number;
}

// ============================================
// Chat & Messaging Types
// ============================================

export interface ChatMessage {
  username: string;
  text: string;
  type: 'chat' | 'system' | 'answer';
  timestamp?: string;
}

export interface ChatMessageData {
  room_id: string;
  user_id: string;
  username: string;
  message: string;
}

// ============================================
// WebSocket Types
// ============================================

export interface WebSocketRequest {
  type: 'subscribe' | 'unsubscribe' | 'message';
  channel: string;
  data?: any;
}

export interface WebSocketResponse {
  type: string;
  channel: string;
  data: any;
}

export interface GameStateData {
  room_id: string;
  current_round: number;
  drawer: string;
  time_left: number;
  players: Player[];
}

export interface TimerUpdateData {
  room_id: string;
  time_left: number;
  round_number: number;
  game_status: GameStatus;
}

export interface CorrectAnswerData {
  room_id: string;
  user_id: string;
  username: string;
  answer: string;
  score: number;
}

// ============================================
// API Response Types
// ============================================

export interface ApiResult<T = any> {
  status: boolean;
  message: string;
  result: T;
}

export interface CreateRoomResponse {
  room_id: string;
  current_word: string;
  current_word_translations?: Record<string, string>;
}

export interface AnswerSubmitResponse {
  is_correct: boolean;
  message: string;
}

// ============================================
// UI State Types
// ============================================

export interface GameState {
  roomId: string;
  isDrawer: boolean;
  currentWord?: string;
  players: Player[];
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
