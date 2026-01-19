package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CurrencyType 재화 타입
type CurrencyType string

const (
	CurrencyTypeCredit  CurrencyType = "credit"   // 일반 재화
	CurrencyTypeHanCoin CurrencyType = "han_coin" // 프리미엄 재화
)

// TransactionType 거래 타입
type TransactionType string

const (
	TransactionTypePurchase    TransactionType = "purchase"    // 상점 구매
	TransactionTypeReward      TransactionType = "reward"      // 게임 보상
	TransactionTypeDailyBonus  TransactionType = "daily_bonus" // 출석 보너스
	TransactionTypeGift        TransactionType = "gift"        // 선물 받음
	TransactionTypeRefund      TransactionType = "refund"      // 환불
	TransactionTypeAdminGrant  TransactionType = "admin_grant" // 관리자 지급
	TransactionTypeLevelUp     TransactionType = "level_up"    // 레벨업 보상
	TransactionTypeAchievement TransactionType = "achievement" // 업적 보상
)

// TransactionStatus 거래 상태
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"   // 대기중
	TransactionStatusCompleted TransactionStatus = "completed" // 완료
	TransactionStatusFailed    TransactionStatus = "failed"    // 실패
	TransactionStatusCancelled TransactionStatus = "cancelled" // 취소됨
)

// Transaction 거래 내역
type Transaction struct {
	ID             primitive.ObjectID     `json:"id" bson:"_id,omitempty"`
	UserID         primitive.ObjectID     `json:"user_id" bson:"user_id"`                       // 사용자 ID
	Type           TransactionType        `json:"type" bson:"type"`                             // 거래 타입
	CurrencyType   CurrencyType           `json:"currency_type" bson:"currency_type"`           // 재화 타입
	Status         TransactionStatus      `json:"status" bson:"status"`                         // 거래 상태
	ChangeAmount   int                    `json:"change_amount" bson:"change_amount"`           // 금액 (양수: 획득, 음수: 차감)
	PreviousCredit int                    `json:"previous_credit" bson:"previous_credit"`       // 거래 전 잔액
	CurrentCredit  int                    `json:"current_credit" bson:"current_credit"`         // 거래 후 잔액
	Description    string                 `json:"description" bson:"description"`               // 거래 설명
	Metadata       map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"` // 추가 데이터
	CreatedAt      time.Time              `json:"created_at" bson:"created_at"`
}

// PurchaseTransaction 구매 거래 메타데이터
type PurchaseTransaction struct {
	ItemID   string `json:"item_id" bson:"item_id"`
	ItemName string `json:"item_name" bson:"item_name"`
	Quantity int    `json:"quantity" bson:"quantity"`
}

// RewardTransaction 보상 거래 메타데이터
type RewardTransaction struct {
	GameRoomID string `json:"game_room_id,omitempty" bson:"game_room_id,omitempty"`
	GameType   string `json:"game_type,omitempty" bson:"game_type,omitempty"`
	Rank       int    `json:"rank,omitempty" bson:"rank,omitempty"`
	IsWinner   bool   `json:"is_winner,omitempty" bson:"is_winner,omitempty"`
}
