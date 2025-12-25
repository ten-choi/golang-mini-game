package service

import (
	"context"
	"draw-and-guess-server/internal/models"
	"draw-and-guess-server/internal/repository"
)

// UserService defines the interface for user business logic
type UserService interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
	GetAllUsers(ctx context.Context) ([]*models.User, error)
	CreateUser(ctx context.Context, username, displayName, email, avatarURL string) (*models.User, error)
	UpdateUser(ctx context.Context, username string, displayName, email, avatarURL *string) (*models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) GetUser(ctx context.Context, username string) (*models.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *userService) GetAllUsers(ctx context.Context) ([]*models.User, error) {
	return s.repo.GetAll(ctx)
}

func (s *userService) CreateUser(ctx context.Context, username, displayName, email, avatarURL string) (*models.User, error) {
	user := &models.User{
		Username:    username,
		DisplayName: displayName,
		Email:       email,
		AvatarURL:   avatarURL,
	}
	return s.repo.Create(ctx, user)
}

func (s *userService) UpdateUser(ctx context.Context, username string, displayName, email, avatarURL *string) (*models.User, error) {
	return s.repo.Update(ctx, username, displayName, email, avatarURL)
}