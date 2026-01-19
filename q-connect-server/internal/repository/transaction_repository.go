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

// TransactionRepository 거래 내역 저장소 인터페이스
type TransactionRepository interface {
	Create(ctx context.Context, transaction *models.Transaction) (*models.Transaction, error)
	GetByID(ctx context.Context, transactionID string) (*models.Transaction, error)
	GetUserTransactions(ctx context.Context, userID string, filter *TransactionFilter) ([]*models.Transaction, error)
	GetUserBalance(ctx context.Context, userID string, currencyType models.CurrencyType) (int, error)
	UpdateStatus(ctx context.Context, transactionID string, status models.TransactionStatus) error
}

type transactionRepository struct {
	collection *mongo.Collection
}

// TransactionFilter 거래 필터
type TransactionFilter struct {
	Type      *models.TransactionType
	Status    *models.TransactionStatus
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Skip      int
}

// NewTransactionRepository 새 거래 저장소 생성
func NewTransactionRepository(collection *mongo.Collection) TransactionRepository {
	return &transactionRepository{collection: collection}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *models.Transaction) (*models.Transaction, error) {
	if transaction.ID.IsZero() {
		transaction.ID = primitive.NewObjectID()
	}
	transaction.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, transaction)
	if err != nil {
		return nil, common.NewInternalError("failed to create transaction", err)
	}

	return transaction, nil
}

func (r *transactionRepository) GetByID(ctx context.Context, transactionID string) (*models.Transaction, error) {
	objID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		return nil, common.NewInternalError("invalid transaction ID", err)
	}

	var transaction models.Transaction
	err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&transaction)
	if err == mongo.ErrNoDocuments {
		return nil, common.NewNotFoundError("transaction not found")
	}
	if err != nil {
		return nil, common.NewInternalError("failed to get transaction", err)
	}

	return &transaction, nil
}

func (r *transactionRepository) GetUserTransactions(ctx context.Context, userID string, filter *TransactionFilter) ([]*models.Transaction, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, common.NewInternalError("invalid user ID", err)
	}

	query := bson.M{"user_id": objID}

	if filter != nil {
		if filter.Type != nil {
			query["type"] = *filter.Type
		}
		if filter.Status != nil {
			query["status"] = *filter.Status
		}
		if filter.StartDate != nil || filter.EndDate != nil {
			dateQuery := bson.M{}
			if filter.StartDate != nil {
				dateQuery["$gte"] = *filter.StartDate
			}
			if filter.EndDate != nil {
				dateQuery["$lte"] = *filter.EndDate
			}
			query["created_at"] = dateQuery
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	if filter != nil {
		if filter.Limit > 0 {
			opts.SetLimit(int64(filter.Limit))
		}
		if filter.Skip > 0 {
			opts.SetSkip(int64(filter.Skip))
		}
	}

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, common.NewInternalError("failed to query transactions", err)
	}
	defer cursor.Close(ctx)

	var transactions []*models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, common.NewInternalError("failed to decode transactions", err)
	}

	return transactions, nil
}

func (r *transactionRepository) GetUserBalance(ctx context.Context, userID string, currencyType models.CurrencyType) (int, error) {
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return 0, common.NewInternalError("invalid user ID", err)
	}

	// 가장 최근 거래의 current_credit을 가져옴 (재화 타입별)
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var transaction models.Transaction
	err = r.collection.FindOne(ctx, bson.M{
		"user_id":       objID,
		"currency_type": currencyType,
		"status":        models.TransactionStatusCompleted,
	}, opts).Decode(&transaction)

	if err == mongo.ErrNoDocuments {
		return 0, nil // 거래 내역이 없으면 0
	}
	if err != nil {
		return 0, common.NewInternalError("failed to get user balance", err)
	}

	return transaction.CurrentCredit, nil
}

func (r *transactionRepository) UpdateStatus(ctx context.Context, transactionID string, status models.TransactionStatus) error {
	objID, err := primitive.ObjectIDFromHex(transactionID)
	if err != nil {
		return common.NewInternalError("invalid transaction ID", err)
	}

	update := bson.M{"$set": bson.M{"status": status}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return common.NewInternalError("failed to update transaction status", err)
	}
	if result.MatchedCount == 0 {
		return common.NewNotFoundError("transaction not found")
	}

	return nil
}
