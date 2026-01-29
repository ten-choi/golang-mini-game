package common

// WebSocket 채널 이름 상수
const (
	ChannelLobby      = "lobby"
	ChannelGamePrefix = "game/"
	ChannelUserPrefix = "user/" // 개인 채널 (초대, DM 등)
)

// 게임 타입 문자열 상수
const (
	GameTypeOX        = "OX"
	GameTypeQA        = "QA"
	GameTypeWordchain = "WORDCHAIN"
	GameTypeDrawing   = "DRAWING"
)

// 게임 타이밍 상수
const (
	RoundDelaySeconds = 6 // 라운드 종료 후 다음 라운드까지 대기 시간 (초)
)

// 유저 상태 상수
const (
	UserStatusLobby   = "lobby"
	UserStatusInGame  = "ingame"
	UserStatusOffline = "offline"
)

// Redis 키 접두사
const (
	RedisKeyUserStatus = "user:status:" // user:status:{userId}
	RedisKeyInvitation = "invitation:"  // invitation:{invitationId}
	RedisUserStatusTTL = 3600           // 1시간
	RedisInvitationTTL = 86400          // 24시간
)
