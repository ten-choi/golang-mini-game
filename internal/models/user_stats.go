package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserStats struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`

	TotalGames int `bson:"total_games" json:"total_games"`
	TotalScore int `bson:"total_score" json:"total_score"`

	GameStats map[string]DetailedRecord `bson:"game_stats" json:"game_stats"`

	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
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
	BestScore int64 `bson:"best_score" json:"best_score"`
}

type TeamRecord struct {
	Wins   int `bson:"wins" json:"wins"`
	Loses  int `bson:"loses" json:"loses"`
	MVPCnt int `bson:"mvp_cnt" json:"mvp_cnt"` // 팀전만의 특화 지표
}
