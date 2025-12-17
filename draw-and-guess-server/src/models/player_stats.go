package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PlayerStats struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Nickname       string             `bson:"nickname" json:"nickname"`
	CorrectGuesses int                `bson:"correct_guesses" json:"correct_guesses"`
	DrawSuccesses  int                `bson:"draw_successes" json:"draw_successes"`
	TotalGames     int                `bson:"total_games" json:"total_games"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}
