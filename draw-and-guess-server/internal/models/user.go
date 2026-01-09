package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Nickname  string             `json:"nickname" bson:"nickname"`
	AvatarURL string             `json:"avatar_url" bson:"avatar_url"`
	Level     int                `json:"level" bson:"level"`
	Credit    int                `json:"credit" bson:"credit"`     // 일반 재화 (게임 플레이로 획득)
	HanCoin   int                `json:"han_coin" bson:"han_coin"` // 프리미엄 재화 (외부 시스템 관리)
	GuildID   primitive.ObjectID `json:"guild_id,omitempty" bson:"guild_id,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
