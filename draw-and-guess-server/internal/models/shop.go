package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ItemType 아이템 타입
type ItemType string

const (
	ItemTypeAvatar     ItemType = "avatar"     // 아바타
	ItemTypeEmoji      ItemType = "emoji"      // 이모지
	ItemTypeBackground ItemType = "background" // 배경
	ItemTypeFrame      ItemType = "frame"      // 프레임
	ItemTypeBadge      ItemType = "badge"      // 뱃지
	ItemTypeTheme      ItemType = "theme"      // 테마
)

// ItemRarity 아이템 등급
type ItemRarity string

const (
	RarityCommon    ItemRarity = "common"    // 일반
	RarityUncommon  ItemRarity = "uncommon"  // 고급
	RarityRare      ItemRarity = "rare"      // 희귀
	RarityEpic      ItemRarity = "epic"      // 영웅
	RarityLegendary ItemRarity = "legendary" // 전설
)

// ShopItem 상점 아이템
type ShopItem struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name         string             `json:"name" bson:"name"`                   // 아이템 이름
	Description  string             `json:"description" bson:"description"`     // 아이템 설명
	ItemType     ItemType           `json:"item_type" bson:"item_type"`         // 아이템 타입
	Rarity       ItemRarity         `json:"rarity" bson:"rarity"`               // 등급
	Price        int                `json:"price" bson:"price"`                 // 가격
	CurrencyType CurrencyType       `json:"currency_type" bson:"currency_type"` // 결제 재화 타입
	ImageURL     string             `json:"image_url" bson:"image_url"`         // 아이템 이미지 URL
	IsAvailable  bool               `json:"is_available" bson:"is_available"`   // 판매 가능 여부
	IsFeatured   bool               `json:"is_featured" bson:"is_featured"`     // 추천 상품 여부
	Stock        int                `json:"stock" bson:"stock"`                 // 재고 (-1: 무제한)
	RequireLevel int                `json:"require_level" bson:"require_level"` // 필요 레벨
	Tags         []string           `json:"tags" bson:"tags"`                   // 태그 (검색용)
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// UserInventory 사용자 인벤토리 (구매한 아이템)
type UserInventory struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`           // 사용자 ID
	ItemID      primitive.ObjectID `json:"item_id" bson:"item_id"`           // 아이템 ID
	ItemName    string             `json:"item_name" bson:"item_name"`       // 아이템 이름 (캐시)
	ItemType    ItemType           `json:"item_type" bson:"item_type"`       // 아이템 타입 (캐시)
	Quantity    int                `json:"quantity" bson:"quantity"`         // 수량
	IsEquipped  bool               `json:"is_equipped" bson:"is_equipped"`   // 장착 여부
	PurchasedAt time.Time          `json:"purchased_at" bson:"purchased_at"` // 구매 시간
}
