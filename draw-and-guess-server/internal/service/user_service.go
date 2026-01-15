package service

import (
	"context"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
)

// CreateUserDTO represents data for creating a new user
type CreateUserDTO struct {
	Nickname  string
	AvatarURL string // Empty string if not provided
	Level     int    // Default 0
	Credit    int    // Default 0
}

// UpdateUserDTO represents data for updating a user
type UpdateUserDTO struct {
	AvatarURL *string
	Level     *int
	Credit    *int
}

// UserService defines the interface for user business logic
type UserService interface {
	GetUser(ctx context.Context, nickname string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	CreateUser(ctx context.Context, dto CreateUserDTO) (*models.User, error)
	UpdateUser(ctx context.Context, nickname string, dto UpdateUserDTO) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUser(ctx context.Context, nickname string) (*models.User, error) {
	return s.repo.GetByNickname(ctx, nickname)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *userService) CreateUser(ctx context.Context, dto CreateUserDTO) (*models.User, error) {
	user := &models.User{
		Nickname:  dto.Nickname,
		AvatarURL: dto.AvatarURL,
		Level:     dto.Level,
		Credit:    dto.Credit,
	}
	return s.repo.Create(ctx, user)
}

func (s *userService) UpdateUser(ctx context.Context, nickname string, dto UpdateUserDTO) (*models.User, error) {
	return s.repo.Update(ctx, nickname, dto.AvatarURL, dto.Level, dto.Credit)
}
