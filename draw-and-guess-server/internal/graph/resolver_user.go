package graph

import (
	"context"
	"draw-and-guess-server/internal/graph/model"
	"draw-and-guess-server/internal/service"
	"fmt"
)

// ========================================
// User Mutations
// ========================================

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, input model.CreateUserInput) (*model.User, error) {
	// GraphQL input -> Service DTO
	dto := service.CreateUserDTO{
		Nickname:  input.Nickname,
		AvatarURL: "",
		Level:     0,
		Credit:    0,
	}
	if input.AvatarURL != nil {
		dto.AvatarURL = *input.AvatarURL
	}

	created, err := r.UserService.CreateUser(ctx, dto)
	if err != nil {
		return nil, err
	}

	// Domain model -> GraphQL model
	return &model.User{
		ID:        created.ID.Hex(),
		Nickname:  created.Nickname,
		AvatarURL: &created.AvatarURL,
		Level:     int32(created.Level),
		Credit:    int32(created.Credit),
		HanCoin:   int32(created.HanCoin),
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, nickname string, input model.UpdateUserInput) (*model.User, error) {
	// GraphQL input -> Service DTO
	dto := service.UpdateUserDTO{
		AvatarURL: input.AvatarURL,
	}
	if input.Level != nil {
		level := int(*input.Level)
		dto.Level = &level
	}
	if input.Credit != nil {
		credit := int(*input.Credit)
		dto.Credit = &credit
	}
	// HanCoin은 외부 재화로 이 게임에서 업데이트하지 않음

	updated, err := r.UserService.UpdateUser(ctx, nickname, dto)
	if err != nil {
		return nil, err
	}

	// Domain model -> GraphQL model
	return &model.User{
		ID:        updated.ID.Hex(),
		Nickname:  updated.Nickname,
		AvatarURL: &updated.AvatarURL,
		Level:     int32(updated.Level),
		Credit:    int32(updated.Credit),
		HanCoin:   int32(updated.HanCoin),
		CreatedAt: updated.CreatedAt,
		UpdatedAt: updated.UpdatedAt,
	}, nil
}

// ========================================
// User Queries
// ========================================

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, nickname string) (*model.User, error) {
	user, err := r.UserService.GetUser(ctx, nickname)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:        user.ID.Hex(),
		Nickname:  user.Nickname,
		AvatarURL: &user.AvatarURL,
		Level:     int32(user.Level),
		Credit:    int32(user.Credit),
		HanCoin:   int32(user.HanCoin),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// Users is the resolver for the users field.
func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	users, err := r.UserService.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*model.User, len(users))
	for i, u := range users {
		result[i] = &model.User{
			ID:        u.ID.Hex(),
			Nickname:  u.Nickname,
			AvatarURL: &u.AvatarURL,
			Level:     int32(u.Level),
			Credit:    int32(u.Credit),
			HanCoin:   int32(u.HanCoin),
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}
	return result, nil
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, nickname string) (bool, error) {
	// TODO: Implement delete user logic
	// This would require adding DeleteUser to UserService
	return false, fmt.Errorf("not implemented yet")
}
