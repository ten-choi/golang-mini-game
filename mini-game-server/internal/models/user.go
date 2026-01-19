package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HangeID   string             `json:"hangeId" bson:"hange_id"`
	Name      string             `json:"name" bson:"name"`
	AvatarURL string             `json:"avatarUrl" bson:"avatar_url"`
	Level     int                `json:"level" bson:"level"`
	Credit    int                `json:"credit" bson:"credit"` // 일반 재화 (게임 플레이로 획듍)
	GuildID   primitive.ObjectID `json:"guildId,omitempty" bson:"guild_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updated_at"`
}
