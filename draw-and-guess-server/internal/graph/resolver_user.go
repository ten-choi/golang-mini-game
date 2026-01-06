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
		Username:    input.Username,
		DisplayName: input.DisplayName,
		Email:       "",
		AvatarURL:   "",
	}
	if input.Email != nil {
		dto.Email = *input.Email
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
		ID:          fmt.Sprintf("%d", created.ID),
		Username:    created.Username,
		DisplayName: created.DisplayName,
		Email:       &created.Email,
		AvatarURL:   &created.AvatarURL,
		CreatedAt:   created.CreatedAt,
		UpdatedAt:   created.UpdatedAt,
	}, nil
}

// UpdateUser is the resolver for the updateUser field.
func (r *mutationResolver) UpdateUser(ctx context.Context, username string, input model.UpdateUserInput) (*model.User, error) {
	// GraphQL input -> Service DTO
	dto := service.UpdateUserDTO{
		DisplayName: input.DisplayName,
		Email:       input.Email,
		AvatarURL:   input.AvatarURL,
	}

	updated, err := r.UserService.UpdateUser(ctx, username, dto)
	if err != nil {
		return nil, err
	}

	// Domain model -> GraphQL model
	return &model.User{
		ID:          fmt.Sprintf("%d", updated.ID),
		Username:    updated.Username,
		DisplayName: updated.DisplayName,
		Email:       &updated.Email,
		AvatarURL:   &updated.AvatarURL,
		CreatedAt:   updated.CreatedAt,
		UpdatedAt:   updated.UpdatedAt,
	}, nil
}

// ========================================
// User Queries
// ========================================

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, username string) (*model.User, error) {
	user, err := r.UserService.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}

	return &model.User{
		ID:          fmt.Sprintf("%d", user.ID),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Email:       &user.Email,
		AvatarURL:   &user.AvatarURL,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
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
			ID:          fmt.Sprintf("%d", u.ID),
			Username:    u.Username,
			DisplayName: u.DisplayName,
			Email:       &u.Email,
			AvatarURL:   &u.AvatarURL,
			CreatedAt:   u.CreatedAt,
			UpdatedAt:   u.UpdatedAt,
		}
	}
	return result, nil
}

// DeleteUser is the resolver for the deleteUser field.
func (r *mutationResolver) DeleteUser(ctx context.Context, username string) (bool, error) {
	// TODO: Implement delete user logic
	// This would require adding DeleteUser to UserService
	return false, fmt.Errorf("not implemented yet")
}
