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
// Drawing Types (aligned with backend)
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
  player_id?: string;
  line_width?: number;
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
// WebSocket Types (aligned with backend)
// ============================================

// Base WebSocket Messages
export interface WebSocketRequest {
  type: 'subscribe' | 'unsubscribe' | 'message' | 'chat' | 'drawing' | 'game_action';
  channel: string;
  data?: any;
}

export interface WebSocketResponse {
  type: string;
  channel: string;
  data: any;
}

// Server Message Format (ws_message.go)
export interface WSSuccessMessage {
  type: string;
  payload: any;
}

export interface WSErrorMessage {
  type: 'ERROR';
  code: string;
  message: string;
}

// Game Event Types
export interface GameEventPayload {
  eventType: 'player_joined' | 'player_left' | 'game_started' | 'round_started' | 'round_ended' | 'game_ended';
  roomId: string;
  data: any;
}

export interface ChatMessagePayload {
  roomId: string;
  playerId: string;
  playerName: string;
  message: string;
}

export interface DrawingEventPayload {
  roomId: string;
  playerId: string;
  strokeId?: string;
  points?: Point[];
  color?: string;
  lineWidth?: number;
  action: 'start' | 'draw' | 'end' | 'clear' | 'undo';
}

// Legacy Game State Types (websocket_dto.go)
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

export interface RoundStartData {
  room_id: string;
  round: number;
  drawer: string;
  topic?: string;
  topic_hint: string;
  time_limit: number;
}

export interface RoundEndData {
  room_id: string;
  round: number;
  topic: string;
  winners: string[];
  scoreboard: Player[];
}

export interface GameEndData {
  room_id: string;
  winner: Player;
  final_score: Player[];
}

export interface PlayerJoinedData {
  room_id: string;
  user_id: string;
  username: string;
}

export interface PlayerLeftData {
  room_id: string;
  user_id: string;
  username: string;
}

export interface ErrorResponse {
  type: 'error';
  channel: string;
  data: ErrorMessage;
}

export interface ErrorMessage {
  code: string;
  message: string;
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
