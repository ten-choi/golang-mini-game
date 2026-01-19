package service

import (
	"context"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
	"time"
)

// CreateUserDTO represents data for creating a new user
type CreateUserDTO struct {
	HangeId   string // 회원 고유 ID (로그인용)
	Name      string // 표시 이름 (HangeId와 동일한 값)
	AvatarURL string // Empty string if not provided
	Level     int    // Default 0
	Credit    int    // Default 0
}

// UpdateUserDTO represents data for updating a user
type UpdateUserDTO struct {
	Name      *string
	AvatarURL *string
	Level     *int
	Credit    *int
}

// UserService defines the interface for user business logic
type UserService interface {
	GetUserByName(ctx context.Context, name string) (*models.User, error)
	GetUserByHangeId(ctx context.Context, hangeId string) (*models.User, error) // 로그인용
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	CreateUser(ctx context.Context, dto CreateUserDTO) (*models.User, error)
	UpdateUser(ctx context.Context, name string, dto UpdateUserDTO) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUserByName(ctx context.Context, name string) (*models.User, error) {
	return s.repo.GetByUserName(ctx, name)
}

// GetUserByHangeId retrieves user by HangeId (for login)
func (s *userService) GetUserByHangeId(ctx context.Context, hangeId string) (*models.User, error) {
	return s.repo.GetByHangeId(ctx, hangeId)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *userService) CreateUser(ctx context.Context, dto CreateUserDTO) (*models.User, error) {
	user := &models.User{
		HangeId:   dto.HangeId,
		Name:      dto.Name,
		AvatarURL: dto.AvatarURL,
		Level:     dto.Level,
		Credit:    dto.Credit,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.repo.Create(ctx, user)
}

func (s *userService) UpdateUser(ctx context.Context, name string, dto UpdateUserDTO) (*models.User, error) {
	return s.repo.Update(ctx, name, dto.AvatarURL, dto.Level, dto.Credit)
}
