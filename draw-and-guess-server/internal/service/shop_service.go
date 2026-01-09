package service

import (
	"context"
	"fmt"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ShopService 상점 서비스 인터페이스
type ShopService interface {
	// Shop Item
	GetItem(ctx context.Context, itemID string) (*models.ShopItem, error)
	GetAllItems(ctx context.Context, filter *repository.ShopItemFilter) ([]*models.ShopItem, error)
	GetFeaturedItems(ctx context.Context) ([]*models.ShopItem, error)
	GetItemsByType(ctx context.Context, itemType models.ItemType) ([]*models.ShopItem, error)
	CreateItem(ctx context.Context, item *models.ShopItem) (*models.ShopItem, error)
	UpdateItem(ctx context.Context, itemID string, item *models.ShopItem) (*models.ShopItem, error)
	DeleteItem(ctx context.Context, itemID string) error

	// Purchase & Inventory
	PurchaseItem(ctx context.Context, userID, itemID string, quantity int) error
	GetUserInventory(ctx context.Context, userID string, itemType *models.ItemType) ([]*models.UserInventory, error)
	EquipItem(ctx context.Context, userID, itemID string) error
	UnequipItem(ctx context.Context, userID, itemID string) error
}

type shopService struct {
	shopRepo        repository.ShopRepository
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
}

// NewShopService 새 상점 서비스 생성
func NewShopService(shopRepo repository.ShopRepository, transactionRepo repository.TransactionRepository, userRepo repository.UserRepository) ShopService {
	return &shopService{
		shopRepo:        shopRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}

func (s *shopService) GetItem(ctx context.Context, itemID string) (*models.ShopItem, error) {
	return s.shopRepo.GetItemByID(ctx, itemID)
}

func (s *shopService) GetAllItems(ctx context.Context, filter *repository.ShopItemFilter) ([]*models.ShopItem, error) {
	return s.shopRepo.GetAllItems(ctx, filter)
}

func (s *shopService) GetFeaturedItems(ctx context.Context) ([]*models.ShopItem, error) {
	return s.shopRepo.GetFeaturedItems(ctx)
}

func (s *shopService) GetItemsByType(ctx context.Context, itemType models.ItemType) ([]*models.ShopItem, error) {
	return s.shopRepo.GetItemsByType(ctx, itemType)
}

func (s *shopService) CreateItem(ctx context.Context, item *models.ShopItem) (*models.ShopItem, error) {
	// 기본값 설정
	if item.Stock == 0 {
		item.Stock = -1 // 무제한
	}
	if !item.IsAvailable {
		item.IsAvailable = true
	}

	return s.shopRepo.CreateItem(ctx, item)
}

func (s *shopService) UpdateItem(ctx context.Context, itemID string, item *models.ShopItem) (*models.ShopItem, error) {
	return s.shopRepo.UpdateItem(ctx, itemID, item)
}

func (s *shopService) DeleteItem(ctx context.Context, itemID string) error {
	return s.shopRepo.DeleteItem(ctx, itemID)
}

func (s *shopService) PurchaseItem(ctx context.Context, userID, itemID string, quantity int) error {
	if quantity <= 0 {
		return common.NewInternalError("quantity must be positive", nil)
	}

	// 1. 아이템 정보 조회
	item, err := s.shopRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return err
	}

	if !item.IsAvailable {
		return common.NewInternalError("item is not available", nil)
	}

	// 2. 재고 확인
	if item.Stock != -1 && item.Stock < quantity {
		return common.NewInternalError("insufficient stock", nil)
	}

	// 3. 사용자 정보 조회
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return common.NewInternalError("invalid user ID", err)
	}

	// Nickname으로 조회해야 하므로 임시로 스킵하거나 수정 필요
	// user, err := s.userRepo.GetByNickname(ctx, userID) // userID를 nickname으로 사용?
	// 여기서는 직접 credit 조회가 필요한데, User 모델에서 ID로 조회하는 메서드 필요

	// 4. 가격 계산
	totalPrice := item.Price * quantity

	// 5. 잔액 확인 (Transaction을 통해 조회 - 재화 타입별)
	currentBalance, err := s.transactionRepo.GetUserBalance(ctx, userID, item.CurrencyType)
	if err != nil {
		return err
	}

	if currentBalance < totalPrice {
		return common.NewInternalError(fmt.Sprintf("insufficient balance. need %d, have %d", totalPrice, currentBalance), nil)
	}

	// 6. 이미 보유 중인지 확인
	existingItem, err := s.shopRepo.GetUserItem(ctx, userID, itemID)
	if err != nil {
		return err
	}

	// 7. 거래 생성
	transaction := &models.Transaction{
		UserID:         userObjID,
		Type:           models.TransactionTypePurchase,
		CurrencyType:   item.CurrencyType, // 아이템의 결제 재화 타입
		Status:         models.TransactionStatusCompleted,
		ChangeAmount:   -totalPrice,
		PreviousCredit: currentBalance,
		CurrentCredit:  currentBalance - totalPrice,
		Description:    fmt.Sprintf("Purchased %s x%d", item.Name, quantity),
		Metadata: map[string]interface{}{
			"item_id":   itemID,
			"item_name": item.Name,
			"quantity":  quantity,
		},
	}

	_, err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		return err
	}

	// 8. 인벤토리에 추가 또는 수량 증가
	if existingItem != nil {
		err = s.shopRepo.UpdateInventoryItem(ctx, userID, itemID, quantity, nil)
	} else {
		itemObjID, _ := primitive.ObjectIDFromHex(itemID)
		inventory := &models.UserInventory{
			UserID:     userObjID,
			ItemID:     itemObjID,
			ItemName:   item.Name,
			ItemType:   item.ItemType,
			Quantity:   quantity,
			IsEquipped: false,
		}
		err = s.shopRepo.AddItemToInventory(ctx, inventory)
	}

	if err != nil {
		// 거래 취소 처리
		_ = s.transactionRepo.UpdateStatus(ctx, transaction.ID.Hex(), models.TransactionStatusCancelled)
		return err
	}

	// 9. 재고 차감 (무제한이 아닌 경우)
	if item.Stock != -1 {
		err = s.shopRepo.UpdateStock(ctx, itemID, -quantity)
		if err != nil {
			return err
		}
	}

	// 10. 사용자 크레딧 업데이트 (User 모델에 반영)
	// userRepo에 UpdateCredit 메서드 필요
	// 임시로 스킵

	return nil
}

func (s *shopService) GetUserInventory(ctx context.Context, userID string, itemType *models.ItemType) ([]*models.UserInventory, error) {
	return s.shopRepo.GetUserInventory(ctx, userID, itemType)
}

func (s *shopService) EquipItem(ctx context.Context, userID, itemID string) error {
	// 아이템 보유 확인
	item, err := s.shopRepo.GetUserItem(ctx, userID, itemID)
	if err != nil {
		return err
	}
	if item == nil {
		return common.NewNotFoundError("item not found in inventory")
	}

	// 같은 타입의 다른 아이템 장착 해제
	inventory, err := s.shopRepo.GetUserInventory(ctx, userID, &item.ItemType)
	if err != nil {
		return err
	}

	for _, inv := range inventory {
		if inv.IsEquipped && inv.ItemID.Hex() != itemID {
			equipped := false
			_ = s.shopRepo.UpdateInventoryItem(ctx, userID, inv.ItemID.Hex(), 0, &equipped)
		}
	}

	// 새 아이템 장착
	equipped := true
	return s.shopRepo.UpdateInventoryItem(ctx, userID, itemID, 0, &equipped)
}

func (s *shopService) UnequipItem(ctx context.Context, userID, itemID string) error {
	equipped := false
	return s.shopRepo.UpdateInventoryItem(ctx, userID, itemID, 0, &equipped)
}
