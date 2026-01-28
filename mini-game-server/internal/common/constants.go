package common

// WebSocket 채널 이름 상수
const (
	ChannelLobby      = "lobby"
	ChannelGamePrefix = "game/"
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
