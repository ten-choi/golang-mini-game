package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserStats struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id" json:"userId"`

	TotalGames int `bson:"total_games" json:"totalGames"`
	TotalScore int `bson:"total_score" json:"totalScore"`

	GameStats map[string]DetailedRecord `bson:"game_stats" json:"gameStats"`

	UpdatedAt time.Time `json:"updatedAt" bson:"updated_at"`
	CreatedAt time.Time `json:"createdAt" bson:"created_at"`
}

type DetailedRecord struct {
	// 개인전 기록
	Solo SoloRecord `bson:"solo,omitempty" json:"solo,omitempty"`
	// 팀전 기록
	Team TeamRecord `bson:"team,omitempty" json:"team,omitempty"`
}

type SoloRecord struct {
	Wins      int   `bson:"wins" json:"wins"`
	Loses     int   `bson:"loses" json:"loses"`
	BestScore int64 `bson:"best_score" json:"bestScore"`
}

type TeamRecord struct {
	Wins   int `bson:"wins" json:"wins"`
	Loses  int `bson:"loses" json:"loses"`
	MVPCnt int `bson:"mvp_cnt" json:"mvpCnt"` // 팀전만의 특화 지표
}
