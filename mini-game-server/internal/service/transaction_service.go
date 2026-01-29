package service

import (
	"context"
	"fmt"

	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TransactionService transaction service interface
type TransactionService interface {
	GetTransaction(ctx context.Context, transactionID string) (*models.Transaction, error)
	GetUserTransactions(ctx context.Context, userID string, filter *repository.TransactionFilter) ([]*models.Transaction, error)
	GetUserBalance(ctx context.Context, userID string, currencyType models.CurrencyType) (int, error)

	// Grant/deduct credit/diamond
	AddCurrency(ctx context.Context, userID string, amount int, currencyType models.CurrencyType, txType models.TransactionType, description string, metadata map[string]interface{}) error
	DeductCurrency(ctx context.Context, userID string, amount int, currencyType models.CurrencyType, txType models.TransactionType, description string, metadata map[string]interface{}) error

	// Game reward
	GrantGameReward(ctx context.Context, userID string, amount int, currencyType models.CurrencyType, gameRoomID, gameType string, rank int, isWinner bool) error
}

type transactionService struct {
	transactionRepo repository.TransactionRepository
}

// NewTransactionService create new transaction service
func NewTransactionService(transactionRepo repository.TransactionRepository) TransactionService {
	return &transactionService{
		transactionRepo: transactionRepo,
	}
}

func (s *transactionService) GetTransaction(ctx context.Context, transactionID string) (*models.Transaction, error) {
	return s.transactionRepo.GetByID(ctx, transactionID)
}

func (s *transactionService) GetUserTransactions(ctx context.Context, userID string, filter *repository.TransactionFilter) ([]*models.Transaction, error) {
	return s.transactionRepo.GetUserTransactions(ctx, userID, filter)
}

func (s *transactionService) GetUserBalance(ctx context.Context, userID string, currencyType models.CurrencyType) (int, error) {
	return s.transactionRepo.GetUserBalance(ctx, userID, currencyType)
}

func (s *transactionService) AddCurrency(ctx context.Context, userID string, amount int, currencyType models.CurrencyType, txType models.TransactionType, description string, metadata map[string]interface{}) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	currentBalance, err := s.transactionRepo.GetUserBalance(ctx, userID, currencyType)
	if err != nil {
		return err
	}

	transaction := &models.Transaction{
		UserID:         userObjID,
		Type:           txType,
		CurrencyType:   currencyType,
		Status:         models.TransactionStatusCompleted,
		ChangeAmount:   amount,
		PreviousCredit: currentBalance,
		CurrentCredit:  currentBalance + amount,
		Description:    description,
		Metadata:       metadata,
	}

	_, err = s.transactionRepo.Create(ctx, transaction)
	return err
}

func (s *transactionService) DeductCurrency(ctx context.Context, userID string, amount int, currencyType models.CurrencyType, txType models.TransactionType, description string, metadata map[string]interface{}) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	currentBalance, err := s.transactionRepo.GetUserBalance(ctx, userID, currencyType)
	if err != nil {
		return err
	}

	if currentBalance < amount {
		return fmt.Errorf("insufficient balance: have %d, need %d", currentBalance, amount)
	}

	transaction := &models.Transaction{
		UserID:         userObjID,
		Type:           txType,
		CurrencyType:   currencyType,
		Status:         models.TransactionStatusCompleted,
		ChangeAmount:   -amount,
		PreviousCredit: currentBalance,
		CurrentCredit:  currentBalance - amount,
		Description:    description,
		Metadata:       metadata,
	}

	_, err = s.transactionRepo.Create(ctx, transaction)
	return err
}

func (s *transactionService) GrantGameReward(ctx context.Context, userID string, amount int, currencyType models.CurrencyType, gameRoomID, gameType string, rank int, isWinner bool) error {
	metadata := map[string]interface{}{
		"game_room_id": gameRoomID,
		"game_type":    gameType,
		"rank":         rank,
		"is_winner":    isWinner,
	}

	description := fmt.Sprintf("Game reward - %s (Rank: %d)", gameType, rank)
	if isWinner {
		description = fmt.Sprintf("Victory reward - %s 🏆", gameType)
	}

	return s.AddCurrency(ctx, userID, amount, currencyType, models.TransactionTypeReward, description, metadata)
}
