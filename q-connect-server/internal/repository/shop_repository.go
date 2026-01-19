package repository

import (
	"context"
	"time"

	"draw-and-guess-server/internal/common"
	"draw-and-guess-server/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ShopRepository 상점 저장소 인터페이스
type ShopRepository interface {
	// ShopItem 관련
	GetItemByID(ctx context.Context, itemID string) (*models.ShopItem, error)
	GetAllItems(ctx context.Context, filter *ShopItemFilter) ([]*models.ShopItem, error)
	GetFeaturedItems(ctx context.Context) ([]*models.ShopItem, error)
	GetItemsByType(ctx context.Context, itemType models.ItemType) ([]*models.ShopItem, error)
	CreateItem(ctx context.Context, item *models.ShopItem) (*models.ShopItem, error)
	UpdateItem(ctx context.Context, itemID string, item *models.ShopItem) (*models.ShopItem, error)
	DeleteItem(ctx context.Context, itemID string) error
	UpdateStock(ctx context.Context, itemID string, quantity int) error

	// UserInventory 관련
	GetUserInventory(ctx context.Context, userID string, itemType *models.ItemType) ([]*models.UserInventory, error)
	GetUserItem(ctx context.Context, userID, itemID string) (*models.UserInventory, error)
	AddItemToInventory(ctx context.Context, inventory *models.UserInventory) error
	UpdateInventoryItem(ctx context.Context, userID, itemID string, quantity int, isEquipped *bool) error
	RemoveItemFromInventory(ctx context.Context, userID, itemID string) error
}

type shopRepository struct {
	itemCollection      *mongo.Collection
	inventoryCollection *mongo.Collection
}

// ShopItemFilter 상점 아이템 필터
type ShopItemFilter struct {
	ItemType     *models.ItemType
	Rarity       *models.ItemRarity
	MinPrice     *int
	MaxPrice     *int
	IsAvailable  *bool
	RequireLevel *int
	Tags         []string
}

// NewShopRepository 새 상점 저장소 생성
func NewShopRepository(itemCollection, inventoryCollection *mongo.Collection) ShopRepository {
	return &shopRepository{
		itemCollection:      itemCollection,
		inventoryCollection: inventoryCollection,
	}
}

// ========== ShopItem Methods ==========

func (r *shopRepository) GetItemByID(ctx context.Context, itemID string) (*models.ShopItem, error) {
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return nil, common.NewInternalError("invalid item ID", err)
	}

	var item models.ShopItem
	err = r.itemCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&item)
	if err == mongo.ErrNoDocuments {
		return nil, common.NewNotFoundError("item not found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get item", err)
	}

	return &item, nil
}

func (r *shopRepository) GetAllItems(ctx context.Context, filter *ShopItemFilter) ([]*models.ShopItem, error) {
	query := bson.M{}

	if filter != nil {
		if filter.ItemType != nil {
			query["item_type"] = *filter.ItemType
		}
		if filter.Rarity != nil {
			query["rarity"] = *filter.Rarity
		}
		if filter.IsAvailable != nil {
			query["is_available"] = *filter.IsAvailable
		}
		if filter.MinPrice != nil || filter.MaxPrice != nil {
			priceQuery := bson.M{}
			if filter.MinPrice != nil {
				priceQuery["$gte"] = *filter.MinPrice
			}
			if filter.MaxPrice != nil {
				priceQuery["$lte"] = *filter.MaxPrice
			}
			query["price"] = priceQuery
		}
		if filter.RequireLevel != nil {
			query["require_level"] = bson.M{"$lte": *filter.RequireLevel}
		}
		if len(filter.Tags) > 0 {
			query["tags"] = bson.M{"$in": filter.Tags}
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "is_featured", Value: -1}, {Key: "created_at", Value: -1}})
	cursor, err := r.itemCollection.Find(ctx, query, opts)
	if err != nil {
		return nil, common.NewInternalError("failed to query items", err)
	}
	defer cursor.Close(ctx)

	var items []*models.ShopItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, common.NewInternalError("failed to decode items", err)
	}

	return items, nil
}

func (r *shopRepository) GetFeaturedItems(ctx context.Context) ([]*models.ShopItem, error) {
	query := bson.M{"is_featured": true, "is_available": true}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(10)

	cursor, err := r.itemCollection.Find(ctx, query, opts)
	if err != nil {
		return nil, common.NewInternalError("failed to query featured items", err)
	}
	defer cursor.Close(ctx)

	var items []*models.ShopItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, common.NewInternalError("failed to decode featured items", err)
	}

	return items, nil
}

func (r *shopRepository) GetItemsByType(ctx context.Context, itemType models.ItemType) ([]*models.ShopItem, error) {
	query := bson.M{"item_type": itemType, "is_available": true}
	opts := options.Find().SetSort(bson.D{{Key: "price", Value: 1}})

	cursor, err := r.itemCollection.Find(ctx, query, opts)
	if err != nil {
		return nil, common.NewInternalError("failed to query items by type", err)
	}
	defer cursor.Close(ctx)

	var items []*models.ShopItem
	if err := cursor.All(ctx, &items); err != nil {
		return nil, common.NewInternalError("failed to decode items", err)
	}

	return items, nil
}

func (r *shopRepository) CreateItem(ctx context.Context, item *models.ShopItem) (*models.ShopItem, error) {
	if item.ID.IsZero() {
		item.ID = primitive.NewObjectID()
	}
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	_, err := r.itemCollection.InsertOne(ctx, item)
	if err != nil {
		return nil, common.NewInternalError("failed to create item", err)
	}

	return item, nil
}

func (r *shopRepository) UpdateItem(ctx context.Context, itemID string, item *models.ShopItem) (*models.ShopItem, error) {
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return nil, common.NewInternalError("invalid item ID", err)
	}

	item.UpdatedAt = time.Now()
	update := bson.M{"$set": item}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updated models.ShopItem
	err = r.itemCollection.FindOneAndUpdate(ctx, bson.M{"_id": objID}, update, opts).Decode(&updated)
	if err == mongo.ErrNoDocuments {
		return nil, common.NewNotFoundError("item not found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to update item", err)
	}

	return &updated, nil
}

func (r *shopRepository) DeleteItem(ctx context.Context, itemID string) error {
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return common.NewInternalError("invalid item ID", err)
	}

	result, err := r.itemCollection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return common.NewInternalError("failed to delete item", err)
	}
	if result.DeletedCount == 0 {
		return common.NewNotFoundError("item not found")
	}

	return nil
}

func (r *shopRepository) UpdateStock(ctx context.Context, itemID string, quantity int) error {
	objID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return common.NewInternalError("invalid item ID", err)
	}

	update := bson.M{
		"$inc": bson.M{"stock": quantity},
		"$set": bson.M{"updated_at": time.Now()},
	}

	result, err := r.itemCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return common.NewInternalError("failed to update stock", err)
	}
	if result.MatchedCount == 0 {
		return common.NewNotFoundError("item not found")
	}

	return nil
}

// ========== UserInventory Methods ==========

func (r *shopRepository) GetUserInventory(ctx context.Context, userID string, itemType *models.ItemType) ([]*models.UserInventory, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, common.NewInternalError("invalid user ID", err)
	}

	query := bson.M{"user_id": objID}
	if itemType != nil {
		query["item_type"] = *itemType
	}

	opts := options.Find().SetSort(bson.D{{Key: "purchased_at", Value: -1}})
	cursor, err := r.inventoryCollection.Find(ctx, query, opts)
	if err != nil {
		return nil, common.NewInternalError("failed to query inventory", err)
	}
	defer cursor.Close(ctx)

	var inventory []*models.UserInventory
	if err := cursor.All(ctx, &inventory); err != nil {
		return nil, common.NewInternalError("failed to decode inventory", err)
	}

	return inventory, nil
}

func (r *shopRepository) GetUserItem(ctx context.Context, userID, itemID string) (*models.UserInventory, error) {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, common.NewInternalError("invalid user ID", err)
	}
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return nil, common.NewInternalError("invalid item ID", err)
	}

	var item models.UserInventory
	err = r.inventoryCollection.FindOne(ctx, bson.M{"user_id": userObjID, "item_id": itemObjID}).Decode(&item)
	if err == mongo.ErrNoDocuments {
		return nil, nil // 아이템이 없으면 nil 반환 (에러 아님)
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get user item", err)
	}

	return &item, nil
}

func (r *shopRepository) AddItemToInventory(ctx context.Context, inventory *models.UserInventory) error {
	if inventory.ID.IsZero() {
		inventory.ID = primitive.NewObjectID()
	}
	inventory.PurchasedAt = time.Now()

	_, err := r.inventoryCollection.InsertOne(ctx, inventory)
	if err != nil {
		return common.NewInternalError("failed to add item to inventory", err)
	}

	return nil
}

func (r *shopRepository) UpdateInventoryItem(ctx context.Context, userID, itemID string, quantity int, isEquipped *bool) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return common.NewInternalError("invalid user ID", err)
	}
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return common.NewInternalError("invalid item ID", err)
	}

	update := bson.M{}
	if quantity != 0 {
		update["$inc"] = bson.M{"quantity": quantity}
	}
	if isEquipped != nil {
		if update["$set"] == nil {
			update["$set"] = bson.M{}
		}
		update["$set"].(bson.M)["is_equipped"] = *isEquipped
	}

	result, err := r.inventoryCollection.UpdateOne(ctx, bson.M{"user_id": userObjID, "item_id": itemObjID}, update)
	if err != nil {
		return common.NewInternalError("failed to update inventory item", err)
	}
	if result.MatchedCount == 0 {
		return common.NewNotFoundError("inventory item not found")
	}

	return nil
}

func (r *shopRepository) RemoveItemFromInventory(ctx context.Context, userID, itemID string) error {
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return common.NewInternalError("invalid user ID", err)
	}
	itemObjID, err := primitive.ObjectIDFromHex(itemID)
	if err != nil {
		return common.NewInternalError("invalid item ID", err)
	}

	result, err := r.inventoryCollection.DeleteOne(ctx, bson.M{"user_id": userObjID, "item_id": itemObjID})
	if err != nil {
		return common.NewInternalError("failed to remove item from inventory", err)
	}
	if result.DeletedCount == 0 {
		return common.NewNotFoundError("inventory item not found")
	}

	return nil
}
