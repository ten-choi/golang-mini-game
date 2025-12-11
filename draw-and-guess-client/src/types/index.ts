export interface GameRoom {
  id: string;
  uuid: string;
  is_active: boolean;
  drawer_user: string;
  players: Player[];
  current_word: string;
  current_word_translations?: Record<string, string>;
  round_number: number;
  time_left: number;
  game_status: 'waiting' | 'playing' | 'finished';
  used_words: string[];
  max_rounds: number;
  winning_score: number;
  created_at: string;
  updated_at: string;
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

export interface Stroke {
  points: Point[];
  color: string;
}

export interface Point {
  x: number;
  y: number;
}

export interface ChatMessage {
  sender: string;
  message: string;
  action: string;
}

export interface GameState {
  roomId: string;
  isDrawer: boolean;
  currentWord?: string;
  players: Player[];
  roundNumber: number;
  timeLeft: number;
  gameStatus: string;
}
