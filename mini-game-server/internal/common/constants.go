package common

// WebSocket channel name constants
const (
	ChannelLobby      = "lobby"
	ChannelGamePrefix = "game/"
	ChannelUserPrefix = "user/" // Personal channel (invitations, DM, etc.)
)

// Game type string constants
const (
	GameTypeOX        = "OX"
	GameTypeQA        = "QA"
	GameTypeWordchain = "WORDCHAIN"
	GameTypeDrawing   = "DRAWING"
)

// Game timing constants
const (
	RoundDelaySeconds     = 6 // Delay time in seconds after round end until next round
	RoomResetDelaySeconds = 5 // Delay time in seconds after game end until room resets to WAITING
)

// Game scoring constants
const (
	WordchainCorrectScore  = 50 // Points awarded for correct wordchain answer
	QuizScorePerDifficulty = 50 // Points multiplier per difficulty level (difficulty × 50)
	QuizMinDifficulty      = 1  // Minimum quiz difficulty level
	QuizMaxDifficulty      = 5  // Maximum quiz difficulty level
)

// User status constants
const (
	UserStatusLobby   = "lobby"
	UserStatusInGame  = "ingame"
	UserStatusOffline = "offline"
)

// Game room status constants
const (
	RoomStatusWaiting  = "WAITING"
	RoomStatusPlaying  = "PLAYING"
	RoomStatusFinished = "FINISHED"
)

// Redis key prefixes
const (
	RedisKeyUserStatus = "user:status:" // user:status:{userId}
	RedisKeyInvitation = "invitation:"  // invitation:{invitationId}
	RedisUserStatusTTL = 3600           // 1 hour
	RedisInvitationTTL = 86400          // 24 hours
)
