package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Nickname     string             `bson:"nickname" json:"nickname"`
	PassWrod     string             `bson:"asdasdasd" json:"asdasdasd"`
	BirthDate    time.Time          `bson:"birth_date" json:"birth_date"`
	ProfileImage string             `bson:"profile_image" json:"profile_image"`
	WinningPoint int                `bson:"winning_point" json:"winning_point"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
}
