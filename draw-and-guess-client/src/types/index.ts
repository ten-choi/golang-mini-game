// ============================================
// Core Game Types (aligned with backend GraphQL schema)
// ============================================

export interface GameRoom {
  id: string;
  name: string;
  gameType: GameType;
  status: GameStatus;
  currentRound: number;
  totalRounds: number;
  roundTimeLimit: number;
  players: Player[];
  maxPlayers: number;
  hostUsername: string;
  usedQuizIds: string[];
  isPrivate: boolean;
  password?: string;
  createdAt: string;
}

export type GameStatus = 'WAITING' | 'PLAYING' | 'FINISHED';
export type GameType = 'OX' | 'QA' | 'WORDCHAIN' | 'DRAWING';

// Quiz Types
export interface GeneralQuiz {
  id: string;
  category: string;
  difficulty: string;
  question: string;
  options: string[];
  answer: number;
  explanation?: string;
  imageUrl?: string;
  usageCount?: number;
  isActive?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface OXQuiz {
  id: string;
  category: string;
  difficulty: string;
  question: string;
  answer: boolean;
  explanation?: string;
  usageCount?: number;
  isActive?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface Player {
  username: string;
  displayName: string;
  score: number;
  isReady: boolean;
}

// User Types
export interface User {
  id: string;
  nickname: string;
  avatarUrl?: string;
  level: number;
  credit: number;
  hanCoin: number;
  guildId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateUserInput {
  nickname: string;
  avatarUrl?: string;
}

export interface UpdateUserInput {
  avatarUrl?: string;
  level?: number;
  credit?: number;
}

// Game Room Input Types
export interface CreateGameRoomInput {
  name: string;
  gameType: GameType;
  maxPlayers: number;
  totalRounds: number;
  roundTimeLimit: number;
  hostUsername: string;
  isPrivate?: boolean;
  password?: string;
}

export interface UpdateGameRoomInput {
  name?: string;
  maxPlayers?: number;
  totalRounds?: number;
  roundTimeLimit?: number;
  isPrivate?: boolean;
  password?: string;
}

// Invitation Types
export type InviteStatus = 'PENDING' | 'ACCEPTED' | 'REJECTED' | 'EXPIRED';

export interface Invitation {
  id: string;
  roomId: string;
  room: GameRoom;
  inviterId: string;
  inviter: User;
  inviteeId: string;
  invitee: User;
  status: InviteStatus;
  createdAt: string;
  expiresAt: string;
}

// Player Stats Types
export interface PlayerStats {
  username: string;
  gameType: string;
  totalGames: number;
  totalWins: number;
  totalScore: number;
  createdAt: string;
  updatedAt: string;
}

// Game Config Types
export interface GameConfig {
  maxPlayers: number;
  roundDuration: number;
  drawingTime: number;
  guessingTime: number;
  roundsPerGame: number;
}

// Wordchain Types
export interface WordchainPrompt {
  word: string;
  hint?: string;
}

// Game Event Types
export type GameEventType = 
  | 'PLAYER_JOINED'
  | 'PLAYER_LEFT'
  | 'PLAYER_READY'
  | 'HOST_CHANGED'
  | 'ROUND_STARTED'
  | 'ROUND_ENDED'
  | 'ANSWER_SUBMITTED'
  | 'CORRECT_ANSWER'
  | 'WRONG_ANSWER'
  | 'GAME_ENDED';

export interface GameEvent {
  type: GameEventType;
  roomId: string;
  username?: string;
  displayName?: string;
  data?: string;
  timestamp: string;
}

export interface ErrorEvent {
  code: string;
  message: string;
  timestamp: string;
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
  id: string;
  roomId: string;
  username: string;
  displayName: string;
  message: string;
  timestamp: string;
}

export interface ChatMessageData {
  roomId: string;
  userId: string;
  username: string;
  message: string;
}

// ============================================
// WebSocket Types (aligned with backend websocket_dto.go)
// ============================================

// Base WebSocket Messages
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

// WebSocket Message Data Types from websocket_dto.go
export interface WSChatMessageData {
  room_id: string;
  user_id: string;
  username: string;
  message: string;
}

export interface WSGameStateData {
  room_id: string;
  current_round: number;
  drawer: string;
  time_left: number;
  players: Array<{
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
  drawer: string;
  topic?: string;
  topic_hint: string;
  time_limit: number;
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
